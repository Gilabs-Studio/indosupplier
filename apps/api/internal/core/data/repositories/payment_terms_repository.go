package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"

	"github.com/gilabs/gims/api/internal/core/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ListParams defines pagination and filtering parameters
type ListParams struct {
	Search  string
	SortBy  string
	SortDir string
	Limit   int
	Offset  int
}

// PaymentTermsRepository defines the interface for payment terms data access
type PaymentTermsRepository interface {
	Create(ctx context.Context, paymentTerms *models.PaymentTerms) error
	FindByID(ctx context.Context, id string) (*models.PaymentTerms, error)
	List(ctx context.Context, params ListParams) ([]models.PaymentTerms, int64, error)
	Update(ctx context.Context, paymentTerms *models.PaymentTerms) error
	Delete(ctx context.Context, id string) error
}

type paymentTermsRepository struct {
	db *gorm.DB
}

// NewPaymentTermsRepository creates a new instance of PaymentTermsRepository
func NewPaymentTermsRepository(db *gorm.DB) PaymentTermsRepository {
	return &paymentTermsRepository{db: db}
}

func (r *paymentTermsRepository) Create(ctx context.Context, paymentTerms *models.PaymentTerms) error {
	return database.GetDB(ctx, r.db).Create(paymentTerms).Error
}

func (r *paymentTermsRepository) FindByID(ctx context.Context, id string) (*models.PaymentTerms, error) {
	var paymentTerms models.PaymentTerms
	err := database.GetDB(ctx, r.db).First(&paymentTerms, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &paymentTerms, nil
}

func (r *paymentTermsRepository) List(ctx context.Context, params ListParams) ([]models.PaymentTerms, int64, error) {
	var paymentTermsList []models.PaymentTerms
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.PaymentTerms{})

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

	if err := query.Find(&paymentTermsList).Error; err != nil {
		return nil, 0, err
	}

	return paymentTermsList, total, nil
}

func (r *paymentTermsRepository) Update(ctx context.Context, paymentTerms *models.PaymentTerms) error {
	return database.GetDB(ctx, r.db).Save(paymentTerms).Error
}

func (r *paymentTermsRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.PaymentTerms{}, "id = ?", id).Error
}
