package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CashBankJournalListParams struct {
	Search        string
	Type          *financeModels.CashBankType
	Status        *financeModels.CashBankStatus
	BankAccountID *string
	StartDate     *time.Time
	EndDate       *time.Time
	SortBy        string
	SortDir       string
	Limit         int
	Offset        int
}

type CashBankJournalRepository interface {
	FindByID(ctx context.Context, id string, withLines bool) (*financeModels.CashBankJournal, error)
	List(ctx context.Context, params CashBankJournalListParams) ([]financeModels.CashBankJournal, int64, error)
}

type cashBankJournalRepository struct {
	db *gorm.DB
}

func NewCashBankJournalRepository(db *gorm.DB) CashBankJournalRepository {
	return &cashBankJournalRepository{db: db}
}

func (r *cashBankJournalRepository) FindByID(ctx context.Context, id string, withLines bool) (*financeModels.CashBankJournal, error) {
	var item financeModels.CashBankJournal
	q := database.GetDB(ctx, r.db)
	// Apply tenant + permission scope filtering
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())
	if withLines {
		q = q.Preload("Lines").Preload("Lines.ChartOfAccount")
	}
	if err := q.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

var cashBankAllowedSort = map[string]clause.OrderByColumn{
	"created_at": {
		Column: clause.Column{Table: "cash_bank_journals", Name: "created_at"},
	},
	"updated_at": {
		Column: clause.Column{Table: "cash_bank_journals", Name: "updated_at"},
	},
	"transaction_date": {
		Column: clause.Column{Table: "cash_bank_journals", Name: "transaction_date"},
	},
	"status": {
		Column: clause.Column{Table: "cash_bank_journals", Name: "status"},
	},
	"type": {
		Column: clause.Column{Table: "cash_bank_journals", Name: "type"},
	},
	"total_amount": {
		Column: clause.Column{Table: "cash_bank_journals", Name: "total_amount"},
	},
}

func (r *cashBankJournalRepository) List(ctx context.Context, params CashBankJournalListParams) ([]financeModels.CashBankJournal, int64, error) {
	var items []financeModels.CashBankJournal
	var total int64

	q := database.GetDB(ctx, r.db).Model(&financeModels.CashBankJournal{})

	// Apply scope-based data filtering (OWN/DIVISION/AREA/ALL)
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())

	if s := strings.TrimSpace(params.Search); s != "" {
		like := "%" + s + "%"
		q = q.Where("cash_bank_journals.description ILIKE ?", like)
	}
	if params.Type != nil {
		q = q.Where("cash_bank_journals.type = ?", *params.Type)
	}
	if params.Status != nil {
		q = q.Where("cash_bank_journals.status = ?", *params.Status)
	}
	if params.BankAccountID != nil {
		q = q.Where("cash_bank_journals.bank_account_id = ?", *params.BankAccountID)
	}
	if params.StartDate != nil {
		q = q.Where("cash_bank_journals.transaction_date >= ?", *params.StartDate)
	}
	if params.EndDate != nil {
		q = q.Where("cash_bank_journals.transaction_date <= ?", *params.EndDate)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortCol := cashBankAllowedSort[params.SortBy]
	if sortCol.Column.Name == "" {
		sortCol = cashBankAllowedSort["transaction_date"]
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

	// Preload lines so that list responses include line count and details for UI detail view.
	q = q.Preload("Lines").Preload("Lines.ChartOfAccount")

	if err := q.Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
