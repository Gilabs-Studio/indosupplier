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

// BusinessTypeRepository defines the interface for business type data access
type BusinessTypeRepository interface {
	FindByID(ctx context.Context, id string) (*models.BusinessType, error)
	FindByName(ctx context.Context, name string) (*models.BusinessType, error)
	List(ctx context.Context, req *dto.ListBusinessTypesRequest) ([]models.BusinessType, int64, error)
	Create(ctx context.Context, b *models.BusinessType) error
	Update(ctx context.Context, b *models.BusinessType) error
	Delete(ctx context.Context, id string) error
}

type businessTypeRepository struct {
	db *gorm.DB
}

// NewBusinessTypeRepository creates a new BusinessTypeRepository
func NewBusinessTypeRepository(db *gorm.DB) BusinessTypeRepository {
	return &businessTypeRepository{db: db}
}

func (r *businessTypeRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *businessTypeRepository) FindByID(ctx context.Context, id string) (*models.BusinessType, error) {
	var businessType models.BusinessType
	err := r.getDB(ctx).Where("id = ?", id).First(&businessType).Error
	if err != nil {
		return nil, err
	}
	return &businessType, nil
}

func (r *businessTypeRepository) FindByName(ctx context.Context, name string) (*models.BusinessType, error) {
	var businessType models.BusinessType
	err := r.getDB(ctx).Where("name = ?", name).First(&businessType).Error
	if err != nil {
		return nil, err
	}
	return &businessType, nil
}

func (r *businessTypeRepository) List(ctx context.Context, req *dto.ListBusinessTypesRequest) ([]models.BusinessType, int64, error) {
	var businessTypes []models.BusinessType
	var total int64

	query := r.getDB(ctx).Model(&models.BusinessType{})

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
		Find(&businessTypes).Error
	if err != nil {
		return nil, 0, err
	}

	return businessTypes, total, nil
}

func (r *businessTypeRepository) Create(ctx context.Context, b *models.BusinessType) error {
	return r.getDB(ctx).Create(b).Error
}

func (r *businessTypeRepository) Update(ctx context.Context, b *models.BusinessType) error {
	return r.getDB(ctx).Save(b).Error
}

func (r *businessTypeRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Where("id = ?", id).Delete(&models.BusinessType{}).Error
}
