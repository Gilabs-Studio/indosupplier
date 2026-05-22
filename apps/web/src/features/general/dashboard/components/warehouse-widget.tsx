"use client";

import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Warehouse } from "lucide-react";
import { useMemo } from "react";
import {
  PieChart,
  Pie,
  Cell,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from "recharts";
import { formatCurrency } from "@/lib/utils";
import type { WarehouseOverviewData } from "../types";

interface WarehouseWidgetProps {
  readonly data?: WarehouseOverviewData;
}

const PIE_COLORS = {
  inStock: "#22c55e",
  lowStock: "#f59e0b",
  outOfStock: "#ef4444",
};

export function WarehouseWidget({ data }: WarehouseWidgetProps) {
  const t = useTranslations("dashboard");
  const warehouses = useMemo(() => data?.warehouses ?? [], [data?.warehouses]);

  // Aggregate stock status totals across all warehouses for the pie chart
  const pieData = useMemo(() => {
    const totals = warehouses.reduce(
      (acc, wh) => ({
        inStock: acc.inStock + (wh.in_stock_count ?? 0),
        lowStock: acc.lowStock + (wh.low_stock_count ?? 0),
        outOfStock: acc.outOfStock + (wh.out_of_stock_count ?? 0),
      }),
      { inStock: 0, lowStock: 0, outOfStock: 0 },
    );

    return [
      { name: t("widgets.warehouse_overview.inStock"), value: totals.inStock, color: PIE_COLORS.inStock },
      { name: t("widgets.warehouse_overview.lowStock"), value: totals.lowStock, color: PIE_COLORS.lowStock },
      { name: t("widgets.warehouse_overview.outOfStock"), value: totals.outOfStock, color: PIE_COLORS.outOfStock },
    ].filter((d) => d.value > 0);
  }, [warehouses, t]);

  const hasPieData = pieData.length > 0;

  return (
    <Card className="h-full">
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base font-semibold">
            {t("widgets.warehouse_overview.title")}
          </CardTitle>
          {data?.total_stock_value !== undefined && (
            <span className="text-sm font-semibold text-primary">
              {formatCurrency(data.total_stock_value)}
            </span>
          )}
        </div>
      </CardHeader>
      <CardContent>
        {warehouses.length === 0 ? (
          <div className="flex h-36 items-center justify-center">
            <p className="text-sm text-muted-foreground">{t("noData")}</p>
          </div>
        ) : (
          <div className="space-y-4">
            {/* Aggregate stock status pie chart */}
            {hasPieData && (
              <div>
                <p className="mb-1 text-xs font-medium text-muted-foreground">
                  {t("widgets.warehouse_overview.stockStatus")}
                </p>
                <ResponsiveContainer width="100%" height={180}>
                  <PieChart>
                    <Pie
                      data={pieData}
                      cx="50%"
                      cy="50%"
                      innerRadius={50}
                      outerRadius={75}
                      paddingAngle={3}
                      dataKey="value"
                    >
                      {pieData.map((entry) => (
                        <Cell key={entry.name} fill={entry.color} />
                      ))}
                    </Pie>
                    <Tooltip
                      formatter={(value: number) => [value, t("widgets.warehouse_overview.items")]}
                    />
                    <Legend
                      iconType="circle"
                      iconSize={8}
                      formatter={(value: string) => (
                        <span className="text-xs">{value}</span>
                      )}
                    />
                  </PieChart>
                </ResponsiveContainer>
              </div>
            )}

            {/* Per-warehouse breakdown */}
            <div className="space-y-2">
              {warehouses.slice(0, 4).map((wh) => {
                const inStock = wh.in_stock_count ?? 0;
                const lowStock = wh.low_stock_count ?? 0;
                const outOfStock = wh.out_of_stock_count ?? 0;
                const total = inStock + lowStock + outOfStock;

                return (
                  <div key={wh.id} className="rounded-lg border p-3">
                    <div className="mb-2 flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Warehouse className="h-4 w-4 text-muted-foreground" />
                        <span className="text-sm font-medium">{wh.name}</span>
                      </div>
                      <span className="text-xs text-muted-foreground">
                        {wh.item_count} {t("widgets.warehouse_overview.items")}
                      </span>
                    </div>

                    {total > 0 ? (
                      <div className="flex flex-wrap gap-2">
                        {inStock > 0 && (
                          <Badge variant="success">
                            {inStock} {t("widgets.warehouse_overview.inStock")}
                          </Badge>
                        )}
                        {lowStock > 0 && (
                          <Badge variant="warning">
                            {lowStock} {t("widgets.warehouse_overview.lowStock")}
                          </Badge>
                        )}
                        {outOfStock > 0 && (
                          <Badge variant="destructive">
                            {outOfStock} {t("widgets.warehouse_overview.outOfStock")}
                          </Badge>
                        )}
                      </div>
                    ) : (
                      <p className="text-xs text-muted-foreground">
                        {formatCurrency(wh.stock_value)}
                      </p>
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
