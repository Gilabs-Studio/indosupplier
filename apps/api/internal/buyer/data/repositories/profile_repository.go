package repositories

import (
	"context"

	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
	userModels "github.com/gilabs/indosupplier/api/internal/user/data/models"
)

type ProfileRepository interface {
	GetProfileByUserID(ctx context.Context, userID string) (*models.BuyerProfile, error)
	GetUserEmail(ctx context.Context, userID string) (string, error)
	UpdateProfile(ctx context.Context, profile *models.BuyerProfile) error
	ListDocuments(ctx context.Context, buyerProfileID string) ([]models.BuyerDocument, error)
	CreateDocument(ctx context.Context, document *models.BuyerDocument) error
}

type profileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) ProfileRepository {
	return &profileRepository{db: db}
}

func (r *profileRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *profileRepository) GetProfileByUserID(ctx context.Context, userID string) (*models.BuyerProfile, error) {
	var profile models.BuyerProfile
	if err := r.getDB(ctx).Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *profileRepository) GetUserEmail(ctx context.Context, userID string) (string, error) {
	var user userModels.User
	if err := r.getDB(ctx).Select("email").Where("id = ?", userID).First(&user).Error; err != nil {
		return "", err
	}
	return user.Email, nil
}

func (r *profileRepository) UpdateProfile(ctx context.Context, profile *models.BuyerProfile) error {
	return r.getDB(ctx).Save(profile).Error
}

func (r *profileRepository) ListDocuments(ctx context.Context, buyerProfileID string) ([]models.BuyerDocument, error) {
	var documents []models.BuyerDocument
	err := r.getDB(ctx).
		Where("buyer_profile_id = ?", buyerProfileID).
		Order("created_at DESC").
		Find(&documents).Error
	return documents, err
}

func (r *profileRepository) CreateDocument(ctx context.Context, document *models.BuyerDocument) error {
	return r.getDB(ctx).Create(document).Error
}
