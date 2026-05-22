"use client";

import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatCurrency } from "@/lib/utils";
import type { TopProductRow } from "../types";

interface TopProductsWidgetProps {
  readonly data?: TopProductRow[];
}

export function TopProductsWidget({ data }: TopProductsWidgetProps) {
  const t = useTranslations("dashboard");
  const rows = data ?? [];

  return (
    <Card className="h-full">
      <CardHeader className="pb-2">
        <CardTitle className="text-base font-semibold">
          {t("widgets.top_products.title")}
        </CardTitle>
      </CardHeader>
      <CardContent>
        {rows.length === 0 ? (
          <div className="flex h-36 items-center justify-center">
            <p className="text-sm text-muted-foreground">{t("noData")}</p>
          </div>
        ) : (
          <div className="space-y-2">
            {rows.slice(0, 6).map((row, i) => (
              <div
                key={row.id || `tp-${i}`}
                className="flex items-center justify-between rounded-lg border p-3"
              >
                <div className="flex items-center gap-3">
                  <span className="flex h-7 w-7 items-center justify-center rounded-full bg-chart-5/10 text-xs font-bold text-chart-5">
                    {i + 1}
                  </span>
                  <div>
                    <p className="text-sm font-medium">{row.name}</p>
                    <p className="text-xs text-muted-foreground">{row.sku}</p>
                  </div>
                </div>
                <div className="text-right">
                  <p className="text-sm font-semibold">{formatCurrency(row.revenue)}</p>
                  <p className="text-xs text-muted-foreground">
                    {row.quantity_sold} {t("widgets.top_products.sold")}
                  </p>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
