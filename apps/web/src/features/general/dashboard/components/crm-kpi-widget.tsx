"use client";

import { useTranslations } from "next-intl";
import { ContactRound, UserPlus, Users2, GitBranch, CalendarCheck } from "lucide-react";
import type { WidgetConfig } from "../types";
import { useDashboardWidgetOverview } from "../hooks/use-dashboard";
import { useUserPermission } from "@/hooks/use-user-permission";
import { WIDGET_REGISTRY } from "../config/widget-registry";
import { WidgetAsyncState } from "./widget-async-state";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

interface CrmKpiWidgetProps {
  readonly widget: WidgetConfig;
}

export function CrmKpiWidget({ widget }: CrmKpiWidgetProps) {
  const t = useTranslations("dashboard");
  const tAuthBlock = useTranslations("auth.block");
  const registry = WIDGET_REGISTRY[widget.type];
  const canView = useUserPermission(registry?.permission ?? "");
  const { data, isLoading, isError, refetch } = useDashboardWidgetOverview(widget.type, {
    enabled: canView,
  });

  const crm = data?.crm_summary;
  const recentLeads = (crm?.recent_leads ?? []).slice(0, 5);
  const pipelineStages = (crm?.pipeline_stages ?? []).slice(0, 5);

  if (!canView) {
    return (
      <Card className="h-full border-dashed">
        <CardContent className="flex min-h-40 flex-col items-center justify-center gap-2 p-6 text-center">
          <Users2 className="h-4 w-4 text-muted-foreground" aria-hidden="true" />
          <p className="text-xs text-muted-foreground">{tAuthBlock("description")}</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <WidgetAsyncState isLoading={isLoading} isError={isError} onRetry={refetch}>
      {widget.type === "crm_total_contacts" && (
        <Card className="h-full">
          <CardHeader className="space-y-1 pb-2">
            <CardDescription className="flex items-center gap-2">
              <ContactRound className="h-4 w-4" />
              {t("widgets.crm_total_contacts.title")}
            </CardDescription>
            <CardTitle className="text-2xl lg:text-3xl">{crm?.total_contacts ?? 0}</CardTitle>
          </CardHeader>
          <CardContent />
        </Card>
      )}

      {widget.type === "crm_active_leads" && (
        <Card className="h-full">
          <CardHeader className="space-y-1 pb-2">
            <CardDescription className="flex items-center gap-2">
              <UserPlus className="h-4 w-4" />
              {t("widgets.crm_active_leads.title")}
            </CardDescription>
            <CardTitle className="text-2xl lg:text-3xl">{crm?.active_leads ?? 0}</CardTitle>
          </CardHeader>
          <CardContent className="pt-0">
            <p className="text-xs text-muted-foreground">
              {crm?.leads_this_month ?? 0} {t("widgets.crm_active_leads.thisMonth")}
            </p>
          </CardContent>
        </Card>
      )}

      {widget.type === "crm_leads_list" && (
        <Card className="h-full">
          <CardHeader className="space-y-1 pb-2">
            <CardDescription className="flex items-center gap-2">
              <Users2 className="h-4 w-4" />
              {t("widgets.crm_leads_list.title")}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-2 pt-0">
            {recentLeads.length > 0 ? (
              <div className="space-y-2">
                {recentLeads.map((lead) => (
                  <div key={lead.id} className="rounded-md border border-border/60 px-3 py-2">
                    <div className="flex items-start justify-between gap-2">
                      <div className="min-w-0">
                        <p className="truncate text-sm font-medium leading-tight">{lead.name}</p>
                        <p className="truncate text-xs text-muted-foreground">{lead.company_name || lead.code}</p>
                      </div>
                      {lead.lead_status ? (
                        <span
                          className={cn(
                            "shrink-0 rounded-full px-2 py-0.5 text-[10px] font-medium",
                            lead.lead_status_color ? "bg-muted text-foreground" : "bg-muted text-muted-foreground",
                          )}
                        >
                          {lead.lead_status}
                        </span>
                      ) : null}
                    </div>
                    <div className="mt-1 flex items-center justify-between gap-2 text-[11px] text-muted-foreground">
                      <span>{lead.assigned_to || "-"}</span>
                      <span>{new Intl.DateTimeFormat("id-ID", { day: "2-digit", month: "short" }).format(new Date(lead.created_at))}</span>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-xs text-muted-foreground">{t("widgets.crm_leads_list.noLeads")}</p>
            )}
          </CardContent>
        </Card>
      )}

      {widget.type === "crm_pipeline_summary" && (
        <Card className="h-full">
          <CardHeader className="space-y-1 pb-2">
            <CardDescription className="flex items-center gap-2">
              <GitBranch className="h-4 w-4" />
              {t("widgets.crm_pipeline_summary.title")}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-2 pt-0">
            {pipelineStages.length > 0 ? (
              pipelineStages.map((stage) => (
                <div key={stage.stage_id} className="space-y-1">
                  <div className="flex items-center justify-between gap-2 text-xs">
                    <span className="truncate text-muted-foreground">{stage.stage_name}</span>
                    <span className="font-semibold">{stage.item_count}</span>
                  </div>
                  <div className="h-1.5 rounded-full bg-muted">
                    <div
                      className="h-1.5 rounded-full bg-primary"
                      style={{ width: `${Math.min(100, stage.item_count * 20)}%` }}
                    />
                  </div>
                </div>
              ))
            ) : (
              <p className="text-xs text-muted-foreground">{t("widgets.crm_pipeline_summary.noPipelineData")}</p>
            )}
          </CardContent>
        </Card>
      )}

      {widget.type === "crm_activity_summary" && (
        <Card className="h-full">
          <CardHeader className="space-y-1 pb-2">
            <CardDescription className="flex items-center gap-2">
              <CalendarCheck className="h-4 w-4" />
              {t("widgets.crm_activity_summary.title")}
            </CardDescription>
            <CardTitle className="text-2xl lg:text-3xl">{crm?.activities_today ?? 0}</CardTitle>
          </CardHeader>
          <CardContent className="pt-0">
            <p className="text-xs text-muted-foreground">{t("widgets.crm_activity_summary.activitiesToday")}</p>
          </CardContent>
        </Card>
      )}
    </WidgetAsyncState>
  );
}
