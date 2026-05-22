package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"
	"time"

	"github.com/gilabs/gims/api/internal/finance/data/models"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AssetAttachmentRepository defines the interface for asset attachment data access
type AssetAttachmentRepository interface {
	// CRUD operations
	Create(ctx context.Context, attachment *models.AssetAttachment) error
	GetByID(ctx context.Context, id string) (*models.AssetAttachment, error)
	GetByAssetID(ctx context.Context, assetID string) ([]models.AssetAttachment, error)
	Update(ctx context.Context, attachment *models.AssetAttachment) error
	Delete(ctx context.Context, id string) error

	// Bulk operations
	DeleteByAssetID(ctx context.Context, assetID string) error

	// Statistics
	CountByAssetID(ctx context.Context, assetID string) (int64, error)
	GetTotalSizeByAssetID(ctx context.Context, assetID string) (int64, error)
}

// assetAttachmentRepository implements AssetAttachmentRepository
type assetAttachmentRepository struct {
	db *gorm.DB
}

// NewAssetAttachmentRepository creates a new instance of AssetAttachmentRepository
func NewAssetAttachmentRepository(db *gorm.DB) AssetAttachmentRepository {
	return &assetAttachmentRepository{db: db}
}

// Create creates a new attachment record
func (r *assetAttachmentRepository) Create(ctx context.Context, attachment *models.AssetAttachment) error {
	if attachment.ID == uuid.Nil {
		attachment.ID = uuid.New()
	}
	attachment.CreatedAt = time.Now()
	attachment.UpdatedAt = time.Now()
	attachment.UploadedAt = time.Now()

	// Ensure the target asset exists and is within caller's scope
	var asset financeModels.Asset
	assetQ := security.ApplyScopeFilter(database.GetDB(ctx, r.db).Model(&financeModels.Asset{}), ctx, security.FinanceScopeQueryOptions())
	if err := assetQ.Where("id = ?", attachment.AssetID).First(&asset).Error; err != nil {
		return err
	}

	return database.GetDB(ctx, r.db).Create(attachment).Error
}

// GetByID retrieves an attachment by ID
func (r *assetAttachmentRepository) GetByID(ctx context.Context, id string) (*models.AssetAttachment, error) {
	var attachment models.AssetAttachment
	err := database.GetDB(ctx, r.db).
		Where("id = ?", id).
		First(&attachment).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	// verify asset is within caller's scope
	var asset financeModels.Asset
	assetQ := security.ApplyScopeFilter(database.GetDB(ctx, r.db).Model(&financeModels.Asset{}), ctx, security.FinanceScopeQueryOptions())
	if err := assetQ.Where("id = ?", attachment.AssetID).First(&asset).Error; err != nil {
		// treat as not found when asset is out of scope
		return nil, nil
	}

	return &attachment, nil
}

// GetByAssetID retrieves all attachments for an asset
func (r *assetAttachmentRepository) GetByAssetID(ctx context.Context, assetID string) ([]models.AssetAttachment, error) {
	var attachments []models.AssetAttachment
	// ensure asset is within caller's scope
	var asset financeModels.Asset
	assetQ := security.ApplyScopeFilter(database.GetDB(ctx, r.db).Model(&financeModels.Asset{}), ctx, security.FinanceScopeQueryOptions())
	if err := assetQ.Where("id = ?", assetID).First(&asset).Error; err != nil {
		// if asset not found or out of scope, return empty list
		return []models.AssetAttachment{}, nil
	}

	err := database.GetDB(ctx, r.db).
		Where("asset_id = ?", assetID).
		Order("created_at DESC").
		Find(&attachments).Error

	return attachments, err
}

// Update updates an attachment record
func (r *assetAttachmentRepository) Update(ctx context.Context, attachment *models.AssetAttachment) error {
	attachment.UpdatedAt = time.Now()
	// Ensure asset belongs to caller's scope before updating
	var asset financeModels.Asset
	assetQ := security.ApplyScopeFilter(database.GetDB(ctx, r.db).Model(&financeModels.Asset{}), ctx, security.FinanceScopeQueryOptions())
	if err := assetQ.Where("id = ?", attachment.AssetID).First(&asset).Error; err != nil {
		return err
	}
	return database.GetDB(ctx, r.db).Save(attachment).Error
}

// Delete soft deletes an attachment
func (r *assetAttachmentRepository) Delete(ctx context.Context, id string) error {
	// Retrieve attachment and ensure its asset is in scope
	var att models.AssetAttachment
	if err := database.GetDB(ctx, r.db).Where("id = ?", id).First(&att).Error; err != nil {
		return err
	}
	var asset financeModels.Asset
	assetQ := security.ApplyScopeFilter(database.GetDB(ctx, r.db).Model(&financeModels.Asset{}), ctx, security.FinanceScopeQueryOptions())
	if err := assetQ.Where("id = ?", att.AssetID).First(&asset).Error; err != nil {
		return err
	}
	return database.GetDB(ctx, r.db).
		Where("id = ?", id).
		Delete(&models.AssetAttachment{}).Error
}

// DeleteByAssetID deletes all attachments for an asset
func (r *assetAttachmentRepository) DeleteByAssetID(ctx context.Context, assetID string) error {
	// ensure asset is within caller's scope
	var asset financeModels.Asset
	assetQ := security.ApplyScopeFilter(database.GetDB(ctx, r.db).Model(&financeModels.Asset{}), ctx, security.FinanceScopeQueryOptions())
	if err := assetQ.Where("id = ?", assetID).First(&asset).Error; err != nil {
		return err
	}
	return database.GetDB(ctx, r.db).
		Where("asset_id = ?", assetID).
		Delete(&models.AssetAttachment{}).Error
}

// CountByAssetID returns the number of attachments for an asset
func (r *assetAttachmentRepository) CountByAssetID(ctx context.Context, assetID string) (int64, error) {
	var count int64
	// ensure asset is within caller's scope
	var asset financeModels.Asset
	assetQ := security.ApplyScopeFilter(database.GetDB(ctx, r.db).Model(&financeModels.Asset{}), ctx, security.FinanceScopeQueryOptions())
	if err := assetQ.Where("id = ?", assetID).First(&asset).Error; err != nil {
		return 0, err
	}
	err := database.GetDB(ctx, r.db).
		Model(&models.AssetAttachment{}).
		Where("asset_id = ?", assetID).
		Count(&count).Error

	return count, err
}

// GetTotalSizeByAssetID returns the total file size of attachments for an asset
func (r *assetAttachmentRepository) GetTotalSizeByAssetID(ctx context.Context, assetID string) (int64, error) {
	var totalSize int64
	// ensure asset is within caller's scope
	var asset financeModels.Asset
	assetQ := security.ApplyScopeFilter(database.GetDB(ctx, r.db).Model(&financeModels.Asset{}), ctx, security.FinanceScopeQueryOptions())
	if err := assetQ.Where("id = ?", assetID).First(&asset).Error; err != nil {
		return 0, err
	}
	err := database.GetDB(ctx, r.db).
		Model(&models.AssetAttachment{}).
		Where("asset_id = ?", assetID).
		Select("COALESCE(SUM(file_size), 0)").
		Scan(&totalSize).Error

	return totalSize, err
}
