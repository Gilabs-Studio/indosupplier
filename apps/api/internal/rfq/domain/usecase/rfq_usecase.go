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
)

var (
	ErrBuyerProfileNotFound    = errors.New("buyer profile not found")
	ErrSupplierProfileNotFound = errors.New("supplier profile not found")
	ErrRFQNotFound             = errors.New("rfq not found")
	ErrBidNotFound             = errors.New("bid not found")
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

	rfq := &models.RFQ{
		ID:                  uuid.NewString(),
		BuyerProfileID:      buyerID,
		Title:               req.ProductName,
		ProductDescription:  req.Description,
		QuantityValue:       qtyValue,
		QuantityUnit:        req.Unit,
		DestinationLocation: req.TargetPort,
		CategoryID:          categoryIDPtr,
		VisibilityStatus:    "open",
		Mode:                "broadcast",
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

	recipients, err := u.buildRecipients(ctx, categoryID, userID)
	if err != nil {
		return dto.RFQResponse{}, err
	}

	if err := u.rfqRepo.Create(ctx, rfq, attachment, recipients); err != nil {
		return dto.RFQResponse{}, err
	}

	return mapper.ToRFQResponse(rfq, categoryName, 0, attachment), nil
}

func (u *rfqUsecase) GetByID(ctx context.Context, userID string, id string) (dto.RFQResponse, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return dto.RFQResponse{}, err
	}

	resolvedID := mapper.ResolveRFQID(id)

	rfq, attachment, replies, err := u.rfqRepo.FindByID(ctx, resolvedID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.RFQResponse{}, ErrRFQNotFound
		}
		return dto.RFQResponse{}, err
	}

	// Check if this RFQ belongs to the buyer
	if rfq.BuyerProfileID != buyerID {
		return dto.RFQResponse{}, errors.New("unauthorized to view this rfq")
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
			ID:           rec.ID,
			SupplierName: supplier.CompanyName,
			Price:        price,
			MOQ:          moq,
			ResponseTime: responseTimeStr,
			Verified:     supplier.VerificationLevel >= 2,
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

	// Verify RFQ ownership
	var rfq models.RFQ
	if err := u.db.WithContext(ctx).Where("id = ? AND buyer_profile_id = ?", resolvedRFQID, buyerID).First(&rfq).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRFQNotFound
		}
		return err
	}

	return u.rfqRepo.AcceptBid(ctx, resolvedRFQID, bidID)
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
	if rfq.ClosedAt != nil {
		status = "closed"
	} else if recipient.Status == "responded" || recipient.Status == "processing" || recipient.Status == "accepted" {
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
	responses := make([]dto.SupplierRFQResponse, 0, len(recipients))
	for _, recipient := range recipients {
		item, err := u.buildSupplierRFQResponse(ctx, recipient)
		if err == nil {
			responses = append(responses, item)
		}
	}
	return responses, total, nil
}

func (u *rfqUsecase) GetSupplierRFQByID(ctx context.Context, userID string, id string) (dto.SupplierRFQResponse, error) {
	supplierID, err := u.getSupplierProfileID(ctx, userID)
	if err != nil {
		return dto.SupplierRFQResponse{}, err
	}
	resolvedID := mapper.ResolveRFQID(id)
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
	recipient, err := u.rfqRepo.FindForSupplier(ctx, supplierID, resolvedID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRFQNotFound
		}
		return err
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
	return u.rfqRepo.SubmitProposal(ctx, recipient, message)
}
