package repositories

import (
	"context"

	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
)

type FollowingRepository interface {
	Create(ctx context.Context, following *models.SupplierFollowing) error
	DeleteByBuyerAndSupplier(ctx context.Context, buyerProfileID string, supplierProfileID string) error
	FindByBuyerAndSupplier(ctx context.Context, buyerProfileID string, supplierProfileID string) (*models.SupplierFollowing, error)
	ListByBuyer(ctx context.Context, buyerProfileID string) ([]models.SupplierFollowing, error)
}

type followingRepository struct {
	db *gorm.DB
}

func NewFollowingRepository(db *gorm.DB) FollowingRepository {
	return &followingRepository{db: db}
}

func (r *followingRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *followingRepository) Create(ctx context.Context, following *models.SupplierFollowing) error {
	return r.getDB(ctx).Create(following).Error
}

func (r *followingRepository) DeleteByBuyerAndSupplier(ctx context.Context, buyerProfileID string, supplierProfileID string) error {
	return r.getDB(ctx).
		Where("buyer_profile_id = ? AND supplier_profile_id = ?", buyerProfileID, supplierProfileID).
		Delete(&models.SupplierFollowing{}).Error
}

func (r *followingRepository) FindByBuyerAndSupplier(ctx context.Context, buyerProfileID string, supplierProfileID string) (*models.SupplierFollowing, error) {
	var following models.SupplierFollowing
	if err := r.getDB(ctx).
		Where("buyer_profile_id = ? AND supplier_profile_id = ?", buyerProfileID, supplierProfileID).
		First(&following).Error; err != nil {
		return nil, err
	}
	return &following, nil
}

func (r *followingRepository) ListByBuyer(ctx context.Context, buyerProfileID string) ([]models.SupplierFollowing, error) {
	var followings []models.SupplierFollowing
	err := r.getDB(ctx).
		Where("buyer_profile_id = ?", buyerProfileID).
		Order("created_at DESC").
		Find(&followings).Error
	return followings, err
}
