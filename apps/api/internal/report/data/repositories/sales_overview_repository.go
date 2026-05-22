package repositories

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"

	"github.com/gilabs/gims/api/internal/core/utils"
	"gorm.io/gorm"
)

// SalesOverviewRepository defines data access for sales overview reports.
// It queries existing tables (employees, sales_orders, delivery_orders, etc.)
// without creating new models — this domain is purely read-only / aggregation.
type SalesOverviewRepository interface {
	// ListSalesRepPerformance returns paginated sales rep performance data
	ListSalesRepPerformance(ctx context.Context, params ListPerformanceParams) ([]SalesRepPerformanceRow, utils.PaginationResult, error)

	// GetMonthlySalesOverview returns monthly aggregated data
	GetMonthlySalesOverview(ctx context.Context, startDate, endDate time.Time) ([]MonthlySalesRow, error)

	// GetSalesRepDetail returns a single sales rep's identity + statistics
	GetSalesRepDetail(ctx context.Context, employeeID string, startDate, endDate time.Time) (*SalesRepDetailRow, error)

	// GetSalesRepCheckInLocations returns paginated visit check-in locations
	GetSalesRepCheckInLocations(ctx context.Context, employeeID string, params CheckInParams) ([]CheckInLocationRow, int, error)

	// GetSalesRepProducts returns paginated product sales for a rep
	GetSalesRepProducts(ctx context.Context, employeeID string, params ProductParams) ([]ProductSalesRow, utils.PaginationResult, error)

	// GetSalesRepCustomers returns paginated customer data for a rep
	GetSalesRepCustomers(ctx context.Context, employeeID string, params CustomerParams) ([]CustomerSalesRow, utils.PaginationResult, error)

	// GetMonthlyTargets returns monthly target amounts within a date range, scoped by caller's RBAC.
	GetMonthlyTargets(ctx context.Context, startDate, endDate time.Time) ([]MonthlyTargetRow, error)

	// GetEmployeeTargetAmounts returns a map of employee_id → aggregated monthly target for the given
	// date range. Single batch query — callers must not loop per-employee to avoid N+1.
	GetEmployeeTargetAmounts(ctx context.Context, employeeIDs []string, startDate, endDate time.Time) (map[string]float64, error)
}

// --- Parameter structs ---

type ListPerformanceParams struct {
	Search    string
	StartDate time.Time
	EndDate   time.Time
	Page      int
	PerPage   int
	SortBy    string
	Order     string
}

type CheckInParams struct {
	StartDate time.Time
	EndDate   time.Time
	Page      int
	PerPage   int
}

type ProductParams struct {
	StartDate time.Time
	EndDate   time.Time
	Page      int
	PerPage   int
	SortBy    string
	Order     string
}

type CustomerParams struct {
	StartDate time.Time
	EndDate   time.Time
	Page      int
	PerPage   int
	SortBy    string
	Order     string
}

// --- Result row structs (raw query results) ---

type SalesRepPerformanceRow struct {
	EmployeeID      string
	EmployeeCode    string
	Name            string
	Email           string
	AvatarURL       string
	PositionName    string
	DivisionName    string
	TotalRevenue    float64
	TotalOrders     int
	TotalDeliveries int
	TotalInvoices   int
	VisitsCompleted int
	TasksCompleted  int
}

type MonthlySalesRow struct {
	Month           int
	Year            int
	TotalRevenue    float64
	TotalCashIn     float64
	TotalOrders     int
	TotalVisits     int
	TotalDeliveries int
}

type MonthlyTargetRow struct {
	Month        int
	Year         int
	TargetAmount float64
}

type SalesRepDetailRow struct {
	EmployeeID   string
	EmployeeCode string
	Name         string
	Email        string
	AvatarURL    string
	PositionName string
	DivisionName string
	// Stats
	TotalRevenue    float64
	TotalOrders     int
	VisitsCompleted int
	TasksCompleted  int
	// Target data
	TargetAmount float64
	// Previous period stats for comparison
	PrevRevenue float64
	PrevOrders  int
	PrevVisits  int
}

type CheckInLocationRow struct {
	VisitID   string
	VisitCode string
	VisitDate time.Time
	CheckInAt *time.Time
	Latitude  *float64
	Longitude *float64
	Address   string
	// Customer info (from sales visit's company)
	CompanyID   *string
	CompanyName string
	Purpose     string
}

type ProductSalesRow struct {
	ProductID    string
	ProductName  string
	ProductSKU   string
	ProductImage string
	CategoryName string
	TotalQty     float64
	TotalRevenue float64
	LastSoldDate *time.Time
}

type CustomerSalesRow struct {
	CustomerID   string
	CustomerName string
	CustomerCode string
	CustomerType string
	CityName     string
	TotalRevenue float64
	TotalOrders  int
	IsActive     bool
}

// --- Implementation ---

type salesOverviewRepository struct {
	db *gorm.DB
}

func NewSalesOverviewRepository(db *gorm.DB) SalesOverviewRepository {
	return &salesOverviewRepository{db: db}
}

func (r *salesOverviewRepository) ListSalesRepPerformance(ctx context.Context, params ListPerformanceParams) ([]SalesRepPerformanceRow, utils.PaginationResult, error) {
	tenantID, tenantFilter, tenantErr := resolveReportTenantScope(ctx)
	if tenantErr != nil {
		return nil, utils.PaginationResult{}, tenantErr
	}

	// Enforce pagination limits
	if params.PerPage <= 0 || params.PerPage > 100 {
		params.PerPage = 20
	}
	if params.Page <= 0 {
		params.Page = 1
	}

	// Base query: include all employees (do not restrict to Sales Representative position)
	baseQuery := `
		FROM employees e
		LEFT JOIN job_positions jp ON jp.id = e.job_position_id
		LEFT JOIN divisions d ON d.id = e.division_id
		LEFT JOIN users u ON u.id = e.user_id
		LEFT JOIN (
			SELECT so.sales_rep_id,
				COALESCE(SUM(so.total_amount), 0) AS total_revenue,
				COUNT(so.id) AS total_orders
			FROM sales_orders so
			WHERE so.deleted_at IS NULL
				AND (@tenantFilter = false OR so.tenant_id = @tenantID)
				AND so.status NOT IN ('draft', 'cancelled')
				AND so.order_date BETWEEN @startDate AND @endDate
			GROUP BY so.sales_rep_id
		) so_agg ON so_agg.sales_rep_id = e.id
		LEFT JOIN (
			SELECT so2.sales_rep_id,
				COUNT(dord.id) AS total_deliveries
			FROM delivery_orders dord
			INNER JOIN sales_orders so2 ON so2.id = dord.sales_order_id AND so2.deleted_at IS NULL AND (@tenantFilter = false OR so2.tenant_id = @tenantID)
			WHERE dord.deleted_at IS NULL
				AND dord.status IN ('delivered', 'shipped')
				AND dord.delivery_date BETWEEN @startDate AND @endDate
			GROUP BY so2.sales_rep_id
		) do_agg ON do_agg.sales_rep_id = e.id
		LEFT JOIN (
			SELECT so3.sales_rep_id,
				COUNT(ci.id) AS total_invoices
			FROM customer_invoices ci
			INNER JOIN sales_orders so3 ON so3.id = ci.sales_order_id AND so3.deleted_at IS NULL AND (@tenantFilter = false OR so3.tenant_id = @tenantID)
			WHERE ci.deleted_at IS NULL
				AND ci.status NOT IN ('draft', 'cancelled')
				AND ci.invoice_date BETWEEN @startDate AND @endDate
			GROUP BY so3.sales_rep_id
		) ci_agg ON ci_agg.sales_rep_id = e.id
		LEFT JOIN (
			SELECT sv.employee_id,
				COUNT(CASE WHEN sv.status = 'completed' THEN 1 END) AS visits_completed,
				COUNT(sv.id) AS tasks_completed
			FROM sales_visits sv
			WHERE sv.deleted_at IS NULL
				AND (@tenantFilter = false OR sv.tenant_id = @tenantID)
				AND sv.visit_date BETWEEN @startDate AND @endDate
			GROUP BY sv.employee_id
		) sv_agg ON sv_agg.employee_id = e.id
		WHERE e.deleted_at IS NULL
			AND (@tenantFilter = false OR e.tenant_id = @tenantID)
			AND e.is_active = true
	`

	queryParams := withReportTenantParams(map[string]interface{}{
		"startDate": params.StartDate,
		"endDate":   params.EndDate,
	}, tenantID, tenantFilter)

	// Add search filter (prefix search for index efficiency)
	if params.Search != "" {
		baseQuery += ` AND (e.name ILIKE @search OR e.employee_code ILIKE @search)`
		queryParams["search"] = params.Search + "%"
	}

	// Apply RBAC scope to restrict which employees are visible
	scopeSQL, scopeParams := buildScopeConditionRaw(ctx, "e.id")
	if scopeSQL != "" {
		baseQuery += "\n\t\t" + scopeSQL
		for k, v := range scopeParams {
			queryParams[k] = v
		}
	}

	// Count total
	var total int64
	countSQL := "SELECT COUNT(*) " + baseQuery
	if err := database.GetDB(ctx, r.db).Raw(countSQL, queryParams).Scan(&total).Error; err != nil {
		return nil, utils.PaginationResult{}, fmt.Errorf("failed to count sales rep performance: %w", err)
	}

	// Sort mapping
	sortColumn := "COALESCE(so_agg.total_revenue, 0)"
	switch params.SortBy {
	case "name":
		sortColumn = "e.name"
	case "orders":
		sortColumn = "COALESCE(so_agg.total_orders, 0)"
	case "visits":
		sortColumn = "COALESCE(sv_agg.visits_completed, 0)"
	case "revenue":
		sortColumn = "COALESCE(so_agg.total_revenue, 0)"
	}
	sortOrder := "DESC"
	if strings.EqualFold(params.Order, "asc") {
		sortOrder = "ASC"
	}

	offset := (params.Page - 1) * params.PerPage

	selectSQL := fmt.Sprintf(`
		SELECT
			e.id AS employee_id,
			e.employee_code,
			e.name,
			COALESCE(e.email, '') AS email,
			COALESCE(u.avatar_url, '') AS avatar_url,
			COALESCE(jp.name, '') AS position_name,
			COALESCE(d.name, '') AS division_name,
			COALESCE(so_agg.total_revenue, 0) AS total_revenue,
			COALESCE(so_agg.total_orders, 0) AS total_orders,
			COALESCE(do_agg.total_deliveries, 0) AS total_deliveries,
			COALESCE(ci_agg.total_invoices, 0) AS total_invoices,
			COALESCE(sv_agg.visits_completed, 0) AS visits_completed,
			COALESCE(sv_agg.tasks_completed, 0) AS tasks_completed
		%s
		ORDER BY %s %s, e.name ASC
		LIMIT @limit OFFSET @offset
	`, baseQuery, sortColumn, sortOrder)

	queryParams["limit"] = params.PerPage
	queryParams["offset"] = offset

	var rows []SalesRepPerformanceRow
	if err := database.GetDB(ctx, r.db).Raw(selectSQL, queryParams).Scan(&rows).Error; err != nil {
		return nil, utils.PaginationResult{}, fmt.Errorf("failed to list sales rep performance: %w", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.PerPage)))
	pagination := utils.PaginationResult{
		Page:       params.Page,
		PerPage:    params.PerPage,
		Total:      int(total),
		TotalPages: totalPages,
	}

	return rows, pagination, nil
}

func (r *salesOverviewRepository) GetMonthlySalesOverview(ctx context.Context, startDate, endDate time.Time) ([]MonthlySalesRow, error) {
	tenantID, tenantFilter, tenantErr := resolveReportTenantScope(ctx)
	if tenantErr != nil {
		return nil, tenantErr
	}

	// Build RBAC scope conditions; both share the same named params.
	orderScopeSQL, scopeParams := buildScopeConditionRaw(ctx, "so.sales_rep_id")
	visitScopeSQL, _ := buildScopeConditionRaw(ctx, "sv.employee_id")

	query := fmt.Sprintf(`
		SELECT
			EXTRACT(MONTH FROM so.order_date)::int AS month,
			EXTRACT(YEAR FROM so.order_date)::int AS year,
			COALESCE(SUM(so.total_amount), 0) AS total_revenue,
			COALESCE(payment_counts.total_cash_in, 0) AS total_cash_in,
			COUNT(so.id) AS total_orders,
			COALESCE(visit_counts.total_visits, 0) AS total_visits,
			COALESCE(delivery_counts.total_deliveries, 0) AS total_deliveries
		FROM sales_orders so
		LEFT JOIN (
			SELECT
				EXTRACT(MONTH FROM sv.visit_date)::int AS month,
				EXTRACT(YEAR FROM sv.visit_date)::int AS year,
				COUNT(CASE WHEN sv.status = 'completed' THEN 1 END) AS total_visits
			FROM sales_visits sv
			WHERE sv.deleted_at IS NULL
				AND (@tenantFilter = false OR sv.tenant_id = @tenantID)
				AND sv.visit_date BETWEEN @startDate AND @endDate
				%s
			GROUP BY EXTRACT(MONTH FROM sv.visit_date), EXTRACT(YEAR FROM sv.visit_date)
		) visit_counts ON visit_counts.month = EXTRACT(MONTH FROM so.order_date) AND visit_counts.year = EXTRACT(YEAR FROM so.order_date)
		LEFT JOIN (
			SELECT
				EXTRACT(MONTH FROM sp.payment_date::date)::int AS month,
				EXTRACT(YEAR FROM sp.payment_date::date)::int AS year,
				COALESCE(SUM(sp.amount), 0) AS total_cash_in
			FROM sales_payments sp
			WHERE sp.deleted_at IS NULL
				AND (@tenantFilter = false OR sp.tenant_id = @tenantID)
				AND sp.status = 'CONFIRMED'
				AND sp.payment_date::date BETWEEN @startDate AND @endDate
			GROUP BY EXTRACT(MONTH FROM sp.payment_date::date), EXTRACT(YEAR FROM sp.payment_date::date)
		) payment_counts ON payment_counts.month = EXTRACT(MONTH FROM so.order_date) AND payment_counts.year = EXTRACT(YEAR FROM so.order_date)
		LEFT JOIN (
			SELECT
				EXTRACT(MONTH FROM dord.delivery_date)::int AS month,
				EXTRACT(YEAR FROM dord.delivery_date)::int AS year,
				COUNT(dord.id) AS total_deliveries
			FROM delivery_orders dord
			WHERE dord.deleted_at IS NULL
				AND (@tenantFilter = false OR dord.tenant_id = @tenantID)
				AND dord.status IN ('delivered', 'shipped')
				AND dord.delivery_date BETWEEN @startDate AND @endDate
			GROUP BY EXTRACT(MONTH FROM dord.delivery_date), EXTRACT(YEAR FROM dord.delivery_date)
		) delivery_counts ON delivery_counts.month = EXTRACT(MONTH FROM so.order_date) AND delivery_counts.year = EXTRACT(YEAR FROM so.order_date)
		WHERE so.deleted_at IS NULL
			AND (@tenantFilter = false OR so.tenant_id = @tenantID)
			AND so.status NOT IN ('draft', 'cancelled')
			AND so.order_date BETWEEN @startDate AND @endDate
			%s
		GROUP BY EXTRACT(MONTH FROM so.order_date), EXTRACT(YEAR FROM so.order_date),
			payment_counts.total_cash_in, visit_counts.total_visits, delivery_counts.total_deliveries
		ORDER BY year, month
	`, visitScopeSQL, orderScopeSQL)

	qp := withReportTenantParams(map[string]interface{}{
		"startDate": startDate,
		"endDate":   endDate,
	}, tenantID, tenantFilter)
	for k, v := range scopeParams {
		qp[k] = v
	}

	var rows []MonthlySalesRow
	if err := database.GetDB(ctx, r.db).Raw(query, qp).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to get monthly sales overview: %w", err)
	}

	return rows, nil
}

func (r *salesOverviewRepository) GetMonthlyTargets(ctx context.Context, startDate, endDate time.Time) ([]MonthlyTargetRow, error) {
	tenantID, tenantFilter, tenantErr := resolveReportTenantScope(ctx)
	if tenantErr != nil {
		return nil, tenantErr
	}
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	scopeSQL, scopeParams := buildScopeConditionRaw(ctx, "st.employee_id")

	// Aggregate monthly targets from the annual sales target rows.
	// If a monthly row is missing or zero, fall back to total_target/12 so the
	// report still shows a non-zero target when the annual target exists.
	query := fmt.Sprintf(`
		WITH months AS (
			SELECT generate_series(1, 12)::int AS month
		)
		SELECT
			months.month,
			st.year,
			COALESCE(
				SUM(COALESCE(NULLIF(mst.target_amount, 0), st.total_target / 12.0)),
				0
			) AS target_amount
		FROM sales_targets st
		CROSS JOIN months
		LEFT JOIN monthly_sales_targets mst ON mst.sales_target_id = st.id
			AND mst.month = months.month
			AND mst.deleted_at IS NULL
		WHERE st.deleted_at IS NULL
			AND (@tenantFilter = false OR st.tenant_id = @tenantID)
			AND st.year BETWEEN EXTRACT(YEAR FROM '%s'::date)::int AND EXTRACT(YEAR FROM '%s'::date)::int
			AND MAKE_DATE(st.year::int, months.month::int, 1) <= '%s'::date
			AND (MAKE_DATE(st.year::int, months.month::int, 1) + INTERVAL '1 month' - INTERVAL '1 day') >= '%s'::date
			%s
		GROUP BY months.month, st.year
		ORDER BY st.year, months.month
	`, startDateStr, endDateStr, endDateStr, startDateStr, scopeSQL)

	qp := withReportTenantParams(map[string]interface{}{
	}, tenantID, tenantFilter)
	for k, v := range scopeParams {
		qp[k] = v
	}

	var rows []MonthlyTargetRow
	if err := database.GetDB(ctx, r.db).Raw(query, qp).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to get monthly targets: %w", err)
	}
	return rows, nil
}

func (r *salesOverviewRepository) GetSalesRepDetail(ctx context.Context, employeeID string, startDate, endDate time.Time) (*SalesRepDetailRow, error) {
	tenantID, tenantFilter, tenantErr := resolveReportTenantScope(ctx)
	if tenantErr != nil {
		return nil, tenantErr
	}

	// Calculate previous period (same duration, shifted back)
	duration := endDate.Sub(startDate)
	prevEnd := startDate.Add(-1 * 24 * time.Hour) // day before start
	prevStart := prevEnd.Add(-duration)

	query := `
		SELECT
			e.id AS employee_id,
			e.employee_code,
			e.name,
			COALESCE(e.email, '') AS email,
			COALESCE(u.avatar_url, '') AS avatar_url,
			COALESCE(jp.name, '') AS position_name,
			COALESCE(d.name, '') AS division_name,
			COALESCE(curr_orders.total_revenue, 0) AS total_revenue,
			COALESCE(curr_orders.total_orders, 0) AS total_orders,
			COALESCE(curr_visits.visits_completed, 0) AS visits_completed,
			COALESCE(curr_visits.tasks_completed, 0) AS tasks_completed,
			COALESCE(target_agg.target_amount, 0) AS target_amount,
			COALESCE(prev_orders.total_revenue, 0) AS prev_revenue,
			COALESCE(prev_orders.total_orders, 0) AS prev_orders,
			COALESCE(prev_visits.visits_completed, 0) AS prev_visits
		FROM employees e
		LEFT JOIN job_positions jp ON jp.id = e.job_position_id
		LEFT JOIN divisions d ON d.id = e.division_id
		LEFT JOIN users u ON u.id = e.user_id
		LEFT JOIN (
			SELECT so.sales_rep_id,
				SUM(so.total_amount) AS total_revenue,
				COUNT(so.id) AS total_orders
			FROM sales_orders so
			WHERE so.deleted_at IS NULL
				AND (@tenantFilter = false OR so.tenant_id = @tenantID)
				AND so.status NOT IN ('draft', 'cancelled')
				AND so.order_date BETWEEN @startDate AND @endDate
			GROUP BY so.sales_rep_id
		) curr_orders ON curr_orders.sales_rep_id = e.id
		LEFT JOIN (
			SELECT sv.employee_id,
				COUNT(CASE WHEN sv.status = 'completed' THEN 1 END) AS visits_completed,
				COUNT(sv.id) AS tasks_completed
			FROM sales_visits sv
			WHERE sv.deleted_at IS NULL
				AND (@tenantFilter = false OR sv.tenant_id = @tenantID)
				AND sv.visit_date BETWEEN @startDate AND @endDate
			GROUP BY sv.employee_id
		) curr_visits ON curr_visits.employee_id = e.id
		LEFT JOIN (
			SELECT st.employee_id,
				COALESCE(SUM(COALESCE(NULLIF(mst.target_amount, 0), st.total_target / 12.0)), 0) AS target_amount
			FROM sales_targets st
			LEFT JOIN monthly_sales_targets mst ON mst.sales_target_id = st.id
				AND mst.deleted_at IS NULL
			WHERE st.deleted_at IS NULL
				AND (@tenantFilter = false OR st.tenant_id = @tenantID)
				AND st.year = @targetYear
			GROUP BY st.employee_id
		) target_agg ON target_agg.employee_id = e.id
		LEFT JOIN (
			SELECT so.sales_rep_id,
				SUM(so.total_amount) AS total_revenue,
				COUNT(so.id) AS total_orders
			FROM sales_orders so
			WHERE so.deleted_at IS NULL
				AND (@tenantFilter = false OR so.tenant_id = @tenantID)
				AND so.status NOT IN ('draft', 'cancelled')
				AND so.order_date BETWEEN @prevStart AND @prevEnd
			GROUP BY so.sales_rep_id
		) prev_orders ON prev_orders.sales_rep_id = e.id
		LEFT JOIN (
			SELECT sv.employee_id,
				COUNT(CASE WHEN sv.status = 'completed' THEN 1 END) AS visits_completed
			FROM sales_visits sv
			WHERE sv.deleted_at IS NULL
				AND (@tenantFilter = false OR sv.tenant_id = @tenantID)
				AND sv.visit_date BETWEEN @prevStart AND @prevEnd
			GROUP BY sv.employee_id
		) prev_visits ON prev_visits.employee_id = e.id
		WHERE e.id = @employeeID
			AND e.deleted_at IS NULL
			AND (@tenantFilter = false OR e.tenant_id = @tenantID)
	`

	var row SalesRepDetailRow
	if err := database.GetDB(ctx, r.db).Raw(query, withReportTenantParams(map[string]interface{}{
		"employeeID":  employeeID,
		"startDate":   startDate,
		"endDate":     endDate,
		"prevStart":   prevStart,
		"prevEnd":     prevEnd,
		"targetYear":  startDate.Year(),
	}, tenantID, tenantFilter)).Scan(&row).Error; err != nil {
		return nil, fmt.Errorf("failed to get sales rep detail: %w", err)
	}

	if row.EmployeeID == "" {
		return nil, nil
	}

	return &row, nil
}

func (r *salesOverviewRepository) GetSalesRepCheckInLocations(ctx context.Context, employeeID string, params CheckInParams) ([]CheckInLocationRow, int, error) {
	tenantID, tenantFilter, tenantErr := resolveReportTenantScope(ctx)
	if tenantErr != nil {
		return nil, 0, tenantErr
	}

	if params.PerPage <= 0 || params.PerPage > 100 {
		params.PerPage = 20
	}
	if params.Page <= 0 {
		params.Page = 1
	}

	baseWhere := `
		WHERE sv.employee_id = @employeeID
			AND sv.deleted_at IS NULL
			AND (@tenantFilter = false OR sv.tenant_id = @tenantID)
			AND sv.check_in_at IS NOT NULL
			AND sv.visit_date BETWEEN @startDate AND @endDate
	`

	queryParams := withReportTenantParams(map[string]interface{}{
		"employeeID": employeeID,
		"startDate":  params.StartDate,
		"endDate":    params.EndDate,
	}, tenantID, tenantFilter)

	// Count total
	var total int64
	countSQL := `SELECT COUNT(*) FROM sales_visits sv ` + baseWhere
	if err := database.GetDB(ctx, r.db).Raw(countSQL, queryParams).Scan(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count check-in locations: %w", err)
	}

	offset := (params.Page - 1) * params.PerPage

	selectSQL := `
		SELECT
			sv.id AS visit_id,
			sv.code AS visit_code,
			sv.visit_date,
			sv.check_in_at,
			sv.latitude,
			sv.longitude,
			COALESCE(sv.address, '') AS address,
			sv.company_id,
			COALESCE(c.name, '') AS company_name,
			COALESCE(sv.purpose, '') AS purpose
		FROM sales_visits sv
		LEFT JOIN companies c ON c.id = sv.company_id
	` + baseWhere + `
		ORDER BY sv.visit_date DESC, sv.check_in_at DESC
		LIMIT @limit OFFSET @offset
	`
	queryParams["limit"] = params.PerPage
	queryParams["offset"] = offset

	var rows []CheckInLocationRow
	if err := database.GetDB(ctx, r.db).Raw(selectSQL, queryParams).Scan(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get check-in locations: %w", err)
	}

	return rows, int(total), nil
}

func (r *salesOverviewRepository) GetSalesRepProducts(ctx context.Context, employeeID string, params ProductParams) ([]ProductSalesRow, utils.PaginationResult, error) {
	tenantID, tenantFilter, tenantErr := resolveReportTenantScope(ctx)
	if tenantErr != nil {
		return nil, utils.PaginationResult{}, tenantErr
	}

	if params.PerPage <= 0 || params.PerPage > 100 {
		params.PerPage = 20
	}
	if params.Page <= 0 {
		params.Page = 1
	}

	baseQuery := `
		FROM sales_order_items soi
		INNER JOIN sales_orders so ON so.id = soi.sales_order_id
			AND so.deleted_at IS NULL
			AND (@tenantFilter = false OR so.tenant_id = @tenantID)
			AND so.status NOT IN ('draft', 'cancelled')
			AND so.sales_rep_id = @employeeID
			AND so.order_date BETWEEN @startDate AND @endDate
		INNER JOIN products p ON p.id = soi.product_id AND p.deleted_at IS NULL AND (@tenantFilter = false OR p.tenant_id = @tenantID)
		LEFT JOIN product_categories pc ON pc.id = p.category_id
		WHERE soi.deleted_at IS NULL
	`

	queryParams := withReportTenantParams(map[string]interface{}{
		"employeeID": employeeID,
		"startDate":  params.StartDate,
		"endDate":    params.EndDate,
	}, tenantID, tenantFilter)

	// Count distinct products
	var total int64
	countSQL := `SELECT COUNT(DISTINCT soi.product_id) ` + baseQuery
	if err := database.GetDB(ctx, r.db).Raw(countSQL, queryParams).Scan(&total).Error; err != nil {
		return nil, utils.PaginationResult{}, fmt.Errorf("failed to count products: %w", err)
	}

	// Sort mapping
	sortColumn := "SUM(soi.subtotal)"
	switch params.SortBy {
	case "total_quantity":
		sortColumn = "SUM(soi.quantity)"
	case "name":
		sortColumn = "p.name"
	case "revenue":
		sortColumn = "SUM(soi.subtotal)"
	}
	sortOrder := "DESC"
	if strings.EqualFold(params.Order, "asc") {
		sortOrder = "ASC"
	}

	offset := (params.Page - 1) * params.PerPage

	selectSQL := fmt.Sprintf(`
		SELECT
			soi.product_id,
			p.name AS product_name,
			COALESCE(p.sku, '') AS product_sku,
			COALESCE(p.image_url, '') AS product_image,
			COALESCE(pc.name, '') AS category_name,
			SUM(soi.quantity) AS total_qty,
			SUM(soi.subtotal) AS total_revenue,
			MAX(so.order_date) AS last_sold_date
		%s
		GROUP BY soi.product_id, p.name, p.sku, p.image_url, pc.name
		ORDER BY %s %s
		LIMIT @limit OFFSET @offset
	`, baseQuery, sortColumn, sortOrder)

	queryParams["limit"] = params.PerPage
	queryParams["offset"] = offset

	var rows []ProductSalesRow
	if err := database.GetDB(ctx, r.db).Raw(selectSQL, queryParams).Scan(&rows).Error; err != nil {
		return nil, utils.PaginationResult{}, fmt.Errorf("failed to get sales rep products: %w", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.PerPage)))
	pagination := utils.PaginationResult{
		Page:       params.Page,
		PerPage:    params.PerPage,
		Total:      int(total),
		TotalPages: totalPages,
	}

	return rows, pagination, nil
}

func (r *salesOverviewRepository) GetSalesRepCustomers(ctx context.Context, employeeID string, params CustomerParams) ([]CustomerSalesRow, utils.PaginationResult, error) {
	tenantID, tenantFilter, tenantErr := resolveReportTenantScope(ctx)
	if tenantErr != nil {
		return nil, utils.PaginationResult{}, tenantErr
	}

	if params.PerPage <= 0 || params.PerPage > 100 {
		params.PerPage = 20
	}
	if params.Page <= 0 {
		params.Page = 1
	}

	baseQuery := `
		FROM sales_orders so
		INNER JOIN customers cust ON cust.id = so.customer_id AND cust.deleted_at IS NULL AND (@tenantFilter = false OR cust.tenant_id = @tenantID)
		LEFT JOIN customer_types ct ON ct.id = cust.customer_type_id
		LEFT JOIN cities city ON city.id = cust.city_id
		WHERE so.deleted_at IS NULL
			AND (@tenantFilter = false OR so.tenant_id = @tenantID)
			AND so.status NOT IN ('draft', 'cancelled')
			AND so.sales_rep_id = @employeeID
			AND so.order_date BETWEEN @startDate AND @endDate
	`

	queryParams := withReportTenantParams(map[string]interface{}{
		"employeeID": employeeID,
		"startDate":  params.StartDate,
		"endDate":    params.EndDate,
	}, tenantID, tenantFilter)

	// Count distinct customers
	var total int64
	countSQL := `SELECT COUNT(DISTINCT so.customer_id) ` + baseQuery
	if err := database.GetDB(ctx, r.db).Raw(countSQL, queryParams).Scan(&total).Error; err != nil {
		return nil, utils.PaginationResult{}, fmt.Errorf("failed to count customers: %w", err)
	}

	sortColumn := "SUM(so.total_amount)"
	switch params.SortBy {
	case "orders":
		sortColumn = "COUNT(so.id)"
	case "name":
		sortColumn = "cust.name"
	case "revenue":
		sortColumn = "SUM(so.total_amount)"
	}
	sortOrder := "DESC"
	if strings.EqualFold(params.Order, "asc") {
		sortOrder = "ASC"
	}

	offset := (params.Page - 1) * params.PerPage

	selectSQL := fmt.Sprintf(`
		SELECT
			cust.id AS customer_id,
			cust.name AS customer_name,
			COALESCE(cust.code, '') AS customer_code,
			COALESCE(ct.name, '') AS customer_type,
			COALESCE(city.name, '') AS city_name,
			SUM(so.total_amount) AS total_revenue,
			COUNT(so.id) AS total_orders,
			cust.is_active
		%s
		GROUP BY cust.id, cust.name, cust.code, ct.name, city.name, cust.is_active
		ORDER BY %s %s
		LIMIT @limit OFFSET @offset
	`, baseQuery, sortColumn, sortOrder)

	queryParams["limit"] = params.PerPage
	queryParams["offset"] = offset

	var rows []CustomerSalesRow
	if err := database.GetDB(ctx, r.db).Raw(selectSQL, queryParams).Scan(&rows).Error; err != nil {
		return nil, utils.PaginationResult{}, fmt.Errorf("failed to get sales rep customers: %w", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.PerPage)))
	pagination := utils.PaginationResult{
		Page:       params.Page,
		PerPage:    params.PerPage,
		Total:      int(total),
		TotalPages: totalPages,
	}

	return rows, pagination, nil
}

type employeeTargetAmountRow struct {
	EmployeeID   string
	TargetAmount float64
}

// GetEmployeeTargetAmounts fetches the sum of monthly targets for each given employee
// within the date range in a single batch query. Returns a map of employee_id → total target.
func (r *salesOverviewRepository) GetEmployeeTargetAmounts(ctx context.Context, employeeIDs []string, startDate, endDate time.Time) (map[string]float64, error) {
	if len(employeeIDs) == 0 {
		return map[string]float64{}, nil
	}

	tenantID, tenantFilter, tenantErr := resolveReportTenantScope(ctx)
	if tenantErr != nil {
		return nil, tenantErr
	}

	var rows []employeeTargetAmountRow
	placeholders := make([]string, len(employeeIDs))
	params := map[string]interface{}{
		"tenantID":     tenantID,
		"tenantFilter": tenantFilter,
		"startYear":    startDate.Year(),
		"endYear":      endDate.Year(),
	}
	for i, employeeID := range employeeIDs {
		placeholder := fmt.Sprintf("@employeeID%d", i)
		placeholders[i] = placeholder
		params[fmt.Sprintf("employeeID%d", i)] = employeeID
	}

	query := fmt.Sprintf(`
		SELECT
			st.employee_id,
			COALESCE(SUM(st.total_target), 0) AS target_amount
		FROM sales_targets st
		WHERE st.deleted_at IS NULL
			AND (@tenantFilter = false OR st.tenant_id = @tenantID)
			AND st.employee_id IN (%s)
			AND st.year BETWEEN @startYear AND @endYear
		GROUP BY st.employee_id
	`, strings.Join(placeholders, ", "))

	if err := database.GetDB(ctx, r.db).Raw(query, params).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to get employee target amounts: %w", err)
	}

	result := make(map[string]float64, len(rows))
	for _, row := range rows {
		result[row.EmployeeID] = row.TargetAmount
	}
	return result, nil
}

// buildScopeConditionRaw extracts the RBAC scope from the request context and returns
// a raw SQL fragment (using @-named params) and the corresponding param map to restrict
// query results by the caller's data visibility scope.
// employeeIDColumn is the SQL column reference that holds the employee ID
// (e.g. "e.id", "so.sales_rep_id", "sv.employee_id").
// Returns ("", nil) when scope is ALL or unset — no extra filtering applied.
func buildScopeConditionRaw(ctx context.Context, employeeIDColumn string) (string, map[string]interface{}) {
	scope, _ := ctx.Value("permission_scope").(string)

	switch scope {
	case "OWN":
		empID, _ := ctx.Value("scope_employee_id").(string)
		if empID == "" {
			return "AND 1 = 0", nil // no employee resolved — block all rows for safety
		}
		return "AND " + employeeIDColumn + " = @scopeEmployeeID",
			map[string]interface{}{"scopeEmployeeID": empID}

	case "DIVISION":
		divID, _ := ctx.Value("scope_division_id").(string)
		if divID == "" {
			// No division assigned — fall back to OWN scope
			empID, _ := ctx.Value("scope_employee_id").(string)
			return "AND " + employeeIDColumn + " = @scopeEmployeeID",
				map[string]interface{}{"scopeEmployeeID": empID}
		}
		return "AND " + employeeIDColumn + " IN (SELECT id FROM employees WHERE division_id = @scopeDivisionID AND deleted_at IS NULL)",
			map[string]interface{}{"scopeDivisionID": divID}

	case "AREA":
		areaIDs, _ := ctx.Value("scope_area_ids").([]string)
		if len(areaIDs) == 0 {
			empID, _ := ctx.Value("scope_employee_id").(string)
			return "AND " + employeeIDColumn + " = @scopeEmployeeID",
				map[string]interface{}{"scopeEmployeeID": empID}
		}
		return "AND " + employeeIDColumn + " IN (SELECT employee_id FROM employee_areas WHERE area_id IN @scopeAreaIDs AND deleted_at IS NULL)",
			map[string]interface{}{"scopeAreaIDs": areaIDs}

	case "OUTLET":
		outletIDs, _ := ctx.Value("scope_outlet_ids").([]string)
		if len(outletIDs) == 0 {
			empID, _ := ctx.Value("scope_employee_id").(string)
			return "AND " + employeeIDColumn + " = @scopeEmployeeID",
				map[string]interface{}{"scopeEmployeeID": empID}
		}
		return "AND " + employeeIDColumn + " IN (SELECT employee_id FROM employee_outlets WHERE outlet_id IN @scopeOutletIDs AND deleted_at IS NULL)",
			map[string]interface{}{"scopeOutletIDs": outletIDs}

	default: // ALL or empty — no restriction
		return "", nil
	}
}
