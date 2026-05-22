"use client";

import { ArrowDown, ArrowUp } from "lucide-react";
import { useTranslations } from "next-intl";
import { Card, CardHeader, CardDescription } from "@/components/ui/card";
import type { KpiCardData } from "../types";

interface StatSummaryCardProps {
  readonly label: string;
  readonly data?: KpiCardData;
}

export function StatSummaryCard({ label, data }: StatSummaryCardProps) {
  const t = useTranslations("dashboard");
  const changePercent = data?.change_percent ?? 0;
  const isPositive = changePercent >= 0;
  const hasChange = changePercent !== 0;

  return (
    <Card className="h-full">
      <CardHeader className="space-y-1">
        <CardDescription>{label}</CardDescription>
        <div className="text-2xl font-bold lg:text-3xl">
          {data?.formatted ?? "Rp 0"}
        </div>
        {hasChange && (
          <div className="flex items-center text-xs">
            {isPositive ? (
              <ArrowUp
                className="mr-1 size-3 text-success"
                aria-hidden="true"
              />
            ) : (
              <ArrowDown
                className="mr-1 size-3 text-destructive"
                aria-hidden="true"
              />
            )}
            <span
              className={`font-medium ${isPositive ? "text-success" : "text-destructive"}`}
            >
              {Math.abs(changePercent).toFixed(1)}%
            </span>
            <span className="ml-1 text-muted-foreground">
              {t("compareLastMonth")}
            </span>
          </div>
        )}
      </CardHeader>
    </Card>
  );
}
