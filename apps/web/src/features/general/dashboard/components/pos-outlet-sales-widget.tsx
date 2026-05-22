"use client";

import { startTransition, useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { Store, Filter, DollarSign, ShoppingCart } from "lucide-react";
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
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { usePosOutletOverview } from "../hooks/use-pos-outlet-overview";
import { Button } from "@/components/ui/button";
import { usePermissionScope } from "@/features/master-data/user-management/hooks/use-has-permission";

interface PosOutletSalesWidgetProps {
  readonly widget: WidgetConfig;
}

export function PosOutletSalesWidget({ widget }: PosOutletSalesWidgetProps) {
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

  useEffect(() => {
    if (process.env.NODE_ENV === "production") {
      return;
    }

    console.debug("[Dashboard POS][Outlet Sales]", {
      widgetType: widget.type,
      canView,
      selectedOutletId,
      hasOutletFilter,
      visibleOutletCount: visibleOutlets.length,
      filteredOutletCount: filteredOutlets.length,
      isLoading,
      isError,
      summary: {
        today_total_orders: pos?.today_total_orders ?? 0,
        today_total_revenue: pos?.today_total_revenue ?? 0,
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
    pos?.served_orders,
    pos?.today_total_orders,
    pos?.today_total_revenue,
    pos?.waiting_orders,
    pos?.warning_orders,
    selectedOutletId,
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
              <Store className="h-4 w-4" />
              {t("widgets.pos_outlet_sales.title")}
            </CardDescription>
            {hasOutletFilter && (
              <Select value={selectedOutletId} onValueChange={setSelectedOutletId}>
                <SelectTrigger className="h-7 w-40 text-xs">
                  <Filter className="mr-1 h-3 w-3" />
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {canSelectAllOutlets && (
                    <SelectItem value="all">{t("widgets.pos_outlet_sales.allOutlets")}</SelectItem>
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
          <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
            <div className="rounded-lg border p-3">
              <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                <DollarSign className="h-3 w-3" />
                {t("widgets.pos_outlet_sales.todayRevenue")}
              </div>
              <CardTitle className="mt-1 text-lg">{pos?.today_total_revenue_formatted ?? "Rp 0"}</CardTitle>
            </div>
            <div className="rounded-lg border p-3">
              <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                <ShoppingCart className="h-3 w-3" />
                {t("widgets.pos_outlet_sales.todayOrders")}
              </div>
              <CardTitle className="mt-1 text-lg">{pos?.today_total_orders ?? 0}</CardTitle>
            </div>
            <div className="rounded-lg border p-3">
              <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                <Store className="h-3 w-3" />
                {t("widgets.pos_outlet_sales.servedOrders")}
              </div>
              <CardTitle className="mt-1 text-lg">{pos?.served_orders ?? 0}</CardTitle>
            </div>
            <div className="rounded-lg border p-3">
              <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                <Filter className="h-3 w-3" />
                {t("widgets.pos_outlet_sales.waitingOrders")}
              </div>
              <CardTitle className="mt-1 text-lg">{pos?.waiting_orders ?? 0}</CardTitle>
            </div>
          </div>

          <div className="flex items-center justify-between rounded-lg border bg-muted/20 px-3 py-2 text-xs text-muted-foreground">
            <span>{t("widgets.pos_outlet_sales.warningOrders")}</span>
            <span className="font-medium text-warning">{pos?.warning_orders ?? 0}</span>
          </div>

          {isSingleOutletScope && focusedOutlet ? (
            <FocusedOutletSalesCard
              outlet={focusedOutlet}
              revenueLabel={t("widgets.pos_outlet_sales.revenue")}
              ordersLabel={t("widgets.pos_outlet_sales.orders")}
              servedLabel={t("widgets.pos_outlet_sales.servedOrders")}
              tablesLabel={t("widgets.pos_outlet_sales.tables")}
            />
          ) : (
            <div className="flex-1 overflow-y-auto">
              <table className="w-full text-xs">
                <thead>
                  <tr className="border-b text-muted-foreground">
                    <th className="pb-1.5 text-left font-medium">{t("widgets.pos_outlet_sales.outlet")}</th>
                    <th className="pb-1.5 text-right font-medium">{t("widgets.pos_outlet_sales.revenue")}</th>
                    <th className="pb-1.5 text-right font-medium">{t("widgets.pos_outlet_sales.orders")}</th>
                    <th className="pb-1.5 text-right font-medium">{t("widgets.pos_outlet_sales.tables")}</th>
                  </tr>
                </thead>
                <tbody>
                  {displayedOutlets.map((outlet) => (
                    <tr key={outlet.outlet_id} className="border-b last:border-0">
                      <td className="py-2 font-medium">{outlet.outlet_name}</td>
                      <td className="py-2 text-right">{outlet.today_revenue_formatted}</td>
                      <td className="py-2 text-right">{outlet.order_count}</td>
                      <td className="py-2 text-right">
                        <span
                          className={cn(
                            outlet.tables_occupied > 0 ? "text-primary" : "text-muted-foreground",
                          )}
                        >
                          {outlet.tables_occupied}/{outlet.tables_total}
                        </span>
                      </td>
                    </tr>
                  ))}
                  {filteredOutlets.length === 0 && (
                    <tr>
                      <td colSpan={4} className="py-4 text-center text-muted-foreground">
                        {t("noData")}
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>

              {hasMore && (
                <div className="mt-3 flex justify-center">
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    className="cursor-pointer"
                    onClick={() => setVisibleCount((current) => current + pageSize)}
                  >
                    {t("widgets.pos_outlet_sales.loadMore")}
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

function FocusedOutletSalesCard({
  outlet,
  revenueLabel,
  ordersLabel,
  servedLabel,
  tablesLabel,
}: {
  readonly outlet: POSOutletItem;
  readonly revenueLabel: string;
  readonly ordersLabel: string;
  readonly servedLabel: string;
  readonly tablesLabel: string;
}) {
  return (
    <div className="rounded-lg border bg-muted/10 p-3">
      <p className="truncate text-sm font-semibold">{outlet.outlet_name}</p>
      <div className="mt-2 grid grid-cols-2 gap-2 text-xs text-muted-foreground">
        <div>
          <p>{revenueLabel}</p>
          <p className="text-sm font-semibold text-foreground">{outlet.today_revenue_formatted}</p>
        </div>
        <div>
          <p>{ordersLabel}</p>
          <p className="text-sm font-semibold text-foreground">{outlet.order_count}</p>
        </div>
        <div>
          <p>{servedLabel}</p>
          <p className="text-sm font-semibold text-foreground">{outlet.served_orders}</p>
        </div>
        <div>
          <p>{tablesLabel}</p>
          <p className="text-sm font-semibold text-foreground">{outlet.tables_occupied}/{outlet.tables_total}</p>
        </div>
      </div>
    </div>
  );
}
