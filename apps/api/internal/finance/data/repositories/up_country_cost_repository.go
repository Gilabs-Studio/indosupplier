package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"

	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UpCountryCostListParams struct {
	Search     string
	StartDate  *time.Time
	EndDate    *time.Time
	Status     *financeModels.UpCountryCostStatus
	EmployeeID *string
	Limit      int
	Offset     int
	SortBy     string
	SortDir    string
}

type UpCountryCostStats struct {
	TotalRequests   int64
	PendingApproval int64
	Approved        int64
	TotalAmount     float64
}

type UpCountryCostRepository interface {
	FindByID(ctx context.Context, id string, withRelations bool) (*financeModels.UpCountryCost, error)
	List(ctx context.Context, params UpCountryCostListParams) ([]financeModels.UpCountryCost, int64, error)
	GenerateCode(ctx context.Context, now time.Time) (string, error)
	GetStats(ctx context.Context) (*UpCountryCostStats, error)
}

type upCountryCostRepository struct {
	db *gorm.DB
}

func NewUpCountryCostRepository(db *gorm.DB) UpCountryCostRepository {
	return &upCountryCostRepository{db: db}
}

func (r *upCountryCostRepository) FindByID(ctx context.Context, id string, withRelations bool) (*financeModels.UpCountryCost, error) {
	var item financeModels.UpCountryCost
	q := database.GetDB(ctx, r.db)
	if withRelations {
		q = q.Preload("Employees").Preload("Items")
	}
	if err := q.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *upCountryCostRepository) List(ctx context.Context, params UpCountryCostListParams) ([]financeModels.UpCountryCost, int64, error) {
	var items []financeModels.UpCountryCost
	var total int64

	// Use WithContext (not GetDB) because EmployeeID filter JOINs up_country_cost_employees.
	// GetDB's unqualified WHERE tenant_id=? causes PG ambiguity on JOIN queries.
	q := r.db.WithContext(ctx).Model(&financeModels.UpCountryCost{})
	q = security.ApplyScopeFilter(q, ctx, security.MixedOwnershipScopeQueryOptions("employee_id"))

	if params.EmployeeID != nil && *params.EmployeeID != "" {
		q = q.Joins("JOIN up_country_cost_employees uce ON uce.up_country_cost_id = up_country_costs.id AND uce.deleted_at IS NULL").
			Where("uce.employee_id = ?", *params.EmployeeID)
	}

	if s := strings.TrimSpace(params.Search); s != "" {
		like := "%" + s + "%"
		q = q.Where("up_country_costs.code ILIKE ? OR up_country_costs.purpose ILIKE ? OR up_country_costs.location ILIKE ?", like, like, like)
	}
	if params.Status != nil {
		q = q.Where("up_country_costs.status = ?", *params.Status)
	}
	if params.StartDate != nil {
		q = q.Where("up_country_costs.start_date >= ?", *params.StartDate)
	}
	if params.EndDate != nil {
		q = q.Where("up_country_costs.end_date <= ?", *params.EndDate)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	allowedSort := map[string]clause.OrderByColumn{
		"created_at": {
			Column: clause.Column{Table: "up_country_costs", Name: "created_at"},
		},
		"updated_at": {
			Column: clause.Column{Table: "up_country_costs", Name: "updated_at"},
		},
		"start_date": {
			Column: clause.Column{Table: "up_country_costs", Name: "start_date"},
		},
		"end_date": {
			Column: clause.Column{Table: "up_country_costs", Name: "end_date"},
		},
		"status": {
			Column: clause.Column{Table: "up_country_costs", Name: "status"},
		},
		"amount": {
			Column: clause.Column{Table: "up_country_costs", Name: "amount"},
		},
	}
	sortCol := allowedSort[params.SortBy]
	if sortCol.Column.Name == "" {
		sortCol = allowedSort["created_at"]
	}
	if strings.EqualFold(strings.TrimSpace(params.SortDir), "asc") {
		sortCol.Desc = false
	} else {
		sortCol.Desc = true
	}
	q = q.Order(sortCol)

	if params.Limit > 0 {
		q = q.Limit(params.Limit)
		q = q.Offset(params.Offset)
	}

	// Preload related data required by list UI (participants and item totals)
	if err := q.Preload("Employees").Preload("Items").Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *upCountryCostRepository) GenerateCode(ctx context.Context, now time.Time) (string, error) {
	prefix := "UCC-" + now.Format("200601") + "-"
	var count int64
	if err := database.GetDB(ctx, r.db).Model(&financeModels.UpCountryCost{}).
		Where("code LIKE ?", prefix+"%").
		Count(&count).Error; err != nil {
		return "", err
	}
	return prefix + fmt.Sprintf("%04d", count+1), nil
}

func (r *upCountryCostRepository) GetStats(ctx context.Context) (*UpCountryCostStats, error) {
	type row struct {
		Status string
		Count  int64
		Amount float64
	}

	var rows []row
	q := r.db.WithContext(ctx).
		Model(&financeModels.UpCountryCost{}).
		Select("status, COUNT(*) as count, COALESCE(SUM(ucci.total), 0) as amount").
		Joins("LEFT JOIN (SELECT up_country_cost_id, SUM(amount) as total FROM up_country_cost_items WHERE deleted_at IS NULL GROUP BY up_country_cost_id) ucci ON ucci.up_country_cost_id = up_country_costs.id").
		Where("up_country_costs.deleted_at IS NULL")

	// Ensure tenant scoping is applied with qualified columns to avoid ambiguous tenant_id in JOINs
	q = applyQualifiedTenantFilter(ctx, q, "up_country_costs.tenant_id", "up_country_cost_items.tenant_id")

	q = q.Group("status")

	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	stats := &UpCountryCostStats{}
	for _, row := range rows {
		stats.TotalRequests += row.Count
		stats.TotalAmount += row.Amount
		switch financeModels.UpCountryCostStatus(row.Status) {
		case financeModels.UpCountryCostStatusSubmitted,
			financeModels.UpCountryCostStatusManagerApproved:
			stats.PendingApproval += row.Count
		case financeModels.UpCountryCostStatusFinanceApproved,
			financeModels.UpCountryCostStatusPaid:
			stats.Approved += row.Count
		}
	}
	return stats, nil
}
