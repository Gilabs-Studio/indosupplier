"use client";

import { startTransition, useEffect, useMemo, useState } from "react";
import { useRouter } from "@/i18n/routing";
import { useTranslations } from "next-intl";
import { LayoutGrid, Filter, Users, ArrowRight } from "lucide-react";
import type { WidgetConfig, POSOutletItem } from "../types";
import { useUserPermission } from "@/hooks/use-user-permission";
import { WidgetAsyncState } from "./widget-async-state";
import { WIDGET_REGISTRY } from "../config/widget-registry";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";
import { Progress } from "@/components/ui/progress";
import { Card, CardContent, CardDescription, CardHeader } from "@/components/ui/card";
import { usePosOutletOverview } from "../hooks/use-pos-outlet-overview";
import { Button } from "@/components/ui/button";
import { useFloorLayout } from "@/features/pos/fb/floor-layout/hooks/use-floor-layouts";
import { usePermissionScope } from "@/features/master-data/user-management/hooks/use-has-permission";

interface PosLiveTableWidgetProps {
  readonly widget: WidgetConfig;
}

export function PosLiveTableWidget({ widget }: PosLiveTableWidgetProps) {
  const router = useRouter();
  const t = useTranslations("dashboard");
  const pageSize = 6;
  const registry = WIDGET_REGISTRY[widget.type];
  const canView = useUserPermission(registry?.permission ?? "");
  const permissionScope = usePermissionScope(registry?.permission ?? "");
  const allowAllOutlets = permissionScope?.trim().toUpperCase() === "ALL";
  const [visibleCount, setVisibleCount] = useState(pageSize);
  const {
    posSummary: pos,
    visibleOutlets,
    filteredOutlets,
    selectedOutletId,
    setSelectedOutletId,
    hasOutletFilter,
    allowAllOutlets: canSelectAllOutlets,
    isSingleOutletScope,
    focusedOutlet,
    isLoading,
    isError,
    refetch,
  } = usePosOutletOverview(widget.type, canView, { allowAllOutlets });

  useEffect(() => {
    startTransition(() => {
      setVisibleCount(pageSize);
    });
  }, [selectedOutletId]);

  const displayedOutlets = useMemo(() => filteredOutlets.slice(0, visibleCount), [filteredOutlets, visibleCount]);
  const hasMore = filteredOutlets.length > displayedOutlets.length;

  const openLiveTable = (outlet: POSOutletItem) => {
    router.push(`/pos/fb/live-table?outlet_id=${encodeURIComponent(outlet.outlet_id)}`);
  };

  const totalTables = filteredOutlets.reduce((sum, o) => sum + o.tables_total, 0);
  const occupiedTables = filteredOutlets.reduce((sum, o) => sum + o.tables_occupied, 0);
  const occupancyRate = totalTables > 0 ? Math.round((occupiedTables / totalTables) * 100) : 0;
  const focusedFloorPlanId = isSingleOutletScope ? focusedOutlet?.floor_plan_id ?? "" : "";
  const focusedPlanQuery = useFloorLayout(focusedFloorPlanId, {
    enabled: Boolean(focusedFloorPlanId && isSingleOutletScope),
  });

  const focusedTableNames = useMemo(() => {
    const layoutData = focusedPlanQuery.data?.data?.layout_data;
    if (!layoutData) {
      return [];
    }

    try {
      const parsed = typeof layoutData === "string" ? JSON.parse(layoutData) : layoutData;
      if (!Array.isArray(parsed)) {
        return [];
      }

      const names = parsed
        .filter((obj): obj is { type?: string; label?: string; tableNumber?: number; id?: string } => Boolean(obj && typeof obj === "object"))
        .filter((obj) => obj.type === "table")
        .map((obj) => obj.label || (typeof obj.tableNumber === "number" ? `T${obj.tableNumber}` : obj.id || ""))
        .filter((name): name is string => Boolean(name));

      return Array.from(new Set(names));
    } catch {
      return [];
    }
  }, [focusedPlanQuery.data?.data?.layout_data]);

  useEffect(() => {
    if (process.env.NODE_ENV === "production") {
      return;
    }

    console.debug("[Dashboard POS][Live Table Overview]", {
      widgetType: widget.type,
      canView,
      selectedOutletId,
      hasOutletFilter,
      visibleOutletCount: visibleOutlets.length,
      filteredOutletCount: filteredOutlets.length,
      isLoading,
      isError,
      occupancy: {
        occupiedTables,
        totalTables,
        occupancyRate,
      },
      summary: {
        today_total_orders: pos?.today_total_orders ?? 0,
        served_orders: pos?.served_orders ?? 0,
        waiting_orders: pos?.waiting_orders ?? 0,
        warning_orders: pos?.warning_orders ?? 0,
      },
    });
  }, [
    canView,
    filteredOutlets.length,
    hasOutletFilter,
    isError,
    isLoading,
    occupiedTables,
    occupancyRate,
    pos?.served_orders,
    pos?.today_total_orders,
    pos?.waiting_orders,
    pos?.warning_orders,
    selectedOutletId,
    totalTables,
    visibleOutlets.length,
    widget.type,
  ]);

  if (!canView) return null;

  return (
    <WidgetAsyncState isLoading={isLoading} isError={isError} onRetry={refetch}>
      <Card className="h-full">
        <CardHeader className="space-y-1 pb-2">
          <div className="flex items-center justify-between gap-2">
            <CardDescription className="flex items-center gap-2">
              <LayoutGrid className="h-4 w-4" />
              {t("widgets.pos_live_table_overview.title")}
            </CardDescription>
            {hasOutletFilter && (
              <Select value={selectedOutletId} onValueChange={setSelectedOutletId}>
                <SelectTrigger className="h-7 w-40 text-xs">
                  <Filter className="mr-1 h-3 w-3" />
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {canSelectAllOutlets && (
                    <SelectItem value="all">{t("widgets.pos_live_table_overview.allOutlets")}</SelectItem>
                  )}
                  {visibleOutlets.map((o) => (
                    <SelectItem key={o.outlet_id} value={o.outlet_id}>
                      {o.outlet_name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </div>
        </CardHeader>

        <CardContent className="flex h-full flex-col gap-3 pt-0">
          <div className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">{t("widgets.pos_live_table_overview.occupancy")}</span>
              <span className="font-medium">{occupiedTables}/{totalTables} ({occupancyRate}%)</span>
            </div>
            <Progress value={occupancyRate} className="h-2" />
          </div>

          <div className="grid grid-cols-3 gap-3 text-xs">
            <div className="rounded-lg border px-3 py-2">
              <div className="text-muted-foreground">{t("widgets.pos_live_table_overview.served")}</div>
              <div className="mt-1 text-sm font-semibold">{pos?.served_orders ?? 0}</div>
            </div>
            <div className="rounded-lg border px-3 py-2">
              <div className="text-muted-foreground">{t("widgets.pos_live_table_overview.waiting")}</div>
              <div className="mt-1 text-sm font-semibold">{pos?.waiting_orders ?? 0}</div>
            </div>
            <div className="rounded-lg border px-3 py-2">
              <div className="text-muted-foreground">{t("widgets.pos_live_table_overview.warnings")}</div>
              <div className="mt-1 text-sm font-semibold text-warning">{pos?.warning_orders ?? 0}</div>
            </div>
          </div>

          {isSingleOutletScope && focusedOutlet ? (
            <FocusedLiveTableCard
              outlet={focusedOutlet}
              onOpen={() => openLiveTable(focusedOutlet)}
              title={t("widgets.pos_live_table_overview.focusedTitle")}
              cta={t("widgets.pos_live_table_overview.openLiveTable")}
              waitingLabel={t("widgets.pos_live_table_overview.waiting")}
              warningLabel={t("widgets.pos_live_table_overview.warnings")}
              tableNames={focusedTableNames}
              tablesLabel={t("widgets.pos_live_table_overview.tables")}
              loadingTables={focusedPlanQuery.isLoading}
              loadingTablesLabel={t("widgets.pos_live_table_overview.loadingTables")}
              noTableLabels={t("widgets.pos_live_table_overview.noTableLabels")}
            />
          ) : (
            <div className="flex flex-1 flex-col gap-3 overflow-y-auto">
              <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {displayedOutlets.map((outlet) => {
                  const rate = outlet.tables_total > 0
                    ? Math.round((outlet.tables_occupied / outlet.tables_total) * 100)
                    : 0;
                  return (
                    <button
                      type="button"
                      key={outlet.outlet_id}
                      className="space-y-2 rounded-lg border p-3 text-left transition-colors hover:bg-muted/20"
                      onClick={() => openLiveTable(outlet)}
                    >
                      <p className="truncate text-sm font-medium">{outlet.outlet_name}</p>
                      <div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
                        <span className="flex items-center gap-1">
                          <Users className="h-3 w-3" />
                          {outlet.tables_occupied}/{outlet.tables_total}
                        </span>
                        <span className="flex items-center gap-1">
                          <LayoutGrid className="h-3 w-3" />
                          {outlet.served_orders} {t("widgets.pos_live_table_overview.served")}
                        </span>
                        <span className={cn(
                          "font-medium",
                          rate >= 80 ? "text-warning" : rate >= 50 ? "text-primary" : "text-success",
                        )}>
                          {rate}%
                        </span>
                      </div>
                      <div className="grid grid-cols-3 gap-2 text-[11px] text-muted-foreground">
                        <span>{t("widgets.pos_live_table_overview.waitingShort", { count: outlet.waiting_orders })}</span>
                        <span>{t("widgets.pos_live_table_overview.warningShort", { count: outlet.warning_orders })}</span>
                        <span className="text-right">{outlet.order_count} {t("widgets.pos_live_table_overview.ordersShort")}</span>
                      </div>
                      <Progress value={rate} className="h-1.5" />
                    </button>
                  );
                })}
                {filteredOutlets.length === 0 && (
                  <div className="col-span-full flex items-center justify-center py-6 text-sm text-muted-foreground">
                    {t("noData")}
                  </div>
                )}
              </div>

              {hasMore && (
                <div className="flex justify-center">
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    className="cursor-pointer"
                    onClick={() => setVisibleCount((current) => current + pageSize)}
                  >
                    {t("widgets.pos_live_table_overview.loadMore")}
                  </Button>
                </div>
              )}
            </div>
          )}
        </CardContent>
      </Card>
    </WidgetAsyncState>
  );
}

function FocusedLiveTableCard({
  outlet,
  onOpen,
  title,
  cta,
  waitingLabel,
  warningLabel,
  tableNames,
  tablesLabel,
  loadingTables,
  loadingTablesLabel,
  noTableLabels,
}: {
  readonly outlet: POSOutletItem;
  readonly onOpen: () => void;
  readonly title: string;
  readonly cta: string;
  readonly waitingLabel: string;
  readonly warningLabel: string;
  readonly tableNames: string[];
  readonly tablesLabel: string;
  readonly loadingTables: boolean;
  readonly loadingTablesLabel: string;
  readonly noTableLabels: string;
}) {
  const rate = outlet.tables_total > 0
    ? Math.round((outlet.tables_occupied / outlet.tables_total) * 100)
    : 0;

  return (
    <div className="space-y-3 rounded-lg border bg-muted/10 p-3">
      <div className="space-y-1">
        <p className="text-xs uppercase tracking-wide text-muted-foreground">{title}</p>
        <p className="truncate text-sm font-semibold">{outlet.outlet_name}</p>
      </div>

      <div className="grid grid-cols-3 gap-2 text-xs">
        <div className="rounded-md border px-2 py-1.5">
          <p className="text-muted-foreground">Tables</p>
          <p className="font-semibold">{outlet.tables_occupied}/{outlet.tables_total}</p>
        </div>
        <div className="rounded-md border px-2 py-1.5">
          <p className="text-muted-foreground">{waitingLabel}</p>
          <p className="font-semibold">{outlet.waiting_orders}</p>
        </div>
        <div className="rounded-md border px-2 py-1.5">
          <p className="text-muted-foreground">{warningLabel}</p>
          <p className="font-semibold text-warning">{outlet.warning_orders}</p>
        </div>
      </div>

      <Progress value={rate} className="h-2" />

      <div className="space-y-1">
        <p className="text-xs uppercase tracking-wide text-muted-foreground">{tablesLabel}</p>
        {loadingTables ? (
          <p className="text-xs text-muted-foreground">{loadingTablesLabel}</p>
        ) : tableNames.length > 0 ? (
          <div className="flex flex-wrap gap-1.5">
            {tableNames.slice(0, 12).map((name) => (
              <span key={name} className="rounded-md border bg-card px-2 py-1 text-[11px] font-medium">
                {name}
              </span>
            ))}
          </div>
        ) : (
          <p className="text-xs text-muted-foreground">{noTableLabels}</p>
        )}
      </div>

      <Button type="button" size="sm" className="w-full cursor-pointer" onClick={onOpen}>
        {cta}
        <ArrowRight className="ml-1.5 h-3.5 w-3.5" />
      </Button>
    </div>
  );
}
