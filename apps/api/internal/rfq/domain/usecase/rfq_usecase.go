package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"gorm.io/gorm"

	buyerModels "github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	"github.com/gilabs/indosupplier/api/internal/core/utils"
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

func (u *rfqUsecase) Create(ctx context.Context, userID string, req *dto.CreateRFQRequest) (dto.RFQResponse, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return dto.RFQResponse{}, err
	}

	// Lookup category
	var categoryID string
	var categoryName string
	var category supplierModels.Category
	if err := u.db.WithContext(ctx).Where("slug = ?", req.Category).First(&category).Error; err == nil {
		categoryID = category.ID
		categoryName = category.Name
	} else if err := u.db.WithContext(ctx).Where("name ILIKE ?", "%"+req.Category+"%").First(&category).Error; err == nil {
		categoryID = category.ID
		categoryName = category.Name
	} else {
		// Use a fallback category if none matches
		var fallback supplierModels.Category
		if err := u.db.WithContext(ctx).First(&fallback).Error; err == nil {
			categoryID = fallback.ID
			categoryName = fallback.Name
		}
	}

	qtyValue, _ := strconv.ParseFloat(req.Quantity, 64)

	rfq := &models.RFQ{
		ID:                  uuid.NewString(),
		BuyerProfileID:      buyerID,
		Title:               req.ProductName,
		ProductDescription:  req.Description,
		QuantityValue:       qtyValue,
		QuantityUnit:        req.Unit,
		DestinationLocation: req.TargetPort,
		CategoryID:          categoryID,
		VisibilityStatus:    "open",
		Mode:                "broadcast",
		CreatedAt:           apptime.Now(),
		UpdatedAt:           apptime.Now(),
	}

	var attachment *models.RFQAttachment
	if req.AttachmentURL != "" {
		attachment = &models.RFQAttachment{
			ID:        uuid.NewString(),
			FileURL:   req.AttachmentURL,
			FileName:  "Spesifikasi_Teknis.pdf",
			FileSize:  1500000, // 1.5MB mock
			MimeType:  "application/pdf",
			CreatedAt: apptime.Now(),
			UpdatedAt: apptime.Now(),
		}
	}

	if err := u.rfqRepo.Create(ctx, rfq, attachment); err != nil {
		return dto.RFQResponse{}, err
	}

	// Trigger notifications to suppliers who are in this category (represented in seeder or simple auto-replies)
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
	u.db.WithContext(ctx).
		Table("categories").
		Where("id = ?", rfq.CategoryID).
		Pluck("name", &categoryName)

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
	catIDs := make([]string, len(rfqs))
	for i, r := range rfqs {
		catIDs[i] = r.CategoryID
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
		catName := catMap[rfq.CategoryID]
		if catName == "" {
			catName = utils.DefaultCategoryName
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

		// Look up message of type "offer" to parse price & moq from metadata
		var msg models.RFQMessage
		price := fmt.Sprintf("%s / Kg", utils.FormatMoney(12500, utils.DefaultCurrency()))
		moq := utils.DefaultMOQ

		if err := u.db.WithContext(ctx).
			Where("rfq_id = ? AND sender_id = ? AND message_type = ?", resolvedRFQID, supplier.ID, "offer").
			First(&msg).Error; err == nil {
			var metadataMap map[string]interface{}
			if errJson := json.Unmarshal([]byte(msg.Body), &metadataMap); errJson == nil {
				if p, ok := metadataMap["price"].(string); ok {
					price = p
				}
				if m, ok := metadataMap["moq"].(string); ok {
					moq = m
				}
			} else if errJson2 := json.Unmarshal([]byte(msg.Metadata), &metadataMap); errJson2 == nil {
				if p, ok := metadataMap["price"].(string); ok {
					price = p
				}
				if m, ok := metadataMap["moq"].(string); ok {
					moq = m
				}
			}
		}

		responseTimeStr := fmt.Sprintf("%d Jam", int(supplier.AvgResponseTimeMinutes/60))
		if supplier.AvgResponseTimeMinutes%60 != 0 {
			responseTimeStr = fmt.Sprintf("%.1f Jam", float64(supplier.AvgResponseTimeMinutes)/60)
		}
		if supplier.AvgResponseTimeMinutes == 0 {
			responseTimeStr = "2 Jam"
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
