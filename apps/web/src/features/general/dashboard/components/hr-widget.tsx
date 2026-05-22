"use client";

import { useTranslations } from "next-intl";
import { CalendarCheck2, ClipboardList, Users, Clock, AlertCircle } from "lucide-react";
import type { WidgetConfig } from "../types";
import { useDashboardWidgetOverview } from "../hooks/use-dashboard";
import { useUserPermission } from "@/hooks/use-user-permission";
import { WidgetAsyncState } from "./widget-async-state";
import { WIDGET_REGISTRY } from "../config/widget-registry";
import { Progress } from "@/components/ui/progress";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

interface HrWidgetProps {
  readonly widget: WidgetConfig;
}

export function HrWidget({ widget }: HrWidgetProps) {
  const t = useTranslations("dashboard");
  const registry = WIDGET_REGISTRY[widget.type];
  const canView = useUserPermission(registry?.permission ?? "");
  const { data, isLoading, isError, refetch } = useDashboardWidgetOverview(widget.type, {
    enabled: canView,
  });

  const hr = data?.hr_summary;

  if (!canView) return null;

  return (
    <WidgetAsyncState isLoading={isLoading} isError={isError} onRetry={refetch}>
      {widget.type === "hr_attendance_today" && (
        <Card className="h-full">
          <CardHeader className="space-y-1 pb-2">
            <CardDescription className="flex items-center gap-2">
              <CalendarCheck2 className="h-4 w-4" />
              {t("widgets.hr_attendance_today.title")}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3 pt-0">
            <div className="grid grid-cols-3 gap-3">
              <div className="rounded-lg border p-2 text-center">
                <div className="flex items-center justify-center gap-1 text-success">
                  <Users className="h-3 w-3" />
                </div>
                <p className="mt-1 text-lg font-bold">{hr?.present_today ?? 0}</p>
                <p className="text-[10px] text-muted-foreground">{t("widgets.hr_attendance_today.present")}</p>
              </div>
              <div className="rounded-lg border p-2 text-center">
                <div className="flex items-center justify-center gap-1 text-warning">
                  <Clock className="h-3 w-3" />
                </div>
                <p className="mt-1 text-lg font-bold">{hr?.late_today ?? 0}</p>
                <p className="text-[10px] text-muted-foreground">{t("widgets.hr_attendance_today.late")}</p>
              </div>
              <div className="rounded-lg border p-2 text-center">
                <div className="flex items-center justify-center gap-1 text-destructive">
                  <AlertCircle className="h-3 w-3" />
                </div>
                <p className="mt-1 text-lg font-bold">{hr?.absent_today ?? 0}</p>
                <p className="text-[10px] text-muted-foreground">{t("widgets.hr_attendance_today.absent")}</p>
              </div>
            </div>
            <div className="space-y-1">
              <div className="flex justify-between text-xs text-muted-foreground">
                <span>{t("widgets.hr_attendance_today.attendanceRate")}</span>
                <span className="font-medium">{hr?.attendance_rate ?? 0}%</span>
              </div>
              <Progress value={hr?.attendance_rate ?? 0} className="h-1.5" />
            </div>
          </CardContent>
        </Card>
      )}

      {widget.type === "hr_pending_leaves" && (
        <Card className="h-full">
          <CardHeader className="space-y-1 pb-2">
            <CardDescription className="flex items-center gap-2">
              <ClipboardList className="h-4 w-4" />
              {t("widgets.hr_pending_leaves.title")}
            </CardDescription>
            <CardTitle className="text-2xl lg:text-3xl">{hr?.pending_leaves ?? 0}</CardTitle>
          </CardHeader>
          <CardContent className="pt-0">
            <p className="text-xs text-muted-foreground">
              {t("widgets.hr_pending_leaves.pendingRequests")}
            </p>
          </CardContent>
        </Card>
      )}
    </WidgetAsyncState>
  );
}
