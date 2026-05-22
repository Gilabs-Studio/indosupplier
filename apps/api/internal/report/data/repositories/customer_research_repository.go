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

// CustomerResearchRepository defines data access for customer research reports.
type CustomerResearchRepository interface {
	GetKPIs(ctx context.Context, startDate, endDate time.Time) (CustomerResearchKPIsRow, error)
	ListCustomers(ctx context.Context, params ListCustomersParams) ([]CustomerResearchRow, utils.PaginationResult, error)
	GetRevenueByCustomer(ctx context.Context, params ListCustomersParams) ([]CustomerResearchRow, utils.PaginationResult, error)
	GetPurchaseFrequency(ctx context.Context, params ListCustomersParams) ([]CustomerResearchRow, utils.PaginationResult, error)
	GetRevenueTrend(ctx context.Context, startDate, endDate time.Time, interval string) ([]CustomerRevenueTrendRow, error)
	GetCustomerDetail(ctx context.Context, customerID string, startDate, endDate time.Time) (*CustomerDetailRow, error)
	GetCustomerTopProducts(ctx context.Context, customerID string, startDate, endDate time.Time, limit int) ([]CustomerProductRow, error)
}

// CustomerResearchKPIsRow is raw KPI query result.
type CustomerResearchKPIsRow struct {
	TotalCustomers  int     `gorm:"column:total_customers"`
	ActiveCustomers int     `gorm:"column:active_customers"`
	TotalRevenue    float64 `gorm:"column:total_revenue"`
	TotalOrders     int     `gorm:"column:total_orders"`
}

// ListCustomersParams defines filters for listing customers.
type ListCustomersParams struct {
	StartDate time.Time
	EndDate   time.Time
	Tab       string
	Search    string
	Page      int
	PerPage   int
	SortBy    string
	Order     string
}

// CustomerResearchRow is raw customer row query result.
type CustomerResearchRow struct {
	CustomerID            string  `gorm:"column:customer_id"`
	CustomerName          string  `gorm:"column:customer_name"`
	TotalRevenue          float64 `gorm:"column:total_revenue"`
	TotalOrders           int     `gorm:"column:total_orders"`
	AverageOrderValue     float64 `gorm:"column:average_order_value"`
	LastOrderDate         string  `gorm:"column:last_order_date"`
	ActiveSalesOrderCount int     `gorm:"column:active_sales_order_count"`
}

// CustomerRevenueTrendRow is raw trend query result.
type CustomerRevenueTrendRow struct {
	Period       string  `gorm:"column:period"`
	TotalRevenue float64 `gorm:"column:total_revenue"`
	TotalOrders  int     `gorm:"column:total_orders"`
}

// CustomerDetailRow is raw customer detail query result.
type CustomerDetailRow struct {
	CustomerID        string  `gorm:"column:customer_id"`
	CustomerName      string  `gorm:"column:customer_name"`
	TotalRevenue      float64 `gorm:"column:total_revenue"`
	TotalOrders       int     `gorm:"column:total_orders"`
	AverageOrderValue float64 `gorm:"column:average_order_value"`
	LastOrderDate     string  `gorm:"column:last_order_date"`
}

// CustomerProductRow is raw product row per customer.
type CustomerProductRow struct {
	ProductID    string  `gorm:"column:product_id"`
	ProductCode  string  `gorm:"column:product_code"`
	ProductName  string  `gorm:"column:product_name"`
	TotalQty     float64 `gorm:"column:total_qty"`
	TotalRevenue float64 `gorm:"column:total_revenue"`
	TotalOrders  int     `gorm:"column:total_orders"`
}

type customerResearchRepository struct {
	db *gorm.DB
}

// NewCustomerResearchRepository creates a new CustomerResearchRepository instance.
func NewCustomerResearchRepository(db *gorm.DB) CustomerResearchRepository {
	return &customerResearchRepository{db: db}
}

func (r *customerResearchRepository) GetKPIs(ctx context.Context, startDate, endDate time.Time) (CustomerResearchKPIsRow, error) {
	tenantID, tenantFilter, tenantErr := resolveReportTenantScope(ctx)
	if tenantErr != nil {
		return CustomerResearchKPIsRow{}, tenantErr
	}

	query := `
		SELECT
			COALESCE(COUNT(DISTINCT c.id), 0) AS total_customers,
			COALESCE(COUNT(DISTINCT so.customer_id), 0) AS active_customers,
			COALESCE(SUM(soi.subtotal), 0) AS total_revenue,
			COALESCE(COUNT(DISTINCT so.id), 0) AS total_orders
		FROM customers c
		LEFT JOIN sales_orders so ON so.customer_id = c.id
			AND so.deleted_at IS NULL
			AND (@tenantFilter = false OR so.tenant_id = @tenantID)
			AND so.status NOT IN ('draft', 'cancelled')
			AND so.order_date BETWEEN @startDate AND @endDate
		LEFT JOIN sales_order_items soi ON soi.sales_order_id = so.id
			AND soi.deleted_at IS NULL
		WHERE c.deleted_at IS NULL
			AND (@tenantFilter = false OR c.tenant_id = @tenantID)
	`

	var row CustomerResearchKPIsRow
	if err := database.GetDB(ctx, r.db).Raw(query, withReportTenantParams(map[string]interface{}{
		"startDate": startDate,
		"endDate":   endDate,
	}, tenantID, tenantFilter)).Scan(&row).Error; err != nil {
		return CustomerResearchKPIsRow{}, fmt.Errorf("failed to get customer research KPIs: %w", err)
	}

	if math.IsNaN(row.TotalRevenue) {
		row.TotalRevenue = 0
	}

	return row, nil
}

func (r *customerResearchRepository) ListCustomers(ctx context.Context, params ListCustomersParams) ([]CustomerResearchRow, utils.PaginationResult, error) {
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
		FROM customers c
		LEFT JOIN (
			SELECT
				so.customer_id,
				COALESCE(SUM(soi.subtotal), 0) AS total_revenue,
				COUNT(DISTINCT so.id) AS total_orders,
				CASE WHEN MAX(so.order_date) IS NULL THEN '' ELSE TO_CHAR(MAX(so.order_date), 'YYYY-MM-DD') END AS last_order_date
			FROM sales_orders so
			INNER JOIN sales_order_items soi ON soi.sales_order_id = so.id AND soi.deleted_at IS NULL
			WHERE so.deleted_at IS NULL
				AND (@tenantFilter = false OR so.tenant_id = @tenantID)
				AND so.status NOT IN ('draft', 'cancelled')
				AND so.order_date BETWEEN @startDate AND @endDate
			GROUP BY so.customer_id
		) o ON o.customer_id = c.id
		LEFT JOIN (
			SELECT customer_id, COUNT(DISTINCT id) AS active_sales_order_count
			FROM sales_orders
			WHERE deleted_at IS NULL
				AND (@tenantFilter = false OR tenant_id = @tenantID)
				AND LOWER(status) IN ('submitted', 'approved')
			GROUP BY customer_id
		) aso ON aso.customer_id = c.id
		WHERE c.deleted_at IS NULL
			AND (@tenantFilter = false OR c.tenant_id = @tenantID)
	`

	queryParams := withReportTenantParams(map[string]interface{}{
		"startDate": params.StartDate,
		"endDate":   params.EndDate,
	}, tenantID, tenantFilter)

	if params.Search != "" {
		baseQuery += ` AND (c.name ILIKE @search OR c.code ILIKE @search)`
		queryParams["search"] = params.Search + "%"
	}

	// Only show customers with orders in the selected period (Top Customers only)
	baseQuery += ` AND o.total_orders > 0`

	var total int64
	countSQL := "SELECT COUNT(*) " + baseQuery
	if err := database.GetDB(ctx, r.db).Raw(countSQL, queryParams).Scan(&total).Error; err != nil {
		return nil, utils.PaginationResult{}, fmt.Errorf("failed to count customers: %w", err)
	}

	sortColumn := "total_revenue"
	switch params.SortBy {
	case "name":
		sortColumn = "c.name"
	case "orders":
		sortColumn = "total_orders"
	case "last_order":
		sortColumn = "last_order_date"
	}

	sortOrder := "DESC"
	if strings.EqualFold(params.Order, "asc") {
		sortOrder = "ASC"
	}

	offset := (params.Page - 1) * params.PerPage
	selectSQL := fmt.Sprintf(`
		SELECT
			COALESCE(c.id::text, '') AS customer_id,
			COALESCE(c.name, '') AS customer_name,
			COALESCE(o.total_revenue, 0) AS total_revenue,
			COALESCE(o.total_orders, 0) AS total_orders,
			COALESCE(o.last_order_date, '') AS last_order_date,
			CASE WHEN COALESCE(o.total_orders, 0) > 0 THEN COALESCE(o.total_revenue, 0) / o.total_orders ELSE 0 END AS average_order_value,
			COALESCE(aso.active_sales_order_count, 0) AS active_sales_order_count
		%s
		ORDER BY %s %s, c.name ASC
		LIMIT @limit OFFSET @offset
	`, baseQuery, sortColumn, sortOrder)

	queryParams["limit"] = params.PerPage
	queryParams["offset"] = offset

	var rows []CustomerResearchRow
	if err := database.GetDB(ctx, r.db).Raw(selectSQL, queryParams).Scan(&rows).Error; err != nil {
		return nil, utils.PaginationResult{}, fmt.Errorf("failed to list customers: %w", err)
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

func (r *customerResearchRepository) GetRevenueByCustomer(ctx context.Context, params ListCustomersParams) ([]CustomerResearchRow, utils.PaginationResult, error) {
	params.Tab = "top"
	params.SortBy = "revenue"
	if params.Order == "" {
		params.Order = "desc"
	}
	return r.ListCustomers(ctx, params)
}

func (r *customerResearchRepository) GetPurchaseFrequency(ctx context.Context, params ListCustomersParams) ([]CustomerResearchRow, utils.PaginationResult, error) {
	params.Tab = "top"
	params.SortBy = "orders"
	if params.Order == "" {
		params.Order = "desc"
	}
	return r.ListCustomers(ctx, params)
}

func (r *customerResearchRepository) GetRevenueTrend(ctx context.Context, startDate, endDate time.Time, interval string) ([]CustomerRevenueTrendRow, error) {
	tenantID, tenantFilter, tenantErr := resolveReportTenantScope(ctx)
	if tenantErr != nil {
		return nil, tenantErr
	}

	dateTrunc := "day"
	switch strings.ToLower(interval) {
	case "weekly":
		dateTrunc = "week"
	case "monthly":
		dateTrunc = "month"
	}

	query := fmt.Sprintf(`
		SELECT
			TO_CHAR(DATE_TRUNC('%s', so.order_date), 'YYYY-MM-DD') AS period,
			COALESCE(SUM(soi.subtotal), 0) AS total_revenue,
			COUNT(DISTINCT so.id) AS total_orders
		FROM sales_orders so
		INNER JOIN sales_order_items soi ON soi.sales_order_id = so.id
			AND soi.deleted_at IS NULL
		WHERE so.deleted_at IS NULL
			AND (@tenantFilter = false OR so.tenant_id = @tenantID)
			AND so.status NOT IN ('draft', 'cancelled')
			AND so.order_date BETWEEN @startDate AND @endDate
		GROUP BY DATE_TRUNC('%s', so.order_date)
		ORDER BY DATE_TRUNC('%s', so.order_date)
	`, dateTrunc, dateTrunc, dateTrunc)

	var rows []CustomerRevenueTrendRow
	if err := database.GetDB(ctx, r.db).Raw(query, withReportTenantParams(map[string]interface{}{
		"startDate": startDate,
		"endDate":   endDate,
	}, tenantID, tenantFilter)).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to get revenue trend: %w", err)
	}

	return rows, nil
}

func (r *customerResearchRepository) GetCustomerDetail(ctx context.Context, customerID string, startDate, endDate time.Time) (*CustomerDetailRow, error) {
	tenantID, tenantFilter, tenantErr := resolveReportTenantScope(ctx)
	if tenantErr != nil {
		return nil, tenantErr
	}

	query := `
		SELECT
			COALESCE(c.id::text, '') AS customer_id,
			COALESCE(c.name, '') AS customer_name,
			COALESCE(SUM(soi.subtotal), 0) AS total_revenue,
			COALESCE(COUNT(DISTINCT so.id), 0) AS total_orders,
			CASE WHEN COALESCE(COUNT(DISTINCT so.id), 0) > 0 THEN COALESCE(SUM(soi.subtotal), 0) / COUNT(DISTINCT so.id) ELSE 0 END AS average_order_value,
			CASE WHEN MAX(so.order_date) IS NULL THEN '' ELSE TO_CHAR(MAX(so.order_date), 'YYYY-MM-DD') END AS last_order_date
		FROM customers c
		LEFT JOIN sales_orders so ON so.customer_id = c.id
			AND so.deleted_at IS NULL
			AND (@tenantFilter = false OR so.tenant_id = @tenantID)
			AND so.status NOT IN ('draft', 'cancelled')
			AND so.order_date BETWEEN @startDate AND @endDate
		LEFT JOIN sales_order_items soi ON soi.sales_order_id = so.id AND soi.deleted_at IS NULL
		WHERE c.deleted_at IS NULL
			AND (@tenantFilter = false OR c.tenant_id = @tenantID)
			AND c.id = @customerID
		GROUP BY c.id, c.name
	`

	var row CustomerDetailRow
	if err := database.GetDB(ctx, r.db).Raw(query, withReportTenantParams(map[string]interface{}{
		"customerID": customerID,
		"startDate":  startDate,
		"endDate":    endDate,
	}, tenantID, tenantFilter)).Scan(&row).Error; err != nil {
		return nil, fmt.Errorf("failed to get customer detail: %w", err)
	}

	if row.CustomerID == "" {
		return nil, nil
	}

	return &row, nil
}

func (r *customerResearchRepository) GetCustomerTopProducts(ctx context.Context, customerID string, startDate, endDate time.Time, limit int) ([]CustomerProductRow, error) {
	tenantID, tenantFilter, tenantErr := resolveReportTenantScope(ctx)
	if tenantErr != nil {
		return nil, tenantErr
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := `
		SELECT
			COALESCE(p.id::text, '') AS product_id,
			COALESCE(p.code, '') AS product_code,
			COALESCE(p.name, '') AS product_name,
			COALESCE(SUM(soi.quantity), 0) AS total_qty,
			COALESCE(SUM(soi.subtotal), 0) AS total_revenue,
			COUNT(DISTINCT so.id) AS total_orders
		FROM sales_order_items soi
		INNER JOIN sales_orders so ON so.id = soi.sales_order_id
			AND so.deleted_at IS NULL
			AND (@tenantFilter = false OR so.tenant_id = @tenantID)
			AND so.status NOT IN ('draft', 'cancelled')
			AND so.customer_id = @customerID
			AND so.order_date BETWEEN @startDate AND @endDate
		INNER JOIN products p ON p.id = soi.product_id
			AND p.deleted_at IS NULL
			AND (@tenantFilter = false OR p.tenant_id = @tenantID)
		WHERE soi.deleted_at IS NULL
		GROUP BY p.id, p.code, p.name
		ORDER BY total_revenue DESC
		LIMIT @limit
	`

	var rows []CustomerProductRow
	if err := database.GetDB(ctx, r.db).Raw(query, withReportTenantParams(map[string]interface{}{
		"customerID": customerID,
		"startDate":  startDate,
		"endDate":    endDate,
		"limit":      limit,
	}, tenantID, tenantFilter)).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to get customer top products: %w", err)
	}

	return rows, nil
}
