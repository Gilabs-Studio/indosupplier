"use client";

import { type ReactNode, useMemo } from "react";
import { useTranslations } from "next-intl";
import {
  Timer,
  CalendarClock,
  ArrowDownToLine,
  ArrowUpFromLine,
  Percent,
  Building2,
  TrendingUp,
  BarChart3,
  RefreshCw,
  Gauge,
  Truck,
  Activity as ActivityIcon,
  Clock,
  Lock,
} from "lucide-react";
import type { WidgetConfig, KpiCardData, WidgetType, OwnerKpiMetric } from "../types";
import { formatCurrency } from "@/lib/utils";
import { Card, CardContent } from "@/components/ui/card";
import { useDashboardWidgetOverview, isOverviewWidget, isOwnerKpiWidget } from "../hooks/use-dashboard";
import { useOwnerKpi } from "../hooks/use-owner-kpi";
import dynamic from "next/dynamic";
import { ChartWidget as _ChartWidgetType } from "./chart-widget";
import { BalanceWidget } from "./balance-widget";
import { CostsCategoryWidget } from "./costs-category-widget";
import { InvoicesSummaryWidget } from "./invoices-summary-widget";
import { RecentInvoicesWidget } from "./recent-invoices-widget";
import { DeliveryWidget } from "./delivery-widget";
import { SalesPerformanceWidget } from "./sales-performance-widget";
// Dynamically import heavy widgets to avoid bundling charting and map libraries
const ChartWidget = dynamic<Parameters<typeof _ChartWidgetType>[0]>(() =>
  import("./chart-widget").then((m) => m.ChartWidget),
  {
    ssr: false,
    loading: () => (
      <Card className="h-full border-dashed">
        <CardContent className="flex h-48 items-center justify-center">
          <div className="h-4 w-1/3 animate-pulse rounded bg-muted" />
        </CardContent>
      </Card>
    ),
  },
);

const GeoWidget = dynamic(() => import("./geo-widget").then((m) => m.GeoWidget), { ssr: false });
import { WarehouseWidget } from "./warehouse-widget";
const RevenueBarChartCard = dynamic(() => import("./revenue-bar-chart-card").then((m) => m.RevenueBarChartCard), {
  ssr: false,
  loading: () => (
    <Card className="h-full border-dashed">
      <CardContent className="flex h-48 items-center justify-center">
        <div className="h-4 w-1/3 animate-pulse rounded bg-muted" />
      </CardContent>
    </Card>
  ),
});
import { StatSummaryCard } from "./stat-summary-card";
import { BestSellingCard } from "./best-selling-card";
import { TrackOrderCard } from "./track-order-card";
import { TrackPurchaseOrderCard } from "./track-purchase-order-card";
import { PendingApprovalsSalesWidget } from "./pending-approvals-sales-widget";
import { PendingApprovalsPurchaseWidget } from "./pending-approvals-purchase-widget";
import { TravelPlannerWidget } from "./travel-planner-widget";
import { OwnerKpiCard } from "./owner-kpi-card";
import { OwnerIntelligenceWidget } from "./owner-intelligence-widget";
import { OpexCapexWidget } from "./opex-capex-widget";
import { WidgetAsyncState } from "./widget-async-state";
import { collectAccessibleMenuUrls, isWidgetVisibleByAccessibleMenus, WIDGET_REGISTRY } from "../config/widget-registry";
import { CrmKpiWidget } from "./crm-kpi-widget";
import { PosOutletSalesWidget } from "./pos-outlet-sales-widget";
import { PosLiveTableWidget } from "./pos-live-table-widget";
import { PosCashControlWidget } from "./pos-cash-control-widget";
import { HrWidget } from "./hr-widget";
import { usePermissionScope } from "@/features/master-data/user-management/hooks/use-has-permission";
import { useUserPermissions } from "@/features/master-data/user-management/hooks/use-user-permissions";
import { isDashboardWidgetRestrictedByScope } from "../config/scope-guard";

function formatNumber(value: number): string {
  return new Intl.NumberFormat("id-ID", {
    maximumFractionDigits: 0,
  }).format(value);
}

interface WidgetRendererProps {
  readonly widget: WidgetConfig;
}

function isNonOverviewWidget(type: WidgetType): boolean {
  return (
    type === "track_orders" ||
    type === "track_purchase_orders" ||
    type === "pending_approvals_sales" ||
    type === "pending_approvals_purchase"
  );
}

const KPI_ICONS: Record<string, ReactNode> = {
  kpi_ccc: <Timer className="h-4 w-4" />,
  kpi_dio: <CalendarClock className="h-4 w-4" />,
  kpi_ar_days: <ArrowDownToLine className="h-4 w-4" />,
  kpi_ap_days: <ArrowUpFromLine className="h-4 w-4" />,
  kpi_roe: <Percent className="h-4 w-4" />,
  kpi_roa: <Building2 className="h-4 w-4" />,
  kpi_net_profit_margin: <TrendingUp className="h-4 w-4" />,
  kpi_gross_profit_margin: <BarChart3 className="h-4 w-4" />,
  kpi_inventory_turnover: <RefreshCw className="h-4 w-4" />,
  kpi_asset_turnover: <Gauge className="h-4 w-4" />,
  kpi_cost_per_delivery: <Truck className="h-4 w-4" />,
  kpi_utilization_rate: <ActivityIcon className="h-4 w-4" />,
  kpi_otd_rate: <Clock className="h-4 w-4" />,
};

export function WidgetRenderer({ widget }: WidgetRendererProps) {
  const t = useTranslations("dashboard");
  const tAuthBlock = useTranslations("auth.block");
  const { data: permissionsData, isLoading: isPermissionsLoading } = useUserPermissions();
  const registry = WIDGET_REGISTRY[widget.type];
  const permissionScope = usePermissionScope(registry?.permission ?? "");
  const accessibleMenuUrls = useMemo(
    () => collectAccessibleMenuUrls(permissionsData?.data.menus ?? []),
    [permissionsData?.data.menus],
  );
  const canView = isWidgetVisibleByAccessibleMenus(widget.type, accessibleMenuUrls);
  const isScopeRestricted = isDashboardWidgetRestrictedByScope(widget.type, permissionScope);

  const {
    data,
    isLoading,
    isError,
    refetch,
  } = useDashboardWidgetOverview(widget.type, {
    enabled: canView && !isScopeRestricted && isOverviewWidget(widget.type) && !isOwnerKpiWidget(widget.type),
  });

  const ownerKpi = useOwnerKpi({
    enabled: canView && !isScopeRestricted && isOwnerKpiWidget(widget.type),
  });

  if (isPermissionsLoading) {
    return (
      <Card className="h-full border-dashed">
        <CardContent className="flex min-h-40 items-center p-6">
          <div className="w-full space-y-3">
            <div className="h-5 w-1/3 animate-pulse rounded bg-muted" />
            <div className="h-4 w-full animate-pulse rounded bg-muted" />
            <div className="h-4 w-5/6 animate-pulse rounded bg-muted" />
          </div>
        </CardContent>
      </Card>
    );
  }

  const costItems = data?.costs_by_category ?? [];
  const totalExpenseValue = costItems.reduce((sum, item) => sum + item.amount, 0);
  const totalExpense: KpiCardData = {
    value: totalExpenseValue,
    formatted: formatCurrency(totalExpenseValue),
  };

  // Keep dashboard figures readable and consistent by avoiding compact abbreviations.
  const revenueSummary = data?.kpi?.total_revenue
    ? {
        ...data.kpi.total_revenue,
        formatted: formatCurrency(data.kpi.total_revenue.value),
      }
    : undefined;

  const ordersSummary = data?.kpi?.total_orders
    ? {
        ...data.kpi.total_orders,
        formatted: formatNumber(data.kpi.total_orders.value),
      }
    : undefined;

  const balanceSummary = data?.balance_overview
    ? {
        ...data.balance_overview,
        formatted: formatCurrency(data.balance_overview.value),
      }
    : undefined;

  if (!registry) return null;
  if (!canView) {
    return (
      <Card className="h-full border-dashed">
        <CardContent className="flex min-h-40 flex-col items-center justify-center gap-2 p-6 text-center">
          <Lock className="h-4 w-4 text-muted-foreground" aria-hidden="true" />
          <p className="text-xs text-muted-foreground">{tAuthBlock("description")}</p>
        </CardContent>
      </Card>
    );
  }

  if (isScopeRestricted) {
    return (
      <Card className="h-full border-dashed">
        <CardContent className="flex min-h-40 flex-col items-center justify-center gap-2 p-6 text-center">
          <Lock className="h-4 w-4 text-muted-foreground" aria-hidden="true" />
          <p className="text-sm font-medium">{t("scopeRestrictedTitle")}</p>
          <p className="text-xs text-muted-foreground">{t("scopeRestrictedDescription")}</p>
        </CardContent>
      </Card>
    );
  }

  // ── Non-overview widgets (self-fetching) ──
  if (isNonOverviewWidget(widget.type)) {
    switch (widget.type) {
      case "track_orders":
        return <TrackOrderCard />;
      case "track_purchase_orders":
        return <TrackPurchaseOrderCard />;
      case "pending_approvals_sales":
        return <PendingApprovalsSalesWidget />;
      case "pending_approvals_purchase":
        return <PendingApprovalsPurchaseWidget />;
      default:
        return null;
    }
  }

  // ── Owner KPI widgets ──
  if (isOwnerKpiWidget(widget.type)) {
    const kpiData = ownerKpi?.data;

    // Special full-width widgets
    if (widget.type === "owner_intelligence") {
      return (
        <WidgetAsyncState
          isLoading={Boolean(ownerKpi?.isLoading)}
          isError={Boolean(ownerKpi?.isError)}
          onRetry={() => {
            void ownerKpi?.refetch();
          }}
        >
          <OwnerIntelligenceWidget data={kpiData?.intelligence} />
        </WidgetAsyncState>
      );
    }

    if (widget.type === "kpi_opex_vs_capex") {
      return (
        <WidgetAsyncState
          isLoading={Boolean(ownerKpi?.isLoading)}
          isError={Boolean(ownerKpi?.isError)}
          onRetry={() => {
            void ownerKpi?.refetch();
          }}
        >
          <OpexCapexWidget data={kpiData?.cost_structure} />
        </WidgetAsyncState>
      );
    }

    // Individual KPI cards
    const icon = KPI_ICONS[widget.type];
    if (!icon) {
      return null;
    }

    // Map widget type to the correct metric from owner KPI data
    const metricMap: Record<string, () => OwnerKpiMetric | undefined> = {
      kpi_roe: () => kpiData?.profitability?.roe,
      kpi_net_profit_margin: () => kpiData?.profitability?.net_profit_margin,
      kpi_gross_profit_margin: () => kpiData?.profitability?.gross_profit_margin,
      kpi_inventory_turnover: () => kpiData?.inventory?.inventory_turnover,
      kpi_dio: () => kpiData?.inventory?.dio,
      kpi_ar_days: () => kpiData?.cashflow?.ar_days,
      kpi_ap_days: () => kpiData?.cashflow?.ap_days,
      kpi_ccc: () => kpiData?.cashflow?.ccc,
      kpi_cost_per_delivery: () => kpiData?.logistics?.cost_per_delivery,
      kpi_utilization_rate: () => kpiData?.logistics?.utilization_rate,
      kpi_otd_rate: () => kpiData?.logistics?.otd_rate,
      kpi_roa: () => kpiData?.asset?.roa,
      kpi_asset_turnover: () => kpiData?.asset?.asset_turnover,
    };

    const getMetric = metricMap[widget.type];
    const metric = getMetric?.();

    const formulaKey = `ownerKpi.metrics.${widget.type}.formula` as const;
    const purposeKey = `ownerKpi.metrics.${widget.type}.purpose` as const;

    if (!metric) {
      return (
        <WidgetAsyncState
          isLoading={Boolean(ownerKpi?.isLoading)}
          isError={Boolean(ownerKpi?.isError)}
          onRetry={() => {
            void ownerKpi?.refetch();
          }}
        >
          <OwnerKpiCard
            title={t(registry.titleKey as Parameters<typeof t>[0])}
            metric={{
              value: 0,
              formatted: "—",
              status: "warning",
              status_label: "No Data",
              unit: "",
            }}
            formula={t(formulaKey as Parameters<typeof t>[0])}
            purpose={t(purposeKey as Parameters<typeof t>[0])}
            icon={icon}
          />
        </WidgetAsyncState>
      );
    }

    return (
      <WidgetAsyncState
        isLoading={Boolean(ownerKpi?.isLoading)}
        isError={Boolean(ownerKpi?.isError)}
        onRetry={() => {
          void ownerKpi?.refetch();
        }}
      >
        <OwnerKpiCard
          title={t(registry.titleKey as Parameters<typeof t>[0])}
          metric={metric}
          formula={t(formulaKey as Parameters<typeof t>[0])}
          purpose={t(purposeKey as Parameters<typeof t>[0])}
          icon={icon}
        />
      </WidgetAsyncState>
    );
  }

  // ── Existing widgets (legacy) ──
  let content: ReactNode = null;

  switch (widget.type) {
    // ---- Legacy KPI widgets (rendered as StatSummaryCard for UI consistency) ----
    case "total_revenue":
      content = <StatSummaryCard label={t("widgets.total_revenue.title")} data={data?.kpi?.total_revenue} />;
      break;
    case "total_orders":
      content = <StatSummaryCard label={t("widgets.total_orders.title")} data={data?.kpi?.total_orders} />;
      break;
    case "total_customers":
      content = <StatSummaryCard label={t("widgets.total_customers.title")} data={data?.kpi?.total_customers} />;
      break;
    case "total_products":
      content = <StatSummaryCard label={t("widgets.total_products.title")} data={data?.kpi?.total_products} />;
      break;
    case "employee_count":
      content = <StatSummaryCard label={t("widgets.employee_count.title")} data={data?.kpi?.employee_count} />;
      break;

    // ---- Legacy chart widgets ----
    case "revenue_chart":
      content = <ChartWidget widgetType="revenue_chart" data={data?.revenue_chart} />;
      break;
    case "costs_chart":
      content = <ChartWidget widgetType="costs_chart" data={data?.costs_chart} />;
      break;
    case "revenue_vs_costs":
      content = <ChartWidget widgetType="revenue_vs_costs" data={data?.revenue_vs_costs} />;
      break;
    case "balance_overview":
      content = <BalanceWidget data={data?.balance_overview} />;
      break;
    case "costs_by_category":
      content = <CostsCategoryWidget data={data?.costs_by_category} />;
      break;
    case "invoices_summary":
      content = <InvoicesSummaryWidget data={data?.invoices_summary} />;
      break;
    case "recent_invoices":
      content = <RecentInvoicesWidget data={data?.recent_invoices} />;
      break;
    case "delivery_status":
      content = <DeliveryWidget data={data?.delivery_status} />;
      break;
    case "sales_performance":
      content = <SalesPerformanceWidget data={data?.sales_performance} />;
      break;
    case "geographic_overview":
      content = <GeoWidget data={data?.geographic_overview} />;
      break;
    case "warehouse_overview":
      content = <WarehouseWidget data={data?.warehouse_overview} />;
      break;

    // ---- New composite widgets (reference Sales Dashboard) ----
    case "revenue_bar_chart":
      content = (
        <RevenueBarChartCard
          revenueData={data?.revenue_chart}
          costsData={data?.costs_chart}
        />
      );
      break;
    case "stat_summary_balance":
      content = <StatSummaryCard label={t("stats.totalBalance")} data={balanceSummary} />;
      break;
    case "stat_summary_revenue":
      content = <StatSummaryCard label={t("stats.totalRevenue")} data={revenueSummary} />;
      break;
    case "stat_summary_expense":
      content = <StatSummaryCard label={t("stats.totalExpense")} data={totalExpense} />;
      break;
    case "stat_summary_orders":
      content = <StatSummaryCard label={t("stats.totalOrders")} data={ordersSummary} />;
      break;
    case "best_selling":
      content = <BestSellingCard data={data?.top_products} />;
      break;
    case "travel_planner_overview":
      content = <TravelPlannerWidget />;
      break;

    // ---- CRM widgets ----
    case "crm_total_contacts":
    case "crm_active_leads":
    case "crm_leads_list":
    case "crm_pipeline_summary":
    case "crm_activity_summary":
      return <CrmKpiWidget widget={widget} />;

    // ---- POS widgets ----
    case "pos_outlet_sales":
      return <PosOutletSalesWidget widget={widget} />;
    case "pos_live_table_overview":
      return <PosLiveTableWidget widget={widget} />;
    case "pos_cash_control":
      return <PosCashControlWidget widget={widget} />;

    // ---- HR widgets ----
    case "hr_attendance_today":
    case "hr_pending_leaves":
      return <HrWidget widget={widget} />;

    default:
      content = null;
      break;
  }

  if (!content) return null;

  return (
    <WidgetAsyncState
      isLoading={isLoading}
      isError={isError}
      onRetry={() => {
        void refetch();
      }}
    >
      {content}
    </WidgetAsyncState>
  );
}

