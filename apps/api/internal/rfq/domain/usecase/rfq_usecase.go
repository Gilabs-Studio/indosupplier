package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	buyerModels "github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	"github.com/gilabs/indosupplier/api/internal/rfq/data/models"
	"github.com/gilabs/indosupplier/api/internal/rfq/data/repositories"
	"github.com/gilabs/indosupplier/api/internal/rfq/domain/dto"
	"github.com/gilabs/indosupplier/api/internal/rfq/domain/mapper"
	supplierModels "github.com/gilabs/indosupplier/api/internal/supplier/data/models"
	trustModels "github.com/gilabs/indosupplier/api/internal/trust/data/models"
)

var (
	ErrBuyerProfileNotFound    = errors.New("buyer profile not found")
	ErrSupplierProfileNotFound = errors.New("supplier profile not found")
	ErrRFQNotFound             = errors.New("rfq not found")
	ErrBidNotFound             = errors.New("bid not found")
	ErrRFQAlreadyClosed        = errors.New("rfq is already closed")
	ErrUnauthorizedAccess      = errors.New("unauthorized access")
)

type RFQUsecase interface {
	Create(ctx context.Context, userID string, req *dto.CreateRFQRequest) (dto.RFQResponse, error)
	GetByID(ctx context.Context, userID string, id string) (dto.RFQResponse, error)
	List(ctx context.Context, userID string, status string, page, perPage int) ([]dto.RFQResponse, int64, error)
	GetBids(ctx context.Context, userID string, rfqID string) ([]dto.RFQBidResponse, error)
	AcceptBid(ctx context.Context, userID string, rfqID string, bidID string) error
	ListForSupplier(ctx context.Context, userID string, page, perPage int) ([]dto.SupplierRFQResponse, int64, error)
	GetSupplierRFQByID(ctx context.Context, userID string, id string) (dto.SupplierRFQResponse, error)
	SubmitProposal(ctx context.Context, userID string, rfqID string, req *dto.SubmitRFQProposalRequest) error
}

type rfqUsecase struct {
	db      *gorm.DB
	rfqRepo repositories.RFQRepository
}

func NewRFQUsecase(db *gorm.DB, rfqRepo repositories.RFQRepository) RFQUsecase {
	return &rfqUsecase{
		db:      db,
		rfqRepo: rfqRepo,
	}
}

func (u *rfqUsecase) getBuyerProfileID(ctx context.Context, userID string) (string, error) {
	var buyer buyerModels.BuyerProfile
	if err := u.db.WithContext(ctx).Where("user_id = ?", userID).First(&buyer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrBuyerProfileNotFound
		}
		return "", err
	}
	return buyer.ID, nil
}

func (u *rfqUsecase) getSupplierProfileID(ctx context.Context, userID string) (string, error) {
	var supplier supplierModels.SupplierProfile
	if err := u.db.WithContext(ctx).Where("user_id = ?", userID).First(&supplier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrSupplierProfileNotFound
		}
		return "", err
	}
	return supplier.ID, nil
}

func (u *rfqUsecase) resolveCategory(ctx context.Context, categoryInput string) (string, string) {
	categoryInput = strings.TrimSpace(categoryInput)
	if categoryInput == "" {
		return "", ""
	}

	var category supplierModels.Category
	if err := u.db.WithContext(ctx).Where("slug = ?", categoryInput).First(&category).Error; err == nil {
		return category.ID, category.Name
	}
	if err := u.db.WithContext(ctx).Where("name ILIKE ?", "%"+categoryInput+"%").First(&category).Error; err == nil {
		return category.ID, category.Name
	}
	return "", categoryInput
}

func (u *rfqUsecase) buildRecipients(ctx context.Context, categoryID string, buyerUserID string) ([]models.RFQRecipient, error) {
	query := u.db.WithContext(ctx).
		Model(&supplierModels.SupplierProfile{}).
		Where("supplier_profiles.status = ?", "active").
		Where("supplier_profiles.user_id <> ?", buyerUserID)

	if categoryID != "" {
		query = query.
			Joins("JOIN supplier_categories ON supplier_categories.supplier_profile_id = supplier_profiles.id").
			Where("supplier_categories.category_id = ?", categoryID)
	}

	var suppliers []supplierModels.SupplierProfile
	if err := query.Order("supplier_profiles.verification_level DESC, supplier_profiles.updated_at DESC").Limit(20).Find(&suppliers).Error; err != nil {
		return nil, err
	}

	recipients := make([]models.RFQRecipient, 0, len(suppliers))
	for i, supplier := range suppliers {
		recipients = append(recipients, models.RFQRecipient{
			ID:                uuid.NewString(),
			SupplierProfileID: supplier.ID,
			Status:            "new",
			RankPosition:      i + 1,
			CreatedAt:         apptime.Now(),
			UpdatedAt:         apptime.Now(),
		})
	}
	return recipients, nil
}

func (u *rfqUsecase) Create(ctx context.Context, userID string, req *dto.CreateRFQRequest) (dto.RFQResponse, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return dto.RFQResponse{}, err
	}

	categoryID, categoryName := u.resolveCategory(ctx, req.Category)

	qtyValue, err := strconv.ParseFloat(req.Quantity, 64)
	if err != nil {
		return dto.RFQResponse{}, err
	}

	var categoryIDPtr *string
	if categoryID != "" {
		categoryIDPtr = &categoryID
	}

	mode := "broadcast"
	var recipients []models.RFQRecipient
	if req.TargetSupplierID != "" {
		var targetSupplier supplierModels.SupplierProfile
		if err := u.db.WithContext(ctx).Where("id = ?", req.TargetSupplierID).First(&targetSupplier).Error; err == nil {
			mode = "specific"
			recipients = append(recipients, models.RFQRecipient{
				ID:                uuid.NewString(),
				SupplierProfileID: targetSupplier.ID,
				Status:            "new",
				RankPosition:      1,
				CreatedAt:         apptime.Now(),
				UpdatedAt:         apptime.Now(),
			})
		}
	}

	if len(recipients) == 0 {
		var err error
		recipients, err = u.buildRecipients(ctx, categoryID, userID)
		if err != nil {
			return dto.RFQResponse{}, err
		}
	}

	imageURL := req.ImageURL
	if imageURL == "" && req.ProductID != nil && *req.ProductID != "" {
		var photoRow struct {
			PhotoURL string `gorm:"column:photo_url"`
		}
		if err := u.db.WithContext(ctx).Table("supplier_product_photos").Where("supplier_product_id = ?", *req.ProductID).Order("sort_order ASC").First(&photoRow).Error; err == nil {
			imageURL = photoRow.PhotoURL
		}
	}

	rfq := &models.RFQ{
		ID:                  uuid.NewString(),
		BuyerProfileID:      buyerID,
		Title:               req.ProductName,
		ProductDescription:  req.Description,
		QuantityValue:       qtyValue,
		QuantityUnit:        req.Unit,
		DestinationLocation: req.TargetPort,
		CategoryID:          categoryIDPtr,
		ProductID:           req.ProductID,
		ImageURL:            imageURL,
		VisibilityStatus:    "open",
		Mode:                mode,
		BudgetMax:           req.Budget,
		DeliveryTimeline:    req.DeliveryTimeline,
		CreatedAt:           apptime.Now(),
		UpdatedAt:           apptime.Now(),
	}

	var attachment *models.RFQAttachment
	if req.AttachmentURL != "" {
		fileName := strings.TrimSpace(req.AttachmentName)
		if fileName == "" {
			fileName = filepath.Base(req.AttachmentURL)
		}
		attachment = &models.RFQAttachment{
			ID:        uuid.NewString(),
			FileURL:   req.AttachmentURL,
			FileName:  fileName,
			FileSize:  req.AttachmentSize,
			CreatedAt: apptime.Now(),
			UpdatedAt: apptime.Now(),
		}
	}

	if err := u.rfqRepo.Create(ctx, rfq, attachment, recipients); err != nil {
		return dto.RFQResponse{}, err
	}

	// Send in-app notification to recipient suppliers
	for _, rec := range recipients {
		notif := trustModels.Notification{
			ID:            uuid.NewString(),
			RecipientType: "supplier",
			RecipientID:   rec.SupplierProfileID,
			Type:          "rfq",
			Title:         fmt.Sprintf("Permintaan RFQ Baru: %s", rfq.Title),
			Body:          fmt.Sprintf("Terdapat permintaan RFQ baru untuk produk %s dengan jumlah %s %s.", rfq.Title, req.Quantity, req.Unit),
			Channel:       "in_app",
			IsRead:        false,
			RelatedType:   "rfq",
			RelatedID:     &rfq.ID,
			CreatedAt:     apptime.Now(),
			UpdatedAt:     apptime.Now(),
		}
		_ = u.db.WithContext(ctx).Create(&notif).Error
	}

	return mapper.ToRFQResponse(rfq, categoryName, 0, attachment), nil
}

func (u *rfqUsecase) GetByID(ctx context.Context, userID string, id string) (dto.RFQResponse, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return dto.RFQResponse{}, err
	}

	resolvedID := mapper.ResolveRFQID(id)
	if _, err := uuid.Parse(resolvedID); err != nil {
		return dto.RFQResponse{}, ErrRFQNotFound
	}

	rfq, attachment, replies, err := u.rfqRepo.FindByID(ctx, resolvedID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.RFQResponse{}, ErrRFQNotFound
		}
		return dto.RFQResponse{}, err
	}

	// Check if this RFQ belongs to the buyer
	if rfq.BuyerProfileID != buyerID {
		return dto.RFQResponse{}, ErrUnauthorizedAccess
	}

	// Get Category Name
	var categoryName string
	if rfq.CategoryID != nil && *rfq.CategoryID != "" {
		u.db.WithContext(ctx).
			Table("categories").
			Where("id = ?", *rfq.CategoryID).
			Pluck("name", &categoryName)
	}

	return mapper.ToRFQResponse(rfq, categoryName, replies, attachment), nil
}

func (u *rfqUsecase) List(ctx context.Context, userID string, status string, page, perPage int) ([]dto.RFQResponse, int64, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	rfqs, replies, total, err := u.rfqRepo.List(ctx, buyerID, status, page, perPage)
	if err != nil {
		return nil, 0, err
	}

	// Fetch categories in batch
	catIDs := make([]string, 0, len(rfqs))
	for _, r := range rfqs {
		if r.CategoryID != nil && *r.CategoryID != "" {
			catIDs = append(catIDs, *r.CategoryID)
		}
	}

	var categories []supplierModels.Category
	catMap := make(map[string]string)
	if len(catIDs) > 0 {
		u.db.WithContext(ctx).Where("id IN ?", catIDs).Find(&categories)
		for _, c := range categories {
			catMap[c.ID] = c.Name
		}
	}

	// Fetch attachments in batch
	rfqIDs := make([]string, len(rfqs))
	for i, r := range rfqs {
		rfqIDs[i] = r.ID
	}

	var attachments []models.RFQAttachment
	attachMap := make(map[string]*models.RFQAttachment)
	if len(rfqIDs) > 0 {
		u.db.WithContext(ctx).Where("rfq_id IN ?", rfqIDs).Find(&attachments)
		for i := range attachments {
			attachMap[attachments[i].RFQID] = &attachments[i]
		}
	}

	var response []dto.RFQResponse
	for i, rfq := range rfqs {
		var catName string
		if rfq.CategoryID != nil {
			catName = catMap[*rfq.CategoryID]
		}
		attach := attachMap[rfq.ID]
		response = append(response, mapper.ToRFQResponse(&rfq, catName, replies[i], attach))
	}

	return response, total, nil
}

func (u *rfqUsecase) GetBids(ctx context.Context, userID string, rfqID string) ([]dto.RFQBidResponse, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}

	resolvedRFQID := mapper.ResolveRFQID(rfqID)
	if _, err := uuid.Parse(resolvedRFQID); err != nil {
		return nil, ErrRFQNotFound
	}

	// Verify RFQ ownership
	var rfq models.RFQ
	if err := u.db.WithContext(ctx).Where("id = ? AND buyer_profile_id = ?", resolvedRFQID, buyerID).First(&rfq).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRFQNotFound
		}
		return nil, err
	}

	recipients, err := u.rfqRepo.GetBids(ctx, resolvedRFQID)
	if err != nil {
		return nil, err
	}

	var responses []dto.RFQBidResponse
	for _, rec := range recipients {
		var supplier supplierModels.SupplierProfile
		if err := u.db.WithContext(ctx).Where("id = ?", rec.SupplierProfileID).First(&supplier).Error; err != nil {
			continue
		}

		var msg models.RFQMessage
		price := ""
		moq := ""
		responseTimeStr := ""

		if err := u.db.WithContext(ctx).
			Where("rfq_id = ? AND sender_id = ? AND message_type = ?", resolvedRFQID, supplier.ID, "offer").
			Order("created_at DESC").
			First(&msg).Error; err == nil {
			var metadataMap map[string]interface{}
			if errJson := json.Unmarshal([]byte(msg.Metadata), &metadataMap); errJson == nil {
				if p, ok := metadataMap["price"].(string); ok {
					price = p
				}
				if m, ok := metadataMap["moq"].(string); ok {
					moq = m
				}
				if d, ok := metadataMap["deliveryTime"].(string); ok {
					responseTimeStr = d
				}
			}
		}

		if responseTimeStr == "" && rec.RespondedAt != nil {
			responseTimeStr = fmt.Sprintf("%d menit", int(rec.RespondedAt.Sub(rec.CreatedAt).Minutes()))
		}

		responses = append(responses, dto.RFQBidResponse{
			ID:                rec.ID,
			SupplierName:      supplier.CompanyName,
			SupplierProfileID: supplier.ID,
			Price:             price,
			MOQ:               moq,
			ResponseTime:      responseTimeStr,
			Verified:          supplier.VerificationLevel >= 2,
		})
	}

	return responses, nil
}

func (u *rfqUsecase) AcceptBid(ctx context.Context, userID string, rfqID string, bidID string) error {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return err
	}

	resolvedRFQID := mapper.ResolveRFQID(rfqID)
	if _, err := uuid.Parse(resolvedRFQID); err != nil {
		return ErrRFQNotFound
	}
	if _, err := uuid.Parse(bidID); err != nil {
		return ErrBidNotFound
	}

	// Verify RFQ ownership
	var rfq models.RFQ
	if err := u.db.WithContext(ctx).Where("id = ? AND buyer_profile_id = ?", resolvedRFQID, buyerID).First(&rfq).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRFQNotFound
		}
		return err
	}

	if err := u.rfqRepo.AcceptBid(ctx, resolvedRFQID, bidID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrBidNotFound
		}
		if err.Error() == "rfq is already closed" {
			return ErrRFQAlreadyClosed
		}
		return err
	}

	// Notify winning supplier
	var rec models.RFQRecipient
	if err := u.db.WithContext(ctx).Where("id = ?", bidID).First(&rec).Error; err == nil {
		notif := trustModels.Notification{
			ID:            uuid.NewString(),
			RecipientType: "supplier",
			RecipientID:   rec.SupplierProfileID,
			Type:          "rfq_accepted",
			Title:         fmt.Sprintf("Penawaran Diterima: %s", rfq.Title),
			Body:          fmt.Sprintf("Selamat! Pembeli telah menerima penawaran Anda untuk RFQ %s.", rfq.Title),
			Channel:       "in_app",
			IsRead:        false,
			RelatedType:   "rfq",
			RelatedID:     &rfq.ID,
			CreatedAt:     apptime.Now(),
			UpdatedAt:     apptime.Now(),
		}
		_ = u.db.WithContext(ctx).Create(&notif).Error
	}

	return nil
}

func (u *rfqUsecase) buildSupplierRFQResponse(ctx context.Context, recipient models.RFQRecipient) (dto.SupplierRFQResponse, error) {
	rfq, _, _, err := u.rfqRepo.FindByID(ctx, recipient.RFQID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.SupplierRFQResponse{}, ErrRFQNotFound
		}
		return dto.SupplierRFQResponse{}, err
	}

	var categoryName string
	if rfq.CategoryID != nil && *rfq.CategoryID != "" {
		_ = u.db.WithContext(ctx).Table("categories").Where("id = ?", *rfq.CategoryID).Pluck("name", &categoryName).Error
	}

	var buyer buyerModels.BuyerProfile
	_ = u.db.WithContext(ctx).Where("id = ?", rfq.BuyerProfileID).First(&buyer).Error

	status := "open"
	if recipient.Status == "accepted" {
		status = "accepted"
	} else if rfq.ClosedAt != nil {
		status = "closed"
	} else if recipient.Status == "responded" || recipient.Status == "processing" {
		status = recipient.Status
	}

	res := dto.SupplierRFQResponse{
		ID:             rfq.ID,
		Product:        rfq.Title,
		Category:       categoryName,
		Port:           rfq.DestinationLocation,
		Date:           rfq.CreatedAt.Format("2006-01-02"),
		Budget:         "",
		Status:         status,
		ImageURL:       mapper.ResolveRFQImageURL(rfq.ImageURL, rfq.Title, categoryName),
		Description:    rfq.ProductDescription,
		ShippingTerm:   rfq.PreferredContactMethod,
		TargetDelivery: rfq.DeliveryTimeline,
	}
	if rfq.QuantityValue == float64(int(rfq.QuantityValue)) {
		res.Quantity = fmt.Sprintf("%d %s", int(rfq.QuantityValue), rfq.QuantityUnit)
	} else {
		res.Quantity = fmt.Sprintf("%.2f %s", rfq.QuantityValue, rfq.QuantityUnit)
	}
	if rfq.BudgetMax > 0 {
		res.Budget = fmt.Sprintf("%.0f", rfq.BudgetMax)
	}
	res.Buyer.Name = buyer.CompanyName
	if res.Buyer.Name == "" {
		res.Buyer.Name = buyer.FullName
	}
	res.Buyer.Location = buyer.Address
	res.Buyer.Rating = ""

	if recipient.Status == "responded" || recipient.Status == "accepted" {
		var offerMsg models.RFQMessage
		if err := u.db.WithContext(ctx).
			Where("rfq_id = ? AND sender_id = ? AND message_type = ?", rfq.ID, recipient.SupplierProfileID, "offer").
			Order("created_at DESC").
			First(&offerMsg).Error; err == nil {
			var meta map[string]interface{}
			_ = json.Unmarshal([]byte(offerMsg.Metadata), &meta)
			priceStr, _ := meta["price"].(string)
			moqStr, _ := meta["moq"].(string)
			delivStr, _ := meta["deliveryTime"].(string)
			res.SubmittedOffer = &dto.SupplierSubmittedOfferDto{
				Price:        priceStr,
				MOQ:          moqStr,
				DeliveryTime: delivStr,
				Notes:        offerMsg.Body,
				RespondedAt:  offerMsg.CreatedAt.Format("2006-01-02 15:04"),
			}
		}
	}

	return res, nil
}

func (u *rfqUsecase) ListForSupplier(ctx context.Context, userID string, page, perPage int) ([]dto.SupplierRFQResponse, int64, error) {
	supplierID, err := u.getSupplierProfileID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	recipients, total, err := u.rfqRepo.ListForSupplier(ctx, supplierID, page, perPage)
	if err != nil {
		return nil, 0, err
	}

	if len(recipients) == 0 {
		return []dto.SupplierRFQResponse{}, total, nil
	}

	// Batch fetch all RFQs to eliminate N+1 queries
	rfqIDs := make([]string, 0, len(recipients))
	for _, r := range recipients {
		if r.RFQID != "" {
			rfqIDs = append(rfqIDs, r.RFQID)
		}
	}

	rfqMap := make(map[string]models.RFQ)
	catIDSet := make(map[string]struct{})
	buyerIDSet := make(map[string]struct{})

	if len(rfqIDs) > 0 {
		var rfqs []models.RFQ
		if err := u.db.WithContext(ctx).Where("id IN ?", rfqIDs).Find(&rfqs).Error; err == nil {
			for _, rfq := range rfqs {
				rfqMap[rfq.ID] = rfq
				if rfq.CategoryID != nil && *rfq.CategoryID != "" {
					catIDSet[*rfq.CategoryID] = struct{}{}
				}
				if rfq.BuyerProfileID != "" {
					buyerIDSet[rfq.BuyerProfileID] = struct{}{}
				}
			}
		}
	}

	// Batch fetch Categories
	catMap := make(map[string]string)
	if len(catIDSet) > 0 {
		catIDs := make([]string, 0, len(catIDSet))
		for id := range catIDSet {
			catIDs = append(catIDs, id)
		}
		var categories []supplierModels.Category
		if err := u.db.WithContext(ctx).Where("id IN ?", catIDs).Find(&categories).Error; err == nil {
			for _, c := range categories {
				catMap[c.ID] = c.Name
			}
		}
	}

	// Batch fetch Buyer Profiles
	buyerMap := make(map[string]buyerModels.BuyerProfile)
	if len(buyerIDSet) > 0 {
		buyerIDs := make([]string, 0, len(buyerIDSet))
		for id := range buyerIDSet {
			buyerIDs = append(buyerIDs, id)
		}
		var buyers []buyerModels.BuyerProfile
		if err := u.db.WithContext(ctx).Where("id IN ?", buyerIDs).Find(&buyers).Error; err == nil {
			for _, b := range buyers {
				buyerMap[b.ID] = b
			}
		}
	}

	responses := make([]dto.SupplierRFQResponse, 0, len(recipients))
	for _, recipient := range recipients {
		rfq, exists := rfqMap[recipient.RFQID]
		if !exists {
			continue
		}

		var catName string
		if rfq.CategoryID != nil {
			catName = catMap[*rfq.CategoryID]
		}
		buyer := buyerMap[rfq.BuyerProfileID]

		status := "open"
		if recipient.Status == "accepted" {
			status = "accepted"
		} else if rfq.ClosedAt != nil {
			status = "closed"
		} else if recipient.Status == "responded" || recipient.Status == "processing" {
			status = recipient.Status
		}

		res := dto.SupplierRFQResponse{
			ID:             rfq.ID,
			Product:        rfq.Title,
			Category:       catName,
			Port:           rfq.DestinationLocation,
			Date:           rfq.CreatedAt.Format("2006-01-02"),
			Budget:         "",
			Status:         status,
			ImageURL:       mapper.ResolveRFQImageURL(rfq.ImageURL, rfq.Title, catName),
			Description:    rfq.ProductDescription,
			ShippingTerm:   rfq.PreferredContactMethod,
			TargetDelivery: rfq.DeliveryTimeline,
		}
		if rfq.QuantityValue == float64(int(rfq.QuantityValue)) {
			res.Quantity = fmt.Sprintf("%d %s", int(rfq.QuantityValue), rfq.QuantityUnit)
		} else {
			res.Quantity = fmt.Sprintf("%.2f %s", rfq.QuantityValue, rfq.QuantityUnit)
		}
		if rfq.BudgetMax > 0 {
			res.Budget = fmt.Sprintf("%.0f", rfq.BudgetMax)
		}
		res.Buyer.Name = buyer.CompanyName
		if res.Buyer.Name == "" {
			res.Buyer.Name = buyer.FullName
		}
		res.Buyer.Location = buyer.Address
		res.Buyer.Rating = ""

		responses = append(responses, res)
	}

	return responses, total, nil
}

func (u *rfqUsecase) GetSupplierRFQByID(ctx context.Context, userID string, id string) (dto.SupplierRFQResponse, error) {
	supplierID, err := u.getSupplierProfileID(ctx, userID)
	if err != nil {
		return dto.SupplierRFQResponse{}, err
	}
	resolvedID := mapper.ResolveRFQID(id)
	if _, err := uuid.Parse(resolvedID); err != nil {
		return dto.SupplierRFQResponse{}, ErrRFQNotFound
	}
	recipient, err := u.rfqRepo.FindForSupplier(ctx, supplierID, resolvedID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.SupplierRFQResponse{}, ErrRFQNotFound
		}
		return dto.SupplierRFQResponse{}, err
	}
	return u.buildSupplierRFQResponse(ctx, *recipient)
}

func (u *rfqUsecase) SubmitProposal(ctx context.Context, userID string, rfqID string, req *dto.SubmitRFQProposalRequest) error {
	supplierID, err := u.getSupplierProfileID(ctx, userID)
	if err != nil {
		return err
	}
	resolvedID := mapper.ResolveRFQID(rfqID)
	if _, err := uuid.Parse(resolvedID); err != nil {
		return ErrRFQNotFound
	}
	recipient, err := u.rfqRepo.FindForSupplier(ctx, supplierID, resolvedID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRFQNotFound
		}
		return err
	}

	var rfq models.RFQ
	if err := u.db.WithContext(ctx).Where("id = ?", resolvedID).First(&rfq).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRFQNotFound
		}
		return err
	}
	if rfq.ClosedAt != nil || rfq.VisibilityStatus == "closed" {
		return errors.New("rfq is closed")
	}

	metadata, err := json.Marshal(map[string]string{
		"price":        strings.TrimSpace(req.Price),
		"moq":          strings.TrimSpace(req.MOQ),
		"deliveryTime": strings.TrimSpace(req.DeliveryTime),
	})
	if err != nil {
		return err
	}
	message := &models.RFQMessage{
		ID:          uuid.NewString(),
		RFQID:       resolvedID,
		SenderType:  "supplier",
		SenderID:    supplierID,
		MessageType: "offer",
		Body:        strings.TrimSpace(req.Notes),
		Metadata:    string(metadata),
		CreatedAt:   apptime.Now(),
		UpdatedAt:   apptime.Now(),
	}
	if err := u.rfqRepo.SubmitProposal(ctx, recipient, message); err != nil {
		return err
	}

	// Notify buyer
	var supplier supplierModels.SupplierProfile
	_ = u.db.WithContext(ctx).Where("id = ?", supplierID).First(&supplier).Error
	supplierName := supplier.CompanyName
	if supplierName == "" {
		supplierName = "Supplier"
	}
	notif := trustModels.Notification{
		ID:            uuid.NewString(),
		RecipientType: "buyer",
		RecipientID:   rfq.BuyerProfileID,
		Type:          "rfq_proposal",
		Title:         fmt.Sprintf("Penawaran Baru: %s", rfq.Title),
		Body:          fmt.Sprintf("%s telah mengirimkan penawaran harga sebesar %s untuk RFQ %s.", supplierName, req.Price, rfq.Title),
		Channel:       "in_app",
		IsRead:        false,
		RelatedType:   "rfq",
		RelatedID:     &rfq.ID,
		CreatedAt:     apptime.Now(),
		UpdatedAt:     apptime.Now(),
	}
	_ = u.db.WithContext(ctx).Create(&notif).Error

	return nil
}
