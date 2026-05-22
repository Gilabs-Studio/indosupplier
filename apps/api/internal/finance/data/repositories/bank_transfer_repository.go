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

type BankTransferListParams struct {
	CompanyID         string
	FromBankAccountID string
	ToBankAccountID   string
	Status            *financeModels.BankTransferStatus
	StartDate         *time.Time
	EndDate           *time.Time
	Search            string
	SortBy            string
	SortDir           string
	Limit             int
	Offset            int
}

type BankTransferRepository interface {
	Create(ctx context.Context, item *financeModels.BankTransfer) error
	FindByID(ctx context.Context, id string) (*financeModels.BankTransfer, error)
	List(ctx context.Context, params BankTransferListParams) ([]financeModels.BankTransfer, int64, error)
	Update(ctx context.Context, item *financeModels.BankTransfer) error
}

type bankTransferRepository struct {
	db *gorm.DB
}

func NewBankTransferRepository(db *gorm.DB) BankTransferRepository {
	return &bankTransferRepository{db: db}
}

func (r *bankTransferRepository) Create(ctx context.Context, item *financeModels.BankTransfer) error {
	return database.GetDB(ctx, r.db).Create(item).Error
}

func (r *bankTransferRepository) FindByID(ctx context.Context, id string) (*financeModels.BankTransfer, error) {
	var item financeModels.BankTransfer
	q := security.ApplyScopeFilter(database.GetDB(ctx, r.db).Model(&financeModels.BankTransfer{}), ctx, security.FinanceScopeQueryOptions())
	if err := q.First(&item, "id = ?", strings.TrimSpace(id)).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

var bankTransferAllowedSort = map[string]clause.OrderByColumn{
	"created_at": {
		Column: clause.Column{Table: "bank_transfers", Name: "created_at"},
	},
	"updated_at": {
		Column: clause.Column{Table: "bank_transfers", Name: "updated_at"},
	},
	"date": {
		Column: clause.Column{Table: "bank_transfers", Name: "date"},
	},
	"amount": {
		Column: clause.Column{Table: "bank_transfers", Name: "amount"},
	},
	"status": {
		Column: clause.Column{Table: "bank_transfers", Name: "status"},
	},
}

func (r *bankTransferRepository) List(ctx context.Context, params BankTransferListParams) ([]financeModels.BankTransfer, int64, error) {
	var items []financeModels.BankTransfer
	var total int64

	q := database.GetDB(ctx, r.db).Model(&financeModels.BankTransfer{})
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())

	if companyID := strings.TrimSpace(params.CompanyID); companyID != "" {
		q = q.Where("company_id = ?", companyID)
	}
	if fromID := strings.TrimSpace(params.FromBankAccountID); fromID != "" {
		q = q.Where("from_bank_account_id = ?", fromID)
	}
	if toID := strings.TrimSpace(params.ToBankAccountID); toID != "" {
		q = q.Where("to_bank_account_id = ?", toID)
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
		q = q.Where("transfer_number ILIKE ? OR reference ILIKE ? OR description ILIKE ?", like, like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortCol := bankTransferAllowedSort[params.SortBy]
	if sortCol.Column.Name == "" {
		sortCol = bankTransferAllowedSort["date"]
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

func (r *bankTransferRepository) Update(ctx context.Context, item *financeModels.BankTransfer) error {
	return database.GetDB(ctx, r.db).Save(item).Error
}
