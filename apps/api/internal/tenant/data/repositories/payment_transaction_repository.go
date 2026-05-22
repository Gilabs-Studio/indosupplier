package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/tenant/data/models"
	"gorm.io/gorm"
)

// PaymentTransactionRepository defines data-access operations for PaymentTransaction.
type PaymentTransactionRepository interface {
	Create(ctx context.Context, txn *models.PaymentTransaction) error
	FindByID(ctx context.Context, id string) (*models.PaymentTransaction, error)
	FindByProviderInvoiceID(ctx context.Context, invoiceID string) (*models.PaymentTransaction, error)
	ListByTenantID(ctx context.Context, tenantID string, page, perPage int) ([]*models.PaymentTransaction, int64, error)
	ListBySubscriptionID(ctx context.Context, subscriptionID string, page, perPage int) ([]*models.PaymentTransaction, int64, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) error
	UpdateStatus(ctx context.Context, id string, status models.PaymentTransactionStatus) error
}

type paymentTransactionRepository struct {
	db *gorm.DB
}

// NewPaymentTransactionRepository creates a new PaymentTransactionRepository.
func NewPaymentTransactionRepository(db *gorm.DB) PaymentTransactionRepository {
	return &paymentTransactionRepository{db: db}
}

func (r *paymentTransactionRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *paymentTransactionRepository) Create(ctx context.Context, txn *models.PaymentTransaction) error {
	return r.getDB(ctx).Create(txn).Error
}

func (r *paymentTransactionRepository) FindByID(ctx context.Context, id string) (*models.PaymentTransaction, error) {
	var txn models.PaymentTransaction
	err := r.getDB(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&txn).Error
	if err != nil {
		return nil, err
	}
	return &txn, nil
}

func (r *paymentTransactionRepository) FindByProviderInvoiceID(ctx context.Context, invoiceID string) (*models.PaymentTransaction, error) {
	var txn models.PaymentTransaction
	err := r.getDB(ctx).Where("provider_invoice_id = ? AND deleted_at IS NULL", invoiceID).First(&txn).Error
	if err != nil {
		return nil, err
	}
	return &txn, nil
}

func (r *paymentTransactionRepository) ListByTenantID(ctx context.Context, tenantID string, page, perPage int) ([]*models.PaymentTransaction, int64, error) {
	var txns []*models.PaymentTransaction
	var total int64

	q := r.getDB(ctx).Model(&models.PaymentTransaction{}).Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	err := q.
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&txns).Error
	return txns, total, err
}

func (r *paymentTransactionRepository) ListBySubscriptionID(ctx context.Context, subscriptionID string, page, perPage int) ([]*models.PaymentTransaction, int64, error) {
	var txns []*models.PaymentTransaction
	var total int64

	q := r.getDB(ctx).Model(&models.PaymentTransaction{}).Where("subscription_id = ? AND deleted_at IS NULL", subscriptionID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	err := q.
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&txns).Error
	return txns, total, err
}

func (r *paymentTransactionRepository) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	return r.getDB(ctx).
		Model(&models.PaymentTransaction{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(updates).Error
}

func (r *paymentTransactionRepository) UpdateStatus(ctx context.Context, id string, status models.PaymentTransactionStatus) error {
	return r.getDB(ctx).
		Model(&models.PaymentTransaction{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("status", status).Error
}
