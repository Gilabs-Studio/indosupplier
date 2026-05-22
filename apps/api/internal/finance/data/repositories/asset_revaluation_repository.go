package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
)

// AssetRevaluationRepository defines interface for asset revaluation operations
type AssetRevaluationRepository interface {
	Create(ctx context.Context, revaluation *models.AssetRevaluation) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.AssetRevaluation, error)
	GetByAssetID(ctx context.Context, assetID uuid.UUID) ([]models.AssetRevaluation, error)
	GetByDateRange(ctx context.Context, tenantID uuid.UUID, dateFrom, dateTo string) ([]models.AssetRevaluation, error)
	List(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]models.AssetRevaluation, int64, error)
	Update(ctx context.Context, revaluation *models.AssetRevaluation) error
}

type assetRevaluationRepository struct {
	db *gorm.DB
}

// NewAssetRevaluationRepository creates a new asset revaluation repository
func NewAssetRevaluationRepository(db *gorm.DB) AssetRevaluationRepository {
	return &assetRevaluationRepository{db: db}
}

func (r *assetRevaluationRepository) Create(ctx context.Context, revaluation *models.AssetRevaluation) error {
	db := database.GetDB(ctx, r.db)
	if revaluation.ID == uuid.Nil {
		revaluation.ID = uuid.New()
	}
	return db.WithContext(ctx).Create(revaluation).Error
}

func (r *assetRevaluationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.AssetRevaluation, error) {
	var revaluation models.AssetRevaluation
	err := database.GetDB(ctx, r.db).WithContext(ctx).Where("id = ?", id).First(&revaluation).Error
	if err != nil {
		return nil, err
	}
	return &revaluation, nil
}

func (r *assetRevaluationRepository) GetByAssetID(ctx context.Context, assetID uuid.UUID) ([]models.AssetRevaluation, error) {
	var revaluations []models.AssetRevaluation
	err := database.GetDB(ctx, r.db).WithContext(ctx).
		Where("asset_id = ?", assetID).
		Order("revaluation_date DESC").
		Find(&revaluations).Error
	return revaluations, err
}

func (r *assetRevaluationRepository) GetByDateRange(ctx context.Context, tenantID uuid.UUID, dateFrom, dateTo string) ([]models.AssetRevaluation, error) {
	var revaluations []models.AssetRevaluation
	query := database.GetDB(ctx, r.db).WithContext(ctx).
		Where("tenant_id = ?", tenantID)

	if dateFrom != "" {
		query = query.Where("revaluation_date >= ?", dateFrom)
	}
	if dateTo != "" {
		query = query.Where("revaluation_date <= ?", dateTo)
	}

	err := query.Order("revaluation_date DESC").Find(&revaluations).Error
	return revaluations, err
}

func (r *assetRevaluationRepository) List(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]models.AssetRevaluation, int64, error) {
	var revaluations []models.AssetRevaluation
	var total int64

	db := database.GetDB(ctx, r.db).WithContext(ctx).
		Where("tenant_id = ?", tenantID)

	if err := db.Model(&models.AssetRevaluation{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Offset(offset).Limit(pageSize).Order("revaluation_date DESC").Find(&revaluations).Error
	return revaluations, total, err
}

func (r *assetRevaluationRepository) Update(ctx context.Context, revaluation *models.AssetRevaluation) error {
	return database.GetDB(ctx, r.db).WithContext(ctx).Model(revaluation).Updates(revaluation).Error
}
