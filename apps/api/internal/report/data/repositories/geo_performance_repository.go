package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"

	"gorm.io/gorm"
)

// GeoPerformanceRepository defines data access for geo performance reports.
// Queries existing sales_orders + customers + provinces/cities tables.
type GeoPerformanceRepository interface {
	// GetGeoPerformanceBySalesOrder aggregates revenue from confirmed/approved sales orders by geographic area
	GetGeoPerformanceBySalesOrder(ctx context.Context, params GeoPerformanceParams) ([]GeoPerformanceRow, error)

	// GetGeoPerformanceByPaidInvoice aggregates revenue from paid invoices by geographic area
	GetGeoPerformanceByPaidInvoice(ctx context.Context, params GeoPerformanceParams) ([]GeoPerformanceRow, error)

	// GetSalesRepsForFilter returns sales reps for the filter dropdown
	GetSalesRepsForFilter(ctx context.Context) ([]SalesRepFilterRow, error)
}

// GeoPerformanceParams holds query filters
type GeoPerformanceParams struct {
	StartDate  time.Time
	EndDate    time.Time
	SalesRepID string
	Level      string // "province" or "city"
}

// GeoPerformanceRow represents a single area's aggregated data
type GeoPerformanceRow struct {
	AreaID       string
	AreaName     string
	ParentName   string
	TotalRevenue float64
	TotalOrders  int
}

// SalesRepFilterRow represents a sales rep for filter dropdown
type SalesRepFilterRow struct {
	ID   string
	Name string
	Code string
}

type geoPerformanceRepository struct {
	db *gorm.DB
}

// NewGeoPerformanceRepository creates a new instance
func NewGeoPerformanceRepository(db *gorm.DB) GeoPerformanceRepository {
	return &geoPerformanceRepository{db: db}
}

func (r *geoPerformanceRepository) GetGeoPerformanceBySalesOrder(ctx context.Context, params GeoPerformanceParams) ([]GeoPerformanceRow, error) {
	tenantID, tenantFilter, tenantErr := resolveReportTenantScope(ctx)
	if tenantErr != nil {
		return nil, tenantErr
	}

	var rows []GeoPerformanceRow

	areaSelect, areaJoin, groupBy := r.buildAreaClause(params.Level)

	query := fmt.Sprintf(`
		SELECT %s,
			COALESCE(SUM(so.total_amount), 0) AS total_revenue,
			COUNT(so.id) AS total_orders
		FROM sales_orders so
		INNER JOIN customers c ON so.customer_id = c.id AND c.deleted_at IS NULL
		%s
		WHERE so.deleted_at IS NULL
			AND so.status IN ('approved', 'closed')
			AND so.order_date BETWEEN ? AND ?
	`, areaSelect, areaJoin)

	args := []interface{}{params.StartDate, params.EndDate}
	if tenantFilter {
		query += " AND so.tenant_id = ? AND c.tenant_id = ?"
		args = append(args, tenantID, tenantID)
	}

	if params.SalesRepID != "" {
		query += " AND so.sales_rep_id = ?"
		args = append(args, params.SalesRepID)
	}

	query += fmt.Sprintf(" GROUP BY %s ORDER BY total_revenue DESC", groupBy)

	err := database.GetDB(ctx, r.db).Raw(query, args...).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get geo performance by sales order: %w", err)
	}

	return rows, nil
}

func (r *geoPerformanceRepository) GetGeoPerformanceByPaidInvoice(ctx context.Context, params GeoPerformanceParams) ([]GeoPerformanceRow, error) {
	tenantID, tenantFilter, tenantErr := resolveReportTenantScope(ctx)
	if tenantErr != nil {
		return nil, tenantErr
	}

	var rows []GeoPerformanceRow

	areaSelect, areaJoin, groupBy := r.buildAreaClause(params.Level)

	query := fmt.Sprintf(`
		SELECT %s,
			COALESCE(SUM(ci.paid_amount), 0) AS total_revenue,
			COUNT(DISTINCT ci.sales_order_id) AS total_orders
		FROM customer_invoices ci
		INNER JOIN sales_orders so ON ci.sales_order_id = so.id AND so.deleted_at IS NULL
		INNER JOIN customers c ON so.customer_id = c.id AND c.deleted_at IS NULL
		%s
		WHERE ci.deleted_at IS NULL
			AND ci.status = 'paid'
			AND ci.invoice_date BETWEEN ? AND ?
	`, areaSelect, areaJoin)

	args := []interface{}{params.StartDate, params.EndDate}
	if tenantFilter {
		query += " AND so.tenant_id = ? AND c.tenant_id = ? AND ci.tenant_id = ?"
		args = append(args, tenantID, tenantID, tenantID)
	}

	if params.SalesRepID != "" {
		query += " AND so.sales_rep_id = ?"
		args = append(args, params.SalesRepID)
	}

	query += fmt.Sprintf(" GROUP BY %s ORDER BY total_revenue DESC", groupBy)

	err := database.GetDB(ctx, r.db).Raw(query, args...).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get geo performance by paid invoice: %w", err)
	}

	return rows, nil
}

func (r *geoPerformanceRepository) GetSalesRepsForFilter(ctx context.Context) ([]SalesRepFilterRow, error) {
	tenantID, tenantFilter, tenantErr := resolveReportTenantScope(ctx)
	if tenantErr != nil {
		return nil, tenantErr
	}

	var rows []SalesRepFilterRow

	query := `
		SELECT DISTINCT e.id, e.name, e.employee_code AS code
		FROM employees e
		INNER JOIN sales_orders so ON so.sales_rep_id = e.id AND so.deleted_at IS NULL
		WHERE e.deleted_at IS NULL
		ORDER BY e.name ASC
	`
	args := []interface{}{}
	if tenantFilter {
		query = `
			SELECT DISTINCT e.id, e.name, e.employee_code AS code
			FROM employees e
			INNER JOIN sales_orders so ON so.sales_rep_id = e.id AND so.deleted_at IS NULL
			WHERE e.deleted_at IS NULL
				AND e.tenant_id = ?
				AND so.tenant_id = ?
			ORDER BY e.name ASC
		`
		args = append(args, tenantID, tenantID)
	}

	err := database.GetDB(ctx, r.db).Raw(query, args...).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get sales reps for filter: %w", err)
	}

	return rows, nil
}

// buildAreaClause returns (selectClause, joinClause, groupByClause) based on aggregation level
func (r *geoPerformanceRepository) buildAreaClause(level string) (string, string, string) {
	if level == "city" {
		return `ci2.id AS area_id, ci2.name AS area_name, p.name AS parent_name`,
			`INNER JOIN cities ci2 ON c.city_id = ci2.id
			 INNER JOIN provinces p ON ci2.province_id = p.id`,
			`ci2.id, ci2.name, p.name`
	}

	// Default: province level
	return `p.id AS area_id, p.name AS area_name, '' AS parent_name`,
		`INNER JOIN provinces p ON c.province_id = p.id`,
		`p.id, p.name`
}
