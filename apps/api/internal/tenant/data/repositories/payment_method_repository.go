package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/tenant/data/models"
	"gorm.io/gorm"
)

// PaymentMethodRepository handles persistence for tenant saved payment methods.
type PaymentMethodRepository interface {
	ListByTenant(ctx context.Context, tenantID string) ([]*models.TenantPaymentMethod, error)
	FindByID(ctx context.Context, id, tenantID string) (*models.TenantPaymentMethod, error)
	Create(ctx context.Context, m *models.TenantPaymentMethod) error
	SetDefault(ctx context.Context, id, tenantID string) error
	Delete(ctx context.Context, id, tenantID string) error
}

type paymentMethodRepository struct {
	db *gorm.DB
}

// NewPaymentMethodRepository creates a PaymentMethodRepository.
func NewPaymentMethodRepository(db *gorm.DB) PaymentMethodRepository {
	return &paymentMethodRepository{db: db}
}

func (r *paymentMethodRepository) ListByTenant(ctx context.Context, tenantID string) ([]*models.TenantPaymentMethod, error) {
	var methods []*models.TenantPaymentMethod
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("is_default DESC, created_at DESC").
		Find(&methods).Error
	return methods, err
}

func (r *paymentMethodRepository) FindByID(ctx context.Context, id, tenantID string) (*models.TenantPaymentMethod, error) {
	var m models.TenantPaymentMethod
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).
		First(&m).Error
	return &m, err
}

func (r *paymentMethodRepository) Create(ctx context.Context, m *models.TenantPaymentMethod) error {
	return r.db.WithContext(ctx).Create(m).Error
}

// SetDefault clears the default flag on all methods for the tenant then sets it on the given ID.
func (r *paymentMethodRepository) SetDefault(ctx context.Context, id, tenantID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.TenantPaymentMethod{}).
			Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
			Update("is_default", false).Error; err != nil {
			return err
		}
		return tx.Model(&models.TenantPaymentMethod{}).
			Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).
			Update("is_default", true).Error
	})
}

func (r *paymentMethodRepository) Delete(ctx context.Context, id, tenantID string) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, tenantID).
		Delete(&models.TenantPaymentMethod{}).Error
}
