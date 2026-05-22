package repositories

import (
	"context"
	"database/sql"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ARAgingRow struct {
	InvoiceID       string
	SourceType      string
	Code            string
	InvoiceNumber   *string
	CustomerID      string
	CustomerName    string
	InvoiceDate     time.Time
	DueDate         time.Time
	Amount          float64
	RemainingAmount float64
}

type APAgingRow struct {
	InvoiceID       string
	SourceType      string
	Code            string
	InvoiceNumber   string
	InvoiceDate     time.Time
	DueDate         time.Time
	SupplierID      string
	SupplierName    string
	Amount          float64
	PaidAmount      float64
	RemainingAmount float64
}

type AgingListParams struct {
	Search   string
	AsOfDate time.Time
	Limit    int
	Offset   int
}

type AgingFinanceListParams struct {
	Search    string
	AsOfDate  time.Time
	PartnerID string
	MinAmount float64
}

type AgingReportRepository interface {
	ListARAging(ctx context.Context, params AgingListParams) ([]ARAgingRow, int64, error)
	ListAPAging(ctx context.Context, params AgingListParams) ([]APAgingRow, int64, error)
	ListARAgingFinance(ctx context.Context, params AgingFinanceListParams) ([]ARAgingRow, error)
	ListAPAgingFinance(ctx context.Context, params AgingFinanceListParams) ([]APAgingRow, error)
}

type agingReportRepository struct {
	db *gorm.DB
}

func NewAgingReportRepository(db *gorm.DB) AgingReportRepository {
	return &agingReportRepository{db: db}
}

func (r *agingReportRepository) ListARAging(ctx context.Context, params AgingListParams) ([]ARAgingRow, int64, error) {
	var rows []ARAgingRow
	var total int64

	asOf := params.AsOfDate.Format("2006-01-02")
	search := strings.TrimSpace(params.Search)
	like := "%" + search + "%"

	countSQL := `
		SELECT COUNT(*)
		FROM customer_invoices ci
		LEFT JOIN sales_orders so ON so.id = ci.sales_order_id AND so.deleted_at IS NULL
		LEFT JOIN customers c ON c.id = so.customer_id AND c.deleted_at IS NULL
		WHERE ci.deleted_at IS NULL
			AND UPPER(ci.status) IN ('UNPAID', 'PARTIAL', 'WAITING_PAYMENT')
			AND ci.remaining_amount > 0
			AND ci.invoice_date <= ?::date
			AND (
				? = ''
				OR ci.code ILIKE ?
				OR ci.invoice_number ILIKE ?
				OR COALESCE(c.name, so.customer_name, '') ILIKE ?
			)
	`
	tenantID := strings.TrimSpace(middleware.TenantFromContext(ctx))
	countArgs := []interface{}{asOf, search, like, like, like}
	if tenantID != "" {
		countSQL += " AND ci.tenant_id = ?"
		countArgs = append(countArgs, tenantID)
	}
	if err := r.db.WithContext(ctx).Raw(countSQL, countArgs...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	listSQL := `
		SELECT 
			ci.id as invoice_id, 
			'CUSTOMER_INVOICE' as source_type,
			ci.code, 
			ci.invoice_number, 
			COALESCE(so.customer_id::text, '') as customer_id,
			COALESCE(c.name, so.customer_name, 'Customer') as customer_name,
			ci.invoice_date, 
			COALESCE(ci.due_date, ci.invoice_date) as due_date, 
			ci.amount, 
			ci.remaining_amount
		FROM customer_invoices ci
		LEFT JOIN sales_orders so ON so.id = ci.sales_order_id AND so.deleted_at IS NULL
		LEFT JOIN customers c ON c.id = so.customer_id AND c.deleted_at IS NULL
		WHERE ci.deleted_at IS NULL
			AND UPPER(ci.status) IN ('UNPAID', 'PARTIAL', 'WAITING_PAYMENT')
			AND ci.remaining_amount > 0
			AND ci.invoice_date <= ?::date
			AND (
				? = ''
				OR ci.code ILIKE ?
				OR ci.invoice_number ILIKE ?
				OR COALESCE(c.name, so.customer_name, '') ILIKE ?
			)
		ORDER BY ci.invoice_date DESC
		LIMIT ? OFFSET ?
	`
	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	listArgs := []interface{}{asOf, search, like, like, like, params.Limit, params.Offset}
	if tenantID != "" {
		listSQL = strings.Replace(listSQL, "WHERE ci.deleted_at IS NULL", "WHERE ci.deleted_at IS NULL AND ci.tenant_id = ?", 1)
		// tenant_id placeholder is injected first in WHERE clause, so args must start with tenantID.
		listArgs = []interface{}{tenantID, asOf, search, like, like, like, params.Limit, params.Offset}
	}
	if err := r.db.WithContext(ctx).Raw(listSQL, listArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

type apAgingScanRow struct {
	InvoiceID     string         `gorm:"column:invoice_id"`
	SourceType    string         `gorm:"column:source_type"`
	Code          string         `gorm:"column:code"`
	InvoiceNumber string         `gorm:"column:invoice_number"`
	InvoiceDate   sql.NullTime   `gorm:"column:invoice_date"`
	DueDate       sql.NullTime   `gorm:"column:due_date"`
	SupplierID    string         `gorm:"column:supplier_id"`
	SupplierName  string         `gorm:"column:supplier_name"`
	Amount        float64        `gorm:"column:amount"`
	PaidAmount    float64        `gorm:"column:paid_amount"`
	Remaining     float64        `gorm:"column:remaining_amount"`
}

func (r *agingReportRepository) ListAPAging(ctx context.Context, params AgingListParams) ([]APAgingRow, int64, error) {
	var total int64
	var scanRows []apAgingScanRow

	asOf := params.AsOfDate.Format("2006-01-02")
	search := strings.TrimSpace(params.Search)

	tenantID := strings.TrimSpace(middleware.TenantFromContext(ctx))

	// supplier_invoices invoice_date and due_date are date columns.
	countSQL := `
		SELECT COUNT(*)
		FROM (
			SELECT si.id
			FROM supplier_invoices si
			LEFT JOIN purchase_payments pp
				ON pp.supplier_invoice_id = si.id
				AND pp.status = 'CONFIRMED'
				AND pp.deleted_at IS NULL
			WHERE si.deleted_at IS NULL
				AND si.status IN ('UNPAID','PARTIAL')
				AND si.invoice_date <= ?::date
				AND (
					? = ''
					OR si.code ILIKE ?
					OR si.invoice_number ILIKE ?
				)
			GROUP BY si.id, si.code, si.invoice_number, si.invoice_date, si.due_date, si.supplier_id, si.amount
			HAVING GREATEST(si.amount - COALESCE(SUM(pp.amount), 0), 0) > 0
		) x
	`
	like := "%" + search + "%"
	countArgs2 := []interface{}{asOf, search, like, like}
	if tenantID != "" {
		// inject tenant filter into inner subquery WHERE
		countSQL = strings.Replace(countSQL, "WHERE si.deleted_at IS NULL", "WHERE si.deleted_at IS NULL AND si.tenant_id = ?", 1)
		countArgs2 = []interface{}{tenantID, asOf, search, like, like}
	}
	if err := r.db.WithContext(ctx).Raw(countSQL, countArgs2...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	listSQL := `
		SELECT
			si.id as invoice_id,
			'SUPPLIER_INVOICE' as source_type,
			si.code,
			si.invoice_number,
			si.invoice_date,
			si.due_date,
			si.supplier_id,
			COALESCE(s.name, '') as supplier_name,
			si.amount,
			COALESCE(SUM(pp.amount), 0) as paid_amount,
			GREATEST(si.amount - COALESCE(SUM(pp.amount), 0), 0) as remaining_amount
		FROM supplier_invoices si
		LEFT JOIN suppliers s ON s.id = si.supplier_id AND s.deleted_at IS NULL
		LEFT JOIN purchase_payments pp
			ON pp.supplier_invoice_id = si.id
			AND pp.status = 'CONFIRMED'
			AND pp.deleted_at IS NULL
		WHERE si.deleted_at IS NULL
			AND si.status IN ('UNPAID','PARTIAL')
			AND si.invoice_date <= ?::date
			AND (
				? = ''
				OR si.code ILIKE ?
				OR si.invoice_number ILIKE ?
				OR s.name ILIKE ?
			)
		GROUP BY si.id, si.code, si.invoice_number, si.invoice_date, si.due_date, si.supplier_id, s.name, si.amount
		HAVING GREATEST(si.amount - COALESCE(SUM(pp.amount), 0), 0) > 0
		ORDER BY si.invoice_date DESC
		LIMIT ? OFFSET ?
	`
	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Offset < 0 {
		params.Offset = 0
	}
	listArgs2 := []interface{}{asOf, search, like, like, like, params.Limit, params.Offset}
	if tenantID != "" {
		listSQL = strings.Replace(listSQL, "WHERE si.deleted_at IS NULL", "WHERE si.deleted_at IS NULL AND si.tenant_id = ?", 1)
		listArgs2 = []interface{}{tenantID, asOf, search, like, like, like, params.Limit, params.Offset}
	}
	if err := r.db.WithContext(ctx).Raw(listSQL, listArgs2...).Scan(&scanRows).Error; err != nil {
		return nil, 0, err
	}

	rows := make([]APAgingRow, 0, len(scanRows))
	for _, sr := range scanRows {
		if !sr.InvoiceDate.Valid {
			continue
		}
		invDate := sr.InvoiceDate.Time
		dueDate := invDate
		if sr.DueDate.Valid {
			dueDate = sr.DueDate.Time
		}
		rows = append(rows, APAgingRow{
			InvoiceID:       sr.InvoiceID,
			SourceType:      sr.SourceType,
			Code:            sr.Code,
			InvoiceNumber:   sr.InvoiceNumber,
			InvoiceDate:     invDate,
			DueDate:         dueDate,
			SupplierID:      sr.SupplierID,
			SupplierName:    sr.SupplierName,
			Amount:          sr.Amount,
			PaidAmount:      sr.PaidAmount,
			RemainingAmount: sr.Remaining,
		})
	}

	return rows, total, nil
}

func (r *agingReportRepository) ListARAgingFinance(ctx context.Context, params AgingFinanceListParams) ([]ARAgingRow, error) {
	rows, _, err := r.ListARAging(ctx, AgingListParams{
		Search:   params.Search,
		AsOfDate: params.AsOfDate,
		Limit:    10000,
		Offset:   0,
	})
	if err != nil {
		return nil, err
	}

	partnerID := strings.TrimSpace(params.PartnerID)
	filtered := make([]ARAgingRow, 0, len(rows))
	for _, row := range rows {
		if partnerID != "" && strings.TrimSpace(row.CustomerID) != partnerID {
			continue
		}
		if params.MinAmount > 0 && row.RemainingAmount < params.MinAmount {
			continue
		}
		if strings.TrimSpace(row.SourceType) == "" {
			row.SourceType = "CUSTOMER_INVOICE"
		}
		filtered = append(filtered, row)
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].DueDate.Equal(filtered[j].DueDate) {
			return filtered[i].Code < filtered[j].Code
		}
		return filtered[i].DueDate.Before(filtered[j].DueDate)
	})

	return filtered, nil
}

func (r *agingReportRepository) listNonTradePayableAging(ctx context.Context, params AgingFinanceListParams) ([]APAgingRow, error) {
	asOf := params.AsOfDate.Format("2006-01-02")
	search := strings.TrimSpace(params.Search)
	like := "%" + search + "%"
	tenantID := strings.TrimSpace(middleware.TenantFromContext(ctx))

	query := `
		SELECT
			ntp.id as invoice_id,
			'NON_TRADE_PAYABLE' as source_type,
			ntp.code,
			'' as invoice_number,
			ntp.transaction_date as invoice_date,
			COALESCE(ntp.due_date, ntp.transaction_date) as due_date,
			ntp.id as supplier_id,
			COALESCE(NULLIF(ntp.vendor_name, ''), ntp.code) as supplier_name,
			ntp.amount as amount,
			COALESCE(paid.paid_amount, 0) as paid_amount,
			GREATEST(ntp.amount - COALESCE(paid.paid_amount, 0), 0) as remaining_amount
		FROM non_trade_payables ntp
		LEFT JOIN LATERAL (
			SELECT SUM(je.debit_total) AS paid_amount
			FROM journal_entries je
			WHERE je.deleted_at IS NULL
				AND je.status = 'posted'
				AND UPPER(COALESCE(je.reference_type, '')) = 'NTP_PAYMENT'
				AND (je.reference = ntp.id::text OR je.reference_id = ntp.id::text)
		) paid ON true
		WHERE ntp.deleted_at IS NULL
			AND ntp.status IN ('posted', 'partial', 'approved')
			AND ntp.transaction_date <= ?::date
			AND (
				? = ''
				OR ntp.code ILIKE ?
				OR COALESCE(ntp.vendor_name, '') ILIKE ?
				OR COALESCE(ntp.reference_no, '') ILIKE ?
			)
			AND GREATEST(ntp.amount - COALESCE(paid.paid_amount, 0), 0) > 0
		ORDER BY COALESCE(ntp.due_date, ntp.transaction_date) ASC, ntp.code ASC
	`

	rows := make([]APAgingRow, 0)
	queryArgs := []interface{}{asOf, search, like, like, like}
	if tenantID != "" {
		query = strings.Replace(query, "WHERE ntp.deleted_at IS NULL", "WHERE ntp.deleted_at IS NULL AND ntp.tenant_id = ?", 1)
		queryArgs = []interface{}{tenantID, asOf, search, like, like, like}
	}
	if err := r.db.WithContext(ctx).Raw(query, queryArgs...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *agingReportRepository) ListAPAgingFinance(ctx context.Context, params AgingFinanceListParams) ([]APAgingRow, error) {
	tradeRows, _, err := r.ListAPAging(ctx, AgingListParams{
		Search:   params.Search,
		AsOfDate: params.AsOfDate,
		Limit:    10000,
		Offset:   0,
	})
	if err != nil {
		return nil, err
	}

	ntpRows, err := r.listNonTradePayableAging(ctx, params)
	if err != nil {
		return nil, err
	}

	rows := make([]APAgingRow, 0, len(tradeRows)+len(ntpRows))
	rows = append(rows, tradeRows...)
	rows = append(rows, ntpRows...)

	partnerID := strings.TrimSpace(params.PartnerID)
	filtered := make([]APAgingRow, 0, len(rows))
	for _, row := range rows {
		if partnerID != "" && strings.TrimSpace(row.SupplierID) != partnerID {
			continue
		}
		if params.MinAmount > 0 && row.RemainingAmount < params.MinAmount {
			continue
		}
		if strings.TrimSpace(row.SourceType) == "" {
			row.SourceType = "SUPPLIER_INVOICE"
		}
		filtered = append(filtered, row)
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].DueDate.Equal(filtered[j].DueDate) {
			return filtered[i].Code < filtered[j].Code
		}
		return filtered[i].DueDate.Before(filtered[j].DueDate)
	})

	return filtered, nil
}
