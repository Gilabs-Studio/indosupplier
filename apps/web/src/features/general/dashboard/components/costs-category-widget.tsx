"use client";

import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatCurrency } from "@/lib/utils";
import type { CostCategoryItem } from "../types";

const CATEGORY_COLORS = [
  "bg-chart-1",
  "bg-chart-2",
  "bg-chart-3",
  "bg-chart-4",
  "bg-chart-5",
];

interface CostsCategoryWidgetProps {
  readonly data?: CostCategoryItem[];
}

export function CostsCategoryWidget({ data }: CostsCategoryWidgetProps) {
  const t = useTranslations("dashboard");
  const items = data ?? [];

  return (
    <Card className="h-full">
      <CardHeader className="pb-2">
        <CardTitle className="text-base font-semibold">
          {t("widgets.costs_by_category.title")}
        </CardTitle>
      </CardHeader>
      <CardContent>
        {items.length === 0 ? (
          <div className="flex h-36 items-center justify-center">
            <p className="text-sm text-muted-foreground">{t("noData")}</p>
          </div>
        ) : (
          <div className="space-y-3">
            {items.map((item, i) => (
              <div key={item.category} className="space-y-1">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground">{item.category}</span>
                  <span className="font-medium">{formatCurrency(item.amount)}</span>
                </div>
                <div className="h-2 overflow-hidden rounded-full bg-secondary">
                  <div
                    className={`h-full rounded-full transition-all ${CATEGORY_COLORS[i % CATEGORY_COLORS.length]}`}
                    style={{ width: `${Math.min(item.percentage, 100)}%` }}
                  />
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
