import type {
  WidgetType,
  WidgetRegistryEntry,
  WidgetConfig,
  WidgetColSpan,
  WidgetRowSpan,
  WidgetSize,
} from "../types";

// Maps legacy WidgetSize to a column span for backward-compatible hydration
const SIZE_TO_COL: Record<WidgetSize, WidgetColSpan> = {
  sm: 1,
  md: 2,
  lg: 3,
  xl: 4,
};

/**
 * Resolves the effective colSpan and rowSpan for a widget.
 * Uses explicit overrides when present, otherwise falls back to registry defaults,
 * then to legacy `size` mapping. Safe to call with old DB layouts that lack colSpan/rowSpan.
 */
export function resolveWidgetSpan(
  widget: WidgetConfig,
): { col: WidgetColSpan; row: WidgetRowSpan } {
  const registry = WIDGET_REGISTRY[widget.type];
  const col = (widget.colSpan ??
    registry?.defaultColSpan ??
    SIZE_TO_COL[widget.size] ??
    1) as WidgetColSpan;
  const row = (widget.rowSpan ?? registry?.defaultRowSpan ?? 1) as WidgetRowSpan;
  return { col, row };
}

type AccessibleMenuNode = {
  url: string;
  actions?: Array<{ access: boolean }>;
  children?: AccessibleMenuNode[];
};

function normalizeMenuUrl(url: string): string {
  const trimmed = url.trim();
  if (!trimmed) return "";
  const withLeadingSlash = trimmed.startsWith("/") ? trimmed : `/${trimmed}`;
  return withLeadingSlash.replace(/\/+$/, "") || "/";
}

export function collectAccessibleMenuUrls(menus: AccessibleMenuNode[]): Set<string> {
  const accessibleMenuUrls = new Set<string>();

  const visit = (menu: AccessibleMenuNode): boolean => {
    const childAccessible = (menu.children ?? []).some(visit);
    const ownAccessible = (menu.actions ?? []).some((action) => action.access);
    const isAccessible = ownAccessible || childAccessible;

    if (isAccessible) {
      const normalized = normalizeMenuUrl(menu.url);
      if (normalized) {
        accessibleMenuUrls.add(normalized);
      }
    }

    return isAccessible;
  };

  menus.forEach(visit);
  return accessibleMenuUrls;
}

export function isWidgetVisibleByAccessibleMenus(
  type: WidgetType,
  accessibleMenuUrls: Set<string>,
): boolean {
  const widgetMenuUrl = getWidgetMenuUrl(type);
  if (!widgetMenuUrl) {
    return false;
  }

  const normalizedWidgetMenuUrl = normalizeMenuUrl(widgetMenuUrl);
  if (!normalizedWidgetMenuUrl) {
    return false;
  }

  for (const accessibleMenuUrl of accessibleMenuUrls) {
    if (
      normalizedWidgetMenuUrl === accessibleMenuUrl ||
      normalizedWidgetMenuUrl.startsWith(`${accessibleMenuUrl}/`) ||
      accessibleMenuUrl.startsWith(`${normalizedWidgetMenuUrl}/`)
    ) {
      return true;
    }
  }

  return false;
}

export function buildAccessibleDefaultWidgets(accessibleMenuUrls: Set<string>): WidgetConfig[] {
  return Object.values(WIDGET_REGISTRY)
    .filter((entry) => isWidgetVisibleByAccessibleMenus(entry.type, accessibleMenuUrls))
    .map((entry, index) => ({
      id: `w-${index + 1}`,
      type: entry.type,
      title: "",
      size: entry.defaultSize,
      colSpan: entry.defaultColSpan,
      rowSpan: entry.defaultRowSpan,
      order: index,
      visible: true,
    }));
}

/**
 * Maps each widget type to its associated menu URL.
 * Widget visibility is derived from the role's menu permissions.
 * Example: POS widgets only show if role has /pos menu access.
 */
export function getWidgetMenuUrl(type: WidgetType): string | null {
  const widgetMenuMap: Record<WidgetType, string | null> = {
    // POS widgets → /pos
    pos_outlet_sales: "/pos",
    pos_live_table_overview: "/pos",
    pos_cash_control: "/pos",

    // Sales widgets → /sales
    total_orders: "/sales",
    sales_performance: "/sales",
    best_selling: "/sales",
    track_orders: "/sales",
    stat_summary_orders: "/sales",
    stat_summary_revenue: "/sales",
    revenue_bar_chart: "/sales",

    // Finance widgets → /finance
    total_revenue: "/finance",
    revenue_chart: "/finance",
    costs_chart: "/finance",
    revenue_vs_costs: "/finance",
    balance_overview: "/finance",
    costs_by_category: "/finance",
    stat_summary_balance: "/finance",
    stat_summary_expense: "/finance",
    kpi_ccc: "/finance",
    kpi_roe: "/finance",
    kpi_roa: "/finance",
    kpi_net_profit_margin: "/finance",
    kpi_gross_profit_margin: "/finance",
    kpi_asset_turnover: "/finance",
    kpi_ap_days: "/finance",
    kpi_opex_vs_capex: "/finance",

    // Inventory/Stock widgets → /stock
    total_products: "/stock",
    warehouse_overview: "/stock",
    kpi_dio: "/stock",
    kpi_inventory_turnover: "/stock",

    // Customer Invoice widgets → /sales
    invoices_summary: "/sales",
    recent_invoices: "/sales",
    kpi_ar_days: "/sales",

    // Purchase widgets → /purchase
    track_purchase_orders: "/purchase",
    pending_approvals_purchase: "/purchase",

    // Delivery widgets → /sales
    delivery_status: "/sales",
    kpi_cost_per_delivery: "/sales",
    kpi_utilization_rate: "/sales",
    kpi_otd_rate: "/sales",

    // Master data widgets → /master-data
    employee_count: "/master-data",
    total_customers: "/master-data",

    // CRM widgets → /crm
    crm_total_contacts: "/crm",
    crm_active_leads: "/crm",
    crm_leads_list: "/crm",
    crm_pipeline_summary: "/crm",
    crm_activity_summary: "/crm",

    // Travel Planner widgets → /travel
    travel_planner_overview: "/travel",

    // HR widgets → /hrd
    hr_attendance_today: "/hrd",
    hr_pending_leaves: "/hrd",

    // Special widgets are intentionally not exposed unless a matching menu exists.
    owner_intelligence: null,
    geographic_overview: null,
    pending_approvals_sales: "/sales",
  };

  return widgetMenuMap[type] ?? null;
}

/** Complete registry of all available dashboard widgets */
export const WIDGET_REGISTRY: Record<WidgetType, WidgetRegistryEntry> = {
  // ════════════════════════════════════════════════════════════════════════
  // Legacy / existing widgets
  // ════════════════════════════════════════════════════════════════════════
  total_revenue: {
    type: "total_revenue",
    category: "erp",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "widgets.total_revenue.title",
    descriptionKey: "widgets.total_revenue.description",
    icon: "DollarSign",
    permission: "sales_order.read",
  },
  total_orders: {
    type: "total_orders",
    category: "erp",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "widgets.total_orders.title",
    descriptionKey: "widgets.total_orders.description",
    icon: "ShoppingCart",
    permission: "sales_order.read",
  },
  total_customers: {
    type: "total_customers",
    category: "crm",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "widgets.total_customers.title",
    descriptionKey: "widgets.total_customers.description",
    icon: "Users",
    permission: "customer.read",
  },
  total_products: {
    type: "total_products",
    category: "erp",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "widgets.total_products.title",
    descriptionKey: "widgets.total_products.description",
    icon: "Package",
    permission: "product.read",
  },
  employee_count: {
    type: "employee_count",
    category: "hr",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "widgets.employee_count.title",
    descriptionKey: "widgets.employee_count.description",
    icon: "UserCheck",
    permission: "employee.read",
  },
  revenue_chart: {
    type: "revenue_chart",
    category: "finance",
    defaultSize: "lg",
    defaultColSpan: 3,
    defaultRowSpan: 2,
    minColSpan: 2,
    minRowSpan: 1,
    titleKey: "widgets.revenue_chart.title",
    descriptionKey: "widgets.revenue_chart.description",
    icon: "TrendingUp",
    permission: "sales_order.read",
  },
  costs_chart: {
    type: "costs_chart",
    category: "finance",
    defaultSize: "lg",
    defaultColSpan: 3,
    defaultRowSpan: 2,
    minColSpan: 2,
    minRowSpan: 1,
    titleKey: "widgets.costs_chart.title",
    descriptionKey: "widgets.costs_chart.description",
    icon: "TrendingDown",
    permission: "journal.read",
  },
  revenue_vs_costs: {
    type: "revenue_vs_costs",
    category: "finance",
    defaultSize: "lg",
    defaultColSpan: 3,
    defaultRowSpan: 2,
    minColSpan: 2,
    minRowSpan: 1,
    titleKey: "widgets.revenue_vs_costs.title",
    descriptionKey: "widgets.revenue_vs_costs.description",
    icon: "BarChart3",
    permission: "journal.read",
  },
  balance_overview: {
    type: "balance_overview",
    category: "finance",
    defaultSize: "md",
    defaultColSpan: 2,
    defaultRowSpan: 1,
    titleKey: "widgets.balance_overview.title",
    descriptionKey: "widgets.balance_overview.description",
    icon: "Wallet",
    permission: "journal.read",
  },
  costs_by_category: {
    type: "costs_by_category",
    category: "finance",
    defaultSize: "md",
    defaultColSpan: 2,
    defaultRowSpan: 2,
    minColSpan: 2,
    minRowSpan: 1,
    titleKey: "widgets.costs_by_category.title",
    descriptionKey: "widgets.costs_by_category.description",
    icon: "PieChart",
    permission: "journal.read",
  },
  invoices_summary: {
    type: "invoices_summary",
    category: "finance",
    defaultSize: "md",
    defaultColSpan: 2,
    defaultRowSpan: 1,
    titleKey: "widgets.invoices_summary.title",
    descriptionKey: "widgets.invoices_summary.description",
    icon: "FileText",
    permission: "customer_invoice.read",
  },
  recent_invoices: {
    type: "recent_invoices",
    category: "finance",
    defaultSize: "lg",
    defaultColSpan: 3,
    defaultRowSpan: 2,
    minColSpan: 2,
    minRowSpan: 1,
    titleKey: "widgets.recent_invoices.title",
    descriptionKey: "widgets.recent_invoices.description",
    icon: "Receipt",
    permission: "customer_invoice.read",
  },
  sales_performance: {
    type: "sales_performance",
    category: "erp",
    defaultSize: "lg",
    defaultColSpan: 3,
    defaultRowSpan: 2,
    minColSpan: 2,
    minRowSpan: 1,
    titleKey: "widgets.sales_performance.title",
    descriptionKey: "widgets.sales_performance.description",
    icon: "Award",
    permission: "sales_order.read",
  },
  delivery_status: {
    type: "delivery_status",
    category: "erp",
    defaultSize: "md",
    defaultColSpan: 2,
    defaultRowSpan: 1,
    titleKey: "widgets.delivery_status.title",
    descriptionKey: "widgets.delivery_status.description",
    icon: "Truck",
    permission: "delivery_order.read",
  },
  geographic_overview: {
    type: "geographic_overview",
    category: "other",
    defaultSize: "xl",
    defaultColSpan: 4,
    defaultRowSpan: 3,
    minColSpan: 3,
    minRowSpan: 2,
    titleKey: "widgets.geographic_overview.title",
    descriptionKey: "widgets.geographic_overview.description",
    icon: "Map",
    permission: "report_geo_performance.read",
  },
  warehouse_overview: {
    type: "warehouse_overview",
    category: "erp",
    defaultSize: "lg",
    defaultColSpan: 3,
    defaultRowSpan: 2,
    minColSpan: 2,
    minRowSpan: 1,
    titleKey: "widgets.warehouse_overview.title",
    descriptionKey: "widgets.warehouse_overview.description",
    icon: "Warehouse",
    permission: "inventory.read",
  },
  // Composite widgets for the reference Sales Dashboard layout
  revenue_bar_chart: {
    type: "revenue_bar_chart",
    category: "finance",
    defaultSize: "xl",
    defaultColSpan: 4,
    defaultRowSpan: 2,
    minColSpan: 2,
    minRowSpan: 1,
    titleKey: "revenueChart.title",
    descriptionKey: "revenueChart.subtitle",
    icon: "BarChart3",
    permission: "sales_order.read",
  },
  stat_summary_balance: {
    type: "stat_summary_balance",
    category: "finance",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "stats.totalBalance",
    descriptionKey: "stats.totalBalance",
    icon: "Wallet",
    permission: "journal.read",
  },
  stat_summary_revenue: {
    type: "stat_summary_revenue",
    category: "finance",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "stats.totalRevenue",
    descriptionKey: "stats.totalRevenue",
    icon: "TrendingUp",
    permission: "sales_order.read",
  },
  stat_summary_expense: {
    type: "stat_summary_expense",
    category: "finance",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "stats.totalExpense",
    descriptionKey: "stats.totalExpense",
    icon: "TrendingDown",
    permission: "journal.read",
  },
  stat_summary_orders: {
    type: "stat_summary_orders",
    category: "erp",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "stats.totalOrders",
    descriptionKey: "stats.totalOrders",
    icon: "ShoppingCart",
    permission: "sales_order.read",
  },
  best_selling: {
    type: "best_selling",
    category: "erp",
    defaultSize: "md",
    defaultColSpan: 2,
    defaultRowSpan: 2,
    minColSpan: 2,
    minRowSpan: 1,
    titleKey: "bestSelling.title",
    descriptionKey: "bestSelling.subtitle",
    icon: "Star",
    permission: "sales_order.read",
  },
  track_orders: {
    type: "track_orders",
    category: "erp",
    defaultSize: "md",
    defaultColSpan: 2,
    defaultRowSpan: 2,
    minColSpan: 2,
    minRowSpan: 1,
    titleKey: "trackOrders.title",
    descriptionKey: "trackOrders.subtitle",
    icon: "Truck",
    permission: "sales_order.read",
  },
  track_purchase_orders: {
    type: "track_purchase_orders",
    category: "erp",
    defaultSize: "md",
    defaultColSpan: 2,
    defaultRowSpan: 2,
    minColSpan: 2,
    minRowSpan: 1,
    titleKey: "trackPurchaseOrders.title",
    descriptionKey: "trackPurchaseOrders.subtitle",
    icon: "PackageCheck",
    permission: "purchase_order.read",
  },
  pending_approvals_sales: {
    type: "pending_approvals_sales",
    category: "erp",
    defaultSize: "xl",
    defaultColSpan: 2,
    defaultRowSpan: 2,
    minColSpan: 2,
    minRowSpan: 1,
    titleKey: "widgets.pending_approvals_sales.title",
    descriptionKey: "widgets.pending_approvals_sales.description",
    icon: "ClipboardCheck",
  },
  pending_approvals_purchase: {
    type: "pending_approvals_purchase",
    category: "erp",
    defaultSize: "xl",
    defaultColSpan: 2,
    defaultRowSpan: 2,
    minColSpan: 2,
    minRowSpan: 1,
    titleKey: "widgets.pending_approvals_purchase.title",
    descriptionKey: "widgets.pending_approvals_purchase.description",
    icon: "PackageCheck",
  },
  travel_planner_overview: {
    type: "travel_planner_overview",
    category: "other",
    defaultSize: "xl",
    defaultColSpan: 4,
    defaultRowSpan: 2,
    minColSpan: 2,
    minRowSpan: 2,
    titleKey: "widgets.travel_planner_overview.title",
    descriptionKey: "widgets.travel_planner_overview.description",
    icon: "Route",
    permission: "travel_planner.read",
  },

  // ════════════════════════════════════════════════════════════════════════
  // Owner KPI widgets — Tier 1 (Cashflow — most critical)
  // ════════════════════════════════════════════════════════════════════════
  kpi_ccc: {
    type: "kpi_ccc",
    category: "finance",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "widgets.kpi_ccc.title",
    descriptionKey: "widgets.kpi_ccc.description",
    icon: "Timer",
    permission: "journal.read",
  },
  kpi_dio: {
    type: "kpi_dio",
    category: "finance",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "widgets.kpi_dio.title",
    descriptionKey: "widgets.kpi_dio.description",
    icon: "CalendarClock",
    permission: "inventory.read",
  },
  kpi_ar_days: {
    type: "kpi_ar_days",
    category: "finance",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "widgets.kpi_ar_days.title",
    descriptionKey: "widgets.kpi_ar_days.description",
    icon: "ArrowDownToLine",
    permission: "customer_invoice.read",
  },
  kpi_ap_days: {
    type: "kpi_ap_days",
    category: "finance",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "widgets.kpi_ap_days.title",
    descriptionKey: "widgets.kpi_ap_days.description",
    icon: "ArrowUpFromLine",
    permission: "journal.read",
  },

  // ════════════════════════════════════════════════════════════════════════
  // Owner KPI widgets — Tier 2 (Profitability & Asset)
  // ════════════════════════════════════════════════════════════════════════
  kpi_roe: {
    type: "kpi_roe",
    category: "finance",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "widgets.kpi_roe.title",
    descriptionKey: "widgets.kpi_roe.description",
    icon: "Percent",
    permission: "journal.read",
  },
  kpi_roa: {
    type: "kpi_roa",
    category: "finance",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "widgets.kpi_roa.title",
    descriptionKey: "widgets.kpi_roa.description",
    icon: "Building2",
    permission: "journal.read",
  },
  kpi_net_profit_margin: {
    type: "kpi_net_profit_margin",
    category: "finance",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "widgets.kpi_net_profit_margin.title",
    descriptionKey: "widgets.kpi_net_profit_margin.description",
    icon: "TrendingUp",
    permission: "journal.read",
  },
  kpi_gross_profit_margin: {
    type: "kpi_gross_profit_margin",
    category: "finance",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "widgets.kpi_gross_profit_margin.title",
    descriptionKey: "widgets.kpi_gross_profit_margin.description",
    icon: "BarChart3",
    permission: "journal.read",
  },
  kpi_inventory_turnover: {
    type: "kpi_inventory_turnover",
    category: "finance",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "widgets.kpi_inventory_turnover.title",
    descriptionKey: "widgets.kpi_inventory_turnover.description",
    icon: "RefreshCw",
    permission: "inventory.read",
  },
  kpi_asset_turnover: {
    type: "kpi_asset_turnover",
    category: "finance",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "widgets.kpi_asset_turnover.title",
    descriptionKey: "widgets.kpi_asset_turnover.description",
    icon: "Gauge",
    permission: "journal.read",
  },

  // ════════════════════════════════════════════════════════════════════════
  // Owner KPI widgets — Tier 3 (Logistics & Cost)
  // ════════════════════════════════════════════════════════════════════════
  kpi_cost_per_delivery: {
    type: "kpi_cost_per_delivery",
    category: "finance",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "widgets.kpi_cost_per_delivery.title",
    descriptionKey: "widgets.kpi_cost_per_delivery.description",
    icon: "Truck",
    permission: "delivery_order.read",
  },
  kpi_utilization_rate: {
    type: "kpi_utilization_rate",
    category: "finance",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "widgets.kpi_utilization_rate.title",
    descriptionKey: "widgets.kpi_utilization_rate.description",
    icon: "Activity",
    permission: "delivery_order.read",
  },
  kpi_otd_rate: {
    type: "kpi_otd_rate",
    category: "finance",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "widgets.kpi_otd_rate.title",
    descriptionKey: "widgets.kpi_otd_rate.description",
    icon: "Clock",
    permission: "delivery_order.read",
  },
  kpi_opex_vs_capex: {
    type: "kpi_opex_vs_capex",
    category: "finance",
    defaultSize: "md",
    defaultColSpan: 2,
    defaultRowSpan: 2,
    minColSpan: 2,
    minRowSpan: 1,
    titleKey: "widgets.kpi_opex_vs_capex.title",
    descriptionKey: "widgets.kpi_opex_vs_capex.description",
    icon: "PieChart",
    permission: "journal.read",
  },

  // ════════════════════════════════════════════════════════════════════════
  // Owner Intelligence
  // ════════════════════════════════════════════════════════════════════════
  owner_intelligence: {
    type: "owner_intelligence",
    category: "other",
    defaultSize: "xl",
    defaultColSpan: 4,
    defaultRowSpan: 2,
    minColSpan: 2,
    minRowSpan: 2,
    titleKey: "widgets.owner_intelligence.title",
    descriptionKey: "widgets.owner_intelligence.description",
    icon: "Brain",
  },

  // ════════════════════════════════════════════════════════════════════════
  // CRM Widgets
  // ════════════════════════════════════════════════════════════════════════
  crm_total_contacts: {
    type: "crm_total_contacts",
    category: "crm",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "widgets.crm_total_contacts.title",
    descriptionKey: "widgets.crm_total_contacts.description",
    icon: "ContactRound",
    permission: "crm_lead.read",
  },
  crm_active_leads: {
    type: "crm_active_leads",
    category: "crm",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "widgets.crm_active_leads.title",
    descriptionKey: "widgets.crm_active_leads.description",
    icon: "UserPlus",
    permission: "crm_lead.read",
  },
  crm_leads_list: {
    type: "crm_leads_list",
    category: "crm",
    defaultSize: "md",
    defaultColSpan: 2,
    defaultRowSpan: 2,
    titleKey: "widgets.crm_leads_list.title",
    descriptionKey: "widgets.crm_leads_list.description",
    icon: "ContactRound",
    permission: "crm_lead.read",
  },
  crm_pipeline_summary: {
    type: "crm_pipeline_summary",
    category: "crm",
    defaultSize: "md",
    defaultColSpan: 2,
    defaultRowSpan: 2,
    titleKey: "widgets.crm_pipeline_summary.title",
    descriptionKey: "widgets.crm_pipeline_summary.description",
    icon: "Handshake",
    permission: "crm_deal.read",
  },
  crm_activity_summary: {
    type: "crm_activity_summary",
    category: "crm",
    defaultSize: "md",
    defaultColSpan: 2,
    defaultRowSpan: 1,
    titleKey: "widgets.crm_activity_summary.title",
    descriptionKey: "widgets.crm_activity_summary.description",
    icon: "CalendarCheck",
    permission: "crm_task.read",
  },

  // ════════════════════════════════════════════════════════════════════════
  // POS Widgets
  // ════════════════════════════════════════════════════════════════════════
  pos_outlet_sales: {
    type: "pos_outlet_sales",
    category: "pos",
    defaultSize: "lg",
    defaultColSpan: 3,
    defaultRowSpan: 2,
    minColSpan: 2,
    minRowSpan: 1,
    titleKey: "widgets.pos_outlet_sales.title",
    descriptionKey: "widgets.pos_outlet_sales.description",
    icon: "Store",
    permission: "pos.order.create",
  },
  pos_live_table_overview: {
    type: "pos_live_table_overview",
    category: "pos",
    defaultSize: "xl",
    defaultColSpan: 4,
    defaultRowSpan: 2,
    minColSpan: 2,
    minRowSpan: 1,
    titleKey: "widgets.pos_live_table_overview.title",
    descriptionKey: "widgets.pos_live_table_overview.description",
    icon: "LayoutGrid",
    permission: "pos.order.create",
  },
  pos_cash_control: {
    type: "pos_cash_control",
    category: "pos",
    defaultSize: "lg",
    defaultColSpan: 2,
    defaultRowSpan: 2,
    minColSpan: 2,
    minRowSpan: 2,
    titleKey: "widgets.pos_cash_control.title",
    descriptionKey: "widgets.pos_cash_control.description",
    icon: "ReceiptText",
    permission: "pos.order.read",
  },

  // ════════════════════════════════════════════════════════════════════════
  // HR Widgets
  // ════════════════════════════════════════════════════════════════════════
  hr_attendance_today: {
    type: "hr_attendance_today",
    category: "hr",
    defaultSize: "md",
    defaultColSpan: 2,
    defaultRowSpan: 1,
    titleKey: "widgets.hr_attendance_today.title",
    descriptionKey: "widgets.hr_attendance_today.description",
    icon: "CalendarCheck2",
    permission: "attendance.read",
  },
  hr_pending_leaves: {
    type: "hr_pending_leaves",
    category: "hr",
    defaultSize: "sm",
    defaultColSpan: 1,
    defaultRowSpan: 1,
    titleKey: "widgets.hr_pending_leaves.title",
    descriptionKey: "widgets.hr_pending_leaves.description",
    icon: "ClipboardList",
    permission: "leave_request.read",
  },
};

/**
 * Default layout for new users — Owner KPI Dashboard
 * 
 * Layout structure (4-col grid):
 * Row 0:     Owner Intelligence (full width)
 * Row 1:     Tier 1 KPIs — CCC, DIO, AR Days, AP Days
 * Row 2:     Tier 2 KPIs — ROE, ROA, NPM, GPM
 * Row 3-4:   Revenue chart (3 col) + Cost structure (2 col, side)
 * Row 5:     Tier 3 — Inventory Turnover, Asset Turnover, cost_per_delivery, OTD Rate
 * Row 6-7:   Track Orders + Track Purchase Orders
 */
export const DEFAULT_WIDGETS: WidgetConfig[] = [
  // ── Owner Intelligence (full-width hero) ──
  { id: "w-1",  type: "owner_intelligence",      title: "", size: "xl", colSpan: 4, rowSpan: 2, order: 0,  visible: true },

  // ── Tier 1 KPIs: Cashflow & Inventory (priority) ──
  { id: "w-2",  type: "kpi_ccc",                 title: "", size: "sm", colSpan: 1, rowSpan: 1, order: 1,  visible: true },
  { id: "w-3",  type: "kpi_dio",                 title: "", size: "sm", colSpan: 1, rowSpan: 1, order: 2,  visible: true },
  { id: "w-4",  type: "kpi_ar_days",             title: "", size: "sm", colSpan: 1, rowSpan: 1, order: 3,  visible: true },
  { id: "w-5",  type: "kpi_ap_days",             title: "", size: "sm", colSpan: 1, rowSpan: 1, order: 4,  visible: true },

  // ── Tier 2 KPIs: Profitability & Asset ──
  { id: "w-6",  type: "kpi_roe",                 title: "", size: "sm", colSpan: 1, rowSpan: 1, order: 5,  visible: true },
  { id: "w-7",  type: "kpi_net_profit_margin",   title: "", size: "sm", colSpan: 1, rowSpan: 1, order: 6,  visible: true },
  { id: "w-8",  type: "kpi_gross_profit_margin",  title: "", size: "sm", colSpan: 1, rowSpan: 1, order: 7,  visible: true },
  { id: "w-9",  type: "kpi_roa",                 title: "", size: "sm", colSpan: 1, rowSpan: 1, order: 8,  visible: true },

  // ── Revenue chart + OPEX/CAPEX ──
  { id: "w-10", type: "revenue_bar_chart",        title: "", size: "md", colSpan: 2, rowSpan: 2, order: 9,  visible: true },
  { id: "w-11", type: "kpi_opex_vs_capex",        title: "", size: "md", colSpan: 2, rowSpan: 2, order: 10, visible: true },

  // ── Tier 3: Remaining KPIs ──
  { id: "w-12", type: "kpi_inventory_turnover",   title: "", size: "sm", colSpan: 1, rowSpan: 1, order: 11, visible: true },
  { id: "w-13", type: "kpi_asset_turnover",       title: "", size: "sm", colSpan: 1, rowSpan: 1, order: 12, visible: true },
  { id: "w-14", type: "kpi_otd_rate",             title: "", size: "sm", colSpan: 1, rowSpan: 1, order: 13, visible: true },
  { id: "w-15", type: "kpi_utilization_rate",     title: "", size: "sm", colSpan: 1, rowSpan: 1, order: 14, visible: true },

  // ── Operational widgets ──
  { id: "w-16", type: "track_orders",             title: "", size: "md", colSpan: 2, rowSpan: 2, order: 15, visible: true },
  { id: "w-17", type: "track_purchase_orders",    title: "", size: "md", colSpan: 2, rowSpan: 2, order: 16, visible: true },
];

/** Widget types grouped by category for the picker UI */
export function getWidgetsByCategory() {
  const grouped: Record<string, WidgetRegistryEntry[]> = {};
  for (const entry of Object.values(WIDGET_REGISTRY)) {
    if (!grouped[entry.category]) grouped[entry.category] = [];
    grouped[entry.category].push(entry);
  }
  return grouped;
}
