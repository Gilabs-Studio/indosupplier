package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/geographic/data/models"
	"gorm.io/gorm"
)

// MapDataRepository defines the interface for geographic map data access
type MapDataRepository interface {
	FindProvincesWithGeometry(ctx context.Context) ([]models.Province, error)
	FindCitiesWithGeometryByProvince(ctx context.Context, provinceID string) ([]models.City, error)
	FindDistrictsWithGeometryByCity(ctx context.Context, cityID string) ([]models.District, error)
	FindDistrictsWithGeometryByProvince(ctx context.Context, provinceID string) ([]models.District, error)
	FindAllDistrictsWithGeometry(ctx context.Context) ([]models.District, error)
}

type mapDataRepository struct {
	db *gorm.DB
}

// NewMapDataRepository creates a new MapDataRepository
func NewMapDataRepository(db *gorm.DB) MapDataRepository {
	return &mapDataRepository{db: db}
}

func (r *mapDataRepository) getDB(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

// FindProvincesWithGeometry returns all active provinces with their geometry data
func (r *mapDataRepository) FindProvincesWithGeometry(ctx context.Context) ([]models.Province, error) {
	var provinces []models.Province
	err := r.getDB(ctx).
		Where("is_active = ? AND geometry IS NOT NULL", true).
		Order("name ASC").
		Find(&provinces).Error
	if err != nil {
		return nil, err
	}
	return provinces, nil
}

// FindCitiesWithGeometryByProvince returns cities in a province with their geometry data
func (r *mapDataRepository) FindCitiesWithGeometryByProvince(ctx context.Context, provinceID string) ([]models.City, error) {
	var cities []models.City
	err := r.getDB(ctx).
		Where("province_id = ? AND is_active = ? AND geometry IS NOT NULL", provinceID, true).
		Order("name ASC").
		Find(&cities).Error
	if err != nil {
		return nil, err
	}
	return cities, nil
}

// FindDistrictsWithGeometryByCity returns districts in a city with their geometry data
func (r *mapDataRepository) FindDistrictsWithGeometryByCity(ctx context.Context, cityID string) ([]models.District, error) {
	var districts []models.District
	err := r.getDB(ctx).
		Where("city_id = ? AND is_active = ? AND geometry IS NOT NULL", cityID, true).
		Order("name ASC").
		Find(&districts).Error
	if err != nil {
		return nil, err
	}
	return districts, nil
}

// FindDistrictsWithGeometryByProvince returns all districts in a province with City preloaded
func (r *mapDataRepository) FindDistrictsWithGeometryByProvince(ctx context.Context, provinceID string) ([]models.District, error) {
	var districts []models.District
	err := r.getDB(ctx).
		Joins("JOIN cities ON cities.id = districts.city_id").
		Where("cities.province_id = ? AND districts.is_active = ? AND districts.geometry IS NOT NULL", provinceID, true).
		Preload("City").
		Find(&districts).Error
	if err != nil {
		return nil, err
	}
	return districts, nil
}

// FindAllDistrictsWithGeometry returns all active districts with geometry, with City and Province preloaded
func (r *mapDataRepository) FindAllDistrictsWithGeometry(ctx context.Context) ([]models.District, error) {
	var districts []models.District
	err := r.getDB(ctx).
		Where("is_active = ? AND geometry IS NOT NULL", true).
		Preload("City.Province").
		Find(&districts).Error
	if err != nil {
		return nil, err
	}
	return districts, nil
}
