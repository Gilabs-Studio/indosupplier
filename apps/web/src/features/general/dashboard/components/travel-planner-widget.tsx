"use client";

import { useMemo } from "react";
import { useQueries } from "@tanstack/react-query";
import { ArrowUpRight, Route, CheckCircle2, Package } from "lucide-react";
import { useTranslations } from "next-intl";
import { useRouter } from "@/i18n/routing";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useTravelPlans, travelPlannerKeys } from "@/features/travel/travel-planner/collaborative/hooks/use-travel-planner";
import { travelPlannerService } from "@/features/travel/travel-planner/collaborative/services/travel-planner-service";
import { formatCurrency, formatTime } from "@/lib/utils";
import { WidgetAsyncState } from "./widget-async-state";

type PlanSpendSummary = {
  id: string;
  code: string;
  title: string;
  budget: number;
  spent: number;
  remaining: number;
  usagePercent: number;
};

function calculateUsagePercent(budget: number, spent: number): number {
  if (budget <= 0) return 0;
  return Math.min(Math.round((spent / budget) * 100), 100);
}

interface HalfCircleProgressProps {
  value: number;
}

function HalfCircleProgress({ value }: HalfCircleProgressProps) {
  const safeValue = Math.max(0, Math.min(100, value));
  const radius = 56;
  const stroke = 12;
  const circumference = Math.PI * radius;
  const offset = circumference - (safeValue / 100) * circumference;

  return (
    <div className="relative h-32 w-56">
      <svg viewBox="0 0 160 90" className="h-full w-full">
        <path
          d="M 24 80 A 56 56 0 0 1 136 80"
          fill="none"
          stroke="hsl(var(--muted))"
          strokeWidth={stroke}
          strokeLinecap="round"
        />
        <path
          d="M 24 80 A 56 56 0 0 1 136 80"
          fill="none"
          stroke="hsl(var(--primary))"
          strokeWidth={stroke}
          strokeLinecap="round"
          strokeDasharray={circumference}
          strokeDashoffset={offset}
          className="transition-all duration-500"
        />
      </svg>
      <div className="pointer-events-none absolute inset-0 flex items-end justify-center pb-1">
        <div className="text-center">
          <p className="text-3xl font-semibold leading-none">{safeValue}%</p>
          <p className="mt-1 text-[11px] uppercase tracking-wide text-muted-foreground">Budget Utilization</p>
        </div>
      </div>
    </div>
  );
}

export function TravelPlannerWidget() {
  const t = useTranslations("dashboard");
  const router = useRouter();

  const plansQuery = useTravelPlans({
    page: 1,
    per_page: 10,
  });

  const plans = useMemo(() => plansQuery.data?.data ?? [], [plansQuery.data?.data]);

  const expenseQueries = useQueries({
    queries: plans.map((plan) => ({
      queryKey: travelPlannerKeys.expenses(plan.id),
      queryFn: () => travelPlannerService.listExpenses(plan.id),
      enabled: !!plan.id,
      staleTime: 30_000,
    })),
  });

  // Fetch linked visits for each plan
  const visitQueries = useQueries({
    queries: plans.map((plan) => ({
      queryKey: travelPlannerKeys.visits(plan.id),
      queryFn: () => travelPlannerService.listVisits(plan.id),
      enabled: !!plan.id,
      staleTime: 30_000,
    })),
  });

  const planSummaries = useMemo<PlanSpendSummary[]>(() => {
    return plans.map((plan, index) => {
      const spent = expenseQueries[index]?.data?.data?.total_amount ?? 0;
      const budget = plan.budget_amount ?? 0;
      const remaining = Math.max(budget - spent, 0);
      const usagePercent = calculateUsagePercent(budget, spent);

      return {
        id: plan.id,
        code: plan.code,
        title: plan.title,
        budget,
        spent,
        remaining,
        usagePercent,
      };
    });
  }, [expenseQueries, plans]);

  const totals = useMemo(() => {
    const totalBudget = planSummaries.reduce((sum, item) => sum + item.budget, 0);
    const totalSpent = planSummaries.reduce((sum, item) => sum + item.spent, 0);
    const totalRemaining = Math.max(totalBudget - totalSpent, 0);
    const overallUsage = calculateUsagePercent(totalBudget, totalSpent);

    return {
      totalBudget,
      totalSpent,
      totalRemaining,
      overallUsage,
    };
  }, [planSummaries]);

  // Count total visits
  const totalVisits = useMemo(() => {
    return visitQueries.reduce((sum, query) => sum + (query.data?.data?.length ?? 0), 0);
  }, [visitQueries]);

  // Get all visits from all plans
  const allVisits = useMemo(() => {
    return visitQueries.flatMap((query) => query.data?.data ?? []);
  }, [visitQueries]);

  const isLoading = plansQuery.isLoading || expenseQueries.some((query) => query.isLoading) || visitQueries.some((query) => query.isLoading);
  const isError = plansQuery.isError || expenseQueries.some((query) => query.isError) || visitQueries.some((query) => query.isError);

  return (
    <WidgetAsyncState
      isLoading={isLoading}
      isError={isError}
      onRetry={() => {
        void plansQuery.refetch();
        expenseQueries.forEach((query) => {
          void query.refetch();
        });
        visitQueries.forEach((query) => {
          void query.refetch();
        });
      }}
    >
      <Card className="h-full">
        <CardHeader>
          <div className="flex items-start justify-between gap-2">
            <div>
              <CardTitle>{t("widgets.travel_planner_overview.title")}</CardTitle>
              <CardDescription>{t("widgets.travel_planner_overview.description")}</CardDescription>
            </div>
            <Button
              type="button"
              variant="outline"
              size="sm"
              className="cursor-pointer"
              onClick={() => router.push("/travel/travel-planner")}
            >
              <ArrowUpRight className="h-4 w-4 mr-1" />
              {t("widgets.travel_planner_overview.open")}
            </Button>
          </div>
        </CardHeader>

        <CardContent className="space-y-6">
          {/* Expenses Section */}
          <div className="space-y-3">
            <div className="rounded-xl border bg-muted/20 p-4">
              <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <div className="flex items-center justify-center sm:justify-center sm:flex-1">
                  <HalfCircleProgress value={totals.overallUsage} />
                </div>
                <div className="grid flex-1 grid-cols-1 gap-2 sm:max-w-xs">
                  <div className="rounded-lg border bg-background/80 px-3 py-2">
                    <p className="text-[11px] uppercase tracking-wide text-muted-foreground">{t("widgets.travel_planner_overview.metrics.budget")}</p>
                    <p className="text-sm font-semibold">{formatCurrency(totals.totalBudget)}</p>
                  </div>
                  <div className="rounded-lg border bg-background/80 px-3 py-2">
                    <p className="text-[11px] uppercase tracking-wide text-muted-foreground">{t("widgets.travel_planner_overview.metrics.spent")}</p>
                    <p className="text-sm font-semibold">{formatCurrency(totals.totalSpent)}</p>
                  </div>
                  <div className="rounded-lg border bg-background/80 px-3 py-2">
                    <p className="text-[11px] uppercase tracking-wide text-muted-foreground">{t("widgets.travel_planner_overview.metrics.remaining")}</p>
                    <p className="text-sm font-semibold">{formatCurrency(totals.totalRemaining)}</p>
                  </div>
                </div>
              </div>
            </div>

            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <p className="text-xs text-muted-foreground uppercase tracking-wide">{t("widgets.travel_planner_overview.listTitle")}</p>
                <Badge variant="secondary">{planSummaries.length}</Badge>
              </div>

              {planSummaries.length === 0 ? (
                <div className="rounded-lg border border-dashed p-4 text-sm text-muted-foreground text-center">
                  {t("widgets.travel_planner_overview.empty")}
                </div>
              ) : (
                <div className="space-y-1.5">
                  {planSummaries.map((plan) => (
                    <button
                      key={plan.id}
                      type="button"
                      className="w-full rounded-lg border px-3 py-2 text-left hover:bg-muted/40 transition-colors cursor-pointer"
                      onClick={() => router.push(`/travel/travel-planner?plan_id=${plan.id}`)}
                    >
                      <div className="grid grid-cols-[minmax(0,1fr)_150px_20px] items-center gap-3">
                        <div className="min-w-0">
                          <p className="text-sm font-medium line-clamp-1">{plan.title}</p>
                          <p className="text-[11px] text-muted-foreground mt-0.5">{plan.code}</p>
                        </div>
                        <div className="shrink-0 text-right">
                          <p className="text-xs text-muted-foreground">{formatCurrency(plan.spent)} / {formatCurrency(plan.budget)}</p>
                          <Badge variant="outline" className="mt-1 inline-flex w-11 justify-center text-[10px]">{plan.usagePercent}%</Badge>
                        </div>
                        <Route className="h-4 w-4 text-muted-foreground" />
                      </div>
                      <p className="mt-1 text-[11px] text-muted-foreground">
                        {t("widgets.travel_planner_overview.metrics.remaining")}: {formatCurrency(plan.remaining)}
                      </p>
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* Visits Section */}
          <div className="space-y-2 border-t pt-4">
            <div className="flex items-center justify-between">
              <p className="text-xs text-muted-foreground uppercase tracking-wide">Active Visits</p>
              <Badge variant="secondary">{totalVisits}</Badge>
            </div>

            {allVisits.length === 0 ? (
              <div className="rounded-lg border border-dashed p-4 text-sm text-muted-foreground text-center">
                No active visits
              </div>
            ) : (
              <div className="space-y-1.5">
                {allVisits.slice(0, 5).map((visit) => (
                  <button
                    key={visit.id}
                    type="button"
                    className="w-full rounded-lg border px-3 py-2 text-left hover:bg-muted/40 transition-colors cursor-pointer"
                    onClick={() => router.push(`/crm/visit-report/${visit.id}`)}
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div className="min-w-0 flex-1">
                        <p className="text-sm font-medium line-clamp-1">{visit.customer_name}</p>
                        <p className="text-[11px] text-muted-foreground mt-0.5">{visit.code}</p>
                        <div className="flex flex-wrap gap-1 mt-1">
                          {visit.check_in_at && (
                            <Badge variant="outline" className="text-[10px] gap-1 py-0">
                              <CheckCircle2 className="h-2.5 w-2.5" />
                              {formatTime(visit.check_in_at)}
                            </Badge>
                          )}
                          {visit.product_interest_count > 0 && (
                            <Badge variant="outline" className="text-[10px] gap-1 py-0">
                              <Package className="h-2.5 w-2.5" />
                              {visit.product_interest_count} products
                            </Badge>
                          )}
                        </div>
                      </div>
                      {visit.outcome && (
                        <Badge
                          variant={
                            visit.outcome === "positive" || visit.outcome === "very_positive"
                              ? "default"
                              : visit.outcome === "negative"
                                ? "destructive"
                                : "secondary"
                          }
                          className="shrink-0 text-[10px]"
                        >
                          {visit.outcome}
                        </Badge>
                      )}
                    </div>
                  </button>
                ))}
                {totalVisits > 5 && (
                  <p className="text-[11px] text-muted-foreground text-center py-1">
                    +{totalVisits - 5} more visits
                  </p>
                )}
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </WidgetAsyncState>
  );
}
