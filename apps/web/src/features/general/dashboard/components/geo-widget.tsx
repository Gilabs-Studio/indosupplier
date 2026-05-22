"use client";

import dynamic from "next/dynamic";
import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useMemo } from "react";
import { formatCurrency } from "@/lib/utils";
import type { GeoOverviewData } from "../types";
import { MapIcon } from "lucide-react";

// Load Leaflet map only on the client to avoid SSR issues
const GeoMapLeaflet = dynamic(
  () => import("./geo-map-leaflet").then((m) => ({ default: m.GeoMapLeaflet })),
  { ssr: false, loading: () => <Skeleton className="h-[350px] w-full rounded-lg" /> },
);

interface GeoWidgetProps {
  readonly data?: GeoOverviewData;
}

// Color scale from light to dark (7 steps)
const COLOR_SCALE = [
  "#e0f2fe",
  "#bae6fd",
  "#7dd3fc",
  "#38bdf8",
  "#0ea5e9",
  "#0284c7",
  "#0369a1",
];

export function GeoWidget({ data }: GeoWidgetProps) {
  const t = useTranslations("dashboard");
  const regions = useMemo(() => data?.regions ?? [], [data?.regions]);

  const sortedRegions = useMemo(
    () => [...regions].sort((a, b) => b.value - a.value),
    [regions],
  );

  return (
    <Card className="h-full">
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base font-semibold">
            {t("widgets.geographic_overview.title")}
          </CardTitle>
          {data?.total_value !== undefined && (
            <span className="text-sm font-semibold text-primary">
              {formatCurrency(data.total_value)}
            </span>
          )}
        </div>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
          {/* Leaflet choropleth map or empty placeholder */}
          <div className="lg:col-span-2">
            {regions.length === 0 ? (
              <div className="relative flex h-[350px] items-center justify-center rounded-lg border bg-secondary/30">
                <div className="text-center text-muted-foreground">
                  <MapIcon className="mx-auto mb-2 h-8 w-8 text-muted-foreground/60" />
                  <p className="text-sm">{t("noData")}</p>
                </div>
              </div>
            ) : (
              <>
                <GeoMapLeaflet regions={regions} />
                <div className="mt-2 flex items-center gap-1 text-xs text-muted-foreground">
                  <span>{t("widgets.geographic_overview.low")}</span>
                  {COLOR_SCALE.map((c) => (
                    <div
                      key={c}
                      className="h-3 w-5 rounded-sm"
                      style={{ backgroundColor: c }}
                    />
                  ))}
                  <span>{t("widgets.geographic_overview.high")}</span>
                </div>
              </>
            )}
          </div>

          {/* Region leaderboard */}
          <div className="max-h-[390px] space-y-1 overflow-y-auto lg:col-span-1">
            {sortedRegions.slice(0, 15).map((region, i) => (
              <div
                key={region.code ?? region.name}
                className="flex items-center justify-between rounded-md px-2 py-1.5 transition-colors hover:bg-secondary"
              >
                <div className="flex items-center gap-2">
                  <span className="flex h-5 w-5 items-center justify-center rounded-full bg-primary/10 text-[10px] font-bold text-primary">
                    {i + 1}
                  </span>
                  <span className="text-sm">{region.name}</span>
                </div>
                <span className="text-xs font-medium text-muted-foreground">
                  {region.formatted}
                </span>
              </div>
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
