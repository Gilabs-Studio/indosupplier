package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"

	"github.com/gilabs/gims/api/internal/customer/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CustomerTypeRepository defines the interface for customer type data access
type CustomerTypeRepository interface {
	Create(ctx context.Context, customerType *models.CustomerType) error
	FindByID(ctx context.Context, id string) (*models.CustomerType, error)
	List(ctx context.Context, params ListParams) ([]models.CustomerType, int64, error)
	Update(ctx context.Context, customerType *models.CustomerType) error
	Delete(ctx context.Context, id string) error
}

type customerTypeRepository struct {
	db *gorm.DB
}

// NewCustomerTypeRepository creates a new CustomerTypeRepository
func NewCustomerTypeRepository(db *gorm.DB) CustomerTypeRepository {
	return &customerTypeRepository{db: db}
}

func (r *customerTypeRepository) Create(ctx context.Context, customerType *models.CustomerType) error {
	return database.GetDB(ctx, r.db).Create(customerType).Error
}

func (r *customerTypeRepository) FindByID(ctx context.Context, id string) (*models.CustomerType, error) {
	var customerType models.CustomerType
	err := database.GetDB(ctx, r.db).First(&customerType, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &customerType, nil
}

func (r *customerTypeRepository) List(ctx context.Context, params ListParams) ([]models.CustomerType, int64, error) {
	var customerTypes []models.CustomerType
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.CustomerType{})

	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", search, search)
	}

	if params.ActiveOnly != nil {
		query = query.Where("is_active = ?", *params.ActiveOnly)
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

	if err := query.Find(&customerTypes).Error; err != nil {
		return nil, 0, err
	}

	return customerTypes, total, nil
}

func (r *customerTypeRepository) Update(ctx context.Context, customerType *models.CustomerType) error {
	return database.GetDB(ctx, r.db).Save(customerType).Error
}

func (r *customerTypeRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.CustomerType{}, "id = ?", id).Error
}
