"use client";

import {
  Activity,
  AlertTriangle,
  CheckCircle2,
  XCircle,
  Lightbulb,
  ChevronRight,
  ShieldAlert,
  Package,
  Banknote,
  Building2,
  Truck,
  TrendingUp,
  Coins,
} from "lucide-react";
import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import type { OwnerIntelligence, KpiHealthStatus, BottleneckArea, OwnerInsightItem } from "../types";

// ─── Styling maps ──────────────────────────────────────────────────────────

const HEALTH_ICON: Record<KpiHealthStatus, React.ReactNode> = {
  good: <CheckCircle2 className="h-5 w-5 text-success" />,
  warning: <AlertTriangle className="h-5 w-5 text-warning" />,
  danger: <XCircle className="h-5 w-5 text-destructive" />,
};

const HEALTH_LABEL_KEY: Record<KpiHealthStatus, string> = {
  good: "ownerKpi.healthStatus.good",
  warning: "ownerKpi.healthStatus.warning",
  danger: "ownerKpi.healthStatus.danger",
};

const HEALTH_BG: Record<KpiHealthStatus, string> = {
  good: "bg-success/10",
  warning: "bg-warning/10",
  danger: "bg-destructive/10",
};

const HEALTH_TEXT: Record<KpiHealthStatus, string> = {
  good: "text-success",
  warning: "text-warning",
  danger: "text-destructive",
};

const AREA_ICON: Record<BottleneckArea, React.ReactNode> = {
  inventory: <Package className="h-4 w-4" />,
  cashflow: <Banknote className="h-4 w-4" />,
  asset: <Building2 className="h-4 w-4" />,
  logistics: <Truck className="h-4 w-4" />,
  profitability: <TrendingUp className="h-4 w-4" />,
  cost: <Coins className="h-4 w-4" />,
};

const AREA_LABEL_KEY: Record<BottleneckArea, string> = {
  inventory: "ownerKpi.areas.inventory",
  cashflow: "ownerKpi.areas.cashflow",
  asset: "ownerKpi.areas.asset",
  logistics: "ownerKpi.areas.logistics",
  profitability: "ownerKpi.areas.profitability",
  cost: "ownerKpi.areas.cost",
};

const SEVERITY_BADGE: Record<KpiHealthStatus, string> = {
  good: "bg-success/10 text-success border-transparent",
  warning: "bg-warning/10 text-warning border-transparent",
  danger: "bg-destructive/10 text-destructive border-transparent",
};

// ─── Components ─────────────────────────────────────────────────────────────

function InsightItem({ insight }: { readonly insight: OwnerInsightItem }) {
  const t = useTranslations("dashboard");
  return (
    <div className="flex gap-3 p-3 rounded-lg bg-muted/30 transition-colors">
      <div className="shrink-0 mt-0.5">
        {AREA_ICON[insight.area]}
      </div>
      <div className="flex-1 min-w-0 space-y-1">
        <div className="flex items-center gap-2 flex-wrap">
          <span className="text-sm font-medium">{insight.title}</span>
          <Badge
            variant="outline"
            className={`text-[10px] ${SEVERITY_BADGE[insight.severity]}`}
          >
            {t(AREA_LABEL_KEY[insight.area] as Parameters<typeof t>[0])}
          </Badge>
        </div>
        <p className="text-xs text-muted-foreground leading-relaxed">
          {insight.description}
        </p>
        {insight.action && (
          <div className="flex items-start gap-1.5 mt-1.5">
            <Lightbulb className="h-3 w-3 text-warning mt-0.5 shrink-0" />
            <p className="text-xs text-warning font-medium">
              {insight.action}
            </p>
          </div>
        )}
      </div>
    </div>
  );
}

interface OwnerIntelligenceWidgetProps {
  readonly data?: OwnerIntelligence;
}

export function OwnerIntelligenceWidget({ data }: OwnerIntelligenceWidgetProps) {
  const t = useTranslations("dashboard");

  if (!data) {
    return (
      <Card className="h-full">
        <CardContent className="flex items-center justify-center min-h-40">
          <p className="text-sm text-muted-foreground">
            {t("ownerKpi.noData")}
          </p>
        </CardContent>
      </Card>
    );
  }

  const insights = data.insights ?? [];

  return (
    <Card className="h-full">
      <CardHeader>
        <div className="flex items-center gap-2 text-muted-foreground">
          <Activity className="h-5 w-5 text-primary" />
          <CardTitle className="text-base font-semibold text-foreground">
            {t("ownerKpi.intelligence.title")}
          </CardTitle>
        </div>
      </CardHeader>

      <CardContent className="space-y-4">
        {/* ── Health Summary Banner ── */}
        <div
          className={`flex items-start gap-3 p-4 rounded-xl ${HEALTH_BG[data.overall_health]}`}
        >
          <div className="shrink-0 mt-0.5">
            {HEALTH_ICON[data.overall_health]}
          </div>
          <div className="space-y-1">
            <p className={`text-sm font-bold ${HEALTH_TEXT[data.overall_health]}`}>
              {t(HEALTH_LABEL_KEY[data.overall_health] as Parameters<typeof t>[0])}
            </p>
            <p className="text-xs text-muted-foreground leading-relaxed">
              {data.health_summary}
            </p>
          </div>
        </div>

        {/* ── Bottleneck Indicator ── */}
        <div className="flex items-center gap-3 p-3 rounded-lg bg-muted/30">
          <ShieldAlert className="h-4 w-4 text-muted-foreground shrink-0" />
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                {t("ownerKpi.primaryBottleneck")}
              </span>
              <ChevronRight className="h-3 w-3 text-muted-foreground" />
              <Badge variant="secondary" className="text-xs font-semibold">
                {AREA_ICON[data.primary_bottleneck]}
                <span className="ml-1">
                  {t(AREA_LABEL_KEY[data.primary_bottleneck] as Parameters<typeof t>[0])}
                </span>
              </Badge>
            </div>
            <p className="text-xs text-muted-foreground mt-1 leading-relaxed">
              {data.bottleneck_summary}
            </p>
          </div>
        </div>

        {/* ── Insight Items ── */}
        {insights.length > 0 && (
          <div className="space-y-2">
            <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider px-1">
              {t("ownerKpi.insightsRecommendations")}
            </h4>
            <div className="space-y-2 max-h-64 overflow-y-auto pr-1">
              {insights.map((insight) => (
                <InsightItem key={insight.id} insight={insight} />
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
