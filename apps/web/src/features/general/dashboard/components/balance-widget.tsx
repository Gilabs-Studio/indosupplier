"use client";

import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { TrendingUp, TrendingDown } from "lucide-react";
import type { KpiCardData } from "../types";

interface BalanceWidgetProps {
  readonly data?: KpiCardData & {
    chart_data?: Array<{ period: string; value: number; formatted: string }>;
  };
}

export function BalanceWidget({ data }: BalanceWidgetProps) {
  const t = useTranslations("dashboard");
  const changePercent = data?.change_percent ?? 0;
  const isPositive = changePercent >= 0;
  const chartData = data?.chart_data ?? [];

  return (
    <Card className="h-full">
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base font-semibold">
            {t("widgets.balance_overview.title")}
          </CardTitle>
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
        </div>
        <p className="text-2xl font-bold tracking-tight">
          {data?.formatted ?? "Rp 0"}
        </p>
      </CardHeader>
      <CardContent className="pb-4">
        {chartData.length > 0 ? (
          <ResponsiveContainer width="100%" height={140}>
            <AreaChart data={chartData} margin={{ top: 0, right: 0, left: -30, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" className="stroke-border" vertical={false} />
              <XAxis dataKey="period" tick={{ fontSize: 10 }} className="text-xs" />
              <YAxis tick={false} axisLine={false} />
              <Tooltip
                contentStyle={{
                  borderRadius: 8,
                  border: "1px solid hsl(var(--border))",
                  background: "hsl(var(--popover))",
                  color: "hsl(var(--popover-foreground))",
                  fontSize: 12,
                }}
                formatter={(_: unknown, __: unknown, props: { payload?: { formatted?: string } }) =>
                  props.payload?.formatted ?? ""
                }
                labelFormatter={(label: string) => label}
              />
              <Area
                type="monotone"
                dataKey="value"
                stroke="hsl(var(--chart-1))"
                fill="hsl(var(--chart-1))"
                fillOpacity={0.15}
                strokeWidth={2}
              />
            </AreaChart>
          </ResponsiveContainer>
        ) : (
          <div className="flex h-36 items-center justify-center">
            <p className="text-sm text-muted-foreground">{t("noData")}</p>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
