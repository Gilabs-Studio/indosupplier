package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"

	"github.com/gilabs/gims/api/internal/product/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UnitOfMeasureRepository defines the interface for unit of measure data access
type UnitOfMeasureRepository interface {
	Create(ctx context.Context, uom *models.UnitOfMeasure) error
	FindByID(ctx context.Context, id string) (*models.UnitOfMeasure, error)
	List(ctx context.Context, params ListParams) ([]models.UnitOfMeasure, int64, error)
	Update(ctx context.Context, uom *models.UnitOfMeasure) error
	Delete(ctx context.Context, id string) error
}

type unitOfMeasureRepository struct {
	db *gorm.DB
}

// NewUnitOfMeasureRepository creates a new instance of UnitOfMeasureRepository
func NewUnitOfMeasureRepository(db *gorm.DB) UnitOfMeasureRepository {
	return &unitOfMeasureRepository{db: db}
}

func (r *unitOfMeasureRepository) Create(ctx context.Context, uom *models.UnitOfMeasure) error {
	return database.GetDB(ctx, r.db).Create(uom).Error
}

func (r *unitOfMeasureRepository) FindByID(ctx context.Context, id string) (*models.UnitOfMeasure, error) {
	var uom models.UnitOfMeasure
	err := database.GetDB(ctx, r.db).First(&uom, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &uom, nil
}

func (r *unitOfMeasureRepository) List(ctx context.Context, params ListParams) ([]models.UnitOfMeasure, int64, error) {
	var uoms []models.UnitOfMeasure
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.UnitOfMeasure{})

	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("name ILIKE ? OR symbol ILIKE ? OR description ILIKE ?", search, search, search)
	}

	if params.ActiveOnly {
		query = query.Where("is_active = ?", true)
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

	if err := query.Find(&uoms).Error; err != nil {
		return nil, 0, err
	}

	return uoms, total, nil
}

func (r *unitOfMeasureRepository) Update(ctx context.Context, uom *models.UnitOfMeasure) error {
	return database.GetDB(ctx, r.db).Save(uom).Error
}

func (r *unitOfMeasureRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.UnitOfMeasure{}, "id = ?", id).Error
}
