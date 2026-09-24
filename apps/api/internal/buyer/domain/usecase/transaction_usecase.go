package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	buyerModels "github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	"github.com/gilabs/indosupplier/api/internal/buyer/data/repositories"
	"github.com/gilabs/indosupplier/api/internal/buyer/domain/dto"
	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	"github.com/gilabs/indosupplier/api/internal/core/utils"
	supplierModels "github.com/gilabs/indosupplier/api/internal/supplier/data/models"
)

var (
	ErrBuyerProfileNotFound    = errors.New("buyer profile not found")
	ErrSupplierProfileNotFound = errors.New("supplier profile not found")
	ErrTransactionNotFound     = errors.New("transaction not found")
)

type TransactionUsecase interface {
	Create(ctx context.Context, userID string, req *dto.CreateTransactionRequest) (*dto.TransactionResponse, error)
	GetByID(ctx context.Context, userID string, id string) (*dto.TransactionResponse, error)
	List(ctx context.Context, userID string, req *dto.ListTransactionsRequest) ([]dto.TransactionResponse, *Pagination, error)
}

type Pagination struct {
	Page       int
	PerPage    int
	Total      int64
	TotalPages int
}

type transactionUsecase struct {
	db              *gorm.DB
	transactionRepo repositories.TransactionRepository
}

func NewTransactionUsecase(db *gorm.DB, transactionRepo repositories.TransactionRepository) TransactionUsecase {
	return &transactionUsecase{
		db:              db,
		transactionRepo: transactionRepo,
	}
}

func (u *transactionUsecase) getBuyerProfileID(ctx context.Context, userID string) (string, error) {
	var buyer buyerModels.BuyerProfile
	if err := u.db.WithContext(ctx).Where("user_id = ?", userID).First(&buyer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrBuyerProfileNotFound
		}
		return "", err
	}
	return buyer.ID, nil
}

func (u *transactionUsecase) getSupplierName(ctx context.Context, supplierID string) (string, error) {
	var supplier supplierModels.SupplierProfile
	if err := u.db.WithContext(ctx).Where("id = ?", supplierID).First(&supplier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrSupplierProfileNotFound
		}
		return "", err
	}
	return supplier.CompanyName, nil
}

func (u *transactionUsecase) Create(ctx context.Context, userID string, req *dto.CreateTransactionRequest) (*dto.TransactionResponse, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}

	supplierName, err := u.getSupplierName(ctx, req.SupplierProfileID)
	if err != nil {
		return nil, err
	}

	// Generate PONumber: PO-YYYYMMDD-XXXXXX
	dateStr := apptime.Now().Format("20060102")
	randomStr := uuid.New().String()[:6]
	poNumber := fmt.Sprintf("PO-%s-%s", dateStr, strings.ToUpper(randomStr))

	totalAmount := req.QuantityValue * req.PricePerUnit

	productImage := req.ProductImage
	if productImage == "" {
		var photoURL string
		q := u.db.WithContext(ctx).Table("supplier_products sp").
			Select("spp.file_url").
			Joins("JOIN supplier_product_photos spp ON spp.supplier_product_id = sp.id").
			Where("sp.supplier_profile_id = ?", req.SupplierProfileID)
		if req.ProductID != nil && *req.ProductID != "" {
			q = q.Where("sp.id = ?", *req.ProductID)
		} else {
			q = q.Where("LOWER(sp.name) = LOWER(?)", req.ProductName)
		}
		if err := q.Order("spp.sort_order ASC").Limit(1).Scan(&photoURL).Error; err == nil && photoURL != "" {
			productImage = photoURL
		}
	}
	if productImage == "" {
		productImage = "/images/categories/cat-bahan-baku.webp"
	}

	po := &buyerModels.PurchaseOrder{
		PONumber:          poNumber,
		BuyerProfileID:    buyerID,
		SupplierProfileID: req.SupplierProfileID,
		RFQID:             req.RFQID,
		ProductID:         req.ProductID,
		ProductName:       req.ProductName,
		ProductImage:      productImage,
		QuantityValue:     req.QuantityValue,
		QuantityUnit:      req.QuantityUnit,
		PricePerUnit:      req.PricePerUnit,
		TotalAmount:       totalAmount,
		Status:            "pending",
		PaymentStatus:     "unpaid",
		DeliveryAddress:   req.DeliveryAddress,
		Notes:             req.Notes,
	}

	if err := u.transactionRepo.Create(ctx, po); err != nil {
		return nil, err
	}

	return &dto.TransactionResponse{
		ID:                po.ID,
		PONumber:          po.PONumber,
		BuyerProfileID:    po.BuyerProfileID,
		SupplierProfileID: po.SupplierProfileID,
		SupplierName:      supplierName,
		RFQID:             po.RFQID,
		ProductID:         po.ProductID,
		ProductName:       po.ProductName,
		ProductImage:      po.ProductImage,
		QuantityValue:     po.QuantityValue,
		QuantityUnit:      po.QuantityUnit,
		PricePerUnit:      po.PricePerUnit,
		TotalAmount:       po.TotalAmount,
		Status:            po.Status,
		PaymentStatus:     po.PaymentStatus,
		DeliveryAddress:   po.DeliveryAddress,
		Notes:             po.Notes,
		CreatedAt:         po.CreatedAt,
		UpdatedAt:         po.UpdatedAt,
	}, nil
}

func (u *transactionUsecase) GetByID(ctx context.Context, userID string, id string) (*dto.TransactionResponse, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}

	po, err := u.transactionRepo.FindByIDAndBuyer(ctx, id, buyerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}

	supplierName, err := u.getSupplierName(ctx, po.SupplierProfileID)
	if err != nil {
		supplierName = utils.DefaultSupplierName
	}

	productImage := po.ProductImage
	if productImage == "" {
		productImage = "/images/categories/cat-bahan-baku.webp"
	}

	var reviewCount int64
	u.db.WithContext(ctx).Table("supplier_reviews").Where("purchase_order_id = ?", po.ID).Count(&reviewCount)

	return &dto.TransactionResponse{
		ID:                po.ID,
		PONumber:          po.PONumber,
		BuyerProfileID:    po.BuyerProfileID,
		SupplierProfileID: po.SupplierProfileID,
		SupplierName:      supplierName,
		RFQID:             po.RFQID,
		ProductID:         po.ProductID,
		ProductName:       po.ProductName,
		ProductImage:      productImage,
		QuantityValue:     po.QuantityValue,
		QuantityUnit:      po.QuantityUnit,
		PricePerUnit:      po.PricePerUnit,
		TotalAmount:       po.TotalAmount,
		Status:            po.Status,
		PaymentStatus:     po.PaymentStatus,
		HasReviewed:       reviewCount > 0,
		DeliveryAddress:   po.DeliveryAddress,
		Notes:             po.Notes,
		CreatedAt:         po.CreatedAt,
		UpdatedAt:         po.UpdatedAt,
	}, nil
}

func (u *transactionUsecase) List(ctx context.Context, userID string, req *dto.ListTransactionsRequest) ([]dto.TransactionResponse, *Pagination, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	page, perPage := utils.NormalizePagination(req.Page, req.PerPage, 10)

	poList, total, err := u.transactionRepo.List(ctx, buyerID, req.Status, page, perPage)
	if err != nil {
		return nil, nil, err
	}

	totalPages := utils.TotalPages(total, perPage)

	// Batch fetch supplier names to eliminate N+1 queries
	supplierMap := make(map[string]string)
	if len(poList) > 0 {
		supplierIDs := make([]string, 0, len(poList))
		for _, po := range poList {
			if po.SupplierProfileID != "" {
				supplierIDs = append(supplierIDs, po.SupplierProfileID)
			}
		}
		if len(supplierIDs) > 0 {
			var suppliers []supplierModels.SupplierProfile
			if err := u.db.WithContext(ctx).Where("id IN ?", supplierIDs).Find(&suppliers).Error; err == nil {
				for _, s := range suppliers {
					supplierMap[s.ID] = s.CompanyName
				}
			}
		}
	}

	// Batch check which completed POs have already been reviewed to eliminate N+1 queries
	reviewedPOMap := make(map[string]bool)
	if len(poList) > 0 {
		completedPOIDs := make([]string, 0, len(poList))
		for _, po := range poList {
			if po.Status == "completed" {
				completedPOIDs = append(completedPOIDs, po.ID)
			}
		}
		if len(completedPOIDs) > 0 {
			var reviewedIDs []string
			if err := u.db.WithContext(ctx).Table("supplier_reviews").
				Where("purchase_order_id IN ?", completedPOIDs).
				Pluck("purchase_order_id", &reviewedIDs).Error; err == nil {
				for _, id := range reviewedIDs {
					reviewedPOMap[id] = true
				}
			}
		}
	}

	var responseList []dto.TransactionResponse
	for i := range poList {
		po := &poList[i]
		supplierName := supplierMap[po.SupplierProfileID]
		if supplierName == "" {
			supplierName = utils.DefaultSupplierName
		}

		productImage := po.ProductImage
		if productImage == "" {
			productImage = "/images/categories/cat-bahan-baku.webp"
		}

		responseList = append(responseList, dto.TransactionResponse{
			ID:                po.ID,
			PONumber:          po.PONumber,
			BuyerProfileID:    po.BuyerProfileID,
			SupplierProfileID: po.SupplierProfileID,
			SupplierName:      supplierName,
			RFQID:             po.RFQID,
			ProductID:         po.ProductID,
			ProductName:       po.ProductName,
			ProductImage:      productImage,
			QuantityValue:     po.QuantityValue,
			QuantityUnit:      po.QuantityUnit,
			PricePerUnit:      po.PricePerUnit,
			TotalAmount:       po.TotalAmount,
			Status:            po.Status,
			PaymentStatus:     po.PaymentStatus,
			HasReviewed:       reviewedPOMap[po.ID],
			DeliveryAddress:   po.DeliveryAddress,
			Notes:             po.Notes,
			CreatedAt:         po.CreatedAt,
			UpdatedAt:         po.UpdatedAt,
		})
	}

	pagination := &Pagination{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}

	return responseList, pagination, nil
}
