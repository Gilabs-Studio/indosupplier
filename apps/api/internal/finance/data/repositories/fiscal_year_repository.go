package repositories

import (
	"context"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
)

type FiscalYearListParams struct {
	CompanyID string
	Status    *financeModels.FiscalYearStatus
	Page      int
	PerPage   int
}

type FiscalYearRepository interface {
	Create(ctx context.Context, item *financeModels.FiscalYear) error
	Update(ctx context.Context, item *financeModels.FiscalYear) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*financeModels.FiscalYear, error)
	List(ctx context.Context, params FiscalYearListParams) ([]financeModels.FiscalYear, int64, error)
	FindActiveByCompany(ctx context.Context, companyID string) (*financeModels.FiscalYear, error)
	HasPostedJournalInRange(ctx context.Context, companyID string, startDate, endDate time.Time) (bool, error)
	GetDB(ctx context.Context) *gorm.DB
}

type fiscalYearRepository struct {
	db *gorm.DB
}

func NewFiscalYearRepository(db *gorm.DB) FiscalYearRepository {
	return &fiscalYearRepository{db: db}
}

func (r *fiscalYearRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value("tx").(*gorm.DB); ok && tx != nil {
		return tx
	}
	return database.GetDB(ctx, r.db)
}

func (r *fiscalYearRepository) GetDB(ctx context.Context) *gorm.DB {
	return r.getDB(ctx)
}

func (r *fiscalYearRepository) Create(ctx context.Context, item *financeModels.FiscalYear) error {
	return r.getDB(ctx).Create(item).Error
}

func (r *fiscalYearRepository) Update(ctx context.Context, item *financeModels.FiscalYear) error {
	return r.getDB(ctx).Save(item).Error
}

func (r *fiscalYearRepository) Delete(ctx context.Context, id string) error {
	q := security.ApplyScopeFilter(r.getDB(ctx).Model(&financeModels.FiscalYear{}), ctx, security.FinanceScopeQueryOptions())
	res := q.Where("id = ?", id).Delete(&financeModels.FiscalYear{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *fiscalYearRepository) FindByID(ctx context.Context, id string) (*financeModels.FiscalYear, error) {
	var item financeModels.FiscalYear
	q := security.ApplyScopeFilter(r.getDB(ctx).Model(&financeModels.FiscalYear{}), ctx, security.FinanceScopeQueryOptions())
	if err := q.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *fiscalYearRepository) List(ctx context.Context, params FiscalYearListParams) ([]financeModels.FiscalYear, int64, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PerPage <= 0 {
		params.PerPage = 20
	}

	query := security.ApplyScopeFilter(r.getDB(ctx).Model(&financeModels.FiscalYear{}), ctx, security.FinanceScopeQueryOptions()).Where("company_id = ?", params.CompanyID)
	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	items := make([]financeModels.FiscalYear, 0)
	offset := (params.Page - 1) * params.PerPage
	if err := query.Order("start_date desc").Limit(params.PerPage).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *fiscalYearRepository) FindActiveByCompany(ctx context.Context, companyID string) (*financeModels.FiscalYear, error) {
	var item financeModels.FiscalYear
	q := security.ApplyScopeFilter(r.getDB(ctx).Model(&financeModels.FiscalYear{}), ctx, security.FinanceScopeQueryOptions())
	if err := q.Where("company_id = ? AND status = ?", companyID, financeModels.FiscalYearStatusActive).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *fiscalYearRepository) HasPostedJournalInRange(ctx context.Context, companyID string, startDate, endDate time.Time) (bool, error) {
	db := r.getDB(ctx)
	if !db.Migrator().HasColumn(&financeModels.JournalEntry{}, "company_id") {
		return false, nil
	}

	var count int64
	err := db.Model(&financeModels.JournalEntry{}).
		Where("company_id = ?", companyID).
		Where("status = ?", financeModels.JournalStatusPosted).
		Where("entry_date >= ? AND entry_date <= ?", startDate, endDate).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
