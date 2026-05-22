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

type BankReconciliationListParams struct {
	CompanyID     string
	BankAccountID string
	Status        *financeModels.BankReconciliationStatus
	StartDate     *time.Time
	EndDate       *time.Time
	SortBy        string
	SortDir       string
	Limit         int
	Offset        int
}

type BankReconciliationRepository interface {
	Create(ctx context.Context, item *financeModels.BankReconciliation) error
	FindByID(ctx context.Context, id string, preloadLines bool) (*financeModels.BankReconciliation, error)
	List(ctx context.Context, params BankReconciliationListParams) ([]financeModels.BankReconciliation, int64, error)
	Update(ctx context.Context, item *financeModels.BankReconciliation) error
	CreateStatementLines(ctx context.Context, lines []financeModels.BankStatementLine) error
	UpdateStatementLine(ctx context.Context, item *financeModels.BankStatementLine) error
	FindStatementLineByID(ctx context.Context, reconciliationID, lineID string) (*financeModels.BankStatementLine, error)
}

type bankReconciliationRepository struct {
	db *gorm.DB
}

func NewBankReconciliationRepository(db *gorm.DB) BankReconciliationRepository {
	return &bankReconciliationRepository{db: db}
}

func (r *bankReconciliationRepository) Create(ctx context.Context, item *financeModels.BankReconciliation) error {
	return database.GetDB(ctx, r.db).Create(item).Error
}

func (r *bankReconciliationRepository) FindByID(ctx context.Context, id string, preloadLines bool) (*financeModels.BankReconciliation, error) {
	var item financeModels.BankReconciliation
	q := security.ApplyScopeFilter(database.GetDB(ctx, r.db).Model(&financeModels.BankReconciliation{}), ctx, security.FinanceScopeQueryOptions())
	if preloadLines {
		q = q.Preload("Lines")
	}
	if err := q.First(&item, "id = ?", strings.TrimSpace(id)).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

var bankReconciliationAllowedSort = map[string]clause.OrderByColumn{
	"created_at": {
		Column: clause.Column{Table: "bank_reconciliations", Name: "created_at"},
	},
	"updated_at": {
		Column: clause.Column{Table: "bank_reconciliations", Name: "updated_at"},
	},
	"statement_date": {
		Column: clause.Column{Table: "bank_reconciliations", Name: "statement_date"},
	},
	"status": {
		Column: clause.Column{Table: "bank_reconciliations", Name: "status"},
	},
}

func (r *bankReconciliationRepository) List(ctx context.Context, params BankReconciliationListParams) ([]financeModels.BankReconciliation, int64, error) {
	var items []financeModels.BankReconciliation
	var total int64

	q := database.GetDB(ctx, r.db).Model(&financeModels.BankReconciliation{})
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())

	if companyID := strings.TrimSpace(params.CompanyID); companyID != "" {
		q = q.Where("company_id = ?", companyID)
	}
	if bankAccountID := strings.TrimSpace(params.BankAccountID); bankAccountID != "" {
		q = q.Where("bank_account_id = ?", bankAccountID)
	}
	if params.Status != nil {
		q = q.Where("status = ?", *params.Status)
	}
	if params.StartDate != nil {
		q = q.Where("statement_date >= ?", *params.StartDate)
	}
	if params.EndDate != nil {
		q = q.Where("statement_date <= ?", *params.EndDate)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortCol := bankReconciliationAllowedSort[params.SortBy]
	if sortCol.Column.Name == "" {
		sortCol = bankReconciliationAllowedSort["statement_date"]
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

func (r *bankReconciliationRepository) Update(ctx context.Context, item *financeModels.BankReconciliation) error {
	return database.GetDB(ctx, r.db).Save(item).Error
}

func (r *bankReconciliationRepository) CreateStatementLines(ctx context.Context, lines []financeModels.BankStatementLine) error {
	if len(lines) == 0 {
		return nil
	}
	return database.GetDB(ctx, r.db).Create(&lines).Error
}

func (r *bankReconciliationRepository) UpdateStatementLine(ctx context.Context, item *financeModels.BankStatementLine) error {
	return database.GetDB(ctx, r.db).Save(item).Error
}

func (r *bankReconciliationRepository) FindStatementLineByID(ctx context.Context, reconciliationID, lineID string) (*financeModels.BankStatementLine, error) {
	var item financeModels.BankStatementLine
	if err := database.GetDB(ctx, r.db).
		First(&item, "id = ? AND bank_reconciliation_id = ?", strings.TrimSpace(lineID), strings.TrimSpace(reconciliationID)).Error; err != nil {
		return nil, err
	}
	return &item, nil
}
