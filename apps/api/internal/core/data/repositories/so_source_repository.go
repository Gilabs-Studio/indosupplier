package repositories

import (
	"context"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"

	"github.com/gilabs/indosupplier/api/internal/core/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SOSourceRepository defines the interface for SO source data access
type SOSourceRepository interface {
	Create(ctx context.Context, soSource *models.SOSource) error
	FindByID(ctx context.Context, id string) (*models.SOSource, error)
	List(ctx context.Context, params ListParams) ([]models.SOSource, int64, error)
	Update(ctx context.Context, soSource *models.SOSource) error
	Delete(ctx context.Context, id string) error
}

type soSourceRepository struct {
	db *gorm.DB
}

// NewSOSourceRepository creates a new instance of SOSourceRepository
func NewSOSourceRepository(db *gorm.DB) SOSourceRepository {
	return &soSourceRepository{db: db}
}

func (r *soSourceRepository) Create(ctx context.Context, soSource *models.SOSource) error {
	return database.GetDB(ctx, r.db).Create(soSource).Error
}

func (r *soSourceRepository) FindByID(ctx context.Context, id string) (*models.SOSource, error) {
	var soSource models.SOSource
	err := database.GetDB(ctx, r.db).First(&soSource, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &soSource, nil
}

func (r *soSourceRepository) List(ctx context.Context, params ListParams) ([]models.SOSource, int64, error) {
	var soSources []models.SOSource
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.SOSource{})

	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search, search, search)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

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

	if err := query.Find(&soSources).Error; err != nil {
		return nil, 0, err
	}

	return soSources, total, nil
}

func (r *soSourceRepository) Update(ctx context.Context, soSource *models.SOSource) error {
	return database.GetDB(ctx, r.db).Save(soSource).Error
}

func (r *soSourceRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.SOSource{}, "id = ?", id).Error
}
