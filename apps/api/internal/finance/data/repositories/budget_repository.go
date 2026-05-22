package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"context"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BudgetListParams struct {
	Search    string
	Status    *financeModels.BudgetStatus
	StartDate *time.Time
	EndDate   *time.Time
	SortBy    string
	SortDir   string
	Limit     int
	Offset    int
}

type BudgetRepository interface {
	FindByID(ctx context.Context, id string, withItems bool) (*financeModels.Budget, error)
	List(ctx context.Context, params BudgetListParams) ([]financeModels.Budget, int64, error)
	SyncActuals(ctx context.Context, budgetID string) error
}

type budgetRepository struct {
	db *gorm.DB
}

func NewBudgetRepository(db *gorm.DB) BudgetRepository {
	return &budgetRepository{db: db}
}

func (r *budgetRepository) FindByID(ctx context.Context, id string, withItems bool) (*financeModels.Budget, error) {
	var item financeModels.Budget
	q := database.GetDB(ctx, r.db)
	if withItems {
		q = q.Preload("Items").Preload("Items.ChartOfAccount")
	}
	if err := q.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

var budgetAllowedSort = map[string]clause.OrderByColumn{
	"created_at": {
		Column: clause.Column{Table: "budgets", Name: "created_at"},
	},
	"updated_at": {
		Column: clause.Column{Table: "budgets", Name: "updated_at"},
	},
	"start_date": {
		Column: clause.Column{Table: "budgets", Name: "start_date"},
	},
	"end_date": {
		Column: clause.Column{Table: "budgets", Name: "end_date"},
	},
	"status": {
		Column: clause.Column{Table: "budgets", Name: "status"},
	},
	"total_amount": {
		Column: clause.Column{Table: "budgets", Name: "total_amount"},
	},
}

func (r *budgetRepository) List(ctx context.Context, params BudgetListParams) ([]financeModels.Budget, int64, error) {
	var items []financeModels.Budget
	var total int64

	q := database.GetDB(ctx, r.db).Model(&financeModels.Budget{})
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())
	if s := strings.TrimSpace(params.Search); s != "" {
		like := "%" + s + "%"
		q = q.Where("budgets.name ILIKE ? OR budgets.description ILIKE ?", like, like)
	}
	if params.Status != nil {
		q = q.Where("budgets.status = ?", *params.Status)
	}
	if params.StartDate != nil {
		q = q.Where("budgets.start_date >= ?", *params.StartDate)
	}
	if params.EndDate != nil {
		q = q.Where("budgets.end_date <= ?", *params.EndDate)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortCol := budgetAllowedSort[params.SortBy]
	if sortCol.Column.Name == "" {
		sortCol = budgetAllowedSort["start_date"]
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

func (r *budgetRepository) SyncActuals(ctx context.Context, budgetID string) error {
	// Query to update actual_amount based on journal lines
	// For simplicity and correctness across account types, we join chart_of_accounts
	query := `
		UPDATE budget_items bi
		SET actual_amount = (
			SELECT COALESCE(SUM(
				CASE 
					WHEN coa.type IN ('expense', 'asset') THEN jl.debit - jl.credit
					ELSE jl.credit - jl.debit
				END
			), 0)
			FROM journal_lines jl
			JOIN journal_entries je ON je.id = jl.journal_entry_id
			JOIN chart_of_accounts coa ON coa.id = jl.chart_of_account_id
			WHERE jl.chart_of_account_id = bi.chart_of_account_id
			  AND je.entry_date >= b.start_date
			  AND je.entry_date <= b.end_date
			  AND je.status = 'posted'
		)
		FROM budgets b
		WHERE bi.budget_id = b.id
		  AND b.id = ?
	`
	tenantID := strings.TrimSpace(middleware.TenantFromContext(ctx))
	args := []interface{}{budgetID}
	// inject tenant filter into query when tenant present
	if tenantID != "" {
		// restrict journal_lines and budgets/budget_items to the tenant
		query = strings.Replace(query, "WHERE jl.chart_of_account_id = bi.chart_of_account_id", "WHERE jl.chart_of_account_id = bi.chart_of_account_id AND jl.tenant_id = ?", 1)
		query += " AND bi.tenant_id = ? AND b.tenant_id = ?"
		args = append(args, tenantID, tenantID, tenantID)
		// note: we appended tenantID three times to match the three placeholders above
	}
	return r.db.WithContext(ctx).Exec(query, args...).Error
}
