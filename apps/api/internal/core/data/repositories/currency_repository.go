package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CurrencyRepository interface {
	Create(ctx context.Context, currency *models.Currency) error
	FindByID(ctx context.Context, id string) (*models.Currency, error)
	FindByCode(ctx context.Context, code string) (*models.Currency, error)
	List(ctx context.Context, params ListParams) ([]models.Currency, int64, error)
	Update(ctx context.Context, currency *models.Currency) error
	Delete(ctx context.Context, id string) error
}

type currencyRepository struct {
	db *gorm.DB
}

func NewCurrencyRepository(db *gorm.DB) CurrencyRepository {
	return &currencyRepository{db: db}
}

func (r *currencyRepository) Create(ctx context.Context, currency *models.Currency) error {
	return r.db.WithContext(ctx).Create(currency).Error
}

func (r *currencyRepository) FindByID(ctx context.Context, id string) (*models.Currency, error) {
	var currency models.Currency
	if err := r.db.WithContext(ctx).First(&currency, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &currency, nil
}

func (r *currencyRepository) FindByCode(ctx context.Context, code string) (*models.Currency, error) {
	var currency models.Currency
	if err := r.db.WithContext(ctx).First(&currency, "code = ?", code).Error; err != nil {
		return nil, err
	}
	return &currency, nil
}

func (r *currencyRepository) List(ctx context.Context, params ListParams) ([]models.Currency, int64, error) {
	var items []models.Currency
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Currency{})
	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("code ILIKE ? OR name ILIKE ? OR symbol ILIKE ?", search, search, search)
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
		query = query.Order("code ASC")
	}

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	if err := query.Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *currencyRepository) Update(ctx context.Context, currency *models.Currency) error {
	return r.db.WithContext(ctx).Save(currency).Error
}

func (r *currencyRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Currency{}, "id = ?", id).Error
}
