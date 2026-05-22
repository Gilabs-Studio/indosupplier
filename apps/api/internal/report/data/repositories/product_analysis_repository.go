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

// ProductAnalysisRepository defines data access for product analysis reports.
// Queries existing tables (products, sales_order_items, sales_orders, etc.)
// without creating new models — this domain is purely read-only / aggregation.
type ProductAnalysisRepository interface {
	// ListProductPerformance returns paginated product performance data
	ListProductPerformance(ctx context.Context, params ListProductPerformanceParams) ([]ProductPerformanceRow, utils.PaginationResult, error)

	// GetMonthlyProductSales returns monthly aggregated product sales data
	GetMonthlyProductSales(ctx context.Context, startDate, endDate time.Time) ([]MonthlyProductSalesRow, error)

	// GetProductDetail returns a single product's detail + statistics
	GetProductDetail(ctx context.Context, productID string, startDate, endDate time.Time) (*ProductDetailRow, error)

	// GetProductTopCustomers returns top customers for a product
	GetProductTopCustomers(ctx context.Context, productID string, params ProductCustomerParams) ([]ProductCustomerRow, utils.PaginationResult, error)

	// GetProductTopSalesReps returns top sales reps selling a product
	GetProductTopSalesReps(ctx context.Context, productID string, params ProductSalesRepParams) ([]ProductSalesRepRow, utils.PaginationResult, error)

	// GetProductMonthlyTrend returns monthly sales trend for a specific product
	GetProductMonthlyTrend(ctx context.Context, productID string, startDate, endDate time.Time) ([]MonthlyProductTrendRow, error)

	// ListCategoryPerformance returns paginated category performance aggregated from product sales
	ListCategoryPerformance(ctx context.Context, params ListCategoryPerformanceParams) ([]CategoryPerformanceRow, utils.PaginationResult, error)

	// ListSegmentPerformance returns paginated performance aggregated by product segment
	ListSegmentPerformance(ctx context.Context, params ListDimensionPerformanceParams) ([]DimensionPerformanceRow, utils.PaginationResult, error)

	// ListTypePerformance returns paginated performance aggregated by product type
	ListTypePerformance(ctx context.Context, params ListDimensionPerformanceParams) ([]DimensionPerformanceRow, utils.PaginationResult, error)

	// ListPackagingPerformance returns paginated performance aggregated by packaging
	ListPackagingPerformance(ctx context.Context, params ListDimensionPerformanceParams) ([]DimensionPerformanceRow, utils.PaginationResult, error)

	// ListProcurementTypePerformance returns paginated performance aggregated by procurement type
	ListProcurementTypePerformance(ctx context.Context, params ListDimensionPerformanceParams) ([]DimensionPerformanceRow, utils.PaginationResult, error)
}

// --- Parameter structs ---

// ListProductPerformanceParams filters the product performance list
type ListProductPerformanceParams struct {
	Search     string
	CategoryID string
	StartDate  time.Time
	EndDate    time.Time
	Page       int
	PerPage    int
	SortBy     string
	Order      string
}

// ProductCustomerParams filters top customers for a product
type ProductCustomerParams struct {
	StartDate time.Time
	EndDate   time.Time
	Page      int
	PerPage   int
	SortBy    string
	Order     string
}

// ProductSalesRepParams filters top sales reps for a product
type ProductSalesRepParams struct {
	StartDate time.Time
	EndDate   time.Time
	Page      int
	PerPage   int
	SortBy    string
	Order     string
}

// ListCategoryPerformanceParams filters the category performance list
type ListCategoryPerformanceParams struct {
	Search    string
	StartDate time.Time
	EndDate   time.Time
	Page      int
	PerPage   int
	SortBy    string
	Order     string
}

// --- Result row structs ---

// ProductPerformanceRow is the raw query result for product performance list
type ProductPerformanceRow struct {
	ProductID    string
	ProductCode  string
	ProductName  string
	ProductSKU   string
	ProductImage string
	CategoryName string
	TotalQty     float64
	TotalRevenue float64
	TotalOrders  int
	AvgPrice     float64
}

// MonthlyProductSalesRow is the raw query result for monthly product sales chart
type MonthlyProductSalesRow struct {
	Month        int
	Year         int
	TotalRevenue float64
	TotalQty     float64
	TotalOrders  int
}

// ProductDetailRow is the raw query result for a single product detail
type ProductDetailRow struct {
	ProductID    string
	ProductCode  string
	ProductName  string
	ProductSKU   string
	ProductImage string
	CategoryName string
	BrandName    string
	SellingPrice float64
	CostPrice    float64
	CurrentStock float64
	// Current period stats
	TotalQty     float64
	TotalRevenue float64
	TotalOrders  int
	// Previous period stats for comparison
	PrevQty     float64
	PrevRevenue float64
	PrevOrders  int
}

// ProductCustomerRow is the raw query result for product top customers
type ProductCustomerRow struct {
	CustomerID   string
	CustomerName string
	CustomerCode string
	CustomerType string
	CityName     string
	TotalQty     float64
	TotalRevenue float64
	TotalOrders  int
}

// ProductSalesRepRow is the raw query result for product top sales reps
type ProductSalesRepRow struct {
	EmployeeID   string
	EmployeeCode string
	Name         string
	AvatarURL    string
	PositionName string
	TotalQty     float64
	TotalRevenue float64
	TotalOrders  int
}

// MonthlyProductTrendRow is the raw query result for a specific product trend
type MonthlyProductTrendRow struct {
	Month        int
	Year         int
	TotalRevenue float64
	TotalQty     float64
	TotalOrders  int
}

// CategoryPerformanceRow is the raw query result for category performance list
type CategoryPerformanceRow struct {
	CategoryID   string
	CategoryName string
	ProductCount int
	TotalQty     float64
	TotalRevenue float64
	TotalOrders  int
	AvgPrice     float64
}

// ListDimensionPerformanceParams is a generic params struct for segment/type/packaging/procurement-type performance
type ListDimensionPerformanceParams struct {
	Search    string
	StartDate time.Time
	EndDate   time.Time
	Page      int
	PerPage   int
	SortBy    string
	Order     string
}

// DimensionPerformanceRow is a generic result row for any product dimension performance query
type DimensionPerformanceRow struct {
	DimensionID   string
	DimensionName string
	ProductCount  int
	TotalQty      float64
	TotalRevenue  float64
	TotalOrders   int
	AvgPrice      float64
}

// --- Implementation ---

type productAnalysisRepository struct {
	db *gorm.DB
}

// NewProductAnalysisRepository creates a new ProductAnalysisRepository instance
func NewProductAnalysisRepository(db *gorm.DB) ProductAnalysisRepository {
	return &productAnalysisRepository{db: db}
}

func (r *productAnalysisRepository) ListProductPerformance(ctx context.Context, params ListProductPerformanceParams) ([]ProductPerformanceRow, utils.PaginationResult, error) {
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
		FROM products p
		LEFT JOIN product_categories pc ON pc.id = p.category_id AND pc.deleted_at IS NULL
		LEFT JOIN (
			SELECT soi.product_id,
				COALESCE(SUM(soi.quantity), 0) AS total_qty,
				COALESCE(SUM(soi.subtotal), 0) AS total_revenue,
				COUNT(DISTINCT soi.sales_order_id) AS total_orders
			FROM sales_order_items soi
			INNER JOIN sales_orders so ON so.id = soi.sales_order_id
				AND so.deleted_at IS NULL
				AND (@tenantFilter = false OR so.tenant_id = @tenantID)
				AND so.status NOT IN ('draft', 'cancelled')
				AND so.order_date BETWEEN @startDate AND @endDate
			WHERE soi.deleted_at IS NULL
			GROUP BY soi.product_id
		) soi_agg ON soi_agg.product_id = p.id
		WHERE p.deleted_at IS NULL
			AND (@tenantFilter = false OR p.tenant_id = @tenantID)
			AND p.is_active = true
	`

	queryParams := withReportTenantParams(map[string]interface{}{
		"startDate": params.StartDate,
		"endDate":   params.EndDate,
	}, tenantID, tenantFilter)

	if params.Search != "" {
		baseQuery += ` AND (p.name ILIKE @search OR p.code ILIKE @search OR p.sku ILIKE @search)`
		queryParams["search"] = params.Search + "%"
	}

	if params.CategoryID != "" {
		baseQuery += ` AND p.category_id = @categoryID`
		queryParams["categoryID"] = params.CategoryID
	}

	// Count total
	var total int64
	countSQL := "SELECT COUNT(*) " + baseQuery
	if err := database.GetDB(ctx, r.db).Raw(countSQL, queryParams).Scan(&total).Error; err != nil {
		return nil, utils.PaginationResult{}, fmt.Errorf("failed to count product performance: %w", err)
	}

	// Sort mapping
	sortColumn := "COALESCE(soi_agg.total_revenue, 0)"
	switch params.SortBy {
	case "name":
		sortColumn = "p.name"
	case "qty":
		sortColumn = "COALESCE(soi_agg.total_qty, 0)"
	case "orders":
		sortColumn = "COALESCE(soi_agg.total_orders, 0)"
	case "revenue":
		sortColumn = "COALESCE(soi_agg.total_revenue, 0)"
	}
	sortOrder := "DESC"
	if strings.EqualFold(params.Order, "asc") {
		sortOrder = "ASC"
	}

	offset := (params.Page - 1) * params.PerPage

	selectSQL := fmt.Sprintf(`
		SELECT
			p.id AS product_id,
			COALESCE(p.code, '') AS product_code,
			p.name AS product_name,
			COALESCE(p.sku, '') AS product_sku,
			COALESCE(p.image_url, '') AS product_image,
			COALESCE(pc.name, '') AS category_name,
			COALESCE(soi_agg.total_qty, 0) AS total_qty,
			COALESCE(soi_agg.total_revenue, 0) AS total_revenue,
			COALESCE(soi_agg.total_orders, 0) AS total_orders,
			CASE WHEN COALESCE(soi_agg.total_qty, 0) > 0
				THEN COALESCE(soi_agg.total_revenue, 0) / soi_agg.total_qty
				ELSE 0
			END AS avg_price
		%s
		ORDER BY %s %s, p.name ASC
		LIMIT @limit OFFSET @offset
	`, baseQuery, sortColumn, sortOrder)

	queryParams["limit"] = params.PerPage
	queryParams["offset"] = offset

	var rows []ProductPerformanceRow
	if err := database.GetDB(ctx, r.db).Raw(selectSQL, queryParams).Scan(&rows).Error; err != nil {
		return nil, utils.PaginationResult{}, fmt.Errorf("failed to list product performance: %w", err)
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

func (r *productAnalysisRepository) GetMonthlyProductSales(ctx context.Context, startDate, endDate time.Time) ([]MonthlyProductSalesRow, error) {
	tenantID, tenantFilter, tenantErr := resolveReportTenantScope(ctx)
	if tenantErr != nil {
		return nil, tenantErr
	}

	query := `
		SELECT
			EXTRACT(MONTH FROM so.order_date)::int AS month,
			EXTRACT(YEAR FROM so.order_date)::int AS year,
			COALESCE(SUM(soi.subtotal), 0) AS total_revenue,
			COALESCE(SUM(soi.quantity), 0) AS total_qty,
			COUNT(DISTINCT so.id) AS total_orders
		FROM sales_order_items soi
		INNER JOIN sales_orders so ON so.id = soi.sales_order_id
			AND so.deleted_at IS NULL
			AND (@tenantFilter = false OR so.tenant_id = @tenantID)
			AND so.status NOT IN ('draft', 'cancelled')
			AND so.order_date BETWEEN @startDate AND @endDate
		WHERE soi.deleted_at IS NULL
		GROUP BY EXTRACT(MONTH FROM so.order_date), EXTRACT(YEAR FROM so.order_date)
		ORDER BY year, month
	`

	var rows []MonthlyProductSalesRow
	if err := database.GetDB(ctx, r.db).Raw(query, withReportTenantParams(map[string]interface{}{
		"startDate": startDate,
		"endDate":   endDate,
	}, tenantID, tenantFilter)).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to get monthly product sales: %w", err)
	}

	return rows, nil
}

func (r *productAnalysisRepository) GetProductDetail(ctx context.Context, productID string, startDate, endDate time.Time) (*ProductDetailRow, error) {
	tenantID, tenantFilter, tenantErr := resolveReportTenantScope(ctx)
	if tenantErr != nil {
		return nil, tenantErr
	}

	// Calculate previous period for comparison
	duration := endDate.Sub(startDate)
	prevEnd := startDate.Add(-1 * 24 * time.Hour)
	prevStart := prevEnd.Add(-duration)

	query := `
		SELECT
			p.id AS product_id,
			COALESCE(p.code, '') AS product_code,
			p.name AS product_name,
			COALESCE(p.sku, '') AS product_sku,
			COALESCE(p.image_url, '') AS product_image,
			COALESCE(pc.name, '') AS category_name,
			COALESCE(pb.name, '') AS brand_name,
			COALESCE(p.selling_price, 0) AS selling_price,
			COALESCE(p.cost_price, 0) AS cost_price,
			COALESCE(p.current_stock, 0) AS current_stock,
			COALESCE(curr.total_qty, 0) AS total_qty,
			COALESCE(curr.total_revenue, 0) AS total_revenue,
			COALESCE(curr.total_orders, 0) AS total_orders,
			COALESCE(prev.total_qty, 0) AS prev_qty,
			COALESCE(prev.total_revenue, 0) AS prev_revenue,
			COALESCE(prev.total_orders, 0) AS prev_orders
		FROM products p
		LEFT JOIN product_categories pc ON pc.id = p.category_id AND pc.deleted_at IS NULL
		LEFT JOIN product_brands pb ON pb.id = p.brand_id AND pb.deleted_at IS NULL
		LEFT JOIN (
			SELECT soi.product_id,
				SUM(soi.quantity) AS total_qty,
				SUM(soi.subtotal) AS total_revenue,
				COUNT(DISTINCT soi.sales_order_id) AS total_orders
			FROM sales_order_items soi
			INNER JOIN sales_orders so ON so.id = soi.sales_order_id
				AND (@tenantFilter = false OR so.tenant_id = @tenantID)
				AND so.deleted_at IS NULL AND so.status NOT IN ('draft', 'cancelled')
				AND so.order_date BETWEEN @startDate AND @endDate
			WHERE soi.deleted_at IS NULL
			GROUP BY soi.product_id
		) curr ON curr.product_id = p.id
		LEFT JOIN (
			SELECT soi.product_id,
				SUM(soi.quantity) AS total_qty,
				SUM(soi.subtotal) AS total_revenue,
				COUNT(DISTINCT soi.sales_order_id) AS total_orders
			FROM sales_order_items soi
			INNER JOIN sales_orders so ON so.id = soi.sales_order_id
				AND (@tenantFilter = false OR so.tenant_id = @tenantID)
				AND so.deleted_at IS NULL AND so.status NOT IN ('draft', 'cancelled')
				AND so.order_date BETWEEN @prevStart AND @prevEnd
			WHERE soi.deleted_at IS NULL
			GROUP BY soi.product_id
		) prev ON prev.product_id = p.id
		WHERE p.id = @productID AND p.deleted_at IS NULL
			AND (@tenantFilter = false OR p.tenant_id = @tenantID)
	`

	var row ProductDetailRow
	if err := database.GetDB(ctx, r.db).Raw(query, withReportTenantParams(map[string]interface{}{
		"productID": productID,
		"startDate": startDate,
		"endDate":   endDate,
		"prevStart": prevStart,
		"prevEnd":   prevEnd,
	}, tenantID, tenantFilter)).Scan(&row).Error; err != nil {
		return nil, fmt.Errorf("failed to get product detail: %w", err)
	}

	if row.ProductID == "" {
		return nil, nil
	}

	return &row, nil
}

func (r *productAnalysisRepository) GetProductTopCustomers(ctx context.Context, productID string, params ProductCustomerParams) ([]ProductCustomerRow, utils.PaginationResult, error) {
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
			AND so.order_date BETWEEN @startDate AND @endDate
		LEFT JOIN customers c ON c.id = so.customer_id AND c.deleted_at IS NULL AND (@tenantFilter = false OR c.tenant_id = @tenantID)
		LEFT JOIN customer_types ct ON ct.id = c.customer_type_id AND ct.deleted_at IS NULL
		LEFT JOIN cities ci ON ci.id = c.city_id AND ci.deleted_at IS NULL
		WHERE soi.deleted_at IS NULL
			AND soi.product_id = @productID
	`

	queryParams := withReportTenantParams(map[string]interface{}{
		"productID": productID,
		"startDate": params.StartDate,
		"endDate":   params.EndDate,
	}, tenantID, tenantFilter)

	// Count distinct customers
	var total int64
	countSQL := "SELECT COUNT(DISTINCT so.customer_id) " + baseQuery
	if err := database.GetDB(ctx, r.db).Raw(countSQL, queryParams).Scan(&total).Error; err != nil {
		return nil, utils.PaginationResult{}, fmt.Errorf("failed to count product customers: %w", err)
	}

	sortColumn := "total_revenue"
	switch params.SortBy {
	case "name":
		sortColumn = "customer_name"
	case "qty":
		sortColumn = "total_qty"
	case "orders":
		sortColumn = "total_orders"
	case "revenue":
		sortColumn = "total_revenue"
	}
	sortOrder := "DESC"
	if strings.EqualFold(params.Order, "asc") {
		sortOrder = "ASC"
	}

	offset := (params.Page - 1) * params.PerPage

	selectSQL := fmt.Sprintf(`
		SELECT
			COALESCE(so.customer_id::text, '') AS customer_id,
			COALESCE(c.name, so.customer_name, '') AS customer_name,
			COALESCE(c.code, '') AS customer_code,
			COALESCE(ct.name, '') AS customer_type,
			COALESCE(ci.name, '') AS city_name,
			COALESCE(SUM(soi.quantity), 0) AS total_qty,
			COALESCE(SUM(soi.subtotal), 0) AS total_revenue,
			COUNT(DISTINCT soi.sales_order_id) AS total_orders
		%s
		GROUP BY so.customer_id, c.name, so.customer_name, c.code, ct.name, ci.name
		ORDER BY %s %s
		LIMIT @limit OFFSET @offset
	`, baseQuery, sortColumn, sortOrder)

	queryParams["limit"] = params.PerPage
	queryParams["offset"] = offset

	var rows []ProductCustomerRow
	if err := database.GetDB(ctx, r.db).Raw(selectSQL, queryParams).Scan(&rows).Error; err != nil {
		return nil, utils.PaginationResult{}, fmt.Errorf("failed to get product top customers: %w", err)
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

func (r *productAnalysisRepository) GetProductTopSalesReps(ctx context.Context, productID string, params ProductSalesRepParams) ([]ProductSalesRepRow, utils.PaginationResult, error) {
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
			AND so.order_date BETWEEN @startDate AND @endDate
		INNER JOIN employees e ON e.id = so.sales_rep_id AND e.deleted_at IS NULL AND (@tenantFilter = false OR e.tenant_id = @tenantID)
		LEFT JOIN job_positions jp ON jp.id = e.job_position_id
		LEFT JOIN users u ON u.id = e.user_id
		WHERE soi.deleted_at IS NULL
			AND soi.product_id = @productID
	`

	queryParams := withReportTenantParams(map[string]interface{}{
		"productID": productID,
		"startDate": params.StartDate,
		"endDate":   params.EndDate,
	}, tenantID, tenantFilter)

	// Count distinct sales reps
	var total int64
	countSQL := "SELECT COUNT(DISTINCT so.sales_rep_id) " + baseQuery
	if err := database.GetDB(ctx, r.db).Raw(countSQL, queryParams).Scan(&total).Error; err != nil {
		return nil, utils.PaginationResult{}, fmt.Errorf("failed to count product sales reps: %w", err)
	}

	sortColumn := "total_revenue"
	switch params.SortBy {
	case "name":
		sortColumn = "e.name"
	case "qty":
		sortColumn = "total_qty"
	case "orders":
		sortColumn = "total_orders"
	case "revenue":
		sortColumn = "total_revenue"
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
			COALESCE(u.avatar_url, '') AS avatar_url,
			COALESCE(jp.name, '') AS position_name,
			COALESCE(SUM(soi.quantity), 0) AS total_qty,
			COALESCE(SUM(soi.subtotal), 0) AS total_revenue,
			COUNT(DISTINCT soi.sales_order_id) AS total_orders
		%s
		GROUP BY e.id, e.employee_code, e.name, u.avatar_url, jp.name
		ORDER BY %s %s
		LIMIT @limit OFFSET @offset
	`, baseQuery, sortColumn, sortOrder)

	queryParams["limit"] = params.PerPage
	queryParams["offset"] = offset

	var rows []ProductSalesRepRow
	if err := database.GetDB(ctx, r.db).Raw(selectSQL, queryParams).Scan(&rows).Error; err != nil {
		return nil, utils.PaginationResult{}, fmt.Errorf("failed to get product top sales reps: %w", err)
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

func (r *productAnalysisRepository) GetProductMonthlyTrend(ctx context.Context, productID string, startDate, endDate time.Time) ([]MonthlyProductTrendRow, error) {
	tenantID, tenantFilter, tenantErr := resolveReportTenantScope(ctx)
	if tenantErr != nil {
		return nil, tenantErr
	}

	query := `
		SELECT
			EXTRACT(MONTH FROM so.order_date)::int AS month,
			EXTRACT(YEAR FROM so.order_date)::int AS year,
			COALESCE(SUM(soi.subtotal), 0) AS total_revenue,
			COALESCE(SUM(soi.quantity), 0) AS total_qty,
			COUNT(DISTINCT so.id) AS total_orders
		FROM sales_order_items soi
		INNER JOIN sales_orders so ON so.id = soi.sales_order_id
			AND so.deleted_at IS NULL
			AND (@tenantFilter = false OR so.tenant_id = @tenantID)
			AND so.status NOT IN ('draft', 'cancelled')
			AND so.order_date BETWEEN @startDate AND @endDate
		WHERE soi.deleted_at IS NULL
			AND soi.product_id = @productID
		GROUP BY EXTRACT(MONTH FROM so.order_date), EXTRACT(YEAR FROM so.order_date)
		ORDER BY year, month
	`

	var rows []MonthlyProductTrendRow
	if err := database.GetDB(ctx, r.db).Raw(query, withReportTenantParams(map[string]interface{}{
		"productID": productID,
		"startDate": startDate,
		"endDate":   endDate,
	}, tenantID, tenantFilter)).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to get product monthly trend: %w", err)
	}

	return rows, nil
}

func (r *productAnalysisRepository) ListCategoryPerformance(ctx context.Context, params ListCategoryPerformanceParams) ([]CategoryPerformanceRow, utils.PaginationResult, error) {
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

	// Base FROM clause — aggregate sales at product level first, then group by category
	baseQuery := `
		FROM products p
		LEFT JOIN product_categories pc ON pc.id = p.category_id AND pc.deleted_at IS NULL
		LEFT JOIN (
			SELECT soi.product_id,
				COALESCE(SUM(soi.quantity), 0)              AS total_qty,
				COALESCE(SUM(soi.subtotal), 0)              AS total_revenue,
				COUNT(DISTINCT soi.sales_order_id)          AS total_orders
			FROM sales_order_items soi
			INNER JOIN sales_orders so ON so.id = soi.sales_order_id
				AND so.deleted_at IS NULL
				AND (@tenantFilter = false OR so.tenant_id = @tenantID)
				AND so.status NOT IN ('draft', 'cancelled')
				AND so.order_date BETWEEN @startDate AND @endDate
			WHERE soi.deleted_at IS NULL
			GROUP BY soi.product_id
		) soi_agg ON soi_agg.product_id = p.id
		WHERE p.deleted_at IS NULL
			AND (@tenantFilter = false OR p.tenant_id = @tenantID)
			AND p.is_active = true
	`

	queryParams := withReportTenantParams(map[string]interface{}{
		"startDate": params.StartDate,
		"endDate":   params.EndDate,
	}, tenantID, tenantFilter)

	if params.Search != "" {
		baseQuery += ` AND pc.name ILIKE @search`
		queryParams["search"] = params.Search + "%"
	}

	// Count distinct categories
	countSQL := "SELECT COUNT(DISTINCT COALESCE(pc.id::text, 'uncategorized')) " + baseQuery
	var total int64
	if err := database.GetDB(ctx, r.db).Raw(countSQL, queryParams).Scan(&total).Error; err != nil {
		return nil, utils.PaginationResult{}, fmt.Errorf("failed to count category performance: %w", err)
	}

	// Sort mapping
	sortColumn := "total_revenue"
	switch params.SortBy {
	case "name":
		sortColumn = "category_name"
	case "qty":
		sortColumn = "total_qty"
	case "orders":
		sortColumn = "total_orders"
	case "products":
		sortColumn = "product_count"
	default:
		sortColumn = "total_revenue"
	}
	sortOrder := "DESC"
	if strings.EqualFold(params.Order, "asc") {
		sortOrder = "ASC"
	}

	offset := (params.Page - 1) * params.PerPage

	selectSQL := fmt.Sprintf(`
		SELECT
			COALESCE(pc.id::text, '')                                AS category_id,
			COALESCE(pc.name, 'Uncategorized')                       AS category_name,
			COUNT(DISTINCT p.id)::int                                AS product_count,
			COALESCE(SUM(soi_agg.total_qty), 0)                     AS total_qty,
			COALESCE(SUM(soi_agg.total_revenue), 0)                 AS total_revenue,
			COALESCE(SUM(soi_agg.total_orders), 0)::int             AS total_orders,
			CASE
				WHEN COALESCE(SUM(soi_agg.total_qty), 0) > 0
				THEN COALESCE(SUM(soi_agg.total_revenue), 0) / SUM(soi_agg.total_qty)
				ELSE 0
			END                                                       AS avg_price
		%s
		GROUP BY pc.id, pc.name
		ORDER BY %s %s, category_name ASC
		LIMIT @limit OFFSET @offset
	`, baseQuery, sortColumn, sortOrder)

	queryParams["limit"] = params.PerPage
	queryParams["offset"] = offset

	var rows []CategoryPerformanceRow
	if err := database.GetDB(ctx, r.db).Raw(selectSQL, queryParams).Scan(&rows).Error; err != nil {
		return nil, utils.PaginationResult{}, fmt.Errorf("failed to list category performance: %w", err)
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

// listDimensionPerformance is a generic helper that aggregates product sales by any FK dimension table.
// dimensionTable is the join table name, fkColumn is the product's FK column.
func (r *productAnalysisRepository) listDimensionPerformance(
	ctx context.Context,
	params ListDimensionPerformanceParams,
	dimensionTable,
	fkColumn,
	dimNameCol string,
) ([]DimensionPerformanceRow, utils.PaginationResult, error) {
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

	baseQuery := fmt.Sprintf(`
		FROM products p
		LEFT JOIN %s dim ON dim.id = p.%s AND dim.deleted_at IS NULL
		LEFT JOIN (
			SELECT soi.product_id,
				COALESCE(SUM(soi.quantity), 0)             AS total_qty,
				COALESCE(SUM(soi.subtotal), 0)             AS total_revenue,
				COUNT(DISTINCT soi.sales_order_id)         AS total_orders
			FROM sales_order_items soi
			INNER JOIN sales_orders so ON so.id = soi.sales_order_id
				AND so.deleted_at IS NULL
				AND (@tenantFilter = false OR so.tenant_id = @tenantID)
				AND so.status NOT IN ('draft', 'cancelled')
				AND so.order_date BETWEEN @startDate AND @endDate
			WHERE soi.deleted_at IS NULL
			GROUP BY soi.product_id
		) soi_agg ON soi_agg.product_id = p.id
		WHERE p.deleted_at IS NULL
			AND (@tenantFilter = false OR p.tenant_id = @tenantID)
			AND p.is_active = true
	`, dimensionTable, fkColumn)

	queryParams := withReportTenantParams(map[string]interface{}{
		"startDate": params.StartDate,
		"endDate":   params.EndDate,
	}, tenantID, tenantFilter)

	if params.Search != "" {
		baseQuery += fmt.Sprintf(` AND dim.%s ILIKE @search`, dimNameCol)
		queryParams["search"] = params.Search + "%"
	}

	countSQL := fmt.Sprintf("SELECT COUNT(DISTINCT COALESCE(dim.id::text, 'uncategorized')) %s", baseQuery)
	var total int64
	if err := database.GetDB(ctx, r.db).Raw(countSQL, queryParams).Scan(&total).Error; err != nil {
		return nil, utils.PaginationResult{}, fmt.Errorf("failed to count %s performance: %w", dimensionTable, err)
	}

	sortColumn := "total_revenue"
	switch params.SortBy {
	case "name":
		sortColumn = "dimension_name"
	case "qty":
		sortColumn = "total_qty"
	case "orders":
		sortColumn = "total_orders"
	case "products":
		sortColumn = "product_count"
	default:
		sortColumn = "total_revenue"
	}
	sortOrder := "DESC"
	if strings.EqualFold(params.Order, "asc") {
		sortOrder = "ASC"
	}

	offset := (params.Page - 1) * params.PerPage

	selectSQL := fmt.Sprintf(`
		SELECT
			COALESCE(dim.id::text, '')                               AS dimension_id,
			COALESCE(dim.%s, 'Uncategorized')                        AS dimension_name,
			COUNT(DISTINCT p.id)::int                                AS product_count,
			COALESCE(SUM(soi_agg.total_qty), 0)                     AS total_qty,
			COALESCE(SUM(soi_agg.total_revenue), 0)                 AS total_revenue,
			COALESCE(SUM(soi_agg.total_orders), 0)::int             AS total_orders,
			CASE
				WHEN COALESCE(SUM(soi_agg.total_qty), 0) > 0
				THEN COALESCE(SUM(soi_agg.total_revenue), 0) / SUM(soi_agg.total_qty)
				ELSE 0
			END                                                       AS avg_price
		%s
		GROUP BY dim.id, dim.%s
		ORDER BY %s %s, dimension_name ASC
		LIMIT @limit OFFSET @offset
	`, dimNameCol, baseQuery, dimNameCol, sortColumn, sortOrder)

	queryParams["limit"] = params.PerPage
	queryParams["offset"] = offset

	var rows []DimensionPerformanceRow
	if err := database.GetDB(ctx, r.db).Raw(selectSQL, queryParams).Scan(&rows).Error; err != nil {
		return nil, utils.PaginationResult{}, fmt.Errorf("failed to list %s performance: %w", dimensionTable, err)
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

func (r *productAnalysisRepository) ListSegmentPerformance(ctx context.Context, params ListDimensionPerformanceParams) ([]DimensionPerformanceRow, utils.PaginationResult, error) {
	return r.listDimensionPerformance(ctx, params, "product_segments", "segment_id", "name")
}

func (r *productAnalysisRepository) ListTypePerformance(ctx context.Context, params ListDimensionPerformanceParams) ([]DimensionPerformanceRow, utils.PaginationResult, error) {
	return r.listDimensionPerformance(ctx, params, "product_types", "type_id", "name")
}

func (r *productAnalysisRepository) ListPackagingPerformance(ctx context.Context, params ListDimensionPerformanceParams) ([]DimensionPerformanceRow, utils.PaginationResult, error) {
	return r.listDimensionPerformance(ctx, params, "packagings", "packaging_id", "name")
}

func (r *productAnalysisRepository) ListProcurementTypePerformance(ctx context.Context, params ListDimensionPerformanceParams) ([]DimensionPerformanceRow, utils.PaginationResult, error) {
	return r.listDimensionPerformance(ctx, params, "procurement_types", "procurement_type_id", "name")
}
