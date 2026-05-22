package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"

	"github.com/gilabs/gims/api/internal/product/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProductBrandRepository defines the interface for product brand data access
type ProductBrandRepository interface {
	Create(ctx context.Context, brand *models.ProductBrand) error
	FindByID(ctx context.Context, id string) (*models.ProductBrand, error)
	List(ctx context.Context, params ListParams) ([]models.ProductBrand, int64, error)
	Update(ctx context.Context, brand *models.ProductBrand) error
	Delete(ctx context.Context, id string) error
}

type productBrandRepository struct {
	db *gorm.DB
}

// NewProductBrandRepository creates a new instance of ProductBrandRepository
func NewProductBrandRepository(db *gorm.DB) ProductBrandRepository {
	return &productBrandRepository{db: db}
}

func (r *productBrandRepository) Create(ctx context.Context, brand *models.ProductBrand) error {
	return database.GetDB(ctx, r.db).Create(brand).Error
}

func (r *productBrandRepository) FindByID(ctx context.Context, id string) (*models.ProductBrand, error) {
	var brand models.ProductBrand
	err := database.GetDB(ctx, r.db).First(&brand, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &brand, nil
}

func (r *productBrandRepository) List(ctx context.Context, params ListParams) ([]models.ProductBrand, int64, error) {
	var brands []models.ProductBrand
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.ProductBrand{})

	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", search, search)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = query.Order("is_active DESC")
	if params.SortBy != "" {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Name: params.SortBy},
			Desc:   params.SortDir == "desc",
		})
	} else {
		query = query.Order("name ASC")
	}

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	if err := query.Find(&brands).Error; err != nil {
		return nil, 0, err
	}

	return brands, total, nil
}

func (r *productBrandRepository) Update(ctx context.Context, brand *models.ProductBrand) error {
	return database.GetDB(ctx, r.db).Save(brand).Error
}

func (r *productBrandRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.ProductBrand{}, "id = ?", id).Error
}
