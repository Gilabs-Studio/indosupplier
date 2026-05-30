package repositories

import (
	"context"

	"gorm.io/gorm"
	"github.com/gilabs/indosupplier/api/internal/sysadmin/data/models"
)

type SystemAdminRepository interface {
	Create(ctx context.Context, sa *models.SystemAdmin) error
	FindByEmail(ctx context.Context, email string) (*models.SystemAdmin, error)
	FindByID(ctx context.Context, id string) (*models.SystemAdmin, error)
}

type systemAdminRepository struct {
	db *gorm.DB
}

func NewSystemAdminRepository(db *gorm.DB) SystemAdminRepository {
	return &systemAdminRepository{db: db}
}

func (r *systemAdminRepository) Create(ctx context.Context, sa *models.SystemAdmin) error {
	return r.db.WithContext(ctx).Create(sa).Error
}

func (r *systemAdminRepository) FindByEmail(ctx context.Context, email string) (*models.SystemAdmin, error) {
	var sa models.SystemAdmin
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&sa).Error
	if err != nil {
		return nil, err
	}
	return &sa, nil
}

func (r *systemAdminRepository) FindByID(ctx context.Context, id string) (*models.SystemAdmin, error) {
	var sa models.SystemAdmin
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&sa).Error
	if err != nil {
		return nil, err
	}
	return &sa, nil
}
