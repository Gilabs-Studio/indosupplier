"use client";

import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { FileText, AlertCircle, CheckCircle2, Clock } from "lucide-react";
import type { InvoiceSummaryData } from "../types";

interface InvoicesSummaryWidgetProps {
  readonly data?: InvoiceSummaryData;
}

export function InvoicesSummaryWidget({ data }: InvoicesSummaryWidgetProps) {
  const t = useTranslations("dashboard");

  const items = [
    {
      label: t("widgets.invoices_summary.total"),
      value: data?.total ?? 0,
      icon: FileText,
      color: "text-chart-1",
      bg: "bg-chart-1/10",
    },
    {
      label: t("widgets.invoices_summary.paid"),
      value: data?.paid ?? 0,
      icon: CheckCircle2,
      color: "text-chart-2",
      bg: "bg-chart-2/10",
    },
    {
      label: t("widgets.invoices_summary.unpaid"),
      value: data?.unpaid ?? 0,
      icon: AlertCircle,
      color: "text-chart-3",
      bg: "bg-chart-3/10",
    },
    {
      label: t("widgets.invoices_summary.overdue"),
      value: data?.overdue ?? 0,
      icon: Clock,
      color: "text-chart-4",
      bg: "bg-chart-4/10",
    },
  ];

  return (
    <Card className="h-full">
      <CardHeader className="pb-2">
        <CardTitle className="text-base font-semibold">
          {t("widgets.invoices_summary.title")}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-2 gap-3">
          {items.map((item) => (
            <div
              key={item.label}
              className="flex items-center gap-3 rounded-lg border p-3"
            >
              <div className={`rounded-lg p-2 ${item.bg}`}>
                <item.icon className={`h-4 w-4 ${item.color}`} />
              </div>
              <div>
                <p className="text-xl font-bold">{item.value}</p>
                <p className="text-xs text-muted-foreground">{item.label}</p>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
