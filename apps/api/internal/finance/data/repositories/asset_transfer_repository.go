package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
)

// AssetTransferRepository defines interface for asset transfer operations
type AssetTransferRepository interface {
	Create(ctx context.Context, transfer *models.AssetTransfer) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.AssetTransfer, error)
	GetByAssetID(ctx context.Context, assetID uuid.UUID) ([]models.AssetTransfer, error)
	GetByDateRange(ctx context.Context, tenantID uuid.UUID, dateFrom, dateTo string) ([]models.AssetTransfer, error)
	List(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]models.AssetTransfer, int64, error)
	Update(ctx context.Context, transfer *models.AssetTransfer) error
}

type assetTransferRepository struct {
	db *gorm.DB
}

// NewAssetTransferRepository creates a new asset transfer repository
func NewAssetTransferRepository(db *gorm.DB) AssetTransferRepository {
	return &assetTransferRepository{db: db}
}

func (r *assetTransferRepository) Create(ctx context.Context, transfer *models.AssetTransfer) error {
	db := database.GetDB(ctx, r.db)
	if transfer.ID == uuid.Nil {
		transfer.ID = uuid.New()
	}
	return db.WithContext(ctx).Create(transfer).Error
}

func (r *assetTransferRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.AssetTransfer, error) {
	var transfer models.AssetTransfer
	err := database.GetDB(ctx, r.db).WithContext(ctx).Where("id = ?", id).First(&transfer).Error
	if err != nil {
		return nil, err
	}
	return &transfer, nil
}

func (r *assetTransferRepository) GetByAssetID(ctx context.Context, assetID uuid.UUID) ([]models.AssetTransfer, error) {
	var transfers []models.AssetTransfer
	err := database.GetDB(ctx, r.db).WithContext(ctx).
		Where("asset_id = ?", assetID).
		Order("transfer_date DESC").
		Find(&transfers).Error
	return transfers, err
}

func (r *assetTransferRepository) GetByDateRange(ctx context.Context, tenantID uuid.UUID, dateFrom, dateTo string) ([]models.AssetTransfer, error) {
	var transfers []models.AssetTransfer
	query := database.GetDB(ctx, r.db).WithContext(ctx).
		Where("tenant_id = ?", tenantID)

	if dateFrom != "" {
		query = query.Where("transfer_date >= ?", dateFrom)
	}
	if dateTo != "" {
		query = query.Where("transfer_date <= ?", dateTo)
	}

	err := query.Order("transfer_date DESC").Find(&transfers).Error
	return transfers, err
}

func (r *assetTransferRepository) List(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]models.AssetTransfer, int64, error) {
	var transfers []models.AssetTransfer
	var total int64

	db := database.GetDB(ctx, r.db).WithContext(ctx).
		Where("tenant_id = ?", tenantID)

	if err := db.Model(&models.AssetTransfer{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Offset(offset).Limit(pageSize).Order("transfer_date DESC").Find(&transfers).Error
	return transfers, total, err
}

func (r *assetTransferRepository) Update(ctx context.Context, transfer *models.AssetTransfer) error {
	return database.GetDB(ctx, r.db).WithContext(ctx).Model(transfer).Updates(transfer).Error
}
