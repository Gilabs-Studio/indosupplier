package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/gilabs/gims/api/internal/sales/data/models"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SalesTargetRepository defines the interface for sales target data access
type SalesTargetRepository interface {
	FindByID(ctx context.Context, id string) (*models.SalesTarget, error)
	List(ctx context.Context, req *dto.ListSalesTargetsRequest) ([]models.SalesTarget, int64, error)
	ListAvailableEmployeesByYear(ctx context.Context, year int, includeEmployeeID string) ([]dto.EmployeeResponse, error)
	ExistsByYearAndEmployee(ctx context.Context, year int, employeeID string, excludeID *string) (bool, error)
	Create(ctx context.Context, st *models.SalesTarget) error
	Update(ctx context.Context, st *models.SalesTarget) error
	Delete(ctx context.Context, id string) error
}

type salesTargetRepository struct {
	db *gorm.DB
}

// NewSalesTargetRepository creates a new SalesTargetRepository
func NewSalesTargetRepository(db *gorm.DB) SalesTargetRepository {
	return &salesTargetRepository{db: db}
}

func (r *salesTargetRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *salesTargetRepository) FindByID(ctx context.Context, id string) (*models.SalesTarget, error) {
	var target models.SalesTarget
	err := r.getDB(ctx).
		Preload("Employee").
		Preload("MonthlyTargets").
		Where("id = ?", id).
		First(&target).Error
	if err != nil {
		return nil, err
	}

	targets := []models.SalesTarget{target}
	r.hydrateMonthlySalesActualAmounts(ctx, &targets)
	target = targets[0]
	return &target, nil
}

func (r *salesTargetRepository) List(ctx context.Context, req *dto.ListSalesTargetsRequest) ([]models.SalesTarget, int64, error) {
	var targets []models.SalesTarget
	var total int64

	query := r.db.WithContext(ctx).Model(&models.SalesTarget{})
	var err error
	query, err = applyTenantFilter(ctx, query, "sales_targets.tenant_id")
	if err != nil {
		return nil, 0, err
	}

	// Apply scope-based data filtering (OWN/DIVISION/AREA/ALL)
	// For sales targets, scope is on the employee_id column
	query = security.ApplyScopeFilter(query, ctx, security.SalesTargetScopeQueryOptions())

	// Apply year filter
	if req.Year != nil {
		query = query.Where("sales_targets.year = ?", *req.Year)
	}

	// Apply employee filter
	if req.EmployeeID != "" {
		query = query.Where("sales_targets.employee_id = ?", req.EmployeeID)
	}

	// Apply search filter
	if s := strings.TrimSpace(req.Search); s != "" {
		search := "%" + s + "%"
		query = query.Joins("LEFT JOIN employees ON employees.id = sales_targets.employee_id")
		query = query.Where("employees.name ILIKE ? OR employees.code ILIKE ? OR sales_targets.notes ILIKE ?", search, search, search)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	perPage := req.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage

	// Apply sorting
	allowedSortColumns := map[string]string{
		"year":         "sales_targets.year",
		"employee_id":  "sales_targets.employee_id",
		"total_target": "sales_targets.total_target",
		"created_at":   "sales_targets.created_at",
		"updated_at":   "sales_targets.updated_at",
	}

	sortBy := allowedSortColumns[strings.ToLower(strings.TrimSpace(req.SortBy))]
	if sortBy == "" {
		sortBy = "sales_targets.year"
	}

	isDesc := strings.ToLower(strings.TrimSpace(req.SortDir)) == "desc"
	query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: sortBy}, Desc: isDesc})

	// Execute query with preloads
	err = query.
		Preload("Employee").
		Preload("MonthlyTargets").
		Limit(perPage).
		Offset(offset).
		Find(&targets).Error
	if err != nil {
		return nil, 0, err
	}

	r.hydrateMonthlySalesActualAmounts(ctx, &targets)

	return targets, total, nil
}

func salesTargetEmployeeScopeQueryOptions() security.ScopeQueryOptions {
	return security.ScopeQueryOptions{
		OwnerEmployeeIDColumn: "id",
		DivisionJoinSQL:       "id IN (SELECT id FROM employees WHERE division_id = ? AND deleted_at IS NULL)",
		AreaJoinSQL:           "id IN (SELECT employee_id FROM employee_areas WHERE area_id IN ? AND deleted_at IS NULL)",
		OutletJoinSQL:         "id IN (SELECT employee_id FROM employee_outlets WHERE outlet_id IN ? AND deleted_at IS NULL)",
		WarehouseJoinSQL:      "id IN (SELECT employee_id FROM employee_warehouses WHERE warehouse_id IN ? AND deleted_at IS NULL)",
	}
}

func (r *salesTargetRepository) ListAvailableEmployeesByYear(ctx context.Context, year int, includeEmployeeID string) ([]dto.EmployeeResponse, error) {
	query := r.db.WithContext(ctx).
		Model(&orgModels.Employee{}).
		Select("employees.id, employees.employee_code, employees.name, COALESCE(employees.email, '') AS email, COALESCE(employees.phone, '') AS phone")

	var err error
	query, err = applyTenantFilter(ctx, query, "employees.tenant_id")
	if err != nil {
		return nil, err
	}

	query = security.ApplyScopeFilter(query, ctx, salesTargetEmployeeScopeQueryOptions())
	query = query.Where("employees.deleted_at IS NULL")
	query = query.Where("employees.is_active = ?", true)
	query = query.Joins(
		"LEFT JOIN sales_targets st ON st.employee_id = employees.id AND st.year = ? AND st.deleted_at IS NULL",
		year,
	)

	if strings.TrimSpace(includeEmployeeID) != "" {
		query = query.Where("st.id IS NULL OR employees.id = ?", includeEmployeeID)
	} else {
		query = query.Where("st.id IS NULL")
	}

	query = query.Order("employees.name ASC")

	var employees []dto.EmployeeResponse
	if err := query.Scan(&employees).Error; err != nil {
		return nil, err
	}

	return employees, nil
}

type monthlySalesActualRow struct {
	SalesTargetID string
	Month         int
	ActualAmount  float64
}

// hydrateMonthlySalesActualAmounts aggregates actual revenue from sales_orders per month,
// using the same revenue source as GetMonthlySalesOverview for data consistency.
func (r *salesTargetRepository) hydrateMonthlySalesActualAmounts(ctx context.Context, targets *[]models.SalesTarget) {
	if targets == nil || len(*targets) == 0 {
		return
	}

	targetIDs := make([]string, 0, len(*targets))
	for _, target := range *targets {
		if target.ID != "" {
			targetIDs = append(targetIDs, target.ID)
		}
	}

	if len(targetIDs) == 0 {
		return
	}

	// Single batch query: aggregate orders by (sales_target_id, month) using sales_rep_id + year
	// from the parent sales_target row. Status filter mirrors GetMonthlySalesOverview.
	query := `
		SELECT
			st.id AS sales_target_id,
			EXTRACT(MONTH FROM so.order_date)::int AS month,
			COALESCE(SUM(so.total_amount), 0) AS actual_amount
		FROM sales_targets st
		INNER JOIN sales_orders so
			ON so.sales_rep_id = st.employee_id
			AND so.deleted_at IS NULL
			AND EXTRACT(YEAR FROM so.order_date)::int = st.year
			AND so.status NOT IN ('draft', 'cancelled')
		WHERE st.id IN @targetIDs
		GROUP BY st.id, EXTRACT(MONTH FROM so.order_date)::int
	`

	var rows []monthlySalesActualRow
	if err := r.getDB(ctx).Raw(query, map[string]interface{}{
		"targetIDs": targetIDs,
	}).Scan(&rows).Error; err != nil {
		return
	}

	actualMap := make(map[string]float64, len(rows))
	for _, row := range rows {
		if row.Month < 1 || row.Month > 12 {
			continue
		}
		key := fmt.Sprintf("%s-%d", row.SalesTargetID, row.Month)
		actualMap[key] = row.ActualAmount
	}

	for i := range *targets {
		for j := range (*targets)[i].MonthlyTargets {
			month := (*targets)[i].MonthlyTargets[j].Month
			key := fmt.Sprintf("%s-%d", (*targets)[i].ID, month)
			(*targets)[i].MonthlyTargets[j].ActualAmount = actualMap[key]
			(*targets)[i].MonthlyTargets[j].CalculateSalesAchievement()
		}
	}
}

func (r *salesTargetRepository) ExistsByYearAndEmployee(ctx context.Context, year int, employeeID string, excludeID *string) (bool, error) {
	query := r.getDB(ctx).Model(&models.SalesTarget{}).
		Where("year = ? AND employee_id = ?", year, employeeID)

	if excludeID != nil && *excludeID != "" {
		query = query.Where("id <> ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *salesTargetRepository) Create(ctx context.Context, st *models.SalesTarget) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		months := st.MonthlyTargets
		st.MonthlyTargets = nil

		if err := tx.Create(st).Error; err != nil {
			return err
		}

		if len(months) > 0 {
			for i := range months {
				months[i].SalesTargetID = st.ID
				if err := tx.Create(&months[i]).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (r *salesTargetRepository) Update(ctx context.Context, st *models.SalesTarget) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		// Update target WITHOUT months
		if err := tx.Omit("MonthlyTargets").Save(st).Error; err != nil {
			return err
		}

		// Delete existing months
		if err := tx.Where("sales_target_id = ?", st.ID).Delete(&models.MonthlySalesTarget{}).Error; err != nil {
			return err
		}

		// Create new months
		if len(st.MonthlyTargets) > 0 {
			for i := range st.MonthlyTargets {
				st.MonthlyTargets[i].SalesTargetID = st.ID
				st.MonthlyTargets[i].CreatedAt = apptime.Now()
				st.MonthlyTargets[i].UpdatedAt = apptime.Now()
				if err := tx.Create(&st.MonthlyTargets[i]).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (r *salesTargetRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete months first (cascade handles automatically but explicit is safer)
		if err := tx.Where("sales_target_id = ?", id).Delete(&models.MonthlySalesTarget{}).Error; err != nil {
			return err
		}

		// Delete target
		return tx.Delete(&models.SalesTarget{}, "id = ?", id).Error
	})
}
