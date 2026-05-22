"use client";

import { HelpCircle, TrendingDown } from "lucide-react";
import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import type { CostStructureKpi } from "../types";

interface OpexCapexWidgetProps {
  readonly data?: CostStructureKpi;
}

function BreakdownBar({
  items,
}: {
  readonly items: ReadonlyArray<{ category: string; percentage: number; amount_formatted: string }>;
}) {
  const colors = [
    "bg-blue-500",
    "bg-violet-500",
    "bg-teal-500",
    "bg-amber-500",
    "bg-rose-500",
  ];

  return (
    <div className="space-y-2">
      {/* Stacked bar */}
      <div className="flex h-3 rounded-full overflow-hidden">
        {items.map((item, i) => (
          <div
            key={item.category}
            className={`${colors[i % colors.length]} transition-all`}
            style={{ width: `${item.percentage}%` }}
          />
        ))}
      </div>
      {/* Legend */}
      <div className="flex flex-wrap gap-x-4 gap-y-1">
        {items.map((item, i) => (
          <div key={item.category} className="flex items-center gap-1.5 text-xs">
            <span className={`inline-block h-2 w-2 rounded-full ${colors[i % colors.length]}`} />
            <span className="text-muted-foreground">{item.category}</span>
            <span className="font-medium">{item.amount_formatted}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

export function OpexCapexWidget({ data }: OpexCapexWidgetProps) {
  const t = useTranslations("dashboard");

  if (!data) {
    return (
      <Card className="h-full">
        <CardContent className="flex items-center justify-center min-h-40">
          <p className="text-sm text-muted-foreground">
            {t("ownerKpi.noCostData")}
          </p>
        </CardContent>
      </Card>
    );
  }

  const opexWidth = Math.max(data.opex_ratio, 5);
  const capexWidth = Math.max(data.capex_ratio, 5);

  return (
    <Card className="h-full">
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2 text-muted-foreground">
            <TrendingDown className="h-5 w-5 text-primary" />
            <CardTitle className="text-base font-semibold text-foreground">
              {t("ownerKpi.opex_vs_capex.title")}
            </CardTitle>
          </div>
          <TooltipProvider delayDuration={200}>
            <Tooltip>
              <TooltipTrigger asChild>
                <button
                  type="button"
                  className="cursor-pointer text-muted-foreground/50 hover:text-muted-foreground transition-colors"
                  aria-label="Info: Cost Structure"
                >
                  <HelpCircle className="h-4 w-4" />
                </button>
              </TooltipTrigger>
              <TooltipContent side="top" className="max-w-xs text-xs space-y-1">
                <p className="font-semibold">
                  {t("ownerKpi.metrics.kpi_opex_vs_capex.formula")}
                </p>
                <p className="text-muted-foreground">
                  {t("ownerKpi.metrics.kpi_opex_vs_capex.purpose")}
                </p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      </CardHeader>

      <CardContent className="space-y-5">
        {/* OPEX vs CAPEX summary bar */}
        <div className="space-y-3">
          <div className="flex items-center gap-3">
            <div className="flex h-5 flex-1 rounded-full overflow-hidden bg-muted">
              <div
                className="bg-chart-1 rounded-l-full transition-all flex items-center justify-center"
                style={{ width: `${opexWidth}%` }}
              >
                {opexWidth > 20 && (
                  <span className="text-[10px] font-bold text-white">
                    {t("ownerKpi.opex")} {data.opex_ratio.toFixed(0)}%
                  </span>
                )}
              </div>
              <div
                className="bg-chart-2 rounded-r-full transition-all flex items-center justify-center"
                style={{ width: `${capexWidth}%` }}
              >
                {capexWidth > 20 && (
                  <span className="text-[10px] font-bold text-white">
                    {t("ownerKpi.capex")} {data.capex_ratio.toFixed(0)}%
                  </span>
                )}
              </div>
            </div>
          </div>

          {/* Summary numbers */}
          <div className="grid grid-cols-3 gap-3">
            <div className="text-center p-2 rounded-lg bg-chart-1/10">
              <p className="text-[10px] font-medium text-chart-1 uppercase tracking-wider">
                {t("ownerKpi.opex")}
              </p>
              <p className="text-sm font-bold mt-0.5">
                {data.total_opex_formatted}
              </p>
            </div>
            <div className="text-center p-2 rounded-lg bg-chart-2/10">
              <p className="text-[10px] font-medium text-chart-2 uppercase tracking-wider">
                {t("ownerKpi.capex")}
              </p>
              <p className="text-sm font-bold mt-0.5">
                {data.total_capex_formatted}
              </p>
            </div>
            <div className="text-center p-2 rounded-lg bg-muted/50">
              <p className="text-[10px] font-medium text-muted-foreground uppercase tracking-wider">
                {t("ownerKpi.depreciation")}
              </p>
              <p className="text-sm font-bold mt-0.5">
                {data.depreciation_formatted}
              </p>
            </div>
          </div>
        </div>

        {/* OPEX breakdown */}
        {data.opex_breakdown.length > 0 && (
          <div className="space-y-2">
            <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              {t("ownerKpi.opexBreakdown")}
            </h4>
            <BreakdownBar items={data.opex_breakdown} />
          </div>
        )}

        {/* CAPEX breakdown */}
        {data.capex_breakdown.length > 0 && (
          <div className="space-y-2">
            <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              {t("ownerKpi.capexBreakdown")}
            </h4>
            <BreakdownBar items={data.capex_breakdown} />
          </div>
        )}
      </CardContent>
    </Card>
  );
}
