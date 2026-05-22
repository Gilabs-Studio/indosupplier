package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
)

// AssetDisposalRepository defines interface for asset disposal operations
type AssetDisposalRepository interface {
	Create(ctx context.Context, disposal *models.AssetDisposal) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.AssetDisposal, error)
	GetByAssetID(ctx context.Context, assetID uuid.UUID) ([]models.AssetDisposal, error)
	GetByDateRange(ctx context.Context, tenantID uuid.UUID, dateFrom, dateTo string) ([]models.AssetDisposal, error)
	List(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]models.AssetDisposal, int64, error)
	Update(ctx context.Context, disposal *models.AssetDisposal) error
}

type assetDisposalRepository struct {
	db *gorm.DB
}

// NewAssetDisposalRepository creates a new asset disposal repository
func NewAssetDisposalRepository(db *gorm.DB) AssetDisposalRepository {
	return &assetDisposalRepository{db: db}
}

func (r *assetDisposalRepository) Create(ctx context.Context, disposal *models.AssetDisposal) error {
	db := database.GetDB(ctx, r.db)
	if disposal.ID == uuid.Nil {
		disposal.ID = uuid.New()
	}
	return db.WithContext(ctx).Create(disposal).Error
}

func (r *assetDisposalRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.AssetDisposal, error) {
	var disposal models.AssetDisposal
	err := database.GetDB(ctx, r.db).WithContext(ctx).Where("id = ?", id).First(&disposal).Error
	if err != nil {
		return nil, err
	}
	return &disposal, nil
}

func (r *assetDisposalRepository) GetByAssetID(ctx context.Context, assetID uuid.UUID) ([]models.AssetDisposal, error) {
	var disposals []models.AssetDisposal
	err := database.GetDB(ctx, r.db).WithContext(ctx).
		Where("asset_id = ?", assetID).
		Order("disposal_date DESC").
		Find(&disposals).Error
	return disposals, err
}

func (r *assetDisposalRepository) GetByDateRange(ctx context.Context, tenantID uuid.UUID, dateFrom, dateTo string) ([]models.AssetDisposal, error) {
	var disposals []models.AssetDisposal
	query := database.GetDB(ctx, r.db).WithContext(ctx).
		Where("tenant_id = ?", tenantID)

	if dateFrom != "" {
		query = query.Where("disposal_date >= ?", dateFrom)
	}
	if dateTo != "" {
		query = query.Where("disposal_date <= ?", dateTo)
	}

	err := query.Order("disposal_date DESC").Find(&disposals).Error
	return disposals, err
}

func (r *assetDisposalRepository) List(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]models.AssetDisposal, int64, error) {
	var disposals []models.AssetDisposal
	var total int64

	db := database.GetDB(ctx, r.db).WithContext(ctx).
		Where("tenant_id = ?", tenantID)

	if err := db.Model(&models.AssetDisposal{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Offset(offset).Limit(pageSize).Order("disposal_date DESC").Find(&disposals).Error
	return disposals, total, err
}

func (r *assetDisposalRepository) Update(ctx context.Context, disposal *models.AssetDisposal) error {
	return database.GetDB(ctx, r.db).WithContext(ctx).Model(disposal).Updates(disposal).Error
}
