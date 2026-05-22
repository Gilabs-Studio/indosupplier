package repositories

import (
	"context"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"

	"github.com/gilabs/indosupplier/api/internal/core/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CourierAgencyRepository defines the interface for courier agency data access
type CourierAgencyRepository interface {
	Create(ctx context.Context, courierAgency *models.CourierAgency) error
	FindByID(ctx context.Context, id string) (*models.CourierAgency, error)
	List(ctx context.Context, params ListParams) ([]models.CourierAgency, int64, error)
	Update(ctx context.Context, courierAgency *models.CourierAgency) error
	Delete(ctx context.Context, id string) error
}

type courierAgencyRepository struct {
	db *gorm.DB
}

// NewCourierAgencyRepository creates a new instance of CourierAgencyRepository
func NewCourierAgencyRepository(db *gorm.DB) CourierAgencyRepository {
	return &courierAgencyRepository{db: db}
}

func (r *courierAgencyRepository) Create(ctx context.Context, courierAgency *models.CourierAgency) error {
	return database.GetDB(ctx, r.db).Create(courierAgency).Error
}

func (r *courierAgencyRepository) FindByID(ctx context.Context, id string) (*models.CourierAgency, error) {
	var courierAgency models.CourierAgency
	err := database.GetDB(ctx, r.db).First(&courierAgency, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &courierAgency, nil
}

func (r *courierAgencyRepository) List(ctx context.Context, params ListParams) ([]models.CourierAgency, int64, error) {
	var courierAgencies []models.CourierAgency
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.CourierAgency{})

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

	if err := query.Find(&courierAgencies).Error; err != nil {
		return nil, 0, err
	}

	return courierAgencies, total, nil
}

func (r *courierAgencyRepository) Update(ctx context.Context, courierAgency *models.CourierAgency) error {
	return database.GetDB(ctx, r.db).Save(courierAgency).Error
}

func (r *courierAgencyRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.CourierAgency{}, "id = ?", id).Error
}
