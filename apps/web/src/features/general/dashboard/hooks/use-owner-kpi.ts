import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { useDashboardStore } from "../stores/useDashboardStore";
import { dashboardService } from "../services/dashboard-service";
import type {
  OwnerKpiData,
  OwnerKpiMetric,
  KpiHealthStatus,
  OwnerIntelligence,
  OwnerInsightItem,
  BottleneckArea,
  CostStructureKpi,
  DashboardOverviewData,
} from "../types";

// ─── Health thresholds ──────────────────────────────────────────────────────

interface Threshold {
  good: number;
  warning: number;
  /** If true, higher values are better (e.g. ROE). If false, lower is better (e.g. DIO). */
  higherIsBetter: boolean;
}

const THRESHOLDS: Record<string, Threshold> = {
  roe: { good: 15, warning: 8, higherIsBetter: true },
  net_profit_margin: { good: 10, warning: 5, higherIsBetter: true },
  gross_profit_margin: { good: 30, warning: 15, higherIsBetter: true },
  inventory_turnover: { good: 6, warning: 3, higherIsBetter: true },
  dio: { good: 60, warning: 90, higherIsBetter: false },
  ar_days: { good: 30, warning: 60, higherIsBetter: false },
  ap_days: { good: 45, warning: 30, higherIsBetter: true }, // higher AP days = better leverage
  ccc: { good: 30, warning: 60, higherIsBetter: false },
  cost_per_delivery: { good: 500000, warning: 1000000, higherIsBetter: false },
  utilization_rate: { good: 80, warning: 60, higherIsBetter: true },
  otd_rate: { good: 95, warning: 85, higherIsBetter: true },
  roa: { good: 10, warning: 5, higherIsBetter: true },
  asset_turnover: { good: 1.5, warning: 0.8, higherIsBetter: true },
};

function getHealthStatus(
  value: number,
  thresholdKey: string,
): KpiHealthStatus {
  const t = THRESHOLDS[thresholdKey];
  if (!t) return "good";

  if (t.higherIsBetter) {
    if (value >= t.good) return "good";
    if (value >= t.warning) return "warning";
    return "danger";
  }
  // Lower is better
  if (value <= t.good) return "good";
  if (value <= t.warning) return "warning";
  return "danger";
}

function getStatusLabel(status: KpiHealthStatus): string {
  switch (status) {
    case "good":
      return "Healthy";
    case "warning":
      return "Watch";
    case "danger":
      return "Critical";
  }
}

// ─── Formatting helpers ────────────────────────────────────────────────────

function formatPercent(value: number): string {
  return `${value.toFixed(1)}%`;
}

function formatDays(value: number): string {
  return `${Math.round(value)} days`;
}

function formatTimes(value: number): string {
  return `${value.toFixed(1)}x`;
}

function formatCurrencyShort(value: number): string {
  if (value >= 1_000_000_000) return `Rp ${(value / 1_000_000_000).toFixed(1)}B`;
  if (value >= 1_000_000) return `Rp ${(value / 1_000_000).toFixed(1)}M`;
  if (value >= 1_000) return `Rp ${(value / 1_000).toFixed(0)}K`;
  return `Rp ${value.toFixed(0)}`;
}

type TranslationFn = ReturnType<typeof useTranslations>;

function safeDiv(numerator: number, denominator: number, fallback = 0): number {
  if (!denominator || !isFinite(denominator)) return fallback;
  const result = numerator / denominator;
  return isFinite(result) ? result : fallback;
}

// ─── Build KPI metric ──────────────────────────────────────────────────────

function buildMetric(
  value: number,
  thresholdKey: string,
  unit: string,
  formatter: (v: number) => string,
  changePercent?: number,
): OwnerKpiMetric {
  const status = getHealthStatus(value, thresholdKey);
  return {
    value,
    formatted: formatter(value),
    status,
    status_label: getStatusLabel(status),
    unit,
    change_percent: changePercent,
  };
}

// ─── Intelligence layer ────────────────────────────────────────────────────

function buildIntelligence(data: OwnerKpiData, t: TranslationFn): OwnerIntelligence {
  const insights: OwnerInsightItem[] = [];
  let worstArea: BottleneckArea = "profitability";
  let worstScore = 0;

  const areaScores: Record<BottleneckArea, number> = {
    profitability: 0,
    inventory: 0,
    cashflow: 0,
    logistics: 0,
    asset: 0,
    cost: 0,
  };

  // Score each area based on KPI health
  const scoreStatus = (s: KpiHealthStatus): number =>
    s === "danger" ? 3 : s === "warning" ? 1 : 0;

  // Profitability
  areaScores.profitability =
    scoreStatus(data.profitability.roe.status) +
    scoreStatus(data.profitability.net_profit_margin.status) +
    scoreStatus(data.profitability.gross_profit_margin.status);

  if (data.profitability.roe.status === "danger") {
    const key = "ownerKpi.insights.roe_low";
    insights.push({
      id: "profit-roe",
      area: "profitability",
      severity: "danger",
      title: t(`${key}.title`),
      description: t(`${key}.description`, { formatted: data.profitability.roe.formatted }),
      action: t(`${key}.action`),
    });
  }

  if (data.profitability.gross_profit_margin.status === "danger") {
    const key = "ownerKpi.insights.gpm_low";
    insights.push({
      id: "profit-gpm",
      area: "profitability",
      severity: "danger",
      title: t(`${key}.title`),
      description: t(`${key}.description`, { formatted: data.profitability.gross_profit_margin.formatted }),
      action: t(`${key}.action`),
    });
  }

  // Inventory
  areaScores.inventory =
    scoreStatus(data.inventory.inventory_turnover.status) * 2 +
    scoreStatus(data.inventory.dio.status) * 2;

  if (data.inventory.dio.status !== "good") {
    const key = "ownerKpi.insights.dio_high";
    insights.push({
      id: "inv-dio",
      area: "inventory",
      severity: data.inventory.dio.status,
      title: t(`${key}.title`),
      description: t(`${key}.description`, { formatted: data.inventory.dio.formatted }),
      action: t(`${key}.action`),
    });
  }

  // Cashflow (highest weight — most critical)
  areaScores.cashflow =
    (scoreStatus(data.cashflow.ccc.status) * 3) +
    (scoreStatus(data.cashflow.ar_days.status) * 2) +
    scoreStatus(data.cashflow.ap_days.status);

  if (data.cashflow.ccc.status === "danger") {
    const key = "ownerKpi.insights.ccc_danger";
    insights.push({
      id: "cf-ccc",
      area: "cashflow",
      severity: "danger",
      title: t(`${key}.title`),
      description: t(`${key}.description`, { formatted: data.cashflow.ccc.formatted }),
      action: t(`${key}.action`),
    });
  }

  if (data.cashflow.ar_days.status !== "good") {
    const key = "ownerKpi.insights.ar_slow";
    insights.push({
      id: "cf-ar",
      area: "cashflow",
      severity: data.cashflow.ar_days.status,
      title: t(`${key}.title`),
      description: t(`${key}.description`, { formatted: data.cashflow.ar_days.formatted }),
      action: t(`${key}.action`),
    });
  }

  // Logistics
  if (data.logistics) {
    areaScores.logistics =
      scoreStatus(data.logistics.otd_rate.status) * 2 +
      scoreStatus(data.logistics.utilization_rate.status) +
      scoreStatus(data.logistics.cost_per_delivery.status);

    if (data.logistics.otd_rate.status !== "good") {
      const key = "ownerKpi.insights.otd_low";
      insights.push({
        id: "log-otd",
        area: "logistics",
        severity: data.logistics.otd_rate.status,
        title: t(`${key}.title`),
        description: t(`${key}.description`, { formatted: data.logistics.otd_rate.formatted }),
        action: t(`${key}.action`),
      });
    }
  }

  // Asset
  areaScores.asset =
    scoreStatus(data.asset.roa.status) +
    scoreStatus(data.asset.asset_turnover.status);

  if (data.asset.roa.status === "danger") {
    const key = "ownerKpi.insights.asset_poor";
    insights.push({
      id: "asset-roa",
      area: "asset",
      severity: "danger",
      title: t(`${key}.title`),
      description: t(`${key}.description`, { formatted: data.asset.roa.formatted }),
      action: t(`${key}.action`),
    });
  }

  // Cost
  const opexRatio = data.cost_structure.opex_ratio;
  if (opexRatio > 70) {
    areaScores.cost += 2;
    const key = "ownerKpi.insights.opex_high";
    insights.push({
      id: "cost-opex",
      area: "cost",
      severity: "warning",
      title: t(`${key}.title`),
      description: t(`${key}.description`, { formatted: `${opexRatio.toFixed(0)}%` }),
      action: t(`${key}.action`),
    });
  }

  // Find worst area
  for (const [area, score] of Object.entries(areaScores)) {
    if (score > worstScore) {
      worstScore = score;
      worstArea = area as BottleneckArea;
    }
  }

  // Overall health
  const totalBadMetrics = Object.values(areaScores).reduce((s, v) => s + v, 0);
  const overallHealth: KpiHealthStatus =
    totalBadMetrics >= 8 ? "danger" : totalBadMetrics >= 3 ? "warning" : "good";

  return {
    overall_health: overallHealth,
    health_summary: t(`ownerKpi.healthSummaries.${overallHealth}`),
    primary_bottleneck: worstArea,
    bottleneck_summary: t(`ownerKpi.bottleneckSummaries.${worstArea}`),
    insights: insights.sort((a, b) => {
      const sev = { danger: 0, warning: 1, good: 2 };
      return sev[a.severity] - sev[b.severity];
    }),
    analyzed_at: new Date().toISOString(),
  };
}

// ─── Client-side KPI derivation from existing dashboard data ───────────────

function deriveOwnerKpiFromDashboard(
  data: Partial<DashboardOverviewData>,
  t: TranslationFn,
): OwnerKpiData {
  const revenue = data.kpi?.total_revenue?.value ?? 0;
  const totalOrders = data.kpi?.total_orders?.value ?? 0;

  // Derive approximate financials from available data
  const costsData = data.costs_by_category ?? [];
  const totalCosts = costsData.reduce((s, c) => s + c.amount, 0);
  const grossProfit = revenue - totalCosts * 0.6; // Approximate COGS as 60% of total costs
  const cogs = totalCosts * 0.6;
  const netProfit = revenue - totalCosts;
  const equity = revenue * 2.5; // rough estimate
  const totalAssets = revenue * 3;

  // Inventory estimates from warehouse data
  const warehouseData = data.warehouse_overview;
  const inventoryValue = warehouseData?.total_stock_value ?? revenue * 0.15;
  const avgInventory = inventoryValue * 1.1; // approximate

  // AR / AP estimates from invoices
  const invoiceData = data.invoices_summary;
  const arValue = (invoiceData?.unpaid ?? 0) + (invoiceData?.overdue ?? 0);
  const arEstimate = arValue > 0 ? arValue * 50000 : revenue * 0.08;
  const apEstimate = totalCosts * 0.12;

  // Delivery data
  const deliveryData = data.delivery_status;

  // Profitability
  const roe = safeDiv(netProfit, equity) * 100;
  const npm = safeDiv(netProfit, revenue) * 100;
  const gpm = safeDiv(grossProfit, revenue) * 100;

  // Inventory
  const inventoryTurnover = safeDiv(cogs, avgInventory);
  const dio = safeDiv(inventoryValue, cogs) * 365;

  // Cashflow
  const arDays = safeDiv(arEstimate, revenue) * 365;
  const apDays = safeDiv(apEstimate, cogs) * 365;
  const ccc = dio + arDays - apDays;

  // Asset
  const roa = safeDiv(netProfit, totalAssets) * 100;
  const assetTurnover = safeDiv(revenue, totalAssets);

  // Logistics
  const totalDeliveries = deliveryData?.total ?? totalOrders;
  const deliveredOnTime = deliveryData?.delivered ?? Math.round(totalDeliveries * 0.92);
  const otdRate = safeDiv(deliveredOnTime, totalDeliveries) * 100;
  const deliveryCost = totalCosts * 0.15;
  const costPerDelivery = safeDiv(deliveryCost, totalDeliveries);
  const utilizationRate = 75; // Placeholder — requires fleet data

  // Cost Structure
  const opex = totalCosts * 0.7;
  const capex = totalCosts * 0.3;
  const opexRatio = safeDiv(opex, totalCosts) * 100;
  const capexRatio = safeDiv(capex, totalCosts) * 100;

  const costStructure: CostStructureKpi = {
    total_opex: opex,
    total_opex_formatted: formatCurrencyShort(opex),
    total_capex: capex,
    total_capex_formatted: formatCurrencyShort(capex),
    opex_ratio: opexRatio,
    capex_ratio: capexRatio,
    opex_breakdown: [
      { category: "Salaries", amount: opex * 0.45, amount_formatted: formatCurrencyShort(opex * 0.45), percentage: 45 },
      { category: "Logistics", amount: opex * 0.25, amount_formatted: formatCurrencyShort(opex * 0.25), percentage: 25 },
      { category: "Rent", amount: opex * 0.15, amount_formatted: formatCurrencyShort(opex * 0.15), percentage: 15 },
      { category: "Admin", amount: opex * 0.15, amount_formatted: formatCurrencyShort(opex * 0.15), percentage: 15 },
    ],
    capex_breakdown: [
      { category: "Equipment", amount: capex * 0.5, amount_formatted: formatCurrencyShort(capex * 0.5), percentage: 50 },
      { category: "Vehicles", amount: capex * 0.3, amount_formatted: formatCurrencyShort(capex * 0.3), percentage: 30 },
      { category: "Warehouse", amount: capex * 0.2, amount_formatted: formatCurrencyShort(capex * 0.2), percentage: 20 },
    ],
    depreciation_total: capex * 0.15,
    depreciation_formatted: formatCurrencyShort(capex * 0.15),
  };

  const kpiData: OwnerKpiData = {
    profitability: {
      roe: buildMetric(roe, "roe", "%", formatPercent),
      net_profit_margin: buildMetric(npm, "net_profit_margin", "%", formatPercent),
      gross_profit_margin: buildMetric(gpm, "gross_profit_margin", "%", formatPercent),
    },
    inventory: {
      inventory_turnover: buildMetric(inventoryTurnover, "inventory_turnover", "x", formatTimes),
      dio: buildMetric(dio, "dio", "days", formatDays),
    },
    cashflow: {
      ar_days: buildMetric(arDays, "ar_days", "days", formatDays),
      ap_days: buildMetric(apDays, "ap_days", "days", formatDays),
      ccc: buildMetric(ccc, "ccc", "days", formatDays),
    },
    logistics: {
      cost_per_delivery: buildMetric(costPerDelivery, "cost_per_delivery", "Rp", formatCurrencyShort),
      utilization_rate: buildMetric(utilizationRate, "utilization_rate", "%", formatPercent),
      otd_rate: buildMetric(otdRate, "otd_rate", "%", formatPercent),
    },
    asset: {
      roa: buildMetric(roa, "roa", "%", formatPercent),
      asset_turnover: buildMetric(assetTurnover, "asset_turnover", "x", formatTimes),
    },
    cost_structure: costStructure,
    // Will be populated below
    intelligence: {} as OwnerIntelligence,
  };

  kpiData.intelligence = buildIntelligence(kpiData, t);
  return kpiData;
}

// ─── Main Hook ──────────────────────────────────────────────────────────────

export function useOwnerKpi(options?: { enabled?: boolean }) {
  const t = useTranslations("dashboard");
  const dateFilter = useDashboardStore((s) => s.dateFilter);

  // Try to get dedicated owner-kpi scope from backend
  const ownerKpiQuery = useQuery<Partial<DashboardOverviewData>>({
    queryKey: ["dashboard", "overview-scope", "owner-kpi", dateFilter],
    queryFn: () => dashboardService.getOverviewScope("owner-kpi", dateFilter),
    enabled: options?.enabled ?? true,
    staleTime: 60 * 1000,
    refetchOnWindowFocus: false,
    retry: 1,
  });

  // Fallback: fetch the general dashboard data to derive KPIs client-side
  const fallbackQuery = useQuery<Partial<DashboardOverviewData>>({
    queryKey: ["dashboard", "overview-scope", "kpi", dateFilter],
    queryFn: () => dashboardService.getOverviewScope("kpi", dateFilter),
    enabled: (options?.enabled ?? true) && !ownerKpiQuery.data?.owner_kpi,
    staleTime: 60 * 1000,
    refetchOnWindowFocus: false,
    retry: 1,
  });

  const ownerKpiData = useMemo<OwnerKpiData | undefined>(() => {
    // Prefer backend-calculated KPIs
    if (ownerKpiQuery.data?.owner_kpi) {
      const kpi = ownerKpiQuery.data.owner_kpi;
      // Ensure intelligence is populated
      if (kpi.intelligence?.overall_health) {
        return kpi;
      }
      return {
        ...kpi,
        intelligence: buildIntelligence(kpi, t),
      };
    }

    // Fallback: derive from general dashboard data
    const mergedData = {
      ...fallbackQuery.data,
      ...ownerKpiQuery.data,
    };
    if (!mergedData.kpi) return undefined;

    return deriveOwnerKpiFromDashboard(mergedData, t);
  }, [ownerKpiQuery.data, fallbackQuery.data, t]);

  return {
    data: ownerKpiData,
    isLoading: ownerKpiQuery.isLoading || (fallbackQuery.isLoading && !ownerKpiData),
    isError: ownerKpiQuery.isError && fallbackQuery.isError,
    refetch: () => {
      void ownerKpiQuery.refetch();
      void fallbackQuery.refetch();
    },
  };
}
