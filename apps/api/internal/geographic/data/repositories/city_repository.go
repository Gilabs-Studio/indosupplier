package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/geographic/data/models"
	"github.com/gilabs/gims/api/internal/geographic/domain/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CityRepository defines the interface for city data access
type CityRepository interface {
	FindByID(ctx context.Context, id string) (*models.City, error)
	FindByCode(ctx context.Context, code string) (*models.City, error)
	List(ctx context.Context, req *dto.ListCitiesRequest) ([]models.City, int64, error)
	Create(ctx context.Context, c *models.City) error
	Update(ctx context.Context, c *models.City) error
	Delete(ctx context.Context, id string) error
	HasDistricts(ctx context.Context, cityID string) (bool, error)
}

type cityRepository struct {
	db *gorm.DB
}

// NewCityRepository creates a new CityRepository
func NewCityRepository(db *gorm.DB) CityRepository {
	return &cityRepository{db: db}
}

func (r *cityRepository) getDB(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *cityRepository) FindByID(ctx context.Context, id string) (*models.City, error) {
	var city models.City
	err := r.getDB(ctx).Select("id, province_id, name, code, type, is_active, created_at, updated_at, deleted_at").Preload("Province").Preload("Province.Country").Where("id = ?", id).First(&city).Error
	if err != nil {
		return nil, err
	}
	return &city, nil
}

func (r *cityRepository) FindByCode(ctx context.Context, code string) (*models.City, error) {
	var city models.City
	err := r.getDB(ctx).Select("id, province_id, name, code, type, is_active, created_at, updated_at, deleted_at").Preload("Province").Preload("Province.Country").Where("code = ?", code).First(&city).Error
	if err != nil {
		return nil, err
	}
	return &city, nil
}

func (r *cityRepository) List(ctx context.Context, req *dto.ListCitiesRequest) ([]models.City, int64, error) {
	var cities []models.City
	var total int64

	// Exclude geometry column from list queries for performance
	query := r.getDB(ctx).Model(&models.City{}).Select("id, province_id, name, code, type, is_active, created_at, updated_at, deleted_at").Preload("Province").Preload("Province.Country")

	// Apply search filter
	if req.Search != "" {
		search := "%" + req.Search + "%"
		query = query.Where("cities.name ILIKE ? OR cities.code ILIKE ?", search, search)
	}

	// Apply province filter
	if req.ProvinceID != "" {
		query = query.Where("cities.province_id = ?", req.ProvinceID)
	}

	// Apply type filter
	if req.Type != "" {
		query = query.Where("cities.type = ?", req.Type)
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
			Column: clause.Column{Table: "cities", Name: req.SortBy},
			Desc:   req.SortDir != "asc",
		})
	} else {
		query = query.Order("cities.updated_at DESC")
	}

	err := query.Offset(offset).Limit(perPage).Find(&cities).Error
	if err != nil {
		return nil, 0, err
	}

	return cities, total, nil
}

func (r *cityRepository) Create(ctx context.Context, c *models.City) error {
	return r.getDB(ctx).Create(c).Error
}

func (r *cityRepository) Update(ctx context.Context, c *models.City) error {
	return r.getDB(ctx).Save(c).Error
}

func (r *cityRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Where("id = ?", id).Delete(&models.City{}).Error
}

func (r *cityRepository) HasDistricts(ctx context.Context, cityID string) (bool, error) {
	var count int64
	err := r.getDB(ctx).Model(&models.District{}).Where("city_id = ?", cityID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
