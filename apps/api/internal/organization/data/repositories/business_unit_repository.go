package repositories

import (
	"context"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BusinessUnitRepository defines the interface for business unit data access
type BusinessUnitRepository interface {
	FindByID(ctx context.Context, id string) (*models.BusinessUnit, error)
	FindByName(ctx context.Context, name string) (*models.BusinessUnit, error)
	List(ctx context.Context, req *dto.ListBusinessUnitsRequest) ([]models.BusinessUnit, int64, error)
	Create(ctx context.Context, b *models.BusinessUnit) error
	Update(ctx context.Context, b *models.BusinessUnit) error
	Delete(ctx context.Context, id string) error
}

type businessUnitRepository struct {
	db *gorm.DB
}

// NewBusinessUnitRepository creates a new BusinessUnitRepository
func NewBusinessUnitRepository(db *gorm.DB) BusinessUnitRepository {
	return &businessUnitRepository{db: db}
}

func (r *businessUnitRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *businessUnitRepository) FindByID(ctx context.Context, id string) (*models.BusinessUnit, error) {
	var businessUnit models.BusinessUnit
	err := r.getDB(ctx).Where("id = ?", id).First(&businessUnit).Error
	if err != nil {
		return nil, err
	}
	return &businessUnit, nil
}

func (r *businessUnitRepository) FindByName(ctx context.Context, name string) (*models.BusinessUnit, error) {
	var businessUnit models.BusinessUnit
	err := r.getDB(ctx).Where("name = ?", name).First(&businessUnit).Error
	if err != nil {
		return nil, err
	}
	return &businessUnit, nil
}

func (r *businessUnitRepository) List(ctx context.Context, req *dto.ListBusinessUnitsRequest) ([]models.BusinessUnit, int64, error) {
	var businessUnits []models.BusinessUnit
	var total int64

	query := r.getDB(ctx).Model(&models.BusinessUnit{})

	// Apply search filter
	if searchTerm := strings.TrimSpace(req.Search); searchTerm != "" {
		search := "%" + searchTerm + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", search, search)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	perPage := req.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage

	// Apply sorting
	sortBy := "updated_at"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}

	err := query.Order("is_active DESC").
		Order(clause.OrderByColumn{
			Column: clause.Column{Name: sortBy},
			Desc:   req.SortDir != "asc",
		}).
		Offset(offset).
		Limit(perPage).
		Find(&businessUnits).Error
	if err != nil {
		return nil, 0, err
	}

	return businessUnits, total, nil
}

func (r *businessUnitRepository) Create(ctx context.Context, b *models.BusinessUnit) error {
	return r.getDB(ctx).Create(b).Error
}

func (r *businessUnitRepository) Update(ctx context.Context, b *models.BusinessUnit) error {
	return r.getDB(ctx).Save(b).Error
}

func (r *businessUnitRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Where("id = ?", id).Delete(&models.BusinessUnit{}).Error
}
