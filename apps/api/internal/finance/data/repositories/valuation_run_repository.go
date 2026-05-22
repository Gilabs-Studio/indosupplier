package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"context"
	"strings"
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ValuationRunRepository handles persistence for valuation runs.
type ValuationRunRepository interface {
	Create(ctx context.Context, run *financeModels.ValuationRun) error
	FindByID(ctx context.Context, id string) (*financeModels.ValuationRun, error)
	FindByReferenceID(ctx context.Context, refID string) (*financeModels.ValuationRun, error)
	HasPendingRun(ctx context.Context, valuationType string, periodStart, periodEnd time.Time) (bool, error)
	FindByIDForUpdate(ctx context.Context, tx *gorm.DB, id string) (*financeModels.ValuationRun, error)
	Update(ctx context.Context, run *financeModels.ValuationRun) error
	CreateDetails(ctx context.Context, tx *gorm.DB, details []financeModels.ValuationRunDetail) error
	ListDetails(ctx context.Context, runID string) ([]financeModels.ValuationRunDetail, error)
	List(ctx context.Context, params ValuationRunListParams) ([]financeModels.ValuationRun, int64, error)
}

// ValuationRunListParams holds filters for listing valuation runs.
type ValuationRunListParams struct {
	ValuationType *string
	Status        *string
	StartDate     *time.Time
	EndDate       *time.Time
	SortBy        string
	SortDir       string
	Limit         int
	Offset        int
}

type valuationRunRepository struct {
	db *gorm.DB
}

// NewValuationRunRepository creates a new repository instance.
func NewValuationRunRepository(db *gorm.DB) ValuationRunRepository {
	return &valuationRunRepository{db: db}
}

func (r *valuationRunRepository) Create(ctx context.Context, run *financeModels.ValuationRun) error {
	return database.GetDB(ctx, r.db).Create(run).Error
}

func (r *valuationRunRepository) FindByID(ctx context.Context, id string) (*financeModels.ValuationRun, error) {
	var run financeModels.ValuationRun
	q := security.ApplyScopeFilter(database.GetDB(ctx, r.db).Model(&financeModels.ValuationRun{}), ctx, security.FinanceScopeQueryOptions()).Preload("Details")
	if err := q.First(&run, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &run, nil
}

func (r *valuationRunRepository) FindByReferenceID(ctx context.Context, refID string) (*financeModels.ValuationRun, error) {
	var run financeModels.ValuationRun
	q := security.ApplyScopeFilter(database.GetDB(ctx, r.db).Model(&financeModels.ValuationRun{}), ctx, security.FinanceScopeQueryOptions()).Preload("Details")
	if err := q.First(&run, "reference_id = ?", refID).Error; err != nil {
		return nil, err
	}
	return &run, nil
}

// HasPendingRun checks if there is already a valuation run awaiting completion for the period.
func (r *valuationRunRepository) HasPendingRun(ctx context.Context, valuationType string, periodStart, periodEnd time.Time) (bool, error) {
	var count int64
	err := security.ApplyScopeFilter(database.GetDB(ctx, r.db).Model(&financeModels.ValuationRun{}), ctx, security.FinanceScopeQueryOptions()).
		Where("valuation_type = ? AND status IN ? AND period_start = ? AND period_end = ?",
			valuationType,
			[]financeModels.ValuationRunStatus{
				financeModels.ValuationRunStatusDraft,
				financeModels.ValuationRunStatusPendingApproval,
				financeModels.ValuationRunStatusApproved,
			},
			periodStart,
			periodEnd,
		).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *valuationRunRepository) FindByIDForUpdate(ctx context.Context, tx *gorm.DB, id string) (*financeModels.ValuationRun, error) {
	var run financeModels.ValuationRun
	if err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Details").
		First(&run, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &run, nil
}

func (r *valuationRunRepository) Update(ctx context.Context, run *financeModels.ValuationRun) error {
	return database.GetDB(ctx, r.db).Save(run).Error
}

func (r *valuationRunRepository) CreateDetails(ctx context.Context, tx *gorm.DB, details []financeModels.ValuationRunDetail) error {
	if len(details) == 0 {
		return nil
	}
	return tx.WithContext(ctx).Create(&details).Error
}

func (r *valuationRunRepository) ListDetails(ctx context.Context, runID string) ([]financeModels.ValuationRunDetail, error) {
	items := make([]financeModels.ValuationRunDetail, 0)
	err := database.GetDB(ctx, r.db).
		Where("valuation_run_id = ?", runID).
		Order("created_at asc").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

var valuationRunAllowedSort = map[string]clause.OrderByColumn{
	"created_at": {
		Column: clause.Column{Table: "valuation_runs", Name: "created_at"},
	},
	"updated_at": {
		Column: clause.Column{Table: "valuation_runs", Name: "updated_at"},
	},
	"period_start": {
		Column: clause.Column{Table: "valuation_runs", Name: "period_start"},
	},
	"status": {
		Column: clause.Column{Table: "valuation_runs", Name: "status"},
	},
}

func (r *valuationRunRepository) List(ctx context.Context, params ValuationRunListParams) ([]financeModels.ValuationRun, int64, error) {
	var items []financeModels.ValuationRun
	var total int64

	q := database.GetDB(ctx, r.db).Model(&financeModels.ValuationRun{})
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())

	if params.ValuationType != nil {
		q = q.Where("valuation_runs.valuation_type = ?", *params.ValuationType)
	}
	if params.Status != nil {
		q = q.Where("valuation_runs.status = ?", *params.Status)
	}
	if params.StartDate != nil {
		q = q.Where("valuation_runs.period_start >= ?", *params.StartDate)
	}
	if params.EndDate != nil {
		q = q.Where("valuation_runs.period_end <= ?", *params.EndDate)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortCol := valuationRunAllowedSort[params.SortBy]
	if sortCol.Column.Name == "" {
		sortCol = valuationRunAllowedSort["created_at"]
	}
	if strings.EqualFold(strings.TrimSpace(params.SortDir), "asc") {
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
