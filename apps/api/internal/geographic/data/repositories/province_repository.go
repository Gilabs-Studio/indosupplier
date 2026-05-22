package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/geographic/data/models"
	"github.com/gilabs/gims/api/internal/geographic/domain/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProvinceRepository defines the interface for province data access
type ProvinceRepository interface {
	FindByID(ctx context.Context, id string) (*models.Province, error)
	FindByCode(ctx context.Context, code string) (*models.Province, error)
	List(ctx context.Context, req *dto.ListProvincesRequest) ([]models.Province, int64, error)
	Create(ctx context.Context, p *models.Province) error
	Update(ctx context.Context, p *models.Province) error
	Delete(ctx context.Context, id string) error
	HasCities(ctx context.Context, provinceID string) (bool, error)
}

type provinceRepository struct {
	db *gorm.DB
}

// NewProvinceRepository creates a new ProvinceRepository
func NewProvinceRepository(db *gorm.DB) ProvinceRepository {
	return &provinceRepository{db: db}
}

func (r *provinceRepository) getDB(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *provinceRepository) FindByID(ctx context.Context, id string) (*models.Province, error) {
	var province models.Province
	err := r.getDB(ctx).Select("id, country_id, name, code, is_active, created_at, updated_at, deleted_at").Preload("Country").Where("id = ?", id).First(&province).Error
	if err != nil {
		return nil, err
	}
	return &province, nil
}

func (r *provinceRepository) FindByCode(ctx context.Context, code string) (*models.Province, error) {
	var province models.Province
	err := r.getDB(ctx).Select("id, country_id, name, code, is_active, created_at, updated_at, deleted_at").Preload("Country").Where("code = ?", code).First(&province).Error
	if err != nil {
		return nil, err
	}
	return &province, nil
}

func (r *provinceRepository) List(ctx context.Context, req *dto.ListProvincesRequest) ([]models.Province, int64, error) {
	var provinces []models.Province
	var total int64

	// Exclude geometry column from list queries for performance
	query := r.getDB(ctx).Model(&models.Province{}).Select("id, country_id, name, code, is_active, created_at, updated_at, deleted_at").Preload("Country")

	// Apply search filter
	if req.Search != "" {
		search := "%" + req.Search + "%"
		query = query.Where("provinces.name ILIKE ? OR provinces.code ILIKE ?", search, search)
	}

	// Apply country filter
	if req.CountryID != "" {
		query = query.Where("provinces.country_id = ?", req.CountryID)
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
			Column: clause.Column{Table: "provinces", Name: req.SortBy},
			Desc:   req.SortDir != "asc",
		})
	} else {
		query = query.Order("provinces.updated_at DESC")
	}

	err := query.Offset(offset).Limit(perPage).Find(&provinces).Error
	if err != nil {
		return nil, 0, err
	}

	return provinces, total, nil
}

func (r *provinceRepository) Create(ctx context.Context, p *models.Province) error {
	return r.getDB(ctx).Create(p).Error
}

func (r *provinceRepository) Update(ctx context.Context, p *models.Province) error {
	return r.getDB(ctx).Save(p).Error
}

func (r *provinceRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Where("id = ?", id).Delete(&models.Province{}).Error
}

func (r *provinceRepository) HasCities(ctx context.Context, provinceID string) (bool, error) {
	var count int64
	err := r.getDB(ctx).Model(&models.City{}).Where("province_id = ?", provinceID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
