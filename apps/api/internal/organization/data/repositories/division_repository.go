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

// DivisionRepository defines the interface for division data access
type DivisionRepository interface {
	FindByID(ctx context.Context, id string) (*models.Division, error)
	FindByName(ctx context.Context, name string) (*models.Division, error)
	List(ctx context.Context, req *dto.ListDivisionsRequest) ([]models.Division, int64, error)
	Create(ctx context.Context, d *models.Division) error
	Update(ctx context.Context, d *models.Division) error
	Delete(ctx context.Context, id string) error
}

type divisionRepository struct {
	db *gorm.DB
}

// NewDivisionRepository creates a new DivisionRepository
func NewDivisionRepository(db *gorm.DB) DivisionRepository {
	return &divisionRepository{db: db}
}

func (r *divisionRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *divisionRepository) FindByID(ctx context.Context, id string) (*models.Division, error) {
	var division models.Division
	err := r.getDB(ctx).Where("id = ?", id).First(&division).Error
	if err != nil {
		return nil, err
	}
	return &division, nil
}

func (r *divisionRepository) FindByName(ctx context.Context, name string) (*models.Division, error) {
	var division models.Division
	err := r.getDB(ctx).Where("name = ?", name).First(&division).Error
	if err != nil {
		return nil, err
	}
	return &division, nil
}

func (r *divisionRepository) List(ctx context.Context, req *dto.ListDivisionsRequest) ([]models.Division, int64, error) {
	var divisions []models.Division
	var total int64

	query := r.getDB(ctx).Model(&models.Division{})

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
		Find(&divisions).Error
	if err != nil {
		return nil, 0, err
	}

	return divisions, total, nil
}

func (r *divisionRepository) Create(ctx context.Context, d *models.Division) error {
	return r.getDB(ctx).Create(d).Error
}

func (r *divisionRepository) Update(ctx context.Context, d *models.Division) error {
	return r.getDB(ctx).Save(d).Error
}

func (r *divisionRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Where("id = ?", id).Delete(&models.Division{}).Error
}
