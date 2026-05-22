package repositories

import (
	"context"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/tenant/data/models"
	"gorm.io/gorm"
)

// SubscriptionRepository defines data-access operations for TenantSubscription.
type SubscriptionRepository interface {
	Create(ctx context.Context, sub *models.TenantSubscription) error
	FindActiveByTenantID(ctx context.Context, tenantID string) (*models.TenantSubscription, error)
	// FindByXenditInvoiceID retrieves a subscription by the stored Xendit invoice ID.
	FindByXenditInvoiceID(ctx context.Context, invoiceID string) (*models.TenantSubscription, error)
	ListByTenantID(ctx context.Context, tenantID string) ([]*models.TenantSubscription, error)
	ListAll(ctx context.Context, page, perPage int) ([]*models.TenantSubscription, int64, error)
	ExpireOldSubscriptions(ctx context.Context) error
	// ExtendSubscription pushes the next billing date and records the new invoice ID
	// after a successful recurring payment.
	ExtendSubscription(ctx context.Context, id uint, nextBillingAt time.Time, newInvoiceID string) error
}

type subscriptionRepository struct {
	db *gorm.DB
}

// NewSubscriptionRepository creates a new SubscriptionRepository.
func NewSubscriptionRepository(db *gorm.DB) SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

func (r *subscriptionRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *subscriptionRepository) Create(ctx context.Context, sub *models.TenantSubscription) error {
	return r.getDB(ctx).Create(sub).Error
}

func (r *subscriptionRepository) FindActiveByTenantID(ctx context.Context, tenantID string) (*models.TenantSubscription, error) {
	var sub models.TenantSubscription
	now := time.Now()
	err := r.getDB(ctx).
		Preload("Coupon").
		Where(
			"tenant_id = ? AND deleted_at IS NULL AND status IN ? AND (expires_at IS NULL OR expires_at > ?)",
			tenantID,
			[]string{"active", "trial", "past_due", "suspended", "cancelled"},
			now,
		).
		Order("starts_at DESC").
		First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) ListByTenantID(ctx context.Context, tenantID string) ([]*models.TenantSubscription, error) {
	var subs []*models.TenantSubscription
	err := r.getDB(ctx).
		Preload("Coupon").
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("starts_at DESC").
		Find(&subs).Error
	return subs, err
}

func (r *subscriptionRepository) ListAll(ctx context.Context, page, perPage int) ([]*models.TenantSubscription, int64, error) {
	var subs []*models.TenantSubscription
	var total int64

	q := r.getDB(ctx).Model(&models.TenantSubscription{}).Where("deleted_at IS NULL")
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	err := q.Preload("Coupon").
		Order("created_at DESC").
		Offset(offset).Limit(perPage).
		Find(&subs).Error
	return subs, total, err
}

// ExpireOldSubscriptions marks subscriptions as expired when their expires_at has passed.
func (r *subscriptionRepository) ExpireOldSubscriptions(ctx context.Context) error {
	return r.getDB(ctx).
		Model(&models.TenantSubscription{}).
		Where("status = ? AND expires_at IS NOT NULL AND expires_at < ?", models.SubscriptionActive, time.Now()).
		Update("status", models.SubscriptionExpired).Error
}

func (r *subscriptionRepository) FindByXenditInvoiceID(ctx context.Context, invoiceID string) (*models.TenantSubscription, error) {
	var sub models.TenantSubscription
	err := r.db.WithContext(ctx).
		Where("xendit_invoice_id = ? AND deleted_at IS NULL", invoiceID).
		First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// ExtendSubscription records the result of a successful recurring payment by advancing
// the next billing date and storing the new Xendit invoice ID for the next cycle.
func (r *subscriptionRepository) ExtendSubscription(ctx context.Context, id uint, nextBillingAt time.Time, newInvoiceID string) error {
	return r.db.WithContext(ctx).
		Model(&models.TenantSubscription{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"next_billing_at":   nextBillingAt,
			"xendit_invoice_id": newInvoiceID,
			"status":            models.SubscriptionActive,
		}).Error
}
