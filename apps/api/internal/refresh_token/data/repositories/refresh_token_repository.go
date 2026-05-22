package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/refresh_token/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, rt *models.RefreshToken) error
	FindByTokenID(ctx context.Context, tokenID string) (*models.RefreshToken, error)
	FindByTokenIDForUpdate(ctx context.Context, tokenID string) (*models.RefreshToken, error)
	Revoke(ctx context.Context, tokenID string) error
	DeleteExpired(ctx context.Context) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *refreshTokenRepository) Create(ctx context.Context, rt *models.RefreshToken) error {
	return r.getDB(ctx).Create(rt).Error
}

func (r *refreshTokenRepository) FindByTokenID(ctx context.Context, tokenID string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	err := r.getDB(ctx).Where("token_id = ?", tokenID).First(&rt).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *refreshTokenRepository) FindByTokenIDForUpdate(ctx context.Context, tokenID string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	// Use FOR UPDATE to lock the row
	err := r.getDB(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("token_id = ?", tokenID).First(&rt).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, tokenID string) error {
	now := apptime.Now()
	// Using locking for revoke if in transaction
	return r.getDB(ctx).Model(&models.RefreshToken{}).
		Where("token_id = ?", tokenID).
		Updates(map[string]interface{}{
			"revoked":    true,
			"revoked_at": now,
		}).Error
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.getDB(ctx).
		Where("expires_at < ?", apptime.Now()).
		Delete(&models.RefreshToken{}).Error
}
