package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/geographic/data/models"
	"github.com/gilabs/gims/api/internal/geographic/domain/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CountryRepository defines the interface for country data access
type CountryRepository interface {
	FindByID(ctx context.Context, id string) (*models.Country, error)
	FindByCode(ctx context.Context, code string) (*models.Country, error)
	List(ctx context.Context, req *dto.ListCountriesRequest) ([]models.Country, int64, error)
	Create(ctx context.Context, c *models.Country) error
	Update(ctx context.Context, c *models.Country) error
	Delete(ctx context.Context, id string) error
	HasProvinces(ctx context.Context, countryID string) (bool, error)
}

type countryRepository struct {
	db *gorm.DB
}

// NewCountryRepository creates a new CountryRepository
func NewCountryRepository(db *gorm.DB) CountryRepository {
	return &countryRepository{db: db}
}

func (r *countryRepository) getDB(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *countryRepository) FindByID(ctx context.Context, id string) (*models.Country, error) {
	var country models.Country
	err := r.getDB(ctx).Where("id = ?", id).First(&country).Error
	if err != nil {
		return nil, err
	}
	return &country, nil
}

func (r *countryRepository) FindByCode(ctx context.Context, code string) (*models.Country, error) {
	var country models.Country
	err := r.getDB(ctx).Where("code = ?", code).First(&country).Error
	if err != nil {
		return nil, err
	}
	return &country, nil
}

func (r *countryRepository) List(ctx context.Context, req *dto.ListCountriesRequest) ([]models.Country, int64, error) {
	var countries []models.Country
	var total int64

	query := r.getDB(ctx).Model(&models.Country{})

	// Apply search filter
	if req.Search != "" {
		search := "%" + req.Search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ?", search, search)
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
	query = query.Order("is_active DESC")
	if req.SortBy != "" {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Name: req.SortBy},
			Desc:   req.SortDir != "asc",
		})
	} else {
		query = query.Order("updated_at DESC")
	}

	err := query.Offset(offset).Limit(perPage).Find(&countries).Error
	if err != nil {
		return nil, 0, err
	}

	return countries, total, nil
}

func (r *countryRepository) Create(ctx context.Context, c *models.Country) error {
	return r.getDB(ctx).Create(c).Error
}

func (r *countryRepository) Update(ctx context.Context, c *models.Country) error {
	return r.getDB(ctx).Save(c).Error
}

func (r *countryRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Where("id = ?", id).Delete(&models.Country{}).Error
}

func (r *countryRepository) HasProvinces(ctx context.Context, countryID string) (bool, error) {
	var count int64
	err := r.getDB(ctx).Model(&models.Province{}).Where("country_id = ?", countryID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
