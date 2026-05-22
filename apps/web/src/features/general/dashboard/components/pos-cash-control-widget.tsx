"use client";

import { startTransition, useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { ArrowRight, LayoutGrid } from "lucide-react";
import type { WidgetConfig, POSOutletItem } from "../types";
import { useUserPermission } from "@/hooks/use-user-permission";
import { WidgetAsyncState } from "./widget-async-state";
import { WIDGET_REGISTRY } from "../config/widget-registry";
import { Card, CardContent, CardDescription, CardHeader } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { usePosOutletOverview } from "../hooks/use-pos-outlet-overview";
import { useRouter } from "@/i18n/routing";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { usePermissionScope } from "@/features/master-data/user-management/hooks/use-has-permission";

interface PosCashControlWidgetProps {
  readonly widget: WidgetConfig;
}

export function PosCashControlWidget({ widget }: PosCashControlWidgetProps) {
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
  const totalActiveTables = filteredOutlets.reduce((sum, outlet) => sum + (outlet.active_table_count ?? 0), 0);

  const openLiveTable = (outlet: POSOutletItem) => {
    router.push(`/pos/fb/live-table?outlet_id=${encodeURIComponent(outlet.outlet_id)}`);
  };

  if (!canView) {
    return null;
  }

  return (
    <WidgetAsyncState isLoading={isLoading} isError={isError} onRetry={refetch}>
      <Card className="h-full">
        <CardHeader className="space-y-1 pb-2">
          <div className="flex items-center justify-between gap-2">
            <CardDescription className="flex items-center gap-2">
              <LayoutGrid className="h-4 w-4" />
              {t("widgets.pos_cash_control.title")}
            </CardDescription>
            {hasOutletFilter && (
              <Select value={selectedOutletId} onValueChange={setSelectedOutletId}>
                <SelectTrigger className="h-7 w-40 text-xs">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {canSelectAllOutlets && (
                    <SelectItem value="all">{t("widgets.pos_cash_control.allOutlets")}</SelectItem>
                  )}
                  {visibleOutlets.map((outlet) => (
                    <SelectItem key={outlet.outlet_id} value={outlet.outlet_id}>
                      {outlet.outlet_name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </div>
          <p className="text-xs text-muted-foreground">{t("widgets.pos_cash_control.description")}</p>
        </CardHeader>

        <CardContent className="flex h-full flex-col gap-3 pt-0">
          <div className="rounded-lg border bg-muted/20 px-3 py-2 text-xs text-muted-foreground">
            {t("widgets.pos_cash_control.totalActiveTables", { count: totalActiveTables })}
          </div>

          {isSingleOutletScope && focusedOutlet ? (
            <FocusedActiveTablesCard
              outlet={focusedOutlet}
              title={t("widgets.pos_cash_control.focusedTitle")}
              activeTablesText={t("widgets.pos_cash_control.activeTables", { count: focusedOutlet.active_table_count ?? 0 })}
              openLabel={t("widgets.pos_cash_control.openLiveTable")}
              onOpen={() => openLiveTable(focusedOutlet)}
              emptyLabel={t("widgets.pos_cash_control.noActiveTables")}
            />
          ) : (
            <div className="flex flex-1 flex-col gap-3 overflow-y-auto">
              <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
                {displayedOutlets.map((outlet) => (
                  <div
                    key={outlet.outlet_id}
                    className={cn(
                      "space-y-2 rounded-lg border p-3",
                      selectedOutletId === outlet.outlet_id ? "border-primary bg-primary/5" : "border-border"
                    )}
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div className="min-w-0">
                        <p className="truncate text-sm font-medium">{outlet.outlet_name}</p>
                        <p className="text-xs text-muted-foreground">
                          {t("widgets.pos_cash_control.activeTables", { count: outlet.active_table_count ?? 0 })}
                        </p>
                      </div>
                      <LayoutGrid className="h-4 w-4 shrink-0 text-muted-foreground" />
                    </div>

                    <div className="flex flex-wrap gap-1.5">
                      {(outlet.active_table_labels ?? []).slice(0, 6).map((tableName) => (
                        <span
                          key={`${outlet.outlet_id}-${tableName}`}
                          className="rounded-md border bg-card px-2 py-1 text-[11px] font-medium"
                        >
                          {tableName}
                        </span>
                      ))}
                      {(outlet.active_table_labels ?? []).length === 0 && (
                        <p className="text-xs text-muted-foreground">{t("widgets.pos_cash_control.noActiveTables")}</p>
                      )}
                    </div>

                    <div className="flex items-center gap-2">
                      <Button
                        type="button"
                        size="sm"
                        variant="outline"
                        className="h-7 cursor-pointer"
                        onClick={() => setSelectedOutletId(outlet.outlet_id)}
                      >
                        {t("widgets.pos_cash_control.focus")}
                      </Button>
                      <Button
                        type="button"
                        size="sm"
                        className="h-7 cursor-pointer"
                        onClick={() => openLiveTable(outlet)}
                      >
                        {t("widgets.pos_cash_control.openLiveTable")}
                      </Button>
                    </div>
                  </div>
                ))}

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
                    {t("widgets.pos_cash_control.loadMore")}
                  </Button>
                </div>
              )}
            </div>
          )}

          <div className="text-[11px] text-muted-foreground">
            {t("widgets.pos_cash_control.activeSessions", { count: pos?.active_sessions ?? 0 })}
          </div>
        </CardContent>
      </Card>
    </WidgetAsyncState>
  );
}

function FocusedActiveTablesCard({
  outlet,
  title,
  activeTablesText,
  openLabel,
  onOpen,
  emptyLabel,
}: {
  readonly outlet: POSOutletItem;
  readonly title: string;
  readonly activeTablesText: string;
  readonly openLabel: string;
  readonly onOpen: () => void;
  readonly emptyLabel: string;
}) {
  return (
    <div className="space-y-3 rounded-lg border bg-muted/10 p-3">
      <div className="space-y-1">
        <p className="text-xs uppercase tracking-wide text-muted-foreground">{title}</p>
        <p className="truncate text-sm font-semibold">{outlet.outlet_name}</p>
        <p className="text-xs text-muted-foreground">{activeTablesText}</p>
      </div>

      <div className="flex flex-wrap gap-1.5">
        {(outlet.active_table_labels ?? []).slice(0, 12).map((tableName) => (
          <span key={`${outlet.outlet_id}-${tableName}`} className="rounded-md border bg-card px-2 py-1 text-[11px] font-medium">
            {tableName}
          </span>
        ))}
        {(outlet.active_table_labels ?? []).length === 0 && (
          <p className="text-xs text-muted-foreground">{emptyLabel}</p>
        )}
      </div>

      <Button type="button" size="sm" className="w-full cursor-pointer" onClick={onOpen}>
        {openLabel}
        <ArrowRight className="ml-1.5 h-3.5 w-3.5" />
      </Button>
    </div>
  );
}
