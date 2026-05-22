package repositories

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/core/middleware"
	crmModels "github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/general/domain/dto"
	"gorm.io/gorm"
)

// DashboardRepository defines the data access interface for all dashboard metrics.
type DashboardRepository interface {
	GetRevenueKPI(ctx context.Context, start, end time.Time) (dto.KPICard, error)
	GetOrdersKPI(ctx context.Context, start, end time.Time) (dto.KPICard, error)
	GetCustomersKPI(ctx context.Context) (dto.KPICard, error)
	GetProductsKPI(ctx context.Context) (dto.KPICard, error)
	GetEmployeeCountKPI(ctx context.Context) (dto.KPICard, error)
	GetRevenueChart(ctx context.Context, start, end time.Time) ([]dto.PeriodChartPoint, error)
	GetCostsChart(ctx context.Context, start, end time.Time) ([]dto.PeriodChartPoint, error)
	GetRevenueVsCosts(ctx context.Context, start, end time.Time) ([]dto.PeriodChartPoint, error)
	GetBalance(ctx context.Context, start, end time.Time) (dto.BalanceData, error)
	GetCostsByCategory(ctx context.Context, start, end time.Time) ([]dto.CostCategoryItem, error)
	GetInvoiceSummary(ctx context.Context, start, end time.Time) (dto.InvoiceSummaryData, error)
	GetRecentInvoices(ctx context.Context, limit int) ([]dto.InvoiceRow, error)
	GetSalesPerformance(ctx context.Context, start, end time.Time, limit int) ([]dto.SalesPerformanceRow, error)
	GetTopProducts(ctx context.Context, start, end time.Time, limit int) ([]dto.TopProductRow, error)
	GetDeliveryStatus(ctx context.Context, start, end time.Time) (dto.DeliveryStatusData, error)
	GetGeoOverview(ctx context.Context, start, end time.Time) (dto.GeoOverviewData, error)
	GetWarehouses(ctx context.Context) ([]dto.WarehouseItem, error)
	GetPOSSummary(ctx context.Context, userID, outletID string, start, end time.Time) (*dto.POSSummaryData, error)
	GetHRSummary(ctx context.Context) (*dto.HRSummaryData, error)
	GetCRMSummary(ctx context.Context) (*dto.CRMSummaryData, error)
}

type dashboardRepository struct {
	db *gorm.DB
}

// NewDashboardRepository creates a new DashboardRepository.
func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{db: db}
}

func (r *dashboardRepository) scoped(ctx context.Context, tableName string) *gorm.DB {
	db := r.db.WithContext(ctx)
	if middleware.IsSystemAdmin(ctx) {
		return db
	}
	tenantID := middleware.TenantFromContext(ctx)
	if tenantID == "" {
		return db
	}
	return db.Where(fmt.Sprintf("%s.tenant_id = ?", tableName), tenantID)
}

const (
	approvedClosedWhere = "status IN (?) AND order_date BETWEEN ? AND ? AND deleted_at IS NULL"
	journalJoinEntries  = "JOIN journal_entries ON journal_entries.id = journal_lines.journal_entry_id"
	journalJoinAccounts = "JOIN chart_of_accounts ON chart_of_accounts.id = journal_lines.chart_of_account_id"
	expenseWhere        = "chart_of_accounts.type = ? AND journal_entries.entry_date BETWEEN ? AND ? AND journal_entries.deleted_at IS NULL AND journal_entries.status = ?"
	salesOrderWhere     = "sales_orders.status IN (?) AND sales_orders.order_date BETWEEN ? AND ? AND sales_orders.deleted_at IS NULL"
)

var approvedStatuses = []string{"approved", "closed"}

func (r *dashboardRepository) GetRevenueKPI(ctx context.Context, start, end time.Time) (dto.KPICard, error) {
	outletIDs, err := resolveDashboardOutletIDs(ctx, r.db)
	if err != nil {
		return dto.KPICard{}, err
	}
	if outletIDs != nil {
		if len(outletIDs) == 0 {
			return dto.KPICard{}, nil
		}
		var current float64
		if err := r.scoped(ctx, "pos_orders").Table("pos_orders").
			Where("created_at BETWEEN ? AND ? AND deleted_at IS NULL", start, end).
			Where("outlet_id IN ?", outletIDs).
			Select("COALESCE(SUM(total_amount), 0)").Scan(&current).Error; err != nil {
			return dto.KPICard{}, err
		}
		duration := end.Sub(start)
		prevStart, prevEnd := start.Add(-duration), start.Add(-time.Second)
		var previous float64
		r.scoped(ctx, "pos_orders").Table("pos_orders").
			Where("created_at BETWEEN ? AND ? AND deleted_at IS NULL", prevStart, prevEnd).
			Where("outlet_id IN ?", outletIDs).
			Select("COALESCE(SUM(total_amount), 0)").Scan(&previous)
		return dto.KPICard{Value: current, ChangePercent: calcChange(previous, current)}, nil
	}
	if scopeWhere, scopeParams := buildScopeWhereClause(ctx, "sales_orders"); scopeWhere != "" {
		var current float64
		if err := r.scoped(ctx, "sales_orders").Table("sales_orders").
			Where("order_date BETWEEN ? AND ? AND deleted_at IS NULL", start, end).
			Where(scopeWhere, scopeParams...).
			Select("COALESCE(SUM(total_amount), 0)").Scan(&current).Error; err != nil {
			return dto.KPICard{}, err
		}
		duration := end.Sub(start)
		prevStart, prevEnd := start.Add(-duration), start.Add(-time.Second)
		var previous float64
		r.scoped(ctx, "sales_orders").Table("sales_orders").
			Where("order_date BETWEEN ? AND ? AND deleted_at IS NULL", prevStart, prevEnd).
			Where(scopeWhere, scopeParams...).
			Select("COALESCE(SUM(total_amount), 0)").Scan(&previous)
		return dto.KPICard{Value: current, ChangePercent: calcChange(previous, current)}, nil
	}

	var current float64
	if err := r.scoped(ctx, "sales_orders").Table("sales_orders").
		Where(approvedClosedWhere, approvedStatuses, start, end).
		Select("COALESCE(SUM(total_amount), 0)").Scan(&current).Error; err != nil {
		return dto.KPICard{}, err
	}
	duration := end.Sub(start)
	prevStart, prevEnd := start.Add(-duration), start.Add(-time.Second)
	var previous float64
	r.scoped(ctx, "sales_orders").Table("sales_orders").
		Where(approvedClosedWhere, approvedStatuses, prevStart, prevEnd).
		Select("COALESCE(SUM(total_amount), 0)").Scan(&previous)
	return dto.KPICard{Value: current, ChangePercent: calcChange(previous, current)}, nil
}

func (r *dashboardRepository) GetOrdersKPI(ctx context.Context, start, end time.Time) (dto.KPICard, error) {
	outletIDs, err := resolveDashboardOutletIDs(ctx, r.db)
	if err != nil {
		return dto.KPICard{}, err
	}
	if outletIDs != nil {
		if len(outletIDs) == 0 {
			return dto.KPICard{}, nil
		}
		var current int64
		if err := r.scoped(ctx, "pos_orders").Table("pos_orders").
			Where("created_at BETWEEN ? AND ? AND deleted_at IS NULL", start, end).
			Where("outlet_id IN ?", outletIDs).
			Count(&current).Error; err != nil {
			return dto.KPICard{}, err
		}
		duration := end.Sub(start)
		prevStart, prevEnd := start.Add(-duration), start.Add(-time.Second)
		var previous int64
		r.scoped(ctx, "pos_orders").Table("pos_orders").
			Where("created_at BETWEEN ? AND ? AND deleted_at IS NULL", prevStart, prevEnd).
			Where("outlet_id IN ?", outletIDs).
			Count(&previous)
		return dto.KPICard{Value: float64(current), ChangePercent: calcChange(float64(previous), float64(current))}, nil
	}
	if scopeWhere, scopeParams := buildScopeWhereClause(ctx, "sales_orders"); scopeWhere != "" {
		var current int64
		if err := r.scoped(ctx, "sales_orders").Table("sales_orders").
			Where("order_date BETWEEN ? AND ? AND deleted_at IS NULL", start, end).
			Where(scopeWhere, scopeParams...).
			Count(&current).Error; err != nil {
			return dto.KPICard{}, err
		}
		duration := end.Sub(start)
		prevStart, prevEnd := start.Add(-duration), start.Add(-time.Second)
		var previous int64
		r.scoped(ctx, "sales_orders").Table("sales_orders").
			Where("order_date BETWEEN ? AND ? AND deleted_at IS NULL", prevStart, prevEnd).
			Where(scopeWhere, scopeParams...).
			Count(&previous)
		return dto.KPICard{Value: float64(current), ChangePercent: calcChange(float64(previous), float64(current))}, nil
	}

	var current int64
	if err := r.scoped(ctx, "sales_orders").Table("sales_orders").
		Where("order_date BETWEEN ? AND ? AND deleted_at IS NULL", start, end).
		Count(&current).Error; err != nil {
		return dto.KPICard{}, err
	}
	duration := end.Sub(start)
	prevStart, prevEnd := start.Add(-duration), start.Add(-time.Second)
	var previous int64
	r.scoped(ctx, "sales_orders").Table("sales_orders").
		Where("order_date BETWEEN ? AND ? AND deleted_at IS NULL", prevStart, prevEnd).
		Count(&previous)
	return dto.KPICard{Value: float64(current), ChangePercent: calcChange(float64(previous), float64(current))}, nil
}

func (r *dashboardRepository) GetCustomersKPI(ctx context.Context) (dto.KPICard, error) {
	outletIDs, err := resolveDashboardOutletIDs(ctx, r.db)
	if err != nil {
		return dto.KPICard{}, err
	}
	if outletIDs != nil {
		if len(outletIDs) == 0 {
			return dto.KPICard{}, nil
		}
		var count int64
		if err := r.scoped(ctx, "customers").Table("customers").
			Joins("JOIN pos_orders ON pos_orders.customer_id = customers.id AND pos_orders.deleted_at IS NULL").
			Where("customers.deleted_at IS NULL").
			Where("pos_orders.outlet_id IN ?", outletIDs).
			Distinct("customers.id").
			Count(&count).Error; err != nil {
			return dto.KPICard{}, err
		}
		return dto.KPICard{Value: float64(count)}, nil
	}
	if scopeWhere, scopeParams := buildScopeWhereClause(ctx, "sales_orders"); scopeWhere != "" {
		var count int64
		if err := r.scoped(ctx, "customers").Table("customers").
			Joins("JOIN sales_orders ON sales_orders.customer_id = customers.id AND sales_orders.deleted_at IS NULL").
			Where("customers.deleted_at IS NULL").
			Where(scopeWhere, scopeParams...).
			Distinct("customers.id").
			Count(&count).Error; err != nil {
			return dto.KPICard{}, err
		}
		return dto.KPICard{Value: float64(count)}, nil
	}

	var count int64
	if err := r.scoped(ctx, "customers").Table("customers").
		Where("is_active = ? AND deleted_at IS NULL", true).Count(&count).Error; err != nil {
		return dto.KPICard{}, err
	}
	return dto.KPICard{Value: float64(count)}, nil
}

func (r *dashboardRepository) GetProductsKPI(ctx context.Context) (dto.KPICard, error) {
	outletIDs, err := resolveDashboardOutletIDs(ctx, r.db)
	if err != nil {
		return dto.KPICard{}, err
	}
	if outletIDs != nil {
		if len(outletIDs) == 0 {
			return dto.KPICard{}, nil
		}
		var count int64
		if err := r.scoped(ctx, "pos_order_items").Table("pos_order_items").
			Joins("JOIN pos_orders ON pos_orders.id = pos_order_items.pos_order_id AND pos_orders.deleted_at IS NULL").
			Where("pos_orders.outlet_id IN ?", outletIDs).
			Distinct("pos_order_items.product_id").
			Count(&count).Error; err != nil {
			return dto.KPICard{}, err
		}
		return dto.KPICard{Value: float64(count)}, nil
	}
	if scopeWhere, scopeParams := buildScopeWhereClause(ctx, "sales_orders"); scopeWhere != "" {
		var count int64
		if err := r.scoped(ctx, "sales_order_items").Table("sales_order_items").
			Joins("JOIN sales_orders ON sales_orders.id = sales_order_items.sales_order_id AND sales_orders.deleted_at IS NULL").
			Where(scopeWhere, scopeParams...).
			Distinct("sales_order_items.product_id").
			Count(&count).Error; err != nil {
			return dto.KPICard{}, err
		}
		return dto.KPICard{Value: float64(count)}, nil
	}

	var count int64
	if err := r.scoped(ctx, "products").Table("products").
		Where("status = ? AND deleted_at IS NULL", "approved").Count(&count).Error; err != nil {
		return dto.KPICard{}, err
	}
	return dto.KPICard{Value: float64(count)}, nil
}

func (r *dashboardRepository) GetEmployeeCountKPI(ctx context.Context) (dto.KPICard, error) {
	outletIDs, err := resolveDashboardOutletIDs(ctx, r.db)
	if err != nil {
		return dto.KPICard{}, err
	}
	if outletIDs != nil {
		if len(outletIDs) == 0 {
			return dto.KPICard{}, nil
		}
		var count int64
		if err := r.scoped(ctx, "employees").Table("employees").
			Joins("JOIN employee_outlets ON employee_outlets.employee_id = employees.id AND employee_outlets.deleted_at IS NULL").
			Where("employee_outlets.outlet_id IN ?", outletIDs).
			Where("employees.is_active = ? AND employees.deleted_at IS NULL", true).
			Distinct("employees.id").
			Count(&count).Error; err != nil {
			return dto.KPICard{}, err
		}
		return dto.KPICard{Value: float64(count)}, nil
	}
	if scopeWhere, scopeParams := buildScopeWhereClause(ctx, "employees"); scopeWhere != "" {
		var count int64
		if err := r.scoped(ctx, "employees").Table("employees").
			Where("employees.is_active = ? AND employees.deleted_at IS NULL", true).
			Where(scopeWhere, scopeParams...).
			Distinct("employees.id").
			Count(&count).Error; err != nil {
			return dto.KPICard{}, err
		}
		return dto.KPICard{Value: float64(count)}, nil
	}

	var count int64
	if err := r.scoped(ctx, "employees").Table("employees").
		Where("is_active = ? AND deleted_at IS NULL", true).Count(&count).Error; err != nil {
		return dto.KPICard{}, err
	}
	return dto.KPICard{Value: float64(count)}, nil
}

func (r *dashboardRepository) GetRevenueChart(ctx context.Context, start, end time.Time) ([]dto.PeriodChartPoint, error) {
	type row struct {
		Period string
		Amount float64
	}
	var rows []row
	outletIDs, err := resolveDashboardOutletIDs(ctx, r.db)
	if err != nil {
		return nil, err
	}
	if outletIDs != nil {
		if len(outletIDs) == 0 {
			return []dto.PeriodChartPoint{}, nil
		}
		if err := r.scoped(ctx, "pos_orders").Table("pos_orders").
			Select("TO_CHAR(created_at, 'YYYY-MM') as period, COALESCE(SUM(total_amount), 0) as amount").
			Where("created_at BETWEEN ? AND ? AND deleted_at IS NULL", start, end).
			Where("outlet_id IN ?", outletIDs).
			Group("period").Order("period ASC").Scan(&rows).Error; err != nil {
			return nil, err
		}
		points := make([]dto.PeriodChartPoint, 0, len(rows))
		for _, row := range rows {
			points = append(points, dto.PeriodChartPoint{Period: row.Period, Revenue: row.Amount})
		}
		return points, nil
	}
	if scopeWhere, scopeParams := buildScopeWhereClause(ctx, "sales_orders"); scopeWhere != "" {
		if err := r.scoped(ctx, "sales_orders").Table("sales_orders").
			Select("TO_CHAR(order_date, 'YYYY-MM') as period, COALESCE(SUM(total_amount), 0) as amount").
			Where(approvedClosedWhere, approvedStatuses, start, end).
			Where(scopeWhere, scopeParams...).
			Group("period").Order("period ASC").Scan(&rows).Error; err != nil {
			return nil, err
		}
		points := make([]dto.PeriodChartPoint, 0, len(rows))
		for _, row := range rows {
			points = append(points, dto.PeriodChartPoint{Period: row.Period, Revenue: row.Amount})
		}
		return points, nil
	}
	if err := r.scoped(ctx, "sales_orders").Table("sales_orders").
		Select("TO_CHAR(order_date, 'YYYY-MM') as period, COALESCE(SUM(total_amount), 0) as amount").
		Where(approvedClosedWhere, approvedStatuses, start, end).
		Group("period").Order("period ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	points := make([]dto.PeriodChartPoint, 0, len(rows))
	for _, row := range rows {
		points = append(points, dto.PeriodChartPoint{Period: row.Period, Revenue: row.Amount})
	}
	return points, nil
}

func (r *dashboardRepository) GetCostsChart(ctx context.Context, start, end time.Time) ([]dto.PeriodChartPoint, error) {
	type row struct {
		Period string
		Amount float64
	}
	var rows []row
	if scopeWhere, scopeParams := buildScopeWhereClause(ctx, "journal_entries"); scopeWhere != "" {
		if err := r.scoped(ctx, "journal_lines").Table("journal_lines").
			Joins(journalJoinEntries).Joins(journalJoinAccounts).
			Select("TO_CHAR(journal_entries.entry_date, 'YYYY-MM') as period, COALESCE(SUM(journal_lines.debit), 0) as amount").
			Where(expenseWhere, "EXPENSE", start, end, "posted").
			Where(scopeWhere, scopeParams...).
			Group("period").Order("period ASC").Scan(&rows).Error; err != nil {
			return nil, err
		}
		points := make([]dto.PeriodChartPoint, 0, len(rows))
		for _, row := range rows {
			points = append(points, dto.PeriodChartPoint{Period: row.Period, Costs: row.Amount})
		}
		return points, nil
	}
	if err := r.scoped(ctx, "journal_lines").Table("journal_lines").
		Joins(journalJoinEntries).Joins(journalJoinAccounts).
		Select("TO_CHAR(journal_entries.entry_date, 'YYYY-MM') as period, COALESCE(SUM(journal_lines.debit), 0) as amount").
		Where(expenseWhere, "EXPENSE", start, end, "posted").
		Group("period").Order("period ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	points := make([]dto.PeriodChartPoint, 0, len(rows))
	for _, row := range rows {
		points = append(points, dto.PeriodChartPoint{Period: row.Period, Costs: row.Amount})
	}
	return points, nil
}

func (r *dashboardRepository) GetRevenueVsCosts(ctx context.Context, start, end time.Time) ([]dto.PeriodChartPoint, error) {
	revenue, err := r.GetRevenueChart(ctx, start, end)
	if err != nil {
		return nil, err
	}
	costs, err := r.GetCostsChart(ctx, start, end)
	if err != nil {
		return nil, err
	}
	merged := make(map[string]*dto.PeriodChartPoint)
	for i := range revenue {
		p := revenue[i]
		merged[p.Period] = &dto.PeriodChartPoint{Period: p.Period, Revenue: p.Revenue}
	}
	for i := range costs {
		p := costs[i]
		if existing, ok := merged[p.Period]; ok {
			existing.Costs = p.Costs
		} else {
			merged[p.Period] = &dto.PeriodChartPoint{Period: p.Period, Costs: p.Costs}
		}
	}
	result := make([]dto.PeriodChartPoint, 0, len(merged))
	for _, v := range merged {
		result = append(result, *v)
	}
	sortPeriodChartPoints(result)
	return result, nil
}

func (r *dashboardRepository) GetBalance(ctx context.Context, start, end time.Time) (dto.BalanceData, error) {
	var current float64
	if scopeWhere, scopeParams := buildScopeWhereClause(ctx, "journal_entries"); scopeWhere != "" {
		r.scoped(ctx, "journal_lines").Table("journal_lines").
			Joins(journalJoinEntries).Joins(journalJoinAccounts).
			Select("COALESCE(SUM(journal_lines.debit) - SUM(journal_lines.credit), 0)").
			Where("chart_of_accounts.type IN ? AND journal_entries.deleted_at IS NULL AND journal_entries.status = ?",
				[]string{"CASH_BANK", "ASSET", "CURRENT_ASSET"}, "posted").
			Where(scopeWhere, scopeParams...).
			Scan(&current)
	} else {
		r.scoped(ctx, "journal_lines").Table("journal_lines").
		Joins(journalJoinEntries).Joins(journalJoinAccounts).
		Select("COALESCE(SUM(journal_lines.debit) - SUM(journal_lines.credit), 0)").
		Where("chart_of_accounts.type IN ? AND journal_entries.deleted_at IS NULL AND journal_entries.status = ?",
			[]string{"CASH_BANK", "ASSET", "CURRENT_ASSET"}, "posted").
			Scan(&current)
	}

	type row struct {
		Period string
		Amount float64
	}
	var rows []row
	if scopeWhere, scopeParams := buildScopeWhereClause(ctx, "journal_entries"); scopeWhere != "" {
		r.scoped(ctx, "journal_lines").Table("journal_lines").
			Joins(journalJoinEntries).Joins(journalJoinAccounts).
			Select("TO_CHAR(journal_entries.entry_date, 'YYYY-MM') as period, COALESCE(SUM(journal_lines.debit) - SUM(journal_lines.credit), 0) as amount").
			Where("chart_of_accounts.type IN ? AND journal_entries.entry_date BETWEEN ? AND ? AND journal_entries.deleted_at IS NULL AND journal_entries.status = ?",
				[]string{"CASH_BANK", "ASSET", "CURRENT_ASSET"}, start, end, "posted").
			Where(scopeWhere, scopeParams...).
			Group("period").Order("period ASC").Scan(&rows)
	} else {
		r.scoped(ctx, "journal_lines").Table("journal_lines").
		Joins(journalJoinEntries).Joins(journalJoinAccounts).
		Select("TO_CHAR(journal_entries.entry_date, 'YYYY-MM') as period, COALESCE(SUM(journal_lines.debit) - SUM(journal_lines.credit), 0) as amount").
		Where("chart_of_accounts.type IN ? AND journal_entries.entry_date BETWEEN ? AND ? AND journal_entries.deleted_at IS NULL AND journal_entries.status = ?",
			[]string{"CASH_BANK", "ASSET", "CURRENT_ASSET"}, start, end, "posted").
			Group("period").Order("period ASC").Scan(&rows)
	}

	trend := make([]dto.PeriodChartPoint, 0, len(rows))
	for _, row := range rows {
		trend = append(trend, dto.PeriodChartPoint{Period: row.Period, Amount: row.Amount})
	}
	return dto.BalanceData{Current: current, Trend: trend}, nil
}

func (r *dashboardRepository) GetCostsByCategory(ctx context.Context, start, end time.Time) ([]dto.CostCategoryItem, error) {
	type row struct {
		Category string
		Amount   float64
	}
	var rows []row
	query := r.scoped(ctx, "journal_lines").Table("journal_lines").
		Joins(journalJoinEntries).Joins(journalJoinAccounts).
		Select("chart_of_accounts.name as category, COALESCE(SUM(journal_lines.debit), 0) as amount").
		Where(expenseWhere, "EXPENSE", start, end, "posted").
		Group("chart_of_accounts.name").Order("amount DESC").Limit(10)
	if scopeWhere, scopeParams := buildScopeWhereClause(ctx, "journal_entries"); scopeWhere != "" {
		query = query.Where(scopeWhere, scopeParams...)
	}
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}
	var total float64
	for _, row := range rows {
		total += row.Amount
	}
	items := make([]dto.CostCategoryItem, 0, len(rows))
	for _, row := range rows {
		pct := float64(0)
		if total > 0 {
			pct = (row.Amount / total) * 100
		}
		items = append(items, dto.CostCategoryItem{Category: row.Category, Amount: row.Amount, Percentage: pct})
	}
	return items, nil
}

func (r *dashboardRepository) GetInvoiceSummary(ctx context.Context, start, end time.Time) (dto.InvoiceSummaryData, error) {
	var summary dto.InvoiceSummaryData
	var total, paid, unpaid, overdue int64
	outletIDs, err := resolveDashboardOutletIDs(ctx, r.db)
	if err != nil {
		return dto.InvoiceSummaryData{}, err
	}
	if outletIDs != nil {
		if len(outletIDs) == 0 {
			return summary, nil
		}
		base := r.scoped(ctx, "customer_invoices").Table("customer_invoices").
			Joins("JOIN sales_orders ON sales_orders.id = customer_invoices.sales_order_id").
			Joins("JOIN pos_orders ON pos_orders.sales_order_id = sales_orders.id AND pos_orders.deleted_at IS NULL").
			Where("customer_invoices.deleted_at IS NULL").
			Where("pos_orders.outlet_id IN ?", outletIDs)

		base.Where("invoice_date BETWEEN ? AND ?", start, end).Count(&total)
		base.Where("invoice_date BETWEEN ? AND ? AND status = ?", start, end, "paid").Count(&paid)
		base.Where("invoice_date BETWEEN ? AND ? AND status IN (?)", start, end, []string{"unpaid", "draft", "sent", "approved"}).Count(&unpaid)
		base.Where("invoice_date BETWEEN ? AND ? AND status NOT IN (?) AND due_date < NOW()",
			start, end, []string{"paid", "cancelled"}).Count(&overdue)
	} else {
		scopeWhere, scopeParams := buildScopeWhereClause(ctx, "sales_orders")
		if scopeWhere != "" {
			base := r.scoped(ctx, "customer_invoices").Table("customer_invoices").
				Joins("JOIN sales_orders ON sales_orders.id = customer_invoices.sales_order_id").
				Where("customer_invoices.deleted_at IS NULL").
				Where(scopeWhere, scopeParams...)
			base.Where("invoice_date BETWEEN ? AND ?", start, end).Count(&total)
			base.Where("invoice_date BETWEEN ? AND ? AND status = ?", start, end, "paid").Count(&paid)
			base.Where("invoice_date BETWEEN ? AND ? AND status IN (?)", start, end, []string{"unpaid", "draft", "sent", "approved"}).Count(&unpaid)
			base.Where("invoice_date BETWEEN ? AND ? AND status NOT IN (?) AND due_date < NOW()",
				start, end, []string{"paid", "cancelled"}).Count(&overdue)
		} else {
		r.scoped(ctx, "customer_invoices").Table("customer_invoices").
			Where("invoice_date BETWEEN ? AND ? AND deleted_at IS NULL", start, end).Count(&total)
		r.scoped(ctx, "customer_invoices").Table("customer_invoices").
			Where("invoice_date BETWEEN ? AND ? AND deleted_at IS NULL AND status = ?", start, end, "paid").Count(&paid)
		r.scoped(ctx, "customer_invoices").Table("customer_invoices").
			Where("invoice_date BETWEEN ? AND ? AND deleted_at IS NULL AND status IN (?)", start, end, []string{"unpaid", "draft", "sent", "approved"}).Count(&unpaid)
		r.scoped(ctx, "customer_invoices").Table("customer_invoices").
			Where("invoice_date BETWEEN ? AND ? AND deleted_at IS NULL AND status NOT IN (?) AND due_date < NOW()",
				start, end, []string{"paid", "cancelled"}).Count(&overdue)
		}
	}
	summary.Total = int(total)
	summary.Paid = int(paid)
	summary.Unpaid = int(unpaid)
	summary.Overdue = int(overdue)
	return summary, nil
}

func (r *dashboardRepository) GetRecentInvoices(ctx context.Context, limit int) ([]dto.InvoiceRow, error) {
	type row struct {
		ID           string
		Code         string
		CustomerID   string
		CustomerName string
		Amount       float64
		Status       string
		DueDate      *time.Time
	}
	var rows []row
	query := r.scoped(ctx, "customer_invoices").Table("customer_invoices").
		Joins("LEFT JOIN sales_orders ON sales_orders.id = customer_invoices.sales_order_id").
		Select("customer_invoices.id, customer_invoices.code, COALESCE(sales_orders.customer_id::text, '') as customer_id, COALESCE(sales_orders.customer_name, '') as customer_name, customer_invoices.amount, customer_invoices.status, customer_invoices.due_date").
		Where("customer_invoices.deleted_at IS NULL")

	outletIDs, err := resolveDashboardOutletIDs(ctx, r.db)
	if err != nil {
		return nil, err
	}
	if outletIDs != nil {
		if len(outletIDs) == 0 {
			return []dto.InvoiceRow{}, nil
		}
		query = query.
			Joins("JOIN pos_orders ON pos_orders.sales_order_id = sales_orders.id AND pos_orders.deleted_at IS NULL").
			Where("pos_orders.outlet_id IN ?", outletIDs)
	} else {
		// Apply scope-based filtering via sales_orders
		scopeWhere, scopeParams := buildScopeWhereClause(ctx, "sales_orders")
		if scopeWhere != "" {
			query = query.Where(scopeWhere, scopeParams...)
		}
	}

	if err := query.Order("customer_invoices.created_at DESC").Limit(limit).Scan(&rows).Error; err != nil {
		return nil, err
	}
	invoices := make([]dto.InvoiceRow, 0, len(rows))
	for _, row := range rows {
		dueStr := ""
		if row.DueDate != nil {
			dueStr = row.DueDate.Format("2006-01-02")
		}
		invoices = append(invoices, dto.InvoiceRow{
			ID:         row.ID,
			CustomerID: row.CustomerID,
			Company:    row.CustomerName,
			Contact:    row.Code,
			IssueDate:  dueStr,
			Value:      row.Amount,
			Status:     normalizeInvoiceStatus(row.Status),
		})
	}
	return invoices, nil
}

func normalizeInvoiceStatus(status string) string {
	switch status {
	case "paid":
		return "paid"
	case "overdue":
		return "overdue"
	default:
		return "unpaid"
	}
}

func (r *dashboardRepository) GetSalesPerformance(ctx context.Context, start, end time.Time, limit int) ([]dto.SalesPerformanceRow, error) {
	type row struct {
		EmployeeID string
		Name       string
		Revenue    float64
		Orders     int64
	}
	var rows []row
	outletIDs, err := resolveDashboardOutletIDs(ctx, r.db)
	if err != nil {
		return nil, err
	}
	if outletIDs != nil {
		if len(outletIDs) == 0 {
			return []dto.SalesPerformanceRow{}, nil
		}
		// No outlet-aware sales performance mapping yet; return empty to avoid cross-outlet leakage.
		return []dto.SalesPerformanceRow{}, nil
	}
	query := r.scoped(ctx, "sales_orders").Table("sales_orders").
		Joins("JOIN employees ON employees.id = sales_orders.sales_rep_id").
		Select("employees.id::text as employee_id, employees.name as name, COALESCE(SUM(sales_orders.total_amount), 0) as revenue, COUNT(sales_orders.id) as orders").
		Where(salesOrderWhere, approvedStatuses, start, end)

	// Apply scope-based filtering
	scopeWhere, scopeParams := buildScopeWhereClause(ctx, "employees")
	if scopeWhere != "" {
		query = query.Where(scopeWhere, scopeParams...)
	}

	if err := query.Group("employees.id, employees.name").Order("revenue DESC").Limit(limit).Scan(&rows).Error; err != nil {
		return nil, err
	}
	maxRevenue := float64(0)
	for _, row := range rows {
		if row.Revenue > maxRevenue {
			maxRevenue = row.Revenue
		}
	}
	perfs := make([]dto.SalesPerformanceRow, 0, len(rows))
	for _, row := range rows {
		targetPct := float64(0)
		if maxRevenue > 0 {
			targetPct = (row.Revenue / maxRevenue) * 100
		}
		perfs = append(perfs, dto.SalesPerformanceRow{
			ID:            row.EmployeeID,
			Name:          row.Name,
			Revenue:       row.Revenue,
			Orders:        int(row.Orders),
			TargetPercent: targetPct,
		})
	}
	return perfs, nil
}

func (r *dashboardRepository) GetTopProducts(ctx context.Context, start, end time.Time, limit int) ([]dto.TopProductRow, error) {
	type row struct {
		ProductID    string
		Name         string
		SKU          string
		QuantitySold float64
		Revenue      float64
	}
	var rows []row
	outletIDs, err := resolveDashboardOutletIDs(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var query *gorm.DB
	if outletIDs != nil {
		if len(outletIDs) == 0 {
			return []dto.TopProductRow{}, nil
		}
		query = r.scoped(ctx, "pos_order_items").Table("pos_order_items").
			Joins("JOIN pos_orders ON pos_orders.id = pos_order_items.pos_order_id AND pos_orders.deleted_at IS NULL").
			Select("pos_order_items.product_id::text as product_id, pos_order_items.product_name as name, COALESCE(pos_order_items.product_code, '') as sku, COALESCE(SUM(pos_order_items.quantity), 0) as quantity_sold, COALESCE(SUM(pos_order_items.subtotal), 0) as revenue").
			Where("pos_orders.created_at BETWEEN ? AND ?", start, end).
			Where("pos_orders.outlet_id IN ?", outletIDs)
	} else {
		query = r.scoped(ctx, "sales_order_items").Table("sales_order_items").
			Joins("JOIN sales_orders ON sales_orders.id = sales_order_items.sales_order_id").
			Joins("JOIN products ON products.id = sales_order_items.product_id").
			Select("products.id::text as product_id, products.name as name, COALESCE(products.sku, '') as sku, COALESCE(SUM(sales_order_items.quantity), 0) as quantity_sold, COALESCE(SUM(sales_order_items.subtotal), 0) as revenue").
			Where(salesOrderWhere, approvedStatuses, start, end)

		// Apply scope-based filtering
		scopeWhere, scopeParams := buildScopeWhereClause(ctx, "sales_orders")
		if scopeWhere != "" {
			query = query.Where(scopeWhere, scopeParams...)
		}
	}

	if err := query.Group("products.id, products.name, products.sku").Order("revenue DESC").Limit(limit).Scan(&rows).Error; err != nil {
		return nil, err
	}
	products := make([]dto.TopProductRow, 0, len(rows))
	for _, row := range rows {
		products = append(products, dto.TopProductRow{
			ID:           row.ProductID,
			Name:         row.Name,
			SKU:          row.SKU,
			QuantitySold: row.QuantitySold,
			Revenue:      row.Revenue,
		})
	}
	return products, nil
}

func (r *dashboardRepository) GetDeliveryStatus(ctx context.Context, start, end time.Time) (dto.DeliveryStatusData, error) {
	var data dto.DeliveryStatusData
	var total, pending, inTransit, delivered int64
	r.scoped(ctx, "delivery_orders").Table("delivery_orders").
		Where("delivery_date BETWEEN ? AND ? AND deleted_at IS NULL", start, end).Count(&total)
	r.scoped(ctx, "delivery_orders").Table("delivery_orders").
		Where("delivery_date BETWEEN ? AND ? AND deleted_at IS NULL AND status IN (?)",
			start, end, []string{"draft", "sent", "approved", "prepared"}).Count(&pending)
	r.scoped(ctx, "delivery_orders").Table("delivery_orders").
		Where("delivery_date BETWEEN ? AND ? AND deleted_at IS NULL AND status = ?", start, end, "shipped").Count(&inTransit)
	r.scoped(ctx, "delivery_orders").Table("delivery_orders").
		Where("delivery_date BETWEEN ? AND ? AND deleted_at IS NULL AND status = ?", start, end, "delivered").Count(&delivered)
	data.Total = int(total)
	data.Pending = int(pending)
	data.InTransit = int(inTransit)
	data.Delivered = int(delivered)
	return data, nil
}

func (r *dashboardRepository) GetGeoOverview(ctx context.Context, start, end time.Time) (dto.GeoOverviewData, error) {
	type row struct {
		Code  string
		Name  string
		Count int
		Value float64
	}
	var rows []row
	query := r.scoped(ctx, "sales_orders").Table("sales_orders").
		Joins("JOIN customers ON customers.id = sales_orders.customer_id").
		Joins("LEFT JOIN provinces ON provinces.id = customers.province_id").
		Select("COALESCE(provinces.id::text, 'unknown') as code, COALESCE(provinces.name, 'Unknown') as name, COUNT(sales_orders.id) as count, COALESCE(SUM(sales_orders.total_amount), 0) as value").
		Where(salesOrderWhere, approvedStatuses, start, end).
		Group("provinces.id, provinces.name").Order("value DESC")
	if scopeWhere, scopeParams := buildScopeWhereClause(ctx, "sales_orders"); scopeWhere != "" {
		query = query.Where(scopeWhere, scopeParams...)
	}
	if err := query.Scan(&rows).Error; err != nil {
		return dto.GeoOverviewData{}, err
	}
	regions := make([]dto.GeoRegionData, 0, len(rows))
	var totalValue float64
	for _, row := range rows {
		totalValue += row.Value
		regions = append(regions, dto.GeoRegionData{
			Code:  row.Code,
			Name:  row.Name,
			Value: row.Value,
			Count: row.Count,
		})
	}
	return dto.GeoOverviewData{Regions: regions, TotalValue: totalValue}, nil
}

func (r *dashboardRepository) GetWarehouses(ctx context.Context) ([]dto.WarehouseItem, error) {
	tenantID := contextString(ctx, "tenant_id")
	isSystemAdmin := middleware.IsSystemAdmin(ctx)
	stockTenantFilter := "1=1"
	warehouseStatsTenantFilter := "1=1"
	warehouseTenantFilter := "1=1"
	params := []any{}
	if !isSystemAdmin && tenantID != "" {
		stockTenantFilter = "ib.tenant_id = ? AND p.tenant_id = ?"
		warehouseStatsTenantFilter = "ib2.tenant_id = ?"
		warehouseTenantFilter = "w.tenant_id = ?"
		params = append(params, tenantID, tenantID, tenantID, tenantID)
	}

	type row struct {
		ID              string
		Name            string
		Address         string
		ItemCount       int
		StockValue      float64
		InStockCount    int
		LowStockCount   int
		OutOfStockCount int
	}
	var rows []row
	// Use a CTE mirroring the inventory feature's stock-status logic:
	// out_of_stock: available <= 0
	// low_stock:    available > 0 AND available <= products.min_stock
	// in_stock:     available > products.min_stock (or min_stock = 0 and available > 0)
	query := fmt.Sprintf(`
		WITH stock_levels AS (
			SELECT
				ib.warehouse_id,
				ib.product_id,
				p.min_stock,
				COALESCE(SUM(ib.current_quantity) - SUM(ib.reserved_quantity), 0) AS available
			FROM inventory_batches ib
			JOIN products p ON p.id = ib.product_id
			WHERE ib.deleted_at IS NULL AND ib.is_active = true AND p.deleted_at IS NULL AND %s
			GROUP BY ib.warehouse_id, ib.product_id, p.min_stock
		),
		warehouse_stats AS (
			SELECT
				ib2.warehouse_id,
				COUNT(DISTINCT ib2.product_id)                         AS item_count,
				COALESCE(SUM(ib2.current_quantity * ib2.cost_price), 0) AS stock_value
			FROM inventory_batches ib2
			WHERE ib2.deleted_at IS NULL AND ib2.is_active = true AND %s
			GROUP BY ib2.warehouse_id
		)
		SELECT
			w.id,
			w.name,
			COALESCE(w.address, '')      AS address,
			COALESCE(ws.item_count, 0)   AS item_count,
			COALESCE(ws.stock_value, 0)  AS stock_value,
			COUNT(CASE WHEN sl.available > sl.min_stock OR (sl.min_stock = 0 AND sl.available > 0) THEN 1 END) AS in_stock_count,
			COUNT(CASE WHEN sl.available > 0 AND sl.available <= sl.min_stock AND sl.min_stock > 0 THEN 1 END)  AS low_stock_count,
			COUNT(CASE WHEN sl.available <= 0 THEN 1 END)                                                       AS out_of_stock_count
		FROM warehouses w
		LEFT JOIN warehouse_stats ws ON ws.warehouse_id = w.id
		LEFT JOIN stock_levels sl ON sl.warehouse_id = w.id
		WHERE w.is_active = true AND w.deleted_at IS NULL AND %s
		GROUP BY w.id, w.name, w.address, ws.item_count, ws.stock_value
		ORDER BY w.name ASC
	`, stockTenantFilter, warehouseStatsTenantFilter, warehouseTenantFilter)
	if err := r.db.WithContext(ctx).Raw(query, params...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]dto.WarehouseItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.WarehouseItem{
			ID:              row.ID,
			Name:            row.Name,
			Location:        row.Address,
			StockValue:      row.StockValue,
			ItemCount:       row.ItemCount,
			InStockCount:    row.InStockCount,
			LowStockCount:   row.LowStockCount,
			OutOfStockCount: row.OutOfStockCount,
		})
	}
	return items, nil
}

func calcChange(previous, current float64) float64 {
	if previous == 0 {
		if current > 0 {
			return 100
		}
		return 0
	}
	return ((current - previous) / previous) * 100
}

func sortPeriodChartPoints(points []dto.PeriodChartPoint) {
	for i := 1; i < len(points); i++ {
		key := points[i]
		j := i - 1
		for j >= 0 && points[j].Period > key.Period {
			points[j+1] = points[j]
			j--
		}
		points[j+1] = key
	}
}

func contextString(ctx context.Context, key string) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(key).(string)
	return strings.TrimSpace(value)
}

func permissionScopeFromContext(ctx context.Context, permissionCode string) string {
	if ctx == nil || permissionCode == "" {
		return ""
	}
	if scopeMap, ok := ctx.Value("user_permissions_scope").(map[string]string); ok {
		if scope, found := scopeMap[permissionCode]; found {
			return strings.ToUpper(strings.TrimSpace(scope))
		}
	}
	return ""
}

func currentRequestScope(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if scope, ok := ctx.Value("permission_scope").(string); ok {
		return strings.ToUpper(strings.TrimSpace(scope))
	}
	return ""
}

func isOutletOrWarehouseScope(ctx context.Context) bool {
	scope := currentRequestScope(ctx)
	return scope == "OUTLET" || scope == "WAREHOUSE"
}

func resolveDashboardOutletIDs(ctx context.Context, db *gorm.DB) ([]string, error) {
	if !isOutletOrWarehouseScope(ctx) {
		return nil, nil
	}

	outletIDs, _ := ctx.Value("scope_outlet_ids").([]string)
	if currentRequestScope(ctx) == "OUTLET" {
		// When the request scope is explicitly OUTLET we must enforce outlet-only
		// results. Return an explicit empty slice when no outlets are present so
		// calling code treats this as "no access to any outlet" instead of
		// falling back to tenant/global queries.
		if outletIDs == nil || len(outletIDs) == 0 {
			return []string{}, nil
		}
		return outletIDs, nil
	}

	if len(outletIDs) > 0 {
		return outletIDs, nil
	}

	warehouseIDs, _ := ctx.Value("scope_warehouse_ids").([]string)
	if len(warehouseIDs) == 0 {
		return []string{}, nil
	}

	var resolved []string
	if err := db.WithContext(ctx).
		Table("outlets").
		Where("warehouse_id IN ? AND deleted_at IS NULL", warehouseIDs).
		Pluck("id", &resolved).Error; err != nil {
		return nil, err
	}

	return resolved, nil
}

func hasFullPOSAccess(ctx context.Context) bool {
	role := strings.ToLower(strings.TrimSpace(contextString(ctx, "user_role")))
	if role == "admin" || role == "system_admin" || role == "tenant_owner" || role == "owner" || strings.HasPrefix(role, "tenant_owner_") || strings.HasSuffix(role, "_owner") {
		return true
	}
	for _, permissionCode := range []string{"pos.order.read", "pos.order.create"} {
		if permissionScopeFromContext(ctx, permissionCode) == "ALL" {
			return true
		}
	}
	return false
}

func withPermissionScope(ctx context.Context, permissionCode string) context.Context {
	if ctx == nil {
		return nil
	}

	scope := permissionScopeFromContext(ctx, permissionCode)
	if scope == "" {
		scope = currentRequestScope(ctx)
	}

	if scope == "" {
		return ctx
	}

	return context.WithValue(ctx, "permission_scope", scope)
}

// buildScopeWhereClause builds a WHERE clause for dashboard queries based on the user's permission scope.
// Returns a WHERE condition string and list of parameters to bind.
// If the user has full access (admin role or ALL scope permissions), returns empty strings.
func buildScopeWhereClause(ctx context.Context, tableAlias string) (whereClause string, params []any) {
	if middleware.IsSystemAdmin(ctx) {
		return "", nil
	}

	requestScope := currentRequestScope(ctx)
	if requestScope == "ALL" {
		log.Printf("[dashboard][scope] request-scope full-access table=%s role=%s", tableAlias, contextString(ctx, "user_role"))
		return "", nil
	}

	outletIDs, _ := ctx.Value("scope_outlet_ids").([]string)
	warehouseIDs, _ := ctx.Value("scope_warehouse_ids").([]string)
	areaIDs, _ := ctx.Value("scope_area_ids").([]string)
	divisionID := contextString(ctx, "scope_division_id")
	employeeID := contextString(ctx, "scope_employee_id")
	userID := contextString(ctx, "user_id")

	log.Printf("[dashboard][scope] table=%s request_scope=%s outlet_ids=%d warehouse_ids=%d area_ids=%d division_id=%s employee_id=%s role=%s",
		tableAlias,
		requestScope,
		len(outletIDs),
		len(warehouseIDs),
		len(areaIDs),
		divisionID,
		employeeID,
		contextString(ctx, "user_role"),
	)

	switch tableAlias {
	case "sales_orders":
		switch requestScope {
		case "OUTLET", "WAREHOUSE":
			if len(outletIDs) == 0 {
				return "1 = 0", nil
			}
			return fmt.Sprintf("%s.outlet_id IN ?", tableAlias), []any{outletIDs}
		case "AREA":
			if len(areaIDs) == 0 {
				return "1 = 0", nil
			}
			return fmt.Sprintf("%s.delivery_area_id IN ?", tableAlias), []any{areaIDs}
		case "DIVISION":
			if divisionID == "" {
				return "1 = 0", nil
			}
			return fmt.Sprintf("%s.sales_rep_id IN (SELECT id FROM employees WHERE division_id = ? AND deleted_at IS NULL)", tableAlias), []any{divisionID}
		case "OWN":
			conditions := make([]string, 0, 2)
			params := make([]any, 0, 2)
			if userID != "" {
				conditions = append(conditions, fmt.Sprintf("%s.created_by = ?", tableAlias))
				params = append(params, userID)
			}
			if employeeID != "" {
				conditions = append(conditions, fmt.Sprintf("%s.sales_rep_id = ?", tableAlias))
				params = append(params, employeeID)
			}
			if len(conditions) == 0 {
				return "1 = 0", nil
			}
			return "(" + strings.Join(conditions, " OR ") + ")", params
		}
	case "employees":
		switch requestScope {
		case "OUTLET":
			if len(outletIDs) == 0 {
				return "1 = 0", nil
			}
			return "employees.id IN (SELECT employee_id FROM employee_outlets WHERE outlet_id IN ? AND deleted_at IS NULL)", []any{outletIDs}
		case "WAREHOUSE":
			if len(warehouseIDs) == 0 {
				return "1 = 0", nil
			}
			return "employees.id IN (SELECT employee_id FROM employee_warehouses WHERE warehouse_id IN ? AND deleted_at IS NULL)", []any{warehouseIDs}
		case "AREA":
			if len(areaIDs) == 0 {
				return "1 = 0", nil
			}
			return "employees.area_id IN ?", []any{areaIDs}
		case "DIVISION":
			if divisionID == "" {
				return "1 = 0", nil
			}
			return "employees.division_id = ?", []any{divisionID}
		case "OWN":
			if employeeID == "" {
				return "1 = 0", nil
			}
			return "employees.id = ?", []any{employeeID}
		}
	case "journal_entries":
		switch requestScope {
		case "OUTLET":
			if len(outletIDs) == 0 {
				return "1 = 0", nil
			}
			return "journal_entries.created_by IN (SELECT user_id FROM employees WHERE id IN (SELECT employee_id FROM employee_outlets WHERE outlet_id IN ? AND deleted_at IS NULL) AND user_id IS NOT NULL)", []any{outletIDs}
		case "WAREHOUSE":
			if len(warehouseIDs) == 0 {
				return "1 = 0", nil
			}
			return "journal_entries.created_by IN (SELECT user_id FROM employees WHERE id IN (SELECT employee_id FROM employee_warehouses WHERE warehouse_id IN ? AND deleted_at IS NULL) AND user_id IS NOT NULL)", []any{warehouseIDs}
		case "AREA":
			if len(areaIDs) == 0 {
				return "1 = 0", nil
			}
			return "journal_entries.created_by IN (SELECT user_id FROM employees WHERE area_id IN ? AND user_id IS NOT NULL)", []any{areaIDs}
		case "DIVISION":
			if divisionID == "" {
				return "1 = 0", nil
			}
			return "journal_entries.created_by IN (SELECT user_id FROM employees WHERE division_id = ? AND user_id IS NOT NULL)", []any{divisionID}
		case "OWN":
			if userID == "" {
				return "1 = 0", nil
			}
			return "journal_entries.created_by = ?", []any{userID}
		}
	}

	log.Printf("[dashboard][scope] table=%s strategy=none request_scope=%s role=%s", tableAlias, requestScope, contextString(ctx, "user_role"))
	return "", nil
}

func resolveAccessiblePOSOutletIDs(ctx context.Context, db *gorm.DB, userID string) ([]string, error) {
	if userID == "" || hasFullPOSAccess(ctx) {
		return nil, nil
	}

	// Respect permission scope: if the user's permission scope for POS is
	// explicitly OUTLET, do not broaden access via warehouse assignments.
	posScope := permissionScopeFromContext(ctx, "pos.order.read")
	if posScope == "" {
		posScope = permissionScopeFromContext(ctx, "pos.order.create")
	}
	if posScope == "" {
		posScope = currentRequestScope(ctx)
	}

	employeeID := contextString(ctx, "scope_employee_id")
	if employeeID == "" {
		if err := db.WithContext(ctx).
			Table("employees").
			Select("id").
			Where("user_id = ? AND deleted_at IS NULL", userID).
			Scan(&employeeID).Error; err != nil {
			return nil, err
		}
	}

	var outletIDs []string
	// If the user has an employee record, prefer explicit employee->outlet assignments first
	if employeeID != "" {
		outletIDs, _ = ctx.Value("scope_outlet_ids").([]string)
		if len(outletIDs) > 0 {
			log.Printf("[dashboard][pos] using scope-context outlets user_id=%s employee_id=%s count=%d scope=%s", userID, employeeID, len(outletIDs), posScope)
			return outletIDs, nil
		}

		if err := db.WithContext(ctx).
			Table("employee_outlets").
			Where("employee_id = ? AND deleted_at IS NULL", employeeID).
			Pluck("outlet_id", &outletIDs).Error; err != nil {
			return nil, err
		}
		if len(outletIDs) > 0 {
			return outletIDs, nil
		}

		// Next, prefer outlets where employee is manager
		if err := db.WithContext(ctx).
			Table("outlets").
			Where("manager_id = ? AND deleted_at IS NULL AND is_active = true", employeeID).
			Pluck("id", &outletIDs).Error; err != nil {
			return nil, err
		}
		if len(outletIDs) > 0 {
			return outletIDs, nil
		}
	}

	// If POS scope is strictly OUTLET, do not derive outlets from warehouse assignments.
	if strings.ToUpper(strings.TrimSpace(posScope)) == "OUTLET" {
		return []string{}, nil
	}

	// Otherwise (WAREHOUSE or unspecified), include outlets belonging to user's warehouses.
	if warehouseIDs, _ := ctx.Value("scope_warehouse_ids").([]string); len(warehouseIDs) > 0 {
		if err := db.WithContext(ctx).
			Table("outlets").
			Where("warehouse_id IN ? AND deleted_at IS NULL AND is_active = true", warehouseIDs).
			Distinct().
			Pluck("id", &outletIDs).Error; err != nil {
			return nil, err
		}
		log.Printf("[dashboard][pos] resolved via employee warehouses user_id=%s warehouses=%d outlets=%d scope=%s", userID, len(warehouseIDs), len(outletIDs), posScope)
		return outletIDs, nil
	}

	if err := db.WithContext(ctx).
		Table("outlets").
		Select("outlets.id").
		Joins("JOIN warehouses ON warehouses.id = outlets.warehouse_id").
		Joins("JOIN user_warehouses ON user_warehouses.warehouse_id = warehouses.id").
		Where("user_warehouses.user_id = ? AND user_warehouses.deleted_at IS NULL AND outlets.deleted_at IS NULL AND outlets.is_active = true", userID).
		Distinct().
		Pluck("outlets.id", &outletIDs).Error; err != nil {
		return nil, err
	}

	return outletIDs, nil
}

// GetPOSSummary returns POS sales per outlet for the requested date range with RBAC filtering.
// If userID is non-empty and the user has OUTLET scope, only their assigned outlets are returned.
func (r *dashboardRepository) GetPOSSummary(ctx context.Context, userID, outletID string, start, end time.Time) (*dto.POSSummaryData, error) {
	db := r.scoped(ctx, "outlets")
	tenantID := contextString(ctx, "tenant_id")
	isSystemAdmin := middleware.IsSystemAdmin(ctx)

	// Resolve outlet filter for RBAC and optional outlet selection.
	outletIDs, err := resolveAccessiblePOSOutletIDs(ctx, db, userID)
	if err != nil {
		return nil, err
	}
	if outletIDs != nil && len(outletIDs) == 0 {
		return &dto.POSSummaryData{
			Outlets:                    []dto.POSOutletItem{},
			TodayTotalRevenue:          0,
			TodayTotalRevenueFormatted: "Rp 0",
			TodayTotalOrders:           0,
			ServedOrders:               0,
			WaitingOrders:              0,
			WarningOrders:              0,
			ActiveSessions:             0,
			OpeningCashToday:           0,
			OpeningCashTodayFormatted:  "Rp 0",
			CashSalesToday:             0,
			CashSalesTodayFormatted:    "Rp 0",
			CashChangeToday:            0,
			CashChangeTodayFormatted:   "Rp 0",
			ExpectedCash:               0,
			ExpectedCashFormatted:      "Rp 0",
			ActualCash:                 0,
			ActualCashFormatted:        "Rp 0",
			CashVariance:               0,
			CashVarianceFormatted:      "Rp 0",
		}, nil
	}
	if outletID != "" {
		if hasFullPOSAccess(ctx) {
			outletIDs = []string{outletID}
		} else {
			allowed := false
			for _, id := range outletIDs {
				if id == outletID {
					allowed = true
					break
				}
			}
			if allowed {
				outletIDs = []string{outletID}
			} else {
				outletIDs = []string{}
			}
		}
	}
	if !hasFullPOSAccess(ctx) && len(outletIDs) == 0 {
		log.Printf("[dashboard][pos] empty-access user_id=%s role=%s tenant_id=%s full_access=%t selected_outlet=%s resolved_outlets=0",
			userID,
			contextString(ctx, "user_role"),
			contextString(ctx, "tenant_id"),
			hasFullPOSAccess(ctx),
			outletID,
		)
		return &dto.POSSummaryData{}, nil
	}

	type outletRow struct {
		OutletID          string  `gorm:"column:outlet_id"`
		OutletName        string  `gorm:"column:outlet_name"`
		FloorPlanID       string  `gorm:"column:floor_plan_id"`
		ActiveTableCount  int     `gorm:"column:active_table_count"`
		ActiveTableLabels string  `gorm:"column:active_table_labels"`
		TodayRevenue      float64 `gorm:"column:today_revenue"`
		OrderCount        int     `gorm:"column:order_count"`
		ServedOrders      int     `gorm:"column:served_orders"`
		WaitingOrders     int     `gorm:"column:waiting_orders"`
		WarningOrders     int     `gorm:"column:warning_orders"`
		TablesTotal       int     `gorm:"column:tables_total"`
		TablesOccupied    int     `gorm:"column:tables_occupied"`
		ActiveSessions    int     `gorm:"column:active_sessions"`
		OpeningCashToday  float64 `gorm:"column:opening_cash_today"`
		CashSalesToday    float64 `gorm:"column:cash_sales_today"`
		CashChangeToday   float64 `gorm:"column:cash_change_today"`
		ActualCashToday   float64 `gorm:"column:actual_cash_today"`
	}

	query := `
		WITH outlet_scope AS (
			SELECT o.id, o.name
			FROM outlets o
			WHERE o.deleted_at IS NULL AND o.is_active = true
		), outlet_orders AS (
			SELECT
				po.outlet_id,
				COALESCE(SUM(CASE WHEN po.created_at BETWEEN ? AND ? THEN po.total_amount ELSE 0 END), 0) AS today_revenue,
				COUNT(*) FILTER (WHERE po.created_at BETWEEN ? AND ?) AS order_count,
				COUNT(*) FILTER (WHERE po.created_at BETWEEN ? AND ? AND po.status IN ('SERVED', 'COMPLETED')) AS served_orders,
				COUNT(*) FILTER (WHERE po.created_at BETWEEN ? AND ? AND po.status IN ('IN_PROGRESS', 'READY', 'PARTIAL_SERVED')) AS waiting_orders,
				COUNT(*) FILTER (
					WHERE po.created_at BETWEEN ? AND ?
						AND po.status IN ('IN_PROGRESS', 'READY', 'PARTIAL_SERVED')
						AND po.created_at <= (?::timestamptz - INTERVAL '15 minutes')
				) AS warning_orders
			FROM pos_orders po
			JOIN outlet_scope oscope ON oscope.id = po.outlet_id
			WHERE po.deleted_at IS NULL
			GROUP BY po.outlet_id
		), floor_plan_tables AS (
			SELECT
				fp.outlet_id,
				COUNT(*) FILTER (WHERE obj.value->>'type' = 'table') AS tables_total
			FROM pos_floor_plans fp
			JOIN outlet_scope oscope ON oscope.id = fp.outlet_id
			CROSS JOIN LATERAL jsonb_array_elements(
				CASE
					WHEN jsonb_typeof(COALESCE(fp.layout_data::jsonb, '[]'::jsonb)) = 'array'
					THEN COALESCE(fp.layout_data::jsonb, '[]'::jsonb)
					ELSE '[]'::jsonb
				END
			) obj
			WHERE fp.deleted_at IS NULL
			GROUP BY fp.outlet_id
		), published_floor_plan AS (
			SELECT DISTINCT ON (fp.outlet_id)
				fp.outlet_id,
				fp.id AS floor_plan_id
			FROM pos_floor_plans fp
			JOIN outlet_scope oscope ON oscope.id = fp.outlet_id
			WHERE fp.deleted_at IS NULL AND fp.status = 'published'
			ORDER BY fp.outlet_id, fp.updated_at DESC, fp.created_at DESC
		), outlet_active_tables AS (
			SELECT
				src.outlet_id,
				COUNT(*) AS active_table_count,
				COALESCE(STRING_AGG(src.table_label, '||' ORDER BY src.table_label), '') AS active_table_labels
			FROM (
				SELECT DISTINCT ON (po.outlet_id, TRIM(po.table_label))
					po.outlet_id,
					TRIM(po.table_label) AS table_label,
					po.created_at
				FROM pos_orders po
				JOIN outlet_scope oscope ON oscope.id = po.outlet_id
				WHERE po.deleted_at IS NULL
					AND po.table_label IS NOT NULL
					AND TRIM(po.table_label) <> ''
					AND po.status IN ('PAID', 'PARTIAL_SERVED', 'SERVED')
				ORDER BY po.outlet_id, TRIM(po.table_label), po.created_at DESC
			) src
			GROUP BY src.outlet_id
		), outlet_table_statuses AS (
			SELECT
				ps.outlet_id,
				COUNT(*) FILTER (WHERE UPPER(pts.status) = 'OCCUPIED') AS tables_occupied,
				COUNT(*) FILTER (WHERE UPPER(pts.status) IN ('RESERVED', 'CLEANING')) AS tables_attention
			FROM pos_table_statuses pts
			JOIN pos_sessions ps ON ps.id = pts.session_id AND ps.deleted_at IS NULL
			JOIN outlet_scope oscope ON oscope.id = ps.outlet_id
			WHERE pts.deleted_at IS NULL AND ps.status = 'OPEN'
			GROUP BY ps.outlet_id
		), outlet_sessions AS (
			SELECT
				ps.outlet_id,
				COUNT(*) FILTER (WHERE ps.status = 'OPEN') AS active_sessions,
				COALESCE(SUM(CASE WHEN ps.opened_at BETWEEN ? AND ? THEN ps.opening_cash ELSE 0 END), 0) AS opening_cash_today,
				COALESCE(SUM(CASE WHEN ps.status = 'CLOSED' AND ps.closed_at BETWEEN ? AND ? AND ps.closing_cash IS NOT NULL THEN ps.closing_cash ELSE 0 END), 0) AS actual_cash_today
			FROM pos_sessions ps
			JOIN outlet_scope oscope ON oscope.id = ps.outlet_id
			WHERE ps.deleted_at IS NULL
			GROUP BY ps.outlet_id
		), outlet_payments AS (
			SELECT
				po.outlet_id,
				COALESCE(SUM(CASE WHEN pp.created_at BETWEEN ? AND ? AND pp.method = 'CASH' AND pp.status = 'PAID' THEN pp.amount ELSE 0 END), 0) AS cash_sales_today,
				COALESCE(SUM(CASE WHEN pp.created_at BETWEEN ? AND ? AND pp.method = 'CASH' AND pp.status = 'PAID' THEN pp.change_amount ELSE 0 END), 0) AS cash_change_today
			FROM pos_payments pp
			JOIN pos_orders po ON po.id = pp.order_id AND po.deleted_at IS NULL
			JOIN outlet_scope oscope ON oscope.id = po.outlet_id
			WHERE pp.created_at BETWEEN ? AND ?
			GROUP BY po.outlet_id
		)
		SELECT
			oscope.id AS outlet_id,
			oscope.name AS outlet_name,
			COALESCE(pfp.floor_plan_id::text, '') AS floor_plan_id,
			COALESCE(oat.active_table_count, 0) AS active_table_count,
			COALESCE(oat.active_table_labels, '') AS active_table_labels,
			COALESCE(oo.today_revenue, 0) AS today_revenue,
			COALESCE(oo.order_count, 0) AS order_count,
			COALESCE(oo.served_orders, 0) AS served_orders,
			COALESCE(oo.waiting_orders, 0) AS waiting_orders,
			COALESCE(oo.warning_orders, 0) + COALESCE(ts.tables_attention, 0) AS warning_orders,
			COALESCE(fp.tables_total, 0) AS tables_total,
			COALESCE(ts.tables_occupied, 0) AS tables_occupied,
			COALESCE(os.active_sessions, 0) AS active_sessions,
			COALESCE(os.opening_cash_today, 0) AS opening_cash_today,
			COALESCE(op.cash_sales_today, 0) AS cash_sales_today,
			COALESCE(op.cash_change_today, 0) AS cash_change_today,
			COALESCE(os.actual_cash_today, 0) AS actual_cash_today
		FROM outlet_scope oscope
		LEFT JOIN outlet_orders oo ON oo.outlet_id = oscope.id
		LEFT JOIN floor_plan_tables fp ON fp.outlet_id = oscope.id
		LEFT JOIN published_floor_plan pfp ON pfp.outlet_id = oscope.id
		LEFT JOIN outlet_active_tables oat ON oat.outlet_id = oscope.id
		LEFT JOIN outlet_table_statuses ts ON ts.outlet_id = oscope.id
		LEFT JOIN outlet_sessions os ON os.outlet_id = oscope.id
		LEFT JOIN outlet_payments op ON op.outlet_id = oscope.id
		WHERE 1=1
	`
	params := []any{
		start, end,
		start, end,
		start, end,
		start, end,
		start, end,
		end,
		start, end,
		start, end,
		start, end,
		start, end,
		start, end,
	}
	if len(outletIDs) > 0 {
		query += " AND oscope.id IN ?"
		params = append(params, outletIDs)
	}
	if !isSystemAdmin && tenantID != "" {
		query = strings.Replace(query, "WHERE o.deleted_at IS NULL AND o.is_active = true", "WHERE o.deleted_at IS NULL AND o.is_active = true AND o.tenant_id = ?", 1)
		params = append([]any{tenantID}, params...)
	}
	query += " ORDER BY today_revenue DESC"

	var rows []outletRow
	if err := db.Raw(query, params...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		var outletsInScope int64
		outletScopeQuery := r.db.WithContext(ctx).
			Table("outlets").
			Where("outlets.deleted_at IS NULL")
		if !isSystemAdmin && tenantID != "" {
			outletScopeQuery = outletScopeQuery.Where("outlets.tenant_id = ?", tenantID)
		}
		if hasFullPOSAccess(ctx) {
			if outletID != "" {
				outletScopeQuery = outletScopeQuery.Where("outlets.id = ?", outletID)
			}
		} else {
			if len(outletIDs) > 0 {
				outletScopeQuery = outletScopeQuery.Where("outlets.id IN ?", outletIDs)
			} else {
				outletScopeQuery = outletScopeQuery.Where("1 = 0")
			}
		}
		if err := outletScopeQuery.Count(&outletsInScope).Error; err != nil {
			log.Printf("[dashboard][pos] diagnostics outlet-count failed: %v", err)
		}

		var ordersInRange int64
		orderScopeQuery := r.db.WithContext(ctx).
			Table("pos_orders po").
			Where("po.deleted_at IS NULL AND po.created_at BETWEEN ? AND ?", start, end)
		if !isSystemAdmin && tenantID != "" {
			orderScopeQuery = orderScopeQuery.Where("po.tenant_id = ?", tenantID)
		}
		if hasFullPOSAccess(ctx) {
			if outletID != "" {
				orderScopeQuery = orderScopeQuery.Where("po.outlet_id = ?", outletID)
			}
		} else {
			if len(outletIDs) > 0 {
				orderScopeQuery = orderScopeQuery.Where("po.outlet_id IN ?", outletIDs)
			} else {
				orderScopeQuery = orderScopeQuery.Where("1 = 0")
			}
		}
		if err := orderScopeQuery.Count(&ordersInRange).Error; err != nil {
			log.Printf("[dashboard][pos] diagnostics order-count failed: %v", err)
		}

		log.Printf("[dashboard][pos] empty-summary user_id=%s role=%s tenant_id=%s full_access=%t selected_outlet=%s resolved_outlets=%d outlets_in_scope=%d orders_in_range=%d start=%s end=%s",
			userID,
			contextString(ctx, "user_role"),
			contextString(ctx, "tenant_id"),
			hasFullPOSAccess(ctx),
			outletID,
			len(outletIDs),
			outletsInScope,
			ordersInRange,
			start.Format(time.RFC3339),
			end.Format(time.RFC3339),
		)
	}

	result := &dto.POSSummaryData{}
	for _, row := range rows {
		warningOrders := row.WarningOrders
		if warningOrders < 0 {
			warningOrders = 0
		}
		expectedCash := row.OpeningCashToday + row.CashSalesToday - row.CashChangeToday
		cashVariance := row.ActualCashToday - expectedCash
		item := dto.POSOutletItem{
			OutletID:              row.OutletID,
			OutletName:            row.OutletName,
			FloorPlanID:           row.FloorPlanID,
			ActiveTableCount:      row.ActiveTableCount,
			TodayRevenue:          row.TodayRevenue,
			TodayRevenueFormatted: formatPOSCurrency(row.TodayRevenue),
			OrderCount:            row.OrderCount,
			ServedOrders:          row.ServedOrders,
			WaitingOrders:         row.WaitingOrders,
			WarningOrders:         warningOrders,
			TablesTotal:           row.TablesTotal,
			TablesOccupied:        row.TablesOccupied,
		}
		if row.ActiveTableLabels != "" {
			parts := strings.Split(row.ActiveTableLabels, "||")
			labels := make([]string, 0, len(parts))
			for _, part := range parts {
				clean := strings.TrimSpace(part)
				if clean == "" {
					continue
				}
				labels = append(labels, clean)
			}
			sort.Slice(labels, func(i, j int) bool {
				leftNum, leftErr := strconv.Atoi(strings.TrimPrefix(strings.ToUpper(labels[i]), "T"))
				rightNum, rightErr := strconv.Atoi(strings.TrimPrefix(strings.ToUpper(labels[j]), "T"))
				if leftErr == nil && rightErr == nil {
					return leftNum < rightNum
				}
				return labels[i] < labels[j]
			})
			item.ActiveTableLabels = labels
		}
		result.Outlets = append(result.Outlets, item)
		result.TodayTotalRevenue += row.TodayRevenue
		result.TodayTotalOrders += row.OrderCount
		result.ServedOrders += row.ServedOrders
		result.WaitingOrders += row.WaitingOrders
		result.WarningOrders += warningOrders
		result.ActiveSessions += row.ActiveSessions
		result.OpeningCashToday += row.OpeningCashToday
		result.CashSalesToday += row.CashSalesToday
		result.CashChangeToday += row.CashChangeToday
		result.ActualCash += row.ActualCashToday
		result.ExpectedCash += expectedCash
		result.CashVariance += cashVariance
	}

	result.TodayTotalRevenueFormatted = formatPOSCurrency(result.TodayTotalRevenue)
	result.OpeningCashTodayFormatted = formatPOSCurrency(result.OpeningCashToday)
	result.CashSalesTodayFormatted = formatPOSCurrency(result.CashSalesToday)
	result.CashChangeTodayFormatted = formatPOSCurrency(result.CashChangeToday)
	result.ExpectedCashFormatted = formatPOSCurrency(result.ExpectedCash)
	result.ActualCashFormatted = formatPOSCurrency(result.ActualCash)
	result.CashVarianceFormatted = formatPOSCurrencySigned(result.CashVariance)

	return result, nil
}

func (r *dashboardRepository) GetHRSummary(ctx context.Context) (*dto.HRSummaryData, error) {
	db := r.scoped(ctx, "employees")

	result := &dto.HRSummaryData{}

	// Total active employees for rate calculation
	var totalEmployees int64
	if err := db.Table("employees").Where("deleted_at IS NULL AND status = ?", "active").Count(&totalEmployees).Error; err != nil {
		return nil, err
	}

	// Today's attendance
	var presentToday int64
	if err := r.scoped(ctx, "attendance_records").
		Table("attendance_records").
		Where("date = CURRENT_DATE AND deleted_at IS NULL AND status = ?", "present").
		Count(&presentToday).Error; err != nil {
		return nil, err
	}
	result.PresentToday = int(presentToday)

	var lateToday int64
	if err := r.scoped(ctx, "attendance_records").
		Table("attendance_records").
		Where("date = CURRENT_DATE AND deleted_at IS NULL AND is_late = true").
		Count(&lateToday).Error; err != nil {
		return nil, err
	}
	result.LateToday = int(lateToday)

	if totalEmployees > 0 {
		result.AbsentToday = int(totalEmployees) - result.PresentToday
		rate := float64(result.PresentToday) / float64(totalEmployees) * 100
		result.AttendanceRate = float64(int(rate*10)) / 10 // 1 decimal
	}

	// Pending leaves
	var pendingLeaves int64
	if err := r.scoped(ctx, "leave_requests").
		Table("leave_requests").
		Where("deleted_at IS NULL AND status = ?", "pending").
		Count(&pendingLeaves).Error; err != nil {
		return nil, err
	}
	result.PendingLeaves = int(pendingLeaves)

	return result, nil
}

// GetCRMSummary returns aggregated CRM metrics.
func (r *dashboardRepository) GetCRMSummary(ctx context.Context) (*dto.CRMSummaryData, error) {
	result := &dto.CRMSummaryData{}
	leadScopeCtx := withPermissionScope(ctx, "crm_lead.read")
	dealScopeCtx := withPermissionScope(ctx, "crm_deal.read")
	taskScopeCtx := withPermissionScope(ctx, "crm_task.read")

	contactsDB := security.ApplyScopeFilter(
		r.scoped(leadScopeCtx, "crm_contacts").Model(&crmModels.Contact{}),
		leadScopeCtx,
		security.DefaultScopeQueryOptions(),
	)
	leadsDB := security.ApplyScopeFilter(
		r.scoped(leadScopeCtx, "crm_leads").Model(&crmModels.Lead{}),
		leadScopeCtx,
		security.MixedOwnershipScopeQueryOptions("assigned_to"),
	)
	dealsDB := security.ApplyScopeFilter(
		r.scoped(dealScopeCtx, "crm_deals").Model(&crmModels.Deal{}),
		dealScopeCtx,
		security.MixedOwnershipScopeQueryOptions("assigned_to"),
	)
	schedulesDB := security.ApplyScopeFilter(
		r.scoped(taskScopeCtx, "crm_schedules").Model(&crmModels.Schedule{}),
		taskScopeCtx,
		security.MixedOwnershipScopeQueryOptions("employee_id"),
	)

	var totalContacts int64
	if err := contactsDB.Where("deleted_at IS NULL").Count(&totalContacts).Error; err != nil {
		return nil, err
	}
	result.TotalContacts = int(totalContacts)

	var activeLeads int64
	if err := leadsDB.
		Where("deleted_at IS NULL AND converted_at IS NULL").
		Count(&activeLeads).Error; err != nil {
		return nil, err
	}
	result.ActiveLeads = int(activeLeads)

	var leadsThisMonth int64
	if err := leadsDB.
		Where("deleted_at IS NULL AND created_at >= date_trunc('month', CURRENT_DATE)").
		Count(&leadsThisMonth).Error; err != nil {
		return nil, err
	}
	result.LeadsThisMonth = int(leadsThisMonth)

	var recentLeadRows []crmModels.Lead
	if err := leadsDB.
		Preload("LeadStatus").
		Preload("AssignedEmployee").
		Preload("Deal.PipelineStage").
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(5).
		Find(&recentLeadRows).Error; err != nil {
		return nil, err
	}

	result.RecentLeads = make([]dto.CRMLeadSummary, 0, len(recentLeadRows))
	for _, lead := range recentLeadRows {
		leadName := strings.TrimSpace(strings.Join([]string{lead.FirstName, lead.LastName}, " "))
		if leadName == "" {
			leadName = lead.CompanyName
		}
		statusName := ""
		statusColor := ""
		if lead.LeadStatus != nil {
			statusName = lead.LeadStatus.Name
			statusColor = lead.LeadStatus.Color
		}
		assignedTo := ""
		if lead.AssignedEmployee != nil {
			assignedTo = lead.AssignedEmployee.Name
		}
		result.RecentLeads = append(result.RecentLeads, dto.CRMLeadSummary{
			ID:              lead.ID,
			Code:            lead.Code,
			Name:            leadName,
			CompanyName:     lead.CompanyName,
			LeadStatus:      statusName,
			LeadStatusColor: statusColor,
			AssignedTo:      assignedTo,
			CreatedAt:       lead.CreatedAt.Format("2006-01-02T15:04:05+07:00"),
		})
	}

	var dealsInProgress int64
	if err := dealsDB.
		Where("deleted_at IS NULL AND status = ?", "open").
		Count(&dealsInProgress).Error; err != nil {
		return nil, err
	}
	result.DealsInProgress = int(dealsInProgress)

	var dealsWonThisMonth int64
	if err := dealsDB.
		Where("deleted_at IS NULL AND status = ?", "won").
		Where("COALESCE(actual_close_date::timestamp, updated_at) >= date_trunc('month', CURRENT_DATE)").
		Count(&dealsWonThisMonth).Error; err != nil {
		return nil, err
	}
	result.DealsWonThisMonth = int(dealsWonThisMonth)

	var activitiesToday int64
	if err := schedulesDB.
		Where("deleted_at IS NULL AND status <> ? AND scheduled_at::date = CURRENT_DATE", "cancelled").
		Count(&activitiesToday).Error; err != nil {
		return nil, err
	}
	result.ActivitiesToday = int(activitiesToday)

	type pipelineStageAggregate struct {
		StageID    string
		StageName  string
		StageOrder int
		ItemCount  int
	}

	var stageRows []pipelineStageAggregate
	if err := dealsDB.
		Table("crm_deals").
		Select(`
			crm_pipeline_stages.id AS stage_id,
			crm_pipeline_stages.name AS stage_name,
			crm_pipeline_stages."order" AS stage_order,
			COUNT(crm_deals.id) AS item_count
		`).
		Joins("JOIN crm_pipeline_stages ON crm_pipeline_stages.id = crm_deals.pipeline_stage_id AND crm_pipeline_stages.deleted_at IS NULL").
		Where("crm_deals.deleted_at IS NULL AND crm_deals.status = ?", "open").
		Group("crm_pipeline_stages.id, crm_pipeline_stages.name, crm_pipeline_stages.\"order\"").
		Order("crm_pipeline_stages.\"order\" ASC").
		Limit(5).
		Scan(&stageRows).Error; err != nil {
		return nil, err
	}

	result.PipelineStages = make([]dto.CRMPipelineStageSummary, 0, len(stageRows))
	for _, row := range stageRows {
		result.PipelineStages = append(result.PipelineStages, dto.CRMPipelineStageSummary{
			StageID:    row.StageID,
			StageName:  row.StageName,
			StageOrder: row.StageOrder,
			ItemCount:  row.ItemCount,
		})
	}

	return result, nil
}

func formatPOSCurrency(v float64) string {
	if v >= 1_000_000_000 {
		return fmt.Sprintf("Rp %.1fM", v/1_000_000_000)
	} else if v >= 1_000_000 {
		return fmt.Sprintf("Rp %.1fjt", v/1_000_000)
	} else if v >= 1_000 {
		return fmt.Sprintf("Rp %.0frb", v/1_000)
	}
	return fmt.Sprintf("Rp %.0f", v)
}

func formatPOSCurrencySigned(v float64) string {
	if v < 0 {
		return "-" + formatPOSCurrency(-v)
	}
	return formatPOSCurrency(v)
}
