"use client";

import { useTranslations } from "next-intl";
import { HelpCircle, ArrowUp, ArrowDown } from "lucide-react";
import { Card, CardHeader, CardDescription } from "@/components/ui/card";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { Badge } from "@/components/ui/badge";
import type { OwnerKpiMetric, KpiHealthStatus } from "../types";

const STATUS_STYLES: Record<
  KpiHealthStatus,
  { badge: string; text: string }
> = {
  good: {
    badge: "bg-success/10 text-success border-transparent",
    text: "text-success",
  },
  warning: {
    badge: "bg-warning/10 text-warning border-transparent",
    text: "text-warning",
  },
  danger: {
    badge: "bg-destructive/10 text-destructive border-transparent",
    text: "text-destructive",
  },
};

interface OwnerKpiCardProps {
  readonly title: string;
  readonly metric: OwnerKpiMetric;
  readonly formula: string;
  readonly purpose: string;
  readonly icon?: React.ReactNode;
}

export function OwnerKpiCard({
  title,
  metric,
  formula,
  purpose,
  icon,
}: OwnerKpiCardProps) {
  const t = useTranslations("dashboard");
  const statusStyle = STATUS_STYLES[metric.status];
  const changePercent = metric.change_percent ?? 0;
  const hasChange = changePercent !== 0;
  const isPositive = changePercent >= 0;

  return (
    <Card className="h-full">
      <CardHeader className="space-y-1">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            {icon && (
              <span className={`shrink-0 ${statusStyle.text}`}>
                {icon}
              </span>
            )}
            <CardDescription>{title}</CardDescription>
          </div>

          <TooltipProvider delayDuration={200}>
            <Tooltip>
              <TooltipTrigger asChild>
                <button
                  type="button"
                  className="cursor-pointer shrink-0 text-muted-foreground/50 hover:text-muted-foreground transition-colors"
                  aria-label={`Info: ${title}`}
                >
                  <HelpCircle className="h-4 w-4" />
                </button>
              </TooltipTrigger>
              <TooltipContent
                side="top"
                className="max-w-xs space-y-1.5 text-xs"
              >
                <p className="font-semibold">{title}</p>
                <p className="text-muted-foreground">
                  <span className="font-medium">{t("formula")}:</span> {formula}
                </p>
                <p className="text-muted-foreground">
                  <span className="font-medium">{t("purpose")}:</span> {purpose}
                </p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>

        <div className="text-2xl font-bold lg:text-3xl">
          {metric.formatted}
        </div>

        <div className="flex items-center gap-2 pt-1">
          <Badge
            variant="outline"
            className={`text-[10px] font-medium ${statusStyle.badge}`}
          >
            {t(`status.${metric.status}`)}
          </Badge>

          {hasChange && (
            <div className="flex items-center text-xs">
              {isPositive ? (
                <ArrowUp className="mr-1 size-3 text-success" />
              ) : (
                <ArrowDown className="mr-1 size-3 text-destructive" />
              )}
              <span
                className={`font-medium ${
                  isPositive ? "text-success" : "text-destructive"
                }`}
              >
                {Math.abs(changePercent).toFixed(1)}%
              </span>
              <span className="text-muted-foreground ml-1">{t("compareLastMonth")}</span>
            </div>
          )}
        </div>
      </CardHeader>
    </Card>
  );
}
