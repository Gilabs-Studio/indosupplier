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

type PaymentListParams struct {
	Search    string
	Status    *financeModels.PaymentStatus
	StartDate *time.Time
	EndDate   *time.Time
	SortBy    string
	SortDir   string
	Limit     int
	Offset    int
}

type PaymentRepository interface {
	FindByID(ctx context.Context, id string, withAllocations bool) (*financeModels.Payment, error)
	List(ctx context.Context, params PaymentListParams) ([]financeModels.Payment, int64, error)
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) FindByID(ctx context.Context, id string, withAllocations bool) (*financeModels.Payment, error) {
	var item financeModels.Payment
	q := database.GetDB(ctx, r.db)
	if withAllocations {
		q = q.Preload("Allocations").Preload("Allocations.ChartOfAccount")
	}
	if err := q.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

var paymentAllowedSort = map[string]clause.OrderByColumn{
	"created_at": {
		Column: clause.Column{Table: "payments", Name: "created_at"},
	},
	"updated_at": {
		Column: clause.Column{Table: "payments", Name: "updated_at"},
	},
	"payment_date": {
		Column: clause.Column{Table: "payments", Name: "payment_date"},
	},
	"status": {
		Column: clause.Column{Table: "payments", Name: "status"},
	},
	"total_amount": {
		Column: clause.Column{Table: "payments", Name: "total_amount"},
	},
}

func (r *paymentRepository) List(ctx context.Context, params PaymentListParams) ([]financeModels.Payment, int64, error) {
	var items []financeModels.Payment
	var total int64

	q := database.GetDB(ctx, r.db).Model(&financeModels.Payment{})

	// Apply scope-based data filtering (OWN/DIVISION/AREA/ALL)
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())

	if s := strings.TrimSpace(params.Search); s != "" {
		// Prefix search keeps queries index-friendly on large tables.
		like := "%" + s + "%"
		q = q.Where("payments.description ILIKE ?", like)
	}
	if params.Status != nil {
		q = q.Where("payments.status = ?", *params.Status)
	}
	if params.StartDate != nil {
		q = q.Where("payments.payment_date >= ?", *params.StartDate)
	}
	if params.EndDate != nil {
		q = q.Where("payments.payment_date <= ?", *params.EndDate)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortCol := paymentAllowedSort[params.SortBy]
	if sortCol.Column.Name == "" {
		sortCol = paymentAllowedSort["payment_date"]
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
