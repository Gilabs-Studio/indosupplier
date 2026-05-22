package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"gorm.io/gorm"
)

// ReceivablesRecapRow is a single row of the recap query.
type ReceivablesRecapRow struct {
	CustomerID        string    `json:"customer_id"`
	CustomerName      string    `json:"customer_name"`
	TotalReceivable   float64   `json:"total_receivable"`
	ReturnAmount      float64   `json:"return_amount"`
	DownPayment       float64   `json:"down_payment"`
	PaidAmount        float64   `json:"paid_amount"`
	OutstandingAmount float64   `json:"outstanding_amount"`
	LastTransaction   time.Time `json:"last_transaction"`
	AgingDays         int       `json:"aging_days"`
	AgingCategory     string    `json:"aging_category"`
}

// ReceivablesSummary holds aggregate totals and bucket counts.
type ReceivablesSummary struct {
	TotalCustomers    int     `json:"total_customers"`
	TotalReceivable   float64 `json:"total_receivable"`
	TotalReturn       float64 `json:"total_return"`
	TotalPaid         float64 `json:"total_paid"`
	TotalOutstanding  float64 `json:"total_outstanding"`
	CurrentCount      int     `json:"current_count"`
	Overdue1_30Count  int     `json:"overdue_1_30_count"`
	Overdue31_60Count int     `json:"overdue_31_60_count"`
	Overdue61_90Count int     `json:"overdue_61_90_count"`
	BadDebtCount      int     `json:"bad_debt_count"`
}

type ReceivablesRecapListParams struct {
	TenantID string
	Search   string
	SortBy   string
	SortDir  string
	Limit    int
	Offset   int
}

type ReceivablesRecapRepository struct {
	db *gorm.DB
}

func NewReceivablesRecapRepository(db *gorm.DB) *ReceivablesRecapRepository {
	return &ReceivablesRecapRepository{db: db}
}

// baseCTETemplate is the core CTE used by all methods.
// It adapts to our schema: sales_payments.customer_invoice_id / sales_payments.status.
// Tenant clauses are injected per source table to prevent cross-tenant aggregates.
const baseCTETemplate = `
WITH return_totals AS (
	SELECT
		sr.customer_id AS customer_id,
		COALESCE(SUM(sr.total_amount), 0) AS total_return,
		MAX(sr.created_at) AS last_return_at
	FROM sales_returns sr
	WHERE sr.deleted_at IS NULL
	  AND sr.status IN ('SUBMITTED', 'PROCESSED', 'COMPLETED')%s
	GROUP BY sr.customer_id
),
customer_receivables AS (
    SELECT 
        c.id                              AS customer_id,
        c.name                            AS customer_name,
        COALESCE(SUM(DISTINCT ci.amount), 0)       AS total_receivable,
		COALESCE(rt.total_return, 0)      AS return_amount,
        COALESCE(SUM(DISTINCT CASE WHEN ci.type = 'down_payment' THEN ci.amount ELSE NULL END), 0) AS down_payment,
        COALESCE(SUM(CASE WHEN sp.status = 'CONFIRMED' THEN sp.amount ELSE 0 END), 0)
                                           AS paid_amount,
        COALESCE(SUM(DISTINCT ci.amount), 0) -
			COALESCE(rt.total_return, 0) -
            COALESCE(SUM(CASE WHEN sp.status = 'CONFIRMED' THEN sp.amount ELSE 0 END), 0)
                                           AS outstanding_amount,
        GREATEST(
            COALESCE(MAX(ci.created_at), '1970-01-01'::timestamp),
			COALESCE(MAX(sp.payment_date::timestamp), '1970-01-01'::timestamp),
			COALESCE(MAX(rt.last_return_at), '1970-01-01'::timestamp)
        ) AS last_transaction,
        COALESCE(
            CASE
                WHEN MIN(
                    CASE WHEN ci.status IN ('PENDING','PARTIAL','OVERDUE')
                         AND ci.due_date IS NOT NULL
                    THEN ci.due_date END
                ) IS NOT NULL
                THEN CURRENT_DATE - MIN(
                    CASE WHEN ci.status IN ('PENDING','PARTIAL','OVERDUE')
                         AND ci.due_date IS NOT NULL
                    THEN ci.due_date::date END
                )
                ELSE 0
            END,
        0) AS aging_days
	FROM customers c
	LEFT JOIN sales_orders so ON c.id = so.customer_id AND so.deleted_at IS NULL%s
    LEFT JOIN customer_invoices ci ON so.id = ci.sales_order_id
        AND ci.status NOT IN ('DRAFT','NOT_CREATED')
		AND ci.deleted_at IS NULL%s
    LEFT JOIN sales_payments sp ON ci.id = sp.customer_invoice_id
		AND sp.deleted_at IS NULL%s
	LEFT JOIN return_totals rt ON c.id = rt.customer_id
	WHERE c.deleted_at IS NULL%s
	GROUP BY c.id, c.name, rt.total_return
	HAVING COALESCE(SUM(DISTINCT ci.amount), 0) > 0 OR COALESCE(rt.total_return, 0) > 0
)
SELECT
    customer_id,
    customer_name,
    total_receivable,
	return_amount,
    down_payment,
    paid_amount,
    outstanding_amount,
    last_transaction,
    aging_days,
    CASE
        WHEN outstanding_amount <= 0     THEN 'Paid'
        WHEN aging_days <= 30            THEN 'Current'
        WHEN aging_days <= 60            THEN 'Overdue 1-30'
        WHEN aging_days <= 90            THEN 'Overdue 31-60'
        WHEN aging_days <= 120           THEN 'Overdue 61-90'
        ELSE 'Bad Debt (>90)'
    END AS aging_category
FROM customer_receivables`

// allowedSortCols guards against SQL injection in ORDER BY.
var allowedSortCols = map[string]string{
	"customer_name":      "customer_name",
	"total_receivable":   "total_receivable",
	"return_amount":      "return_amount",
	"down_payment":       "down_payment",
	"paid_amount":        "paid_amount",
	"outstanding_amount": "outstanding_amount",
	"aging_days":         "aging_days",
	"last_transaction":   "last_transaction",
}

func buildReceivablesRecapBaseQuery(ctx context.Context) (string, map[string]interface{}, error) {
	if middleware.IsSystemAdmin(ctx) {
		query := fmt.Sprintf(baseCTETemplate, "", "", "", "", "")
		return query, map[string]interface{}{}, nil
	}

	tenantID := strings.TrimSpace(middleware.TenantFromContext(ctx))
	if tenantID == "" {
		return "", nil, fmt.Errorf("tenant context is required")
	}

	query := fmt.Sprintf(
		baseCTETemplate,
		" AND sr.tenant_id = @tenantID",
		" AND so.tenant_id = @tenantID",
		" AND ci.tenant_id = @tenantID",
		" AND sp.tenant_id = @tenantID",
		" AND c.tenant_id = @tenantID",
	)

	return query, map[string]interface{}{"tenantID": tenantID}, nil
}

func (r *ReceivablesRecapRepository) FindAll(ctx context.Context, params ReceivablesRecapListParams) ([]ReceivablesRecapRow, int64, error) {
	baseQuery, queryParams, err := buildReceivablesRecapBaseQuery(ctx)
	if err != nil {
		return nil, 0, err
	}

	where := ""
	if search := strings.TrimSpace(params.Search); search != "" {
		where = " WHERE customer_name ILIKE @search"
		queryParams["search"] = "%" + search + "%"
	}

	// Count
	countQ := fmt.Sprintf("SELECT COUNT(*) FROM (%s%s) AS cnt", baseQuery, where)
	var total int64
	if err := database.GetDB(ctx, r.db).Raw(countQ, queryParams).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sort
	orderBy := " ORDER BY last_transaction DESC, outstanding_amount DESC, aging_days DESC"
	if sortBy, ok := allowedSortCols[params.SortBy]; ok {
		dir := "ASC"
		if strings.EqualFold(params.SortDir, "desc") {
			dir = "DESC"
		}
		orderBy = fmt.Sprintf(" ORDER BY %s %s", sortBy, dir)
	}

	queryParams["limit"] = params.Limit
	queryParams["offset"] = params.Offset
	finalQ := fmt.Sprintf("%s%s%s LIMIT @limit OFFSET @offset", baseQuery, where, orderBy)

	var rows []ReceivablesRecapRow
	if err := database.GetDB(ctx, r.db).Raw(finalQ, queryParams).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *ReceivablesRecapRepository) GetSummary(ctx context.Context) (*ReceivablesSummary, error) {
	baseQuery, queryParams, err := buildReceivablesRecapBaseQuery(ctx)
	if err != nil {
		return nil, err
	}

	summaryQ := fmt.Sprintf(`
		SELECT
			COUNT(*)                                                           AS total_customers,
			COALESCE(SUM(total_receivable), 0)                                 AS total_receivable,
			COALESCE(SUM(return_amount), 0)                                    AS total_return,
			COALESCE(SUM(paid_amount), 0)                                      AS total_paid,
			COALESCE(SUM(outstanding_amount), 0)                               AS total_outstanding,
			COUNT(*) FILTER (WHERE aging_days <= 30 AND outstanding_amount > 0)  AS current_count,
			COUNT(*) FILTER (WHERE aging_days BETWEEN 31 AND 60 AND outstanding_amount > 0) AS overdue_1_30_count,
			COUNT(*) FILTER (WHERE aging_days BETWEEN 61 AND 90 AND outstanding_amount > 0) AS overdue_31_60_count,
			COUNT(*) FILTER (WHERE aging_days BETWEEN 91 AND 120 AND outstanding_amount > 0) AS overdue_61_90_count,
			COUNT(*) FILTER (WHERE aging_days > 120 AND outstanding_amount > 0) AS bad_debt_count
		FROM (%s) AS recap
	`, baseQuery)

	var s ReceivablesSummary
	if err := database.GetDB(ctx, r.db).Raw(summaryQ, queryParams).Scan(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ReceivablesRecapRepository) FindAllForExport(ctx context.Context, params ReceivablesRecapListParams) ([]ReceivablesRecapRow, error) {
	baseQuery, queryParams, err := buildReceivablesRecapBaseQuery(ctx)
	if err != nil {
		return nil, err
	}

	where := ""
	if search := strings.TrimSpace(params.Search); search != "" {
		where = " WHERE customer_name ILIKE @search"
		queryParams["search"] = "%" + search + "%"
	}

	orderBy := " ORDER BY last_transaction DESC, outstanding_amount DESC, aging_days DESC"
	if sortBy, ok := allowedSortCols[params.SortBy]; ok {
		dir := "ASC"
		if strings.EqualFold(params.SortDir, "desc") {
			dir = "DESC"
		}
		orderBy = fmt.Sprintf(" ORDER BY %s %s", sortBy, dir)
	}

	limit := params.Limit
	if limit <= 0 || limit > 10000 {
		limit = 10000
	}
	queryParams["limit"] = limit
	finalQ := fmt.Sprintf("%s%s%s LIMIT @limit", baseQuery, where, orderBy)

	var rows []ReceivablesRecapRow
	if err := database.GetDB(ctx, r.db).Raw(finalQ, queryParams).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
