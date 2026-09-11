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

	po := &buyerModels.PurchaseOrder{
		PONumber:          poNumber,
		BuyerProfileID:    buyerID,
		SupplierProfileID: req.SupplierProfileID,
		RFQID:             req.RFQID,
		ProductName:       req.ProductName,
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
		ProductName:       po.ProductName,
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

	return &dto.TransactionResponse{
		ID:                po.ID,
		PONumber:          po.PONumber,
		BuyerProfileID:    po.BuyerProfileID,
		SupplierProfileID: po.SupplierProfileID,
		SupplierName:      supplierName,
		RFQID:             po.RFQID,
		ProductName:       po.ProductName,
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

	var responseList []dto.TransactionResponse
	for i := range poList {
		po := &poList[i]
		supplierName := supplierMap[po.SupplierProfileID]
		if supplierName == "" {
			supplierName = utils.DefaultSupplierName
		}

		responseList = append(responseList, dto.TransactionResponse{
			ID:                po.ID,
			PONumber:          po.PONumber,
			BuyerProfileID:    po.BuyerProfileID,
			SupplierProfileID: po.SupplierProfileID,
			SupplierName:      supplierName,
			RFQID:             po.RFQID,
			ProductName:       po.ProductName,
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
