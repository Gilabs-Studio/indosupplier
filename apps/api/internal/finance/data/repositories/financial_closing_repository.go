package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FinancialClosingListParams struct {
	Limit   int
	Offset  int
	SortBy  string
	SortDir string
}

type FinancialClosingRepository interface {
	FindByID(ctx context.Context, id string) (*financeModels.FinancialClosing, error)
	List(ctx context.Context, params FinancialClosingListParams) ([]financeModels.FinancialClosing, int64, error)
	LatestApproved(ctx context.Context) (*financeModels.FinancialClosing, error)
	Delete(ctx context.Context, id string) error
}

type financialClosingRepository struct {
	db *gorm.DB
}

func NewFinancialClosingRepository(db *gorm.DB) FinancialClosingRepository {
	return &financialClosingRepository{db: db}
}

func (r *financialClosingRepository) FindByID(ctx context.Context, id string) (*financeModels.FinancialClosing, error) {
	var item financeModels.FinancialClosing
	q := security.ApplyScopeFilter(database.GetDB(ctx, r.db).Model(&financeModels.FinancialClosing{}), ctx, security.FinanceScopeQueryOptions())
	if err := q.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

var financialClosingAllowedSort = map[string]clause.OrderByColumn{
	"created_at": {
		Column: clause.Column{Table: "financial_closings", Name: "created_at"},
	},
	"updated_at": {
		Column: clause.Column{Table: "financial_closings", Name: "updated_at"},
	},
	"period_end_date": {
		Column: clause.Column{Table: "financial_closings", Name: "period_end_date"},
	},
	"status": {
		Column: clause.Column{Table: "financial_closings", Name: "status"},
	},
}

func (r *financialClosingRepository) List(ctx context.Context, params FinancialClosingListParams) ([]financeModels.FinancialClosing, int64, error) {
	var items []financeModels.FinancialClosing
	var total int64

	q := database.GetDB(ctx, r.db).Model(&financeModels.FinancialClosing{})
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortCol := financialClosingAllowedSort[params.SortBy]
	if sortCol.Column.Name == "" {
		sortCol = financialClosingAllowedSort["period_end_date"]
	}
	sortDir := strings.ToLower(strings.TrimSpace(params.SortDir))
	if sortDir == "asc" {
		sortCol.Desc = false
	} else {
		sortCol.Desc = true
	}
	q = q.Order(sortCol)

	if params.Limit > 0 {
		q = q.Limit(params.Limit)
	}
	if params.Offset > 0 {
		q = q.Offset(params.Offset)
	}

	if err := q.Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *financialClosingRepository) LatestApproved(ctx context.Context) (*financeModels.FinancialClosing, error) {
	var item financeModels.FinancialClosing
	if err := security.ApplyScopeFilter(database.GetDB(ctx, r.db).Model(&financeModels.FinancialClosing{}), ctx, security.FinanceScopeQueryOptions()).
		Where("status = ?", financeModels.FinancialClosingStatusApproved).
		Order("period_end_date desc").
		First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}
func (r *financialClosingRepository) Delete(ctx context.Context, id string) error {
	q := security.ApplyScopeFilter(database.GetDB(ctx, r.db).Model(&financeModels.FinancialClosing{}), ctx, security.FinanceScopeQueryOptions())
	return q.Delete(&financeModels.FinancialClosing{}, "id = ?", id).Error
}
