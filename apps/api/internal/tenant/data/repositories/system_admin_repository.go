package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/tenant/data/models"
	"gorm.io/gorm"
)

// SystemAdminRepository provides access to the system_admins table
type SystemAdminRepository interface {
	FindByID(ctx context.Context, id string) (*models.SystemAdmin, error)
	FindByEmail(ctx context.Context, email string) (*models.SystemAdmin, error)
	Create(ctx context.Context, admin *models.SystemAdmin) error
	Update(ctx context.Context, admin *models.SystemAdmin) error
}

type systemAdminRepository struct {
	db *gorm.DB
}

// NewSystemAdminRepository creates a new SystemAdminRepository
func NewSystemAdminRepository(db *gorm.DB) SystemAdminRepository {
	return &systemAdminRepository{db: db}
}

func (r *systemAdminRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *systemAdminRepository) FindByID(ctx context.Context, id string) (*models.SystemAdmin, error) {
	var admin models.SystemAdmin
	err := r.getDB(ctx).Where("id = ?", id).First(&admin).Error
	return &admin, err
}

func (r *systemAdminRepository) FindByEmail(ctx context.Context, email string) (*models.SystemAdmin, error) {
	var admin models.SystemAdmin
	err := r.getDB(ctx).Where("email = ?", email).First(&admin).Error
	return &admin, err
}

func (r *systemAdminRepository) Create(ctx context.Context, admin *models.SystemAdmin) error {
	return r.getDB(ctx).Create(admin).Error
}

func (r *systemAdminRepository) Update(ctx context.Context, admin *models.SystemAdmin) error {
	return r.getDB(ctx).Save(admin).Error
}
