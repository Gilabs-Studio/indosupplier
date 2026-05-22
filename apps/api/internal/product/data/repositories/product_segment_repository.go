package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"

	"github.com/gilabs/gims/api/internal/product/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProductSegmentRepository defines the interface for product segment data access
type ProductSegmentRepository interface {
	Create(ctx context.Context, segment *models.ProductSegment) error
	FindByID(ctx context.Context, id string) (*models.ProductSegment, error)
	List(ctx context.Context, params ListParams) ([]models.ProductSegment, int64, error)
	Update(ctx context.Context, segment *models.ProductSegment) error
	Delete(ctx context.Context, id string) error
}

type productSegmentRepository struct {
	db *gorm.DB
}

// NewProductSegmentRepository creates a new instance of ProductSegmentRepository
func NewProductSegmentRepository(db *gorm.DB) ProductSegmentRepository {
	return &productSegmentRepository{db: db}
}

func (r *productSegmentRepository) Create(ctx context.Context, segment *models.ProductSegment) error {
	return database.GetDB(ctx, r.db).Create(segment).Error
}

func (r *productSegmentRepository) FindByID(ctx context.Context, id string) (*models.ProductSegment, error) {
	var segment models.ProductSegment
	err := database.GetDB(ctx, r.db).First(&segment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &segment, nil
}

func (r *productSegmentRepository) List(ctx context.Context, params ListParams) ([]models.ProductSegment, int64, error) {
	var segments []models.ProductSegment
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.ProductSegment{})

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

	if err := query.Find(&segments).Error; err != nil {
		return nil, 0, err
	}

	return segments, total, nil
}

func (r *productSegmentRepository) Update(ctx context.Context, segment *models.ProductSegment) error {
	return database.GetDB(ctx, r.db).Save(segment).Error
}

func (r *productSegmentRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.ProductSegment{}, "id = ?", id).Error
}
