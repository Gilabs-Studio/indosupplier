"use client";

import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { formatCurrency } from "@/lib/utils";
import type { SalesPerformanceRow } from "../types";

interface SalesPerformanceWidgetProps {
  readonly data?: SalesPerformanceRow[];
}

export function SalesPerformanceWidget({ data }: SalesPerformanceWidgetProps) {
  const t = useTranslations("dashboard");
  const rows = data ?? [];

  return (
    <Card className="h-full">
      <CardHeader className="pb-2">
        <CardTitle className="text-base font-semibold">
          {t("widgets.sales_performance.title")}
        </CardTitle>
      </CardHeader>
      <CardContent>
        {rows.length === 0 ? (
          <div className="flex h-36 items-center justify-center">
            <p className="text-sm text-muted-foreground">{t("noData")}</p>
          </div>
        ) : (
          <div className="space-y-4">
            {rows.slice(0, 5).map((row, i) => (
              <div key={row.id || `sales-row-${i}`} className="space-y-1.5">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="flex h-6 w-6 items-center justify-center rounded-full bg-primary/10 text-xs font-bold text-primary">
                      {i + 1}
                    </span>
                    <span className="text-sm font-medium">{row.name}</span>
                  </div>
                  <span className="text-sm font-semibold">
                    {formatCurrency(row.revenue)}
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <Progress value={Math.min(row.target_percent, 100)} className="h-1.5" />
                  <span className="min-w-[3rem] text-right text-xs text-muted-foreground">
                    {row.target_percent}%
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
