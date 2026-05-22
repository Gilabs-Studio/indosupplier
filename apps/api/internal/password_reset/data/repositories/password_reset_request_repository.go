package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/password_reset/data/models"
	"gorm.io/gorm"
)

// PasswordResetRequestRepository defines data access for password reset requests.
type PasswordResetRequestRepository interface {
	Create(ctx context.Context, req *models.PasswordResetRequest) error
	FindByToken(ctx context.Context, token string) (*models.PasswordResetRequest, error)
	FindByUserID(ctx context.Context, userID string) (*models.PasswordResetRequest, error)
	FindPendingByUserID(ctx context.Context, userID string) (*models.PasswordResetRequest, error)
	Update(ctx context.Context, req *models.PasswordResetRequest) error
	DeleteByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
}

type passwordResetRequestRepository struct {
	db *gorm.DB
}

func NewPasswordResetRequestRepository(db *gorm.DB) PasswordResetRequestRepository {
	return &passwordResetRequestRepository{db: db}
}

func (r *passwordResetRequestRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *passwordResetRequestRepository) Create(ctx context.Context, req *models.PasswordResetRequest) error {
	return r.getDB(ctx).Create(req).Error
}

func (r *passwordResetRequestRepository) FindByToken(ctx context.Context, token string) (*models.PasswordResetRequest, error) {
	var req models.PasswordResetRequest
	err := r.getDB(ctx).Where("token = ?", token).First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *passwordResetRequestRepository) FindByUserID(ctx context.Context, userID string) (*models.PasswordResetRequest, error) {
	var req models.PasswordResetRequest
	err := r.getDB(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *passwordResetRequestRepository) FindPendingByUserID(ctx context.Context, userID string) (*models.PasswordResetRequest, error) {
	var req models.PasswordResetRequest
	err := r.getDB(ctx).
		Where("user_id = ? AND status = ?", userID, models.PasswordResetStatusPending).
		Order("created_at DESC").
		First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *passwordResetRequestRepository) Update(ctx context.Context, req *models.PasswordResetRequest) error {
	return r.getDB(ctx).Save(req).Error
}

func (r *passwordResetRequestRepository) DeleteByUserID(ctx context.Context, userID string) error {
	return r.getDB(ctx).Where("user_id = ?", userID).Delete(&models.PasswordResetRequest{}).Error
}

func (r *passwordResetRequestRepository) DeleteExpired(ctx context.Context) error {
	return r.getDB(ctx).
		Where("expires_at < NOW() OR status = ?", models.PasswordResetStatusUsed).
		Delete(&models.PasswordResetRequest{}).Error
}
