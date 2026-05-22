package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/geographic/data/models"
	"github.com/gilabs/gims/api/internal/geographic/domain/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DistrictRepository defines the interface for district data access
type DistrictRepository interface {
	FindByID(ctx context.Context, id string) (*models.District, error)
	FindByCode(ctx context.Context, code string) (*models.District, error)
	List(ctx context.Context, req *dto.ListDistrictsRequest) ([]models.District, int64, error)
	Create(ctx context.Context, d *models.District) error
	Update(ctx context.Context, d *models.District) error
	Delete(ctx context.Context, id string) error
	HasVillages(ctx context.Context, districtID string) (bool, error)
}

type districtRepository struct {
	db *gorm.DB
}

// NewDistrictRepository creates a new DistrictRepository
func NewDistrictRepository(db *gorm.DB) DistrictRepository {
	return &districtRepository{db: db}
}

func (r *districtRepository) getDB(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *districtRepository) FindByID(ctx context.Context, id string) (*models.District, error) {
	var district models.District
	err := r.getDB(ctx).Select("id, city_id, name, code, is_active, created_at, updated_at, deleted_at").Preload("City").Preload("City.Province").Preload("City.Province.Country").Where("id = ?", id).First(&district).Error
	if err != nil {
		return nil, err
	}
	return &district, nil
}

func (r *districtRepository) FindByCode(ctx context.Context, code string) (*models.District, error) {
	var district models.District
	err := r.getDB(ctx).Select("id, city_id, name, code, is_active, created_at, updated_at, deleted_at").Preload("City").Preload("City.Province").Preload("City.Province.Country").Where("code = ?", code).First(&district).Error
	if err != nil {
		return nil, err
	}
	return &district, nil
}

func (r *districtRepository) List(ctx context.Context, req *dto.ListDistrictsRequest) ([]models.District, int64, error) {
	var districts []models.District
	var total int64

	// Exclude geometry column from list queries for performance
	query := r.getDB(ctx).Model(&models.District{}).Select("id, city_id, name, code, is_active, created_at, updated_at, deleted_at").Preload("City").Preload("City.Province")

	// Apply search filter
	if req.Search != "" {
		search := "%" + req.Search + "%"
		query = query.Where("districts.name ILIKE ? OR districts.code ILIKE ?", search, search)
	}

	// Apply city filter
	if req.CityID != "" {
		query = query.Where("districts.city_id = ?", req.CityID)
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
			Column: clause.Column{Table: "districts", Name: req.SortBy},
			Desc:   req.SortDir != "asc",
		})
	} else {
		query = query.Order("districts.updated_at DESC")
	}

	err := query.Offset(offset).Limit(perPage).Find(&districts).Error
	if err != nil {
		return nil, 0, err
	}

	return districts, total, nil
}

func (r *districtRepository) Create(ctx context.Context, d *models.District) error {
	return r.getDB(ctx).Create(d).Error
}

func (r *districtRepository) Update(ctx context.Context, d *models.District) error {
	return r.getDB(ctx).Save(d).Error
}

func (r *districtRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Where("id = ?", id).Delete(&models.District{}).Error
}

func (r *districtRepository) HasVillages(ctx context.Context, districtID string) (bool, error) {
	var count int64
	err := r.getDB(ctx).Model(&models.Village{}).Where("district_id = ?", districtID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
