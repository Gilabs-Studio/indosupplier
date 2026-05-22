// Dashboard Widget System Types

/** Available widget categories for the widget picker */
export type WidgetCategory =
  | "erp"
  | "crm"
  | "pos"
  | "hr"
  | "finance"
  | "other";

/** Widget size presets for the grid */
export type WidgetSize = "sm" | "md" | "lg" | "xl";

/** Number of columns a widget occupies in the 4-column grid */
export type WidgetColSpan = 1 | 2 | 3 | 4;

/** Number of rows a widget occupies (controls height) */
export type WidgetRowSpan = 1 | 2 | 3;

/** Widget configuration stored per-user */
export interface WidgetConfig {
  id: string;
  type: WidgetType;
  title: string;
  size: WidgetSize;
  /** Column span override (1–4). Takes precedence over `size` when present. */
  colSpan?: WidgetColSpan;
  /** Row span override (1–3). Controls widget height. Defaults to 1 if absent. */
  rowSpan?: WidgetRowSpan;
  order: number;
  visible: boolean;
}

/** All available widget types in the system */
export type WidgetType =
  // --- Existing / Legacy ---
  | "total_revenue"
  | "total_orders"
  | "total_customers"
  | "total_products"
  | "revenue_chart"
  | "costs_chart"
  | "revenue_vs_costs"
  | "invoices_summary"
  | "recent_invoices"
  | "balance_overview"
  | "costs_by_category"
  | "sales_performance"
  | "geographic_overview"
  | "warehouse_overview"
  | "employee_count"
  | "delivery_status"
  | "revenue_bar_chart"
  | "stat_summary_balance"
  | "stat_summary_revenue"
  | "stat_summary_expense"
  | "stat_summary_orders"
  | "best_selling"
  | "track_orders"
  | "track_purchase_orders"
  | "pending_approvals_sales"
  | "pending_approvals_purchase"
  | "travel_planner_overview"
  // --- Owner KPI: Profitability ---
  | "kpi_roe"
  | "kpi_net_profit_margin"
  | "kpi_gross_profit_margin"
  // --- Owner KPI: Inventory ---
  | "kpi_inventory_turnover"
  | "kpi_dio"
  // --- Owner KPI: Cashflow ---
  | "kpi_ar_days"
  | "kpi_ap_days"
  | "kpi_ccc"
  // --- Owner KPI: Logistics ---
  | "kpi_cost_per_delivery"
  | "kpi_utilization_rate"
  | "kpi_otd_rate"
  // --- Owner KPI: Asset ---
  | "kpi_roa"
  | "kpi_asset_turnover"
  // --- Owner KPI: Cost Structure ---
  | "kpi_opex_vs_capex"
  // --- Owner Intelligence ---
  | "owner_intelligence"
  // --- CRM ---
  | "crm_total_contacts"
  | "crm_active_leads"
  | "crm_leads_list"
  | "crm_pipeline_summary"
  | "crm_activity_summary"
  // --- POS ---
  | "pos_outlet_sales"
  | "pos_live_table_overview"
  | "pos_cash_control"
  // --- HR ---
  | "hr_attendance_today"
  | "hr_pending_leaves";

/** Widget registry entry - metadata describing a widget */
export interface WidgetRegistryEntry {
  type: WidgetType;
  category: WidgetCategory;
  defaultSize: WidgetSize;
  /** Default column span in the 4-column grid. */
  defaultColSpan: WidgetColSpan;
  /** Default row span (height). */
  defaultRowSpan: WidgetRowSpan;
  /** Minimum column span — prevents the widget from being too narrow for its content. */
  minColSpan?: WidgetColSpan;
  /** Minimum row span. */
  minRowSpan?: WidgetRowSpan;
  titleKey: string;
  descriptionKey: string;
  icon: string;
  /** Permission code required to view this widget (e.g. "sales_order.read"). Omit for always-visible widgets. */
  permission?: string;
}

/** Dashboard layout persisted to localStorage */
export interface DashboardLayout {
  widgets: WidgetConfig[];
  updatedAt: string;
}

/** API response for a saved dashboard layout */
export interface DashboardLayoutApiResponse {
  success: boolean;
  data: {
    dashboard_type: string;
    layout_json: string;
  } | null;
}

/** Global date filter for the dashboard */
export interface DashboardDateFilter {
  from: string | null;
  to: string | null;
  year: number;
}

/** Scope selector for partial dashboard overview payloads */
export type DashboardOverviewScope =
  | "kpi"
  | "charts"
  | "balance"
  | "costs"
  | "invoices"
  | "sales-performance"
  | "products"
  | "delivery"
  | "geo"
  | "warehouse"
  | "owner-kpi"
  | "pos"
  | "hr"
  | "crm";

// ---- API Response Types ----

/** KPI summary card data */
export interface KpiCardData {
  value: number;
  formatted: string;
  change_percent?: number;
  previous_value?: number;
  previous_formatted?: string;
}

/** Chart series with period labels */
export interface ChartSeriesData {
  label: string;
  data: number[];
  formatted: string[];
}

/** Period-based chart */
export interface PeriodChartData {
  series: ChartSeriesData[];
  period: string[];
}

/** Category cost breakdown */
export interface CostCategoryItem {
  category: string;
  amount: number;
  amount_formatted: string;
  percentage: number;
}

/** Invoice row */
export interface InvoiceRow {
  id: string;
  customer_id?: string;
  company: string;
  issue_date: string;
  contact: string;
  value: number;
  value_formatted: string;
  status: "unpaid" | "paid" | "overdue";
}

/** Invoice summary counts */
export interface InvoiceSummaryData {
  total: number;
  unpaid: number;
  paid: number;
  overdue: number;
}

/** Sales performance row */
export interface SalesPerformanceRow {
  id: string;
  name: string;
  revenue: number;
  revenue_formatted: string;
  orders: number;
  target_percent: number;
}

/** Top product row */
export interface TopProductRow {
  id: string;
  name: string;
  sku: string;
  quantity_sold: number;
  revenue: number;
  revenue_formatted: string;
}

/** Delivery status breakdown */
export interface DeliveryStatusData {
  total: number;
  pending: number;
  in_transit: number;
  delivered: number;
  change_percent?: number;
}

/** Geographic overview data for choropleth */
export interface GeoOverviewData {
  regions: GeoRegionData[];
  total_value: number;
  total_formatted: string;
}

export interface GeoRegionData {
  name: string;
  code: string;
  value: number;
  formatted: string;
  count: number;
}

/** Warehouse overview */
export interface WarehouseOverviewData {
  warehouses: WarehouseItem[];
  total_stock_value: number;
  total_stock_formatted: string;
}

export interface WarehouseItem {
  id: string;
  name: string;
  location: string;
  stock_value: number;
  stock_formatted: string;
  item_count: number;
  in_stock_count: number;
  low_stock_count: number;
  out_of_stock_count: number;
}

// ---- Owner KPI Types ----

/** Health status for KPI indicators */
export type KpiHealthStatus = "good" | "warning" | "danger";

/** Single owner KPI metric with health assessment */
export interface OwnerKpiMetric {
  value: number;
  formatted: string;
  /** Status: good / warning / danger */
  status: KpiHealthStatus;
  /** Human-readable status label */
  status_label: string;
  /** Change vs previous period (percent) */
  change_percent?: number;
  /** Previous period's value */
  previous_value?: number;
  previous_formatted?: string;
  /** Unit for display (%, days, x, Rp, etc.) */
  unit: string;
}

/** Profitability KPI group */
export interface ProfitabilityKpi {
  roe: OwnerKpiMetric;
  net_profit_margin: OwnerKpiMetric;
  gross_profit_margin: OwnerKpiMetric;
}

/** Inventory KPI group */
export interface InventoryKpi {
  inventory_turnover: OwnerKpiMetric;
  dio: OwnerKpiMetric;
}

/** Cashflow KPI group */
export interface CashflowKpi {
  ar_days: OwnerKpiMetric;
  ap_days: OwnerKpiMetric;
  ccc: OwnerKpiMetric;
}

/** Logistics KPI group */
export interface LogisticsKpi {
  cost_per_delivery: OwnerKpiMetric;
  utilization_rate: OwnerKpiMetric;
  otd_rate: OwnerKpiMetric;
}

/** Asset KPI group */
export interface AssetKpi {
  roa: OwnerKpiMetric;
  asset_turnover: OwnerKpiMetric;
}

/** Cost structure KPI */
export interface CostStructureKpi {
  total_opex: number;
  total_opex_formatted: string;
  total_capex: number;
  total_capex_formatted: string;
  opex_ratio: number;
  capex_ratio: number;
  opex_breakdown: CostBreakdownItem[];
  capex_breakdown: CostBreakdownItem[];
  depreciation_total: number;
  depreciation_formatted: string;
}

export interface CostBreakdownItem {
  category: string;
  amount: number;
  amount_formatted: string;
  percentage: number;
}

/** Bottleneck area identified by intelligence layer */
export type BottleneckArea =
  | "inventory"
  | "cashflow"
  | "asset"
  | "logistics"
  | "profitability"
  | "cost";

/** Owner intelligence layer — auto-generated insights */
export interface OwnerIntelligence {
  /** Overall health: good / warning / danger */
  overall_health: KpiHealthStatus;
  /** Human-readable summary of business health */
  health_summary: string;
  /** Detected primary bottleneck */
  primary_bottleneck: BottleneckArea;
  /** Human-readable bottleneck explanation */
  bottleneck_summary: string;
  /** Individual insight items with severity */
  insights: OwnerInsightItem[];
  /** Timestamp of the analysis */
  analyzed_at: string;
}

export interface OwnerInsightItem {
  id: string;
  area: BottleneckArea;
  severity: KpiHealthStatus;
  title: string;
  description: string;
  /** Recommended action */
  action?: string;
}

/** Combined Owner KPI data from backend */
export interface OwnerKpiData {
  profitability: ProfitabilityKpi;
  inventory: InventoryKpi;
  cashflow: CashflowKpi;
  logistics?: LogisticsKpi;
  asset: AssetKpi;
  cost_structure: CostStructureKpi;
  intelligence: OwnerIntelligence;
}

/** Dashboard overview API response (aggregated) */
export interface DashboardOverviewResponse {
  success: boolean;
  data: DashboardOverviewData;
  timestamp: string;
  request_id: string;
}

export interface DashboardOverviewData {
  kpi: {
    total_revenue: KpiCardData;
    total_orders: KpiCardData;
    total_customers: KpiCardData;
    total_products: KpiCardData;
    employee_count: KpiCardData;
  };
  revenue_chart: PeriodChartData;
  costs_chart: PeriodChartData;
  revenue_vs_costs: PeriodChartData;
  balance_overview: KpiCardData & {
    chart_data: Array<{ period: string; value: number; formatted: string }>;
  };
  costs_by_category: CostCategoryItem[];
  invoices_summary: InvoiceSummaryData;
  recent_invoices: InvoiceRow[];
  sales_performance: SalesPerformanceRow[];
  top_products: TopProductRow[];
  delivery_status: DeliveryStatusData;
  geographic_overview: GeoOverviewData;
  warehouse_overview: WarehouseOverviewData;
  /** Owner KPI data — returned when scope=owner-kpi */
  owner_kpi?: OwnerKpiData;
  /** POS summary — returned when scope=pos */
  pos_summary?: POSSummaryData;
  /** HR summary — returned when scope=hr */
  hr_summary?: HRSummaryData;
  /** CRM summary — returned when scope=crm */
  crm_summary?: CRMSummaryData;
}

// ─── POS Summary Types ──────────────────────────────────────────────────────

export interface POSOutletItem {
  outlet_id: string;
  outlet_name: string;
  floor_plan_id?: string;
  active_table_count: number;
  active_table_labels?: string[];
  today_revenue: number;
  today_revenue_formatted: string;
  order_count: number;
  served_orders: number;
  waiting_orders: number;
  warning_orders: number;
  tables_occupied: number;
  tables_total: number;
}

export interface POSSummaryData {
  today_total_revenue: number;
  today_total_revenue_formatted: string;
  today_total_orders: number;
  served_orders: number;
  waiting_orders: number;
  warning_orders: number;
  active_sessions: number;
  opening_cash_today: number;
  opening_cash_today_formatted: string;
  cash_sales_today: number;
  cash_sales_today_formatted: string;
  cash_change_today: number;
  cash_change_today_formatted: string;
  expected_cash: number;
  expected_cash_formatted: string;
  actual_cash: number;
  actual_cash_formatted: string;
  cash_variance: number;
  cash_variance_formatted: string;
  outlets: POSOutletItem[];
}

// ─── HR Summary Types ────────────────────────────────────────────────────────

export interface HRSummaryData {
  total_employees: number;
  present_today: number;
  absent_today: number;
  late_today: number;
  pending_leaves: number;
  pending_overtime: number;
  attendance_rate: number;
}

// ─── CRM Summary Types ───────────────────────────────────────────────────────

export interface CRMSummaryData {
  total_contacts: number;
  active_leads: number;
  leads_this_month: number;
  deals_in_progress: number;
  deals_won_this_month: number;
  activities_today: number;
  recent_leads: CRMLeadSummary[];
  pipeline_stages: CRMPipelineStageSummary[];
}

export interface CRMLeadSummary {
  id: string;
  code: string;
  name: string;
  company_name: string;
  lead_status: string;
  lead_status_color: string;
  assigned_to: string;
  created_at: string;
}

export interface CRMPipelineStageSummary {
  stage_id: string;
  stage_name: string;
  stage_order: number;
  item_count: number;
}
