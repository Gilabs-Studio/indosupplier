package repositories

import (
	"context"

	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
)

type BookmarkRepository interface {
	Create(ctx context.Context, bookmark *models.Bookmark) error
	Delete(ctx context.Context, id string, buyerProfileID string) error
	FindByIDAndBuyer(ctx context.Context, id string, buyerProfileID string) (*models.Bookmark, error)
	FindByBuyerAndSupplier(ctx context.Context, buyerProfileID string, supplierProfileID string) (*models.Bookmark, error)
	List(ctx context.Context, buyerProfileID string) ([]models.Bookmark, error)
}

type bookmarkRepository struct {
	db *gorm.DB
}

func NewBookmarkRepository(db *gorm.DB) BookmarkRepository {
	return &bookmarkRepository{db: db}
}

func (r *bookmarkRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *bookmarkRepository) Create(ctx context.Context, bookmark *models.Bookmark) error {
	return r.getDB(ctx).Create(bookmark).Error
}

func (r *bookmarkRepository) Delete(ctx context.Context, id string, buyerProfileID string) error {
	return r.getDB(ctx).Where("id = ? AND buyer_profile_id = ?", id, buyerProfileID).Delete(&models.Bookmark{}).Error
}

func (r *bookmarkRepository) FindByIDAndBuyer(ctx context.Context, id string, buyerProfileID string) (*models.Bookmark, error) {
	var b models.Bookmark
	if err := r.getDB(ctx).Where("id = ? AND buyer_profile_id = ?", id, buyerProfileID).First(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *bookmarkRepository) FindByBuyerAndSupplier(ctx context.Context, buyerProfileID string, supplierProfileID string) (*models.Bookmark, error) {
	var b models.Bookmark
	if err := r.getDB(ctx).Where("buyer_profile_id = ? AND supplier_profile_id = ?", buyerProfileID, supplierProfileID).First(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *bookmarkRepository) List(ctx context.Context, buyerProfileID string) ([]models.Bookmark, error) {
	var bookmarks []models.Bookmark
	err := r.getDB(ctx).
		Where("buyer_profile_id = ?", buyerProfileID).
		Order("created_at DESC").
		Find(&bookmarks).Error
	return bookmarks, err
}
