"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import {
  Bar,
  BarChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { formatCurrency } from "@/lib/utils";
import type { PeriodChartData } from "../types";

interface RevenueBarChartCardProps {
  readonly revenueData?: PeriodChartData;
  readonly costsData?: PeriodChartData;
  readonly isLoading?: boolean;
}

export function RevenueBarChartCard({
  revenueData,
  costsData,
  isLoading,
}: RevenueBarChartCardProps) {
  const t = useTranslations("dashboard");
  const [activeKey, setActiveKey] = useState<"revenue" | "costs">("revenue");

  const activeSeries = activeKey === "revenue" ? revenueData : costsData;
  const periods = activeSeries?.period ?? [];
  const seriesLine = activeSeries?.series?.[0];

  const chartData = periods.map((p, i) => ({
    period: p,
    value: seriesLine?.data?.[i] ?? 0,
  }));

  const revenueTotal =
    revenueData?.series?.[0]?.data?.reduce((a, b) => a + b, 0) ?? 0;
  const costsTotal =
    costsData?.series?.[0]?.data?.reduce((a, b) => a + b, 0) ?? 0;

  if (isLoading) {
    return (
      <Card className="h-full">
        <CardContent className="flex h-72 items-center justify-center px-6">
          <Skeleton className="h-48 w-full" />
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="relative flex h-full flex-col overflow-hidden">
      <CardHeader>
        <CardTitle>{t("revenueChart.title")}</CardTitle>
        <CardDescription>
          {t("revenueChart.subtitle")}
        </CardDescription>
        {/* Toggle buttons — anchored top-right on md+ screens */}
        <div className="end-0 top-0 flex divide-x rounded-md border md:absolute md:rounded-none md:rounded-bl-md md:border-e-transparent md:border-t-transparent">
          <button
            data-active={activeKey === "revenue"}
            className="data-[active=true]:bg-muted relative flex flex-1 cursor-pointer flex-col justify-center gap-1 px-6 py-4 text-left"
            onClick={() => setActiveKey("revenue")}
          >
            <span className="text-xs text-muted-foreground">
              {t("revenueChart.revenue")}
            </span>
            <span className="text-lg font-bold sm:text-2xl">
              {formatCurrency(revenueTotal)}
            </span>
          </button>
          <button
            data-active={activeKey === "costs"}
            className="data-[active=true]:bg-muted relative flex flex-1 cursor-pointer flex-col justify-center gap-1 px-6 py-4 text-left"
            onClick={() => setActiveKey("costs")}
          >
            <span className="text-xs text-muted-foreground">
              {t("revenueChart.costs")}
            </span>
            <span className="text-lg font-bold sm:text-2xl">
              {formatCurrency(costsTotal)}
            </span>
          </button>
        </div>
      </CardHeader>
      <CardContent className="flex flex-1 items-center px-6 pb-6">
        <div className="w-full">
          <ResponsiveContainer width="100%" height={186}>
            <BarChart
              data={chartData}
              margin={{ top: 5, right: 10, left: -10, bottom: 0 }}
            >
              <CartesianGrid
                vertical={false}
                strokeDasharray=""
                className="stroke-border/50"
              />
              <XAxis
                dataKey="period"
                tick={{ fontSize: 11 }}
                tickLine={false}
                axisLine={false}
              />
              <YAxis hide />
              <Tooltip
                cursor={{
                  fill: "hsl(var(--border))",
                  fillOpacity: 0.3,
                }}
                contentStyle={{
                  borderRadius: 8,
                  border: "1px solid hsl(var(--border))",
                  background: "hsl(var(--popover))",
                  color: "hsl(var(--popover-foreground))",
                  fontSize: 12,
                }}
                formatter={(v: number) => [formatCurrency(v)]}
                labelFormatter={(label) => String(label)}
              />
              <Bar
                dataKey="value"
                fill={
                  activeKey === "revenue"
                    ? "hsl(var(--chart-2))"
                    : "hsl(var(--chart-1))"
                }
                radius={[5, 5, 0, 0]}
                maxBarSize={20}
              />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  );
}
