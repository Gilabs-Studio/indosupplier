import apiClient from "@/lib/api-client";
import type {
  DashboardOverviewResponse,
  DashboardOverviewData,
  DashboardOverviewScope,
  DashboardDateFilter,
  WidgetConfig,
} from "../types";

function buildDateParams(filter?: DashboardDateFilter) {
  const params: Record<string, string | number> = {};
  if (filter?.from) params.start_date = filter.from;
  if (filter?.to) params.end_date = filter.to;
  if (filter?.year) params.year = filter.year;
  return params;
}

export const dashboardService = {
  async getOverview(
    filter?: DashboardDateFilter,
  ): Promise<DashboardOverviewData> {
    const params = buildDateParams(filter);
    const response = await apiClient.get<DashboardOverviewResponse>(
      "/general/dashboard/overview",
      { params },
    );
    return response.data.data;
  },

  async getOverviewScope(
    scope: DashboardOverviewScope,
    filter?: DashboardDateFilter,
    outletId?: string,
  ): Promise<Partial<DashboardOverviewData>> {
    const params: Record<string, string | number> = {
      ...buildDateParams(filter),
      scope,
    };
    if (outletId) params.outlet_id = outletId;
    const response = await apiClient.get<{ success: boolean; data: Partial<DashboardOverviewData> }>(
      "/general/dashboard/overview",
      { params },
    );
    return response.data.data;
  },

  /** Fetch the current user's saved layout. Returns null if no layout exists (first-time user). */
  async getLayout(dashboardType = "general"): Promise<WidgetConfig[] | null> {
    try {
      const response = await apiClient.get<{
        success: boolean;
        data: { dashboard_type: string; layout_json: string } | null;
      }>("/general/dashboard/layout", { params: { type: dashboardType } });

      if (!response.data.data?.layout_json) return null;
      const parsed = JSON.parse(response.data.data.layout_json) as {
        widgets: WidgetConfig[];
      };
      return Array.isArray(parsed?.widgets) ? parsed.widgets : null;
    } catch {
      // 404 = no saved layout yet, treat as null (use default)
      return null;
    }
  },

  /** Persist the user's current layout to the database. */
  async saveLayout(
    widgets: WidgetConfig[],
    dashboardType = "general",
  ): Promise<void> {
    const layoutJSON = JSON.stringify({
      widgets,
      updatedAt: new Date().toISOString(),
    });
    await apiClient.put("/general/dashboard/layout", {
      dashboard_type: dashboardType,
      layout_json: layoutJSON,
    });
  },
};
