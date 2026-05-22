package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"

	"github.com/gilabs/gims/api/internal/pos/data/models"
	"gorm.io/gorm"
)

// POSPaymentRepository defines data access for POS payments
type POSPaymentRepository interface {
	Create(ctx context.Context, payment *models.POSPayment) error
	Update(ctx context.Context, payment *models.POSPayment) error
	FindByID(ctx context.Context, id string) (*models.POSPayment, error)
	FindByOrderID(ctx context.Context, orderID string) ([]models.POSPayment, error)
	FindByExternalOrderID(ctx context.Context, externalOrderID string) (*models.POSPayment, error)
}

type posPaymentRepository struct {
	db *gorm.DB
}

// NewPOSPaymentRepository creates the concrete implementation
func NewPOSPaymentRepository(db *gorm.DB) POSPaymentRepository {
	return &posPaymentRepository{db: db}
}

func (r *posPaymentRepository) Create(ctx context.Context, payment *models.POSPayment) error {
	return database.GetDB(ctx, r.db).Create(payment).Error
}

func (r *posPaymentRepository) Update(ctx context.Context, payment *models.POSPayment) error {
	return database.GetDB(ctx, r.db).Save(payment).Error
}

func (r *posPaymentRepository) FindByID(ctx context.Context, id string) (*models.POSPayment, error) {
	var payment models.POSPayment
	err := database.GetDB(ctx, r.db).First(&payment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *posPaymentRepository) FindByOrderID(ctx context.Context, orderID string) ([]models.POSPayment, error) {
	var payments []models.POSPayment
	err := database.GetDB(ctx, r.db).
		Where("order_id = ?", orderID).
		Order("created_at DESC").
		Find(&payments).Error
	return payments, err
}

func (r *posPaymentRepository) FindByExternalOrderID(ctx context.Context, externalOrderID string) (*models.POSPayment, error) {
	var payment models.POSPayment
	err := database.GetDB(ctx, r.db).
		Where("external_order_id = ?", externalOrderID).
		First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}
