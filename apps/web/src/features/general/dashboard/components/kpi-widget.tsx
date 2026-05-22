"use client";

import { useTranslations } from "next-intl";
import {
  DollarSign,
  ShoppingCart,
  Users,
  Package,
  UserCheck,
  TrendingUp,
  TrendingDown,
} from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import type { KpiCardData } from "../types";

const ICON_MAP: Record<string, React.ElementType> = {
  DollarSign,
  ShoppingCart,
  Users,
  Package,
  UserCheck,
};

const COLOR_MAP: Record<string, string> = {
  total_revenue: "text-chart-1",
  total_orders: "text-chart-2",
  total_customers: "text-chart-3",
  total_products: "text-chart-5",
  employee_count: "text-chart-4",
};

const BG_MAP: Record<string, string> = {
  total_revenue: "bg-chart-1/10",
  total_orders: "bg-chart-2/10",
  total_customers: "bg-chart-3/10",
  total_products: "bg-chart-5/10",
  employee_count: "bg-chart-4/10",
};

interface KpiWidgetProps {
  readonly widgetType: string;
  readonly data?: KpiCardData;
  readonly iconName: string;
}

export function KpiWidget({ widgetType, data, iconName }: KpiWidgetProps) {
  const t = useTranslations("dashboard");
  const Icon = ICON_MAP[iconName] ?? DollarSign;
  const colorClass = COLOR_MAP[widgetType] ?? "text-primary";
  const bgClass = BG_MAP[widgetType] ?? "bg-primary/10";
  const changePercent = data?.change_percent ?? 0;
  const isPositive = changePercent >= 0;

  return (
    <Card className="relative overflow-hidden h-full">
      <CardContent className="flex items-center gap-4 p-5">
        <div className={`rounded-xl p-3 ${bgClass}`}>
          <Icon className={`h-6 w-6 ${colorClass}`} />
        </div>
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm text-muted-foreground">
            {t(`widgets.${widgetType}.title` as Parameters<typeof t>[0])}
          </p>
          <p className="text-2xl font-bold tracking-tight">
            {data?.formatted ?? "0"}
          </p>
        </div>
        {changePercent !== 0 && (
          <div
            className={`flex items-center gap-0.5 rounded-full px-2 py-0.5 text-xs font-medium ${
              isPositive
                ? "bg-chart-2/10 text-chart-2"
                : "bg-destructive/10 text-destructive"
            }`}
          >
            {isPositive ? (
              <TrendingUp className="h-3 w-3" />
            ) : (
              <TrendingDown className="h-3 w-3" />
            )}
            {isPositive ? "+" : ""}
            {changePercent.toFixed(1)}%
          </div>
        )}
      </CardContent>
    </Card>
  );
}
