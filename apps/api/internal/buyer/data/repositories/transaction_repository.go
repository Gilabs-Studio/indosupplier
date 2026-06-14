package repositories

import (
	"context"

	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *models.PurchaseOrder) error
	FindByID(ctx context.Context, id string) (*models.PurchaseOrder, error)
	FindByIDAndBuyer(ctx context.Context, id string, buyerProfileID string) (*models.PurchaseOrder, error)
	List(ctx context.Context, buyerProfileID string, status string, page int, perPage int) ([]models.PurchaseOrder, int64, error)
	UpdateStatus(ctx context.Context, id string, status string, paymentStatus string) error
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *transactionRepository) Create(ctx context.Context, tx *models.PurchaseOrder) error {
	return r.getDB(ctx).Create(tx).Error
}

func (r *transactionRepository) FindByID(ctx context.Context, id string) (*models.PurchaseOrder, error) {
	var tx models.PurchaseOrder
	if err := r.getDB(ctx).Where("id = ?", id).First(&tx).Error; err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *transactionRepository) FindByIDAndBuyer(ctx context.Context, id string, buyerProfileID string) (*models.PurchaseOrder, error) {
	var tx models.PurchaseOrder
	if err := r.getDB(ctx).Where("id = ? AND buyer_profile_id = ?", id, buyerProfileID).First(&tx).Error; err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *transactionRepository) List(ctx context.Context, buyerProfileID string, status string, page int, perPage int) ([]models.PurchaseOrder, int64, error) {
	var txs []models.PurchaseOrder
	var total int64

	query := r.getDB(ctx).Model(&models.PurchaseOrder{}).Where("buyer_profile_id = ?", buyerProfileID)

	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&txs).Error

	if err != nil {
		return nil, 0, err
	}

	return txs, total, nil
}

func (r *transactionRepository) UpdateStatus(ctx context.Context, id string, status string, paymentStatus string) error {
	updates := map[string]interface{}{}
	if status != "" {
		updates["status"] = status
	}
	if paymentStatus != "" {
		updates["payment_status"] = paymentStatus
	}
	if len(updates) == 0 {
		return nil
	}
	return r.getDB(ctx).Model(&models.PurchaseOrder{}).Where("id = ?", id).Updates(updates).Error
}
