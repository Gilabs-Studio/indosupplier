package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"

	"github.com/gilabs/gims/api/internal/organization/data/models"
	"gorm.io/gorm"
)

type EmployeeAssetRepository interface {
	Create(ctx context.Context, asset *models.EmployeeAsset) error
	Update(ctx context.Context, asset *models.EmployeeAsset) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*models.EmployeeAsset, error)
	FindByEmployeeID(ctx context.Context, employeeID string) ([]*models.EmployeeAsset, error)
	FindBorrowedByEmployeeID(ctx context.Context, employeeID string) ([]*models.EmployeeAsset, error)
	FindByAssetCode(ctx context.Context, assetCode string) (*models.EmployeeAsset, error)
	FindByAssetID(ctx context.Context, assetID string) (*models.EmployeeAsset, error)
}

type employeeAssetRepository struct {
	db *gorm.DB
}

func NewEmployeeAssetRepository(db *gorm.DB) EmployeeAssetRepository {
	return &employeeAssetRepository{db: db}
}

func (r *employeeAssetRepository) Create(ctx context.Context, asset *models.EmployeeAsset) error {
	return database.GetDB(ctx, r.db).Create(asset).Error
}

func (r *employeeAssetRepository) Update(ctx context.Context, asset *models.EmployeeAsset) error {
	return database.GetDB(ctx, r.db).Save(asset).Error
}

func (r *employeeAssetRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.EmployeeAsset{}, "id = ?", id).Error
}

func (r *employeeAssetRepository) FindByID(ctx context.Context, id string) (*models.EmployeeAsset, error) {
	var asset models.EmployeeAsset
	err := database.GetDB(ctx, r.db).First(&asset, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *employeeAssetRepository) FindByEmployeeID(ctx context.Context, employeeID string) ([]*models.EmployeeAsset, error) {
	var assets []*models.EmployeeAsset
	err := database.GetDB(ctx, r.db).
		Where("employee_id = ?", employeeID).
		Order("borrow_date DESC").
		Find(&assets).Error
	if err != nil {
		return nil, err
	}
	return assets, nil
}

func (r *employeeAssetRepository) FindBorrowedByEmployeeID(ctx context.Context, employeeID string) ([]*models.EmployeeAsset, error) {
	var assets []*models.EmployeeAsset
	err := database.GetDB(ctx, r.db).
		Where("employee_id = ? AND return_date IS NULL", employeeID).
		Order("borrow_date DESC").
		Find(&assets).Error
	if err != nil {
		return nil, err
	}
	return assets, nil
}

func (r *employeeAssetRepository) FindByAssetCode(ctx context.Context, assetCode string) (*models.EmployeeAsset, error) {
	var asset models.EmployeeAsset
	err := database.GetDB(ctx, r.db).
		Where("asset_code = ? AND return_date IS NULL", assetCode).
		First(&asset).Error
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *employeeAssetRepository) FindByAssetID(ctx context.Context, assetID string) (*models.EmployeeAsset, error) {
	var asset models.EmployeeAsset
	err := database.GetDB(ctx, r.db).
		Where("asset_id = ?", assetID).
		Order("borrow_date DESC").
		First(&asset).Error
	if err != nil {
		return nil, err
	}
	return &asset, nil
}
