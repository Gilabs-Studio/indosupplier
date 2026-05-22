package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type NonTradePayableListParams struct {
	Search    string
	StartDate *time.Time
	EndDate   *time.Time
	Status    *financeModels.NonTradePayableStatus
	Limit     int
	Offset    int
	SortBy    string
	SortDir   string
}

type NonTradePayableRepository interface {
	FindByID(ctx context.Context, id string) (*financeModels.NonTradePayable, error)
	List(ctx context.Context, params NonTradePayableListParams) ([]financeModels.NonTradePayable, int64, error)
	GenerateCode(ctx context.Context, now time.Time) (string, error)
}

type nonTradePayableRepository struct {
	db *gorm.DB
}

func NewNonTradePayableRepository(db *gorm.DB) NonTradePayableRepository {
	return &nonTradePayableRepository{db: db}
}

func (r *nonTradePayableRepository) FindByID(ctx context.Context, id string) (*financeModels.NonTradePayable, error) {
	var item financeModels.NonTradePayable
	if err := database.GetDB(ctx, r.db).
		Preload("ChartOfAccount").
		First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

var nonTradePayableAllowedSort = map[string]clause.OrderByColumn{
	"created_at": {
		Column: clause.Column{Table: "non_trade_payables", Name: "created_at"},
	},
	"updated_at": {
		Column: clause.Column{Table: "non_trade_payables", Name: "updated_at"},
	},
	"transaction_date": {
		Column: clause.Column{Table: "non_trade_payables", Name: "transaction_date"},
	},
	"amount": {
		Column: clause.Column{Table: "non_trade_payables", Name: "amount"},
	},
}

func (r *nonTradePayableRepository) List(ctx context.Context, params NonTradePayableListParams) ([]financeModels.NonTradePayable, int64, error) {
	var items []financeModels.NonTradePayable
	var total int64

	q := database.GetDB(ctx, r.db).Model(&financeModels.NonTradePayable{}).Preload("ChartOfAccount")
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())
	if s := strings.TrimSpace(params.Search); s != "" {
		like := "%" + s + "%"
		q = q.Where("non_trade_payables.description ILIKE ? OR non_trade_payables.vendor_name ILIKE ? OR non_trade_payables.reference_no ILIKE ? OR non_trade_payables.code ILIKE ?", like, like, like, like)
	}
	if params.Status != nil {
		q = q.Where("non_trade_payables.status = ?", *params.Status)
	}
	if params.StartDate != nil {
		q = q.Where("non_trade_payables.transaction_date >= ?", *params.StartDate)
	}
	if params.EndDate != nil {
		q = q.Where("non_trade_payables.transaction_date <= ?", *params.EndDate)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortCol := nonTradePayableAllowedSort[params.SortBy]
	if sortCol.Column.Name == "" {
		sortCol = nonTradePayableAllowedSort["transaction_date"]
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

func (r *nonTradePayableRepository) GenerateCode(ctx context.Context, now time.Time) (string, error) {
	prefix := "NTP-" + now.Format("200601") + "-"
	var count int64
	if err := database.GetDB(ctx, r.db).Model(&financeModels.NonTradePayable{}).
		Where("code LIKE ?", prefix+"%").
		Count(&count).Error; err != nil {
		return "", err
	}
	return prefix + fmt.Sprintf("%04d", count+1), nil
}
