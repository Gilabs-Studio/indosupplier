package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"
	"errors"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InventorySettingsRepository interface {
	FindByCompanyID(ctx context.Context, companyID string) (*financeModels.InventorySettings, error)
	Upsert(ctx context.Context, item *financeModels.InventorySettings) error
	SetLocked(ctx context.Context, companyID string) error
	GetAverageCostByProduct(ctx context.Context, companyID, productID string) (*financeModels.InventoryAverageCost, error)
	UpsertAverageCost(ctx context.Context, item *financeModels.InventoryAverageCost) error
	GetDB(ctx context.Context) *gorm.DB
}

type inventorySettingsRepository struct {
	db *gorm.DB
}

func NewInventorySettingsRepository(db *gorm.DB) InventorySettingsRepository {
	return &inventorySettingsRepository{db: db}
}

func (r *inventorySettingsRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value("tx").(*gorm.DB); ok && tx != nil {
		return tx
	}
	return database.GetDB(ctx, r.db)
}

func (r *inventorySettingsRepository) GetDB(ctx context.Context) *gorm.DB {
	return r.getDB(ctx)
}

func (r *inventorySettingsRepository) FindByCompanyID(ctx context.Context, companyID string) (*financeModels.InventorySettings, error) {
	var item financeModels.InventorySettings
	if err := r.getDB(ctx).Where("company_id = ?", companyID).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *inventorySettingsRepository) Upsert(ctx context.Context, item *financeModels.InventorySettings) error {
	db := r.getDB(ctx)
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "company_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"valuation_method", "is_locked", "updated_at"}),
	}).Create(item).Error
}

func (r *inventorySettingsRepository) SetLocked(ctx context.Context, companyID string) error {
	return r.getDB(ctx).
		Model(&financeModels.InventorySettings{}).
		Where("company_id = ?", companyID).
		Updates(map[string]interface{}{"is_locked": true}).Error
}

func (r *inventorySettingsRepository) GetAverageCostByProduct(ctx context.Context, companyID, productID string) (*financeModels.InventoryAverageCost, error) {
	var item financeModels.InventoryAverageCost
	err := r.getDB(ctx).Where("company_id = ? AND product_id = ?", companyID, productID).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *inventorySettingsRepository) UpsertAverageCost(ctx context.Context, item *financeModels.InventoryAverageCost) error {
	db := r.getDB(ctx)
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "company_id"}, {Name: "product_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"average_cost", "total_quantity", "total_value", "last_updated", "updated_at"}),
	}).Create(item).Error
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
