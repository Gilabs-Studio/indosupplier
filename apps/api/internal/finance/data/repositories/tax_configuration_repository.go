package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"

	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
)

type TaxConfigurationListParams struct {
	CompanyID string
	TaxType   *financeModels.TaxType
	IsActive  *bool
	Page      int
	PerPage   int
}

type TaxConfigurationRepository interface {
	Create(ctx context.Context, item *financeModels.TaxConfiguration) error
	Update(ctx context.Context, item *financeModels.TaxConfiguration) error
	FindByID(ctx context.Context, id string) (*financeModels.TaxConfiguration, error)
	List(ctx context.Context, params TaxConfigurationListParams) ([]financeModels.TaxConfiguration, int64, error)
	ExistsByTaxCode(ctx context.Context, companyID, taxCode string, excludeID *string) (bool, error)
	GetDB(ctx context.Context) *gorm.DB
}

type taxConfigurationRepository struct {
	db *gorm.DB
}

func NewTaxConfigurationRepository(db *gorm.DB) TaxConfigurationRepository {
	return &taxConfigurationRepository{db: db}
}

func (r *taxConfigurationRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value("tx").(*gorm.DB); ok && tx != nil {
		return tx
	}
	return database.GetDB(ctx, r.db)
}

func (r *taxConfigurationRepository) GetDB(ctx context.Context) *gorm.DB {
	return r.getDB(ctx)
}

func (r *taxConfigurationRepository) Create(ctx context.Context, item *financeModels.TaxConfiguration) error {
	return r.getDB(ctx).Create(item).Error
}

func (r *taxConfigurationRepository) Update(ctx context.Context, item *financeModels.TaxConfiguration) error {
	return r.getDB(ctx).Save(item).Error
}

func (r *taxConfigurationRepository) FindByID(ctx context.Context, id string) (*financeModels.TaxConfiguration, error) {
	var item financeModels.TaxConfiguration
	q := security.ApplyScopeFilter(r.getDB(ctx).Model(&financeModels.TaxConfiguration{}), ctx, security.FinanceScopeQueryOptions())
	if err := q.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *taxConfigurationRepository) List(ctx context.Context, params TaxConfigurationListParams) ([]financeModels.TaxConfiguration, int64, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PerPage <= 0 {
		params.PerPage = 20
	}

	query := security.ApplyScopeFilter(r.getDB(ctx).Model(&financeModels.TaxConfiguration{}), ctx, security.FinanceScopeQueryOptions()).Where("company_id = ?", params.CompanyID)
	if params.TaxType != nil {
		query = query.Where("tax_type = ?", *params.TaxType)
	}
	if params.IsActive != nil {
		query = query.Where("is_active = ?", *params.IsActive)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	items := make([]financeModels.TaxConfiguration, 0)
	offset := (params.Page - 1) * params.PerPage
	if err := query.Order("tax_code asc").Limit(params.PerPage).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *taxConfigurationRepository) ExistsByTaxCode(ctx context.Context, companyID, taxCode string, excludeID *string) (bool, error) {
	query := r.getDB(ctx).Model(&financeModels.TaxConfiguration{}).Where("company_id = ? AND tax_code = ?", companyID, taxCode)
	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
