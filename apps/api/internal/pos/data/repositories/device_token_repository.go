package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/pos/data/models"
	"gorm.io/gorm"
)

type POSDeviceTokenRepository interface {
	Upsert(ctx context.Context, token *models.POSDeviceToken) error
	FindByScope(ctx context.Context, tenantID, outletID string) ([]models.POSDeviceToken, error)
}

type posDeviceTokenRepository struct {
	db *gorm.DB
}

func NewPOSDeviceTokenRepository(db *gorm.DB) POSDeviceTokenRepository {
	return &posDeviceTokenRepository{db: db}
}

func (r *posDeviceTokenRepository) Upsert(ctx context.Context, token *models.POSDeviceToken) error {
	var existing models.POSDeviceToken
	err := database.GetDB(ctx, r.db).
		Where("token = ?", token.Token).
		First(&existing).Error
	if err == nil {
		existing.UserID = token.UserID
		existing.TenantID = token.TenantID
		existing.OutletID = token.OutletID
		existing.Platform = token.Platform
		token.ID = existing.ID
		return database.GetDB(ctx, r.db).Save(&existing).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	return database.GetDB(ctx, r.db).Create(token).Error
}

func (r *posDeviceTokenRepository) FindByScope(ctx context.Context, tenantID, outletID string) ([]models.POSDeviceToken, error) {
	var tokens []models.POSDeviceToken
	err := database.GetDB(ctx, r.db).
		Where("tenant_id = ? AND outlet_id = ?", tenantID, outletID).
		Find(&tokens).Error
	return tokens, err
}
