package repositories

import (
	"context"

	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
	supplierModels "github.com/gilabs/indosupplier/api/internal/supplier/data/models"
	verificationModels "github.com/gilabs/indosupplier/api/internal/verification/data/models"
)

type VerificationRepository interface {
	GetProfileByUserID(ctx context.Context, userID string) (*supplierModels.SupplierProfile, error)
	UpdateProfile(ctx context.Context, profile *supplierModels.SupplierProfile) error
	GetLatestRequestBySupplierProfileID(ctx context.Context, supplierProfileID string) (*verificationModels.VerificationRequest, error)
	SaveVerificationRequest(ctx context.Context, request *verificationModels.VerificationRequest) error
}

type verificationRepository struct {
	db *gorm.DB
}

func NewVerificationRepository(db *gorm.DB) VerificationRepository {
	return &verificationRepository{db: db}
}

func (r *verificationRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *verificationRepository) GetProfileByUserID(ctx context.Context, userID string) (*supplierModels.SupplierProfile, error) {
	var profile supplierModels.SupplierProfile
	if err := r.getDB(ctx).Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *verificationRepository) UpdateProfile(ctx context.Context, profile *supplierModels.SupplierProfile) error {
	return r.getDB(ctx).Save(profile).Error
}

func (r *verificationRepository) GetLatestRequestBySupplierProfileID(ctx context.Context, supplierProfileID string) (*verificationModels.VerificationRequest, error) {
	var request verificationModels.VerificationRequest
	if err := r.getDB(ctx).
		Where("supplier_profile_id = ?", supplierProfileID).
		Order("updated_at DESC").
		First(&request).Error; err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *verificationRepository) SaveVerificationRequest(ctx context.Context, request *verificationModels.VerificationRequest) error {
	return r.getDB(ctx).Save(request).Error
}
