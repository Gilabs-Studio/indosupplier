package dto

// DashboardRequest holds the query parameters for filtering dashboard data.
type DashboardRequest struct {
	StartDate string                 `form:"start_date"`
	EndDate   string                 `form:"end_date"`
	Year      int                    `form:"year"`
	Scope     DashboardOverviewScope `form:"scope"`
	OutletID  string                 `form:"outlet_id"`
	// Populated by handler from auth context for RBAC filtering
	UserID    string `form:"-"`
}

// DashboardOverviewScope controls which section of overview data should be returned.
type DashboardOverviewScope string

const (
	DashboardOverviewScopeKPI              DashboardOverviewScope = "kpi"
	DashboardOverviewScopeCharts           DashboardOverviewScope = "charts"
	DashboardOverviewScopeBalance          DashboardOverviewScope = "balance"
	DashboardOverviewScopeCosts            DashboardOverviewScope = "costs"
	DashboardOverviewScopeInvoices         DashboardOverviewScope = "invoices"
	DashboardOverviewScopeSalesPerformance DashboardOverviewScope = "sales-performance"
	DashboardOverviewScopeProducts         DashboardOverviewScope = "products"
	DashboardOverviewScopeDelivery         DashboardOverviewScope = "delivery"
	DashboardOverviewScopeGeo              DashboardOverviewScope = "geo"
	DashboardOverviewScopeWarehouse        DashboardOverviewScope = "warehouse"
	DashboardOverviewScopeOwnerKPI         DashboardOverviewScope = "owner-kpi"
	DashboardOverviewScopePOS              DashboardOverviewScope = "pos"
	DashboardOverviewScopeHR               DashboardOverviewScope = "hr"
	DashboardOverviewScopeCRM              DashboardOverviewScope = "crm"
)

// IsValid returns true when the scope value is one of the supported scope constants.
func (s DashboardOverviewScope) IsValid() bool {
	switch s {
	case DashboardOverviewScopeKPI,
		DashboardOverviewScopeCharts,
		DashboardOverviewScopeBalance,
		DashboardOverviewScopeCosts,
		DashboardOverviewScopeInvoices,
		DashboardOverviewScopeSalesPerformance,
		DashboardOverviewScopeProducts,
		DashboardOverviewScopeDelivery,
		DashboardOverviewScopeGeo,
		DashboardOverviewScopeWarehouse,
		DashboardOverviewScopeOwnerKPI,
		DashboardOverviewScopePOS,
		DashboardOverviewScopeHR,
		DashboardOverviewScopeCRM:
		return true
	default:
		return false
	}
}

// DashboardOverviewResponse is the top-level dashboard API response.
// Field names exactly match the TypeScript DashboardOverviewData interface.
type DashboardOverviewResponse struct {
	KPI                KPIData               `json:"kpi"`
	RevenueChart       PeriodChartData       `json:"revenue_chart"`
	CostsChart         PeriodChartData       `json:"costs_chart"`
	RevenueVsCosts     PeriodChartData       `json:"revenue_vs_costs"`
	BalanceOverview    BalanceOverviewData   `json:"balance_overview"`
	CostsByCategory    []CostCategoryItem    `json:"costs_by_category"`
	InvoicesSummary    InvoiceSummaryData    `json:"invoices_summary"`
	RecentInvoices     []InvoiceRow          `json:"recent_invoices"`
	SalesPerformance   []SalesPerformanceRow `json:"sales_performance"`
	TopProducts        []TopProductRow       `json:"top_products"`
	DeliveryStatus     DeliveryStatusData    `json:"delivery_status"`
	GeographicOverview GeoOverviewData       `json:"geographic_overview"`
	WarehouseOverview  WarehouseOverviewData `json:"warehouse_overview"`
}

// DashboardScopedOverviewResponse is used when the request includes a scope query.
// Only requested sections are included, reducing payload size.
type DashboardScopedOverviewResponse struct {
	KPI                *KPIData               `json:"kpi,omitempty"`
	RevenueChart       *PeriodChartData       `json:"revenue_chart,omitempty"`
	CostsChart         *PeriodChartData       `json:"costs_chart,omitempty"`
	RevenueVsCosts     *PeriodChartData       `json:"revenue_vs_costs,omitempty"`
	BalanceOverview    *BalanceOverviewData   `json:"balance_overview,omitempty"`
	CostsByCategory    []CostCategoryItem     `json:"costs_by_category,omitempty"`
	InvoicesSummary    *InvoiceSummaryData    `json:"invoices_summary,omitempty"`
	RecentInvoices     []InvoiceRow           `json:"recent_invoices,omitempty"`
	SalesPerformance   []SalesPerformanceRow  `json:"sales_performance,omitempty"`
	TopProducts        []TopProductRow        `json:"top_products,omitempty"`
	DeliveryStatus     *DeliveryStatusData    `json:"delivery_status,omitempty"`
	GeographicOverview *GeoOverviewData       `json:"geographic_overview,omitempty"`
	WarehouseOverview  *WarehouseOverviewData `json:"warehouse_overview,omitempty"`
	POSSummary         *POSSummaryData        `json:"pos_summary,omitempty"`
	HRSummary          *HRSummaryData         `json:"hr_summary,omitempty"`
	CRMSummary         *CRMSummaryData        `json:"crm_summary,omitempty"`
}

// KPIData contains the five main KPI summary cards.
type KPIData struct {
	TotalRevenue   KPICard `json:"total_revenue"`
	TotalOrders    KPICard `json:"total_orders"`
	TotalCustomers KPICard `json:"total_customers"`
	TotalProducts  KPICard `json:"total_products"`
	EmployeeCount  KPICard `json:"employee_count"`
}

// KPICard represents a single KPI metric.
// Formatted is computed by the usecase layer; repository only sets Value and ChangePercent.
type KPICard struct {
	Value         float64 `json:"value"`
	Formatted     string  `json:"formatted"`
	ChangePercent float64 `json:"change_percent"`
}

// ChartSeriesData is a named data series for a time-series chart widget.
type ChartSeriesData struct {
	Label     string    `json:"label"`
	Data      []float64 `json:"data"`
	Formatted []string  `json:"formatted"`
}

// PeriodChartData is the response format consumed by all chart widgets.
type PeriodChartData struct {
	Series []ChartSeriesData `json:"series"`
	Period []string          `json:"period"`
}

// PeriodChartPoint is an internal data point used only by the repository layer.
// It is not serialised to JSON directly.
type PeriodChartPoint struct {
	Period  string
	Revenue float64
	Costs   float64
	Amount  float64
}

// BalanceChartPoint is a single time-series entry for the balance trend chart.
type BalanceChartPoint struct {
	Period    string  `json:"period"`
	Value     float64 `json:"value"`
	Formatted string  `json:"formatted"`
}

// BalanceOverviewData is the response for the balance widget.
type BalanceOverviewData struct {
	Value         float64             `json:"value"`
	Formatted     string              `json:"formatted"`
	ChangePercent float64             `json:"change_percent"`
	ChartData     []BalanceChartPoint `json:"chart_data"`
}

// BalanceData is the raw balance data returned by the repository.
// It is transformed into BalanceOverviewData in the usecase layer.
type BalanceData struct {
	Current float64
	Change  float64
	Trend   []PeriodChartPoint
}

// CostCategoryItem represents a single cost breakdown entry.
type CostCategoryItem struct {
	Category        string  `json:"category"`
	Amount          float64 `json:"amount"`
	AmountFormatted string  `json:"amount_formatted"`
	Percentage      float64 `json:"percentage"`
}

// InvoiceSummaryData contains invoice status counts.
type InvoiceSummaryData struct {
	Total   int `json:"total"`
	Paid    int `json:"paid"`
	Unpaid  int `json:"unpaid"`
	Overdue int `json:"overdue"`
}

// InvoiceRow represents a single recent invoice entry.
// Field names match the TypeScript InvoiceRow interface.
type InvoiceRow struct {
	ID             string  `json:"id"`
	CustomerID     string  `json:"customer_id"` // customer UUID for detail modal
	Company        string  `json:"company"`     // customer name
	Contact        string  `json:"contact"`     // invoice code
	IssueDate      string  `json:"issue_date"`  // due date
	Value          float64 `json:"value"`
	ValueFormatted string  `json:"value_formatted"`
	Status         string  `json:"status"` // "paid" | "unpaid" | "overdue"
}

// SalesPerformanceRow represents a sales representative performance entry.
type SalesPerformanceRow struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Revenue          float64 `json:"revenue"`
	RevenueFormatted string  `json:"revenue_formatted"`
	Orders           int     `json:"orders"`
	TargetPercent    float64 `json:"target_percent"`
}

// TopProductRow represents a top-selling product.
type TopProductRow struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	SKU              string  `json:"sku"`
	QuantitySold     float64 `json:"quantity_sold"`
	Revenue          float64 `json:"revenue"`
	RevenueFormatted string  `json:"revenue_formatted"`
}

// DeliveryStatusData contains delivery status breakdown.
type DeliveryStatusData struct {
	Total     int `json:"total"`
	Pending   int `json:"pending"`
	InTransit int `json:"in_transit"`
	Delivered int `json:"delivered"`
}

// GeoOverviewData contains geographic performance data.
// Field names match the TypeScript GeoOverviewData interface.
type GeoOverviewData struct {
	Regions        []GeoRegionData `json:"regions"`
	TotalValue     float64         `json:"total_value"`
	TotalFormatted string          `json:"total_formatted"`
}

// GeoRegionData represents performance data for a single province/region.
// Field names match the TypeScript GeoRegionData interface.
type GeoRegionData struct {
	Name      string  `json:"name"`      // province name
	Code      string  `json:"code"`      // province id
	Value     float64 `json:"value"`     // revenue
	Formatted string  `json:"formatted"` // formatted revenue
	Count     int     `json:"count"`     // total orders
}

// WarehouseItem represents a single warehouse in the overview widget.
// Field names match the TypeScript WarehouseItem interface.
type WarehouseItem struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	Location           string  `json:"location"`
	StockValue         float64 `json:"stock_value"`
	StockFormatted     string  `json:"stock_formatted"`
	ItemCount          int     `json:"item_count"`
	// Stock status breakdown — consistent with inventory feature status logic
	InStockCount    int `json:"in_stock_count"`
	LowStockCount   int `json:"low_stock_count"`
	OutOfStockCount int `json:"out_of_stock_count"`
}

// WarehouseOverviewData wraps the warehouse list with aggregate totals.
// Field names match the TypeScript WarehouseOverviewData interface.
type WarehouseOverviewData struct {
	Warehouses          []WarehouseItem `json:"warehouses"`
	TotalStockValue     float64         `json:"total_stock_value"`
	TotalStockFormatted string          `json:"total_stock_formatted"`
}

// ---- Dashboard Layout DTOs ----

// DashboardLayoutSaveRequest is the request body for saving a user's layout.
type DashboardLayoutSaveRequest struct {
	DashboardType string `json:"dashboard_type" binding:"required"`
	LayoutJSON    string `json:"layout_json" binding:"required"`
}

// DashboardLayoutResponse is the API response for a user's saved layout.
type DashboardLayoutResponse struct {
	DashboardType string `json:"dashboard_type"`
	LayoutJSON    string `json:"layout_json"`
}

// ---- POS Summary DTOs ----

// POSOutletItem represents a single outlet's sales data.
type POSOutletItem struct {
	OutletID              string  `json:"outlet_id"`
	OutletName            string  `json:"outlet_name"`
	FloorPlanID           string  `json:"floor_plan_id,omitempty"`
	ActiveTableCount      int     `json:"active_table_count"`
	ActiveTableLabels     []string `json:"active_table_labels,omitempty"`
	TodayRevenue          float64 `json:"today_revenue"`
	TodayRevenueFormatted string  `json:"today_revenue_formatted"`
	OrderCount            int     `json:"order_count"`
	ServedOrders          int     `json:"served_orders"`
	WaitingOrders         int     `json:"waiting_orders"`
	WarningOrders         int     `json:"warning_orders"`
	TablesTotal           int     `json:"tables_total"`
	TablesOccupied        int     `json:"tables_occupied"`
}

// POSSummaryData is the aggregated POS dashboard data.
type POSSummaryData struct {
	Outlets                   []POSOutletItem `json:"outlets"`
	TodayTotalRevenue         float64         `json:"today_total_revenue"`
	TodayTotalRevenueFormatted string         `json:"today_total_revenue_formatted"`
	TodayTotalOrders          int             `json:"today_total_orders"`
	ServedOrders              int             `json:"served_orders"`
	WaitingOrders             int             `json:"waiting_orders"`
	WarningOrders             int             `json:"warning_orders"`
	ActiveSessions            int             `json:"active_sessions"`
	OpeningCashToday          float64         `json:"opening_cash_today"`
	OpeningCashTodayFormatted string          `json:"opening_cash_today_formatted"`
	CashSalesToday            float64         `json:"cash_sales_today"`
	CashSalesTodayFormatted   string          `json:"cash_sales_today_formatted"`
	CashChangeToday           float64         `json:"cash_change_today"`
	CashChangeTodayFormatted  string          `json:"cash_change_today_formatted"`
	ExpectedCash              float64         `json:"expected_cash"`
	ExpectedCashFormatted     string          `json:"expected_cash_formatted"`
	ActualCash                float64         `json:"actual_cash"`
	ActualCashFormatted       string          `json:"actual_cash_formatted"`
	CashVariance              float64         `json:"cash_variance"`
	CashVarianceFormatted     string          `json:"cash_variance_formatted"`
}

// ---- HR Summary DTOs ----

// HRSummaryData is the aggregated HR dashboard data.
type HRSummaryData struct {
	PresentToday   int     `json:"present_today"`
	LateToday      int     `json:"late_today"`
	AbsentToday    int     `json:"absent_today"`
	AttendanceRate float64 `json:"attendance_rate"`
	PendingLeaves  int     `json:"pending_leaves"`
}

// ---- CRM Summary DTOs ----

// CRMSummaryData is the aggregated CRM dashboard data.
type CRMSummaryData struct {
	TotalContacts    int `json:"total_contacts"`
	ActiveLeads      int `json:"active_leads"`
	LeadsThisMonth   int `json:"leads_this_month"`
	DealsInProgress  int `json:"deals_in_progress"`
	DealsWonThisMonth int `json:"deals_won_this_month"`
	ActivitiesToday  int `json:"activities_today"`
	RecentLeads      []CRMLeadSummary `json:"recent_leads"`
	PipelineStages   []CRMPipelineStageSummary `json:"pipeline_stages"`
}

// CRMLeadSummary represents a compact lead item for the dashboard widget.
type CRMLeadSummary struct {
	ID            string `json:"id"`
	Code          string `json:"code"`
	Name          string `json:"name"`
	CompanyName   string `json:"company_name"`
	LeadStatus    string `json:"lead_status"`
	LeadStatusColor string `json:"lead_status_color"`
	AssignedTo    string `json:"assigned_to"`
	CreatedAt     string `json:"created_at"`
}

// CRMPipelineStageSummary represents open deal counts grouped by pipeline stage.
type CRMPipelineStageSummary struct {
	StageID     string `json:"stage_id"`
	StageName   string `json:"stage_name"`
	StageOrder  int    `json:"stage_order"`
	ItemCount   int    `json:"item_count"`
}
