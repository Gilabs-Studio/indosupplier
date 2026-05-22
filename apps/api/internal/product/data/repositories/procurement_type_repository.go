package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"

	"github.com/gilabs/gims/api/internal/product/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProcurementTypeRepository defines the interface for procurement type data access
type ProcurementTypeRepository interface {
	Create(ctx context.Context, procurementType *models.ProcurementType) error
	FindByID(ctx context.Context, id string) (*models.ProcurementType, error)
	List(ctx context.Context, params ListParams) ([]models.ProcurementType, int64, error)
	Update(ctx context.Context, procurementType *models.ProcurementType) error
	Delete(ctx context.Context, id string) error
}

type procurementTypeRepository struct {
	db *gorm.DB
}

// NewProcurementTypeRepository creates a new instance of ProcurementTypeRepository
func NewProcurementTypeRepository(db *gorm.DB) ProcurementTypeRepository {
	return &procurementTypeRepository{db: db}
}

func (r *procurementTypeRepository) Create(ctx context.Context, procurementType *models.ProcurementType) error {
	return database.GetDB(ctx, r.db).Create(procurementType).Error
}

func (r *procurementTypeRepository) FindByID(ctx context.Context, id string) (*models.ProcurementType, error) {
	var procurementType models.ProcurementType
	err := database.GetDB(ctx, r.db).First(&procurementType, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &procurementType, nil
}

func (r *procurementTypeRepository) List(ctx context.Context, params ListParams) ([]models.ProcurementType, int64, error) {
	var procurementTypes []models.ProcurementType
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.ProcurementType{})

	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", search, search)
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

	if err := query.Find(&procurementTypes).Error; err != nil {
		return nil, 0, err
	}

	return procurementTypes, total, nil
}

func (r *procurementTypeRepository) Update(ctx context.Context, procurementType *models.ProcurementType) error {
	return database.GetDB(ctx, r.db).Save(procurementType).Error
}

func (r *procurementTypeRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.ProcurementType{}, "id = ?", id).Error
}
