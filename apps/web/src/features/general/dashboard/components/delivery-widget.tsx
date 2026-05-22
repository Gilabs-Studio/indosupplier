"use client";

import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Truck, Clock, ArrowRight, CheckCircle2, TrendingUp, TrendingDown } from "lucide-react";
import type { DeliveryStatusData } from "../types";

interface DeliveryWidgetProps {
  readonly data?: DeliveryStatusData;
}

export function DeliveryWidget({ data }: DeliveryWidgetProps) {
  const t = useTranslations("dashboard");
  const total = data?.total ?? 0;
  const changePercent = data?.change_percent ?? 0;
  const isPositive = changePercent >= 0;

  const statItems = [
    {
      label: t("widgets.delivery_status.pending"),
      value: data?.pending ?? 0,
      icon: Clock,
      color: "text-chart-3",
      bg: "bg-chart-3/10",
    },
    {
      label: t("widgets.delivery_status.inTransit"),
      value: data?.in_transit ?? 0,
      icon: ArrowRight,
      color: "text-chart-1",
      bg: "bg-chart-1/10",
    },
    {
      label: t("widgets.delivery_status.delivered"),
      value: data?.delivered ?? 0,
      icon: CheckCircle2,
      color: "text-chart-2",
      bg: "bg-chart-2/10",
    },
  ];

  return (
    <Card className="h-full">
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base font-semibold">
            {t("widgets.delivery_status.title")}
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
        <div className="flex items-center gap-2">
          <Truck className="h-5 w-5 text-muted-foreground" />
          <span className="text-2xl font-bold">{total}</span>
          <span className="text-sm text-muted-foreground">
            {t("widgets.delivery_status.totalLabel")}
          </span>
        </div>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-3 gap-2">
          {statItems.map((item) => (
            <div
              key={item.label}
              className="flex flex-col items-center gap-1 rounded-lg border p-3"
            >
              <div className={`rounded-lg p-1.5 ${item.bg}`}>
                <item.icon className={`h-4 w-4 ${item.color}`} />
              </div>
              <p className="text-lg font-bold">{item.value}</p>
              <p className="text-center text-xs text-muted-foreground">
                {item.label}
              </p>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
