package repositories

import (
	"context"

	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
	monetizationModels "github.com/gilabs/indosupplier/api/internal/monetization/data/models"
	"github.com/gilabs/indosupplier/api/internal/supplier/data/models"
)

type PortalRepository interface {
	GetProfileByUserID(ctx context.Context, userID string) (*models.SupplierProfile, error)
	UpdateProfile(ctx context.Context, profile *models.SupplierProfile) error
	GetActiveSubscription(ctx context.Context, supplierProfileID string) (*monetizationModels.SupplierSubscription, error)
	GetSubscriptionPlanByID(ctx context.Context, planID string) (*monetizationModels.SubscriptionPlan, error)
	GetSubscriptionPlanByCode(ctx context.Context, code string) (*monetizationModels.SubscriptionPlan, error)
	CreateSubscriptionPlan(ctx context.Context, plan *monetizationModels.SubscriptionPlan) error
	CreateSubscription(ctx context.Context, sub *monetizationModels.SupplierSubscription) error
	UpdateSubscription(ctx context.Context, sub *monetizationModels.SupplierSubscription) error
	GetInvoices(ctx context.Context, supplierProfileID string) ([]monetizationModels.Invoice, error)
	GetPaymentByID(ctx context.Context, paymentID string) (*monetizationModels.Payment, error)
	CreateInvoice(ctx context.Context, inv *monetizationModels.Invoice) error
	CreatePayment(ctx context.Context, pay *monetizationModels.Payment) error
}

type portalRepository struct {
	db *gorm.DB
}

func NewPortalRepository(db *gorm.DB) PortalRepository {
	return &portalRepository{db: db}
}

func (r *portalRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *portalRepository) GetProfileByUserID(ctx context.Context, userID string) (*models.SupplierProfile, error) {
	var profile models.SupplierProfile
	err := r.getDB(ctx).Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *portalRepository) UpdateProfile(ctx context.Context, profile *models.SupplierProfile) error {
	return r.getDB(ctx).Save(profile).Error
}

func (r *portalRepository) GetActiveSubscription(ctx context.Context, supplierProfileID string) (*monetizationModels.SupplierSubscription, error) {
	var sub monetizationModels.SupplierSubscription
	err := r.getDB(ctx).Where("supplier_profile_id = ? AND status = ?", supplierProfileID, "active").First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *portalRepository) GetSubscriptionPlanByID(ctx context.Context, planID string) (*monetizationModels.SubscriptionPlan, error) {
	var plan monetizationModels.SubscriptionPlan
	err := r.getDB(ctx).Where("id = ?", planID).First(&plan).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *portalRepository) GetSubscriptionPlanByCode(ctx context.Context, code string) (*monetizationModels.SubscriptionPlan, error) {
	var plan monetizationModels.SubscriptionPlan
	err := r.getDB(ctx).Where("code = ? AND is_active = ?", code, true).First(&plan).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *portalRepository) CreateSubscriptionPlan(ctx context.Context, plan *monetizationModels.SubscriptionPlan) error {
	return r.getDB(ctx).Create(plan).Error
}

func (r *portalRepository) CreateSubscription(ctx context.Context, sub *monetizationModels.SupplierSubscription) error {
	return r.getDB(ctx).Create(sub).Error
}

func (r *portalRepository) UpdateSubscription(ctx context.Context, sub *monetizationModels.SupplierSubscription) error {
	return r.getDB(ctx).Save(sub).Error
}

func (r *portalRepository) GetInvoices(ctx context.Context, supplierProfileID string) ([]monetizationModels.Invoice, error) {
	var invoices []monetizationModels.Invoice
	// We need to join payments to fetch invoices related to this supplier
	err := r.getDB(ctx).
		Joins("JOIN payments ON payments.id = invoices.payment_id").
		Where("payments.supplier_profile_id = ?", supplierProfileID).
		Order("invoices.created_at DESC").
		Find(&invoices).Error
	if err != nil {
		return nil, err
	}
	return invoices, nil
}

func (r *portalRepository) GetPaymentByID(ctx context.Context, paymentID string) (*monetizationModels.Payment, error) {
	var payment monetizationModels.Payment
	if err := r.getDB(ctx).Where("id = ?", paymentID).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *portalRepository) CreateInvoice(ctx context.Context, inv *monetizationModels.Invoice) error {
	return r.getDB(ctx).Create(inv).Error
}

func (r *portalRepository) CreatePayment(ctx context.Context, pay *monetizationModels.Payment) error {
	return r.getDB(ctx).Create(pay).Error
}
