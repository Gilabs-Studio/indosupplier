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

type CashBankTransactionListParams struct {
	CompanyID     string
	BankAccountID string
	Type          *financeModels.CashBankTransactionType
	Status        *financeModels.CashBankTransactionStatus
	StartDate     *time.Time
	EndDate       *time.Time
	Search        string
	SortBy        string
	SortDir       string
	Limit         int
	Offset        int
}

type CashBankTransactionRepository interface {
	Create(ctx context.Context, tx *financeModels.CashBankTransaction) error
	FindByID(ctx context.Context, id string) (*financeModels.CashBankTransaction, error)
	List(ctx context.Context, params CashBankTransactionListParams) ([]financeModels.CashBankTransaction, int64, error)
	Update(ctx context.Context, tx *financeModels.CashBankTransaction) error
}

type cashBankTransactionRepository struct {
	db *gorm.DB
}

func NewCashBankTransactionRepository(db *gorm.DB) CashBankTransactionRepository {
	return &cashBankTransactionRepository{db: db}
}

func (r *cashBankTransactionRepository) Create(ctx context.Context, tx *financeModels.CashBankTransaction) error {
	return database.GetDB(ctx, r.db).Create(tx).Error
}

func (r *cashBankTransactionRepository) FindByID(ctx context.Context, id string) (*financeModels.CashBankTransaction, error) {
	var item financeModels.CashBankTransaction
	q := security.ApplyScopeFilter(database.GetDB(ctx, r.db).Model(&financeModels.CashBankTransaction{}), ctx, security.FinanceScopeQueryOptions())
	if err := q.First(&item, "id = ?", strings.TrimSpace(id)).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

var cashBankTransactionAllowedSort = map[string]clause.OrderByColumn{
	"created_at": {
		Column: clause.Column{Table: "cash_bank_transactions", Name: "created_at"},
	},
	"updated_at": {
		Column: clause.Column{Table: "cash_bank_transactions", Name: "updated_at"},
	},
	"date": {
		Column: clause.Column{Table: "cash_bank_transactions", Name: "date"},
	},
	"amount": {
		Column: clause.Column{Table: "cash_bank_transactions", Name: "amount"},
	},
	"status": {
		Column: clause.Column{Table: "cash_bank_transactions", Name: "status"},
	},
}

func (r *cashBankTransactionRepository) List(ctx context.Context, params CashBankTransactionListParams) ([]financeModels.CashBankTransaction, int64, error) {
	var items []financeModels.CashBankTransaction
	var total int64

	q := database.GetDB(ctx, r.db).Model(&financeModels.CashBankTransaction{})
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())

	if companyID := strings.TrimSpace(params.CompanyID); companyID != "" {
		q = q.Where("company_id = ?", companyID)
	}
	if bankAccountID := strings.TrimSpace(params.BankAccountID); bankAccountID != "" {
		q = q.Where("bank_account_id = ?", bankAccountID)
	}
	if params.Type != nil {
		q = q.Where("type = ?", *params.Type)
	}
	if params.Status != nil {
		q = q.Where("status = ?", *params.Status)
	}
	if params.StartDate != nil {
		q = q.Where("date >= ?", *params.StartDate)
	}
	if params.EndDate != nil {
		q = q.Where("date <= ?", *params.EndDate)
	}
	if search := strings.TrimSpace(params.Search); search != "" {
		like := "%" + search + "%"
		q = q.Where("reference_number ILIKE ? OR reference ILIKE ? OR description ILIKE ?", like, like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortCol := cashBankTransactionAllowedSort[params.SortBy]
	if sortCol.Column.Name == "" {
		sortCol = cashBankTransactionAllowedSort["date"]
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

func (r *cashBankTransactionRepository) Update(ctx context.Context, tx *financeModels.CashBankTransaction) error {
	return database.GetDB(ctx, r.db).Save(tx).Error
}
