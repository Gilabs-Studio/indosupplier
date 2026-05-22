package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"

	"github.com/gilabs/gims/api/internal/crm/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ContactRoleRepository defines the interface for contact role data access
type ContactRoleRepository interface {
	Create(ctx context.Context, role *models.ContactRole) error
	FindByID(ctx context.Context, id string) (*models.ContactRole, error)
	ExistsByName(ctx context.Context, name string, excludeID string) (bool, error)
	List(ctx context.Context, params ListParams) ([]models.ContactRole, int64, error)
	Update(ctx context.Context, role *models.ContactRole) error
	Delete(ctx context.Context, id string) error
}

type contactRoleRepository struct {
	db *gorm.DB
}

// NewContactRoleRepository creates a new contact role repository
func NewContactRoleRepository(db *gorm.DB) ContactRoleRepository {
	return &contactRoleRepository{db: db}
}

func (r *contactRoleRepository) Create(ctx context.Context, role *models.ContactRole) error {
	return database.GetDB(ctx, r.db).Create(role).Error
}

func (r *contactRoleRepository) FindByID(ctx context.Context, id string) (*models.ContactRole, error) {
	var role models.ContactRole
	err := database.GetDB(ctx, r.db).First(&role, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *contactRoleRepository) ExistsByName(ctx context.Context, name string, excludeID string) (bool, error) {
	var count int64
	q := database.GetDB(ctx, r.db).Model(&models.ContactRole{}).Where("LOWER(name) = LOWER(?)", name)
	if excludeID != "" {
		q = q.Where("id <> ?", excludeID)
	}

	if err := q.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *contactRoleRepository) List(ctx context.Context, params ListParams) ([]models.ContactRole, int64, error) {
	var roles []models.ContactRole
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.ContactRole{})

	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search, search, search)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if params.SortBy != "" {
		query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: params.SortBy}, Desc: params.SortDir == "desc"})
	} else {
		query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: "name"}, Desc: false})
	}

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	if err := query.Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

func (r *contactRoleRepository) Update(ctx context.Context, role *models.ContactRole) error {
	return database.GetDB(ctx, r.db).Save(role).Error
}

func (r *contactRoleRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.ContactRole{}, "id = ?", id).Error
}
