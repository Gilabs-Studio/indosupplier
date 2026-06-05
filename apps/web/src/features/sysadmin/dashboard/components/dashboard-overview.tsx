"use client";

import React, { useEffect, useState, useCallback } from "react";
import { useSysadminStore } from "@/features/sysadmin/auth/stores/use-sysadmin-store";
import { waitingListService } from "@/features/sysadmin/waiting-list/services/waiting-list-service";
import type { WaitingListEntry } from "@/features/sysadmin/waiting-list/types";
import { toast } from "sonner";
import { Link } from "@/i18n/routing";
import { useTranslations } from "next-intl";
import {
  Users,
  Building,
  UserCheck,
  Inbox,
  ArrowRight,
  ShieldCheck,
  Calendar,
  Loader2,
  RefreshCw
} from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

export default function DashboardOverview() {
  const t = useTranslations("sysadminDashboard");
  const { admin } = useSysadminStore();
  const [isLoading, setIsLoading] = useState(true);
  const [recentEntries, setRecentEntries] = useState<WaitingListEntry[]>([]);
  const [stats, setStats] = useState({
    total: 0,
    suppliers: 0,
    buyers: 0,
    pending: 0,
  });

  const loadDashboardData = useCallback(async () => {
    setIsLoading(true);
    try {
      // Load recent 5 entries
      const recent = await waitingListService.list({ page: 1, limit: 5 });
      setRecentEntries(recent.items);

      // Load stats by calculating over a large batch (up to 1000 items)
      const allData = await waitingListService.list({ page: 1, limit: 1000 });
      const supplierCount = allData.items.filter(i => i.company_type === "supplier").length;
      const buyerCount = allData.items.filter(i => i.company_type === "buyer").length;
      const pendingCount = allData.items.filter(i => i.status === "pending").length;

      setStats({
        total: allData.total,
        suppliers: supplierCount,
        buyers: buyerCount,
        pending: pendingCount,
      });
    } catch {
      toast.error(t("errorLoad"));
    } finally {
      setIsLoading(false);
    }
  }, [t]);

  useEffect(() => {
    const timer = setTimeout(() => {
      loadDashboardData();
    }, 0);
    return () => clearTimeout(timer);
  }, [loadDashboardData]);

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "approved":
        return <Badge className="bg-success/15 text-success border border-success/30 font-bold">Approved</Badge>;
      case "contacted":
        return <Badge className="bg-primary/10 text-primary border border-primary/20 font-bold">Contacted</Badge>;
      case "rejected":
        return <Badge variant="destructive" className="font-bold">Rejected</Badge>;
      default:
        return <Badge className="bg-warning/15 text-warning border border-warning/30 font-bold">Pending</Badge>;
    }
  };

  const getCompanyTypeBadge = (type: string) => {
    switch (type) {
      case "supplier":
        return (
          <Badge className="bg-purple/10 text-purple border border-purple/20 font-bold">
            Supplier
          </Badge>
        );
      case "buyer":
        return (
          <Badge className="bg-cyan/10 text-cyan border border-cyan/20 font-bold">
            Buyer
          </Badge>
        );
      default:
        return <Badge variant="secondary" className="font-bold">Other</Badge>;
    }
  };

  if (isLoading) {
    return (
      <div className="py-24 text-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary mx-auto mb-3" />
        <span className="text-sm font-semibold text-muted-foreground">Loading...</span>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      {/* Welcome Banner */}
      <div className="bg-gradient-to-r from-primary to-primary/80 text-primary-foreground rounded-lg p-6 shadow-sm flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
        <div className="space-y-1 text-left">
          <h2 className="text-xl font-bold">{t("title", { name: admin?.name || "System Admin" })}</h2>
          <p className="text-xs text-primary-foreground/95 max-w-2xl font-light">
            {t("subtitle")}
          </p>
        </div>
        <button
          onClick={loadDashboardData}
          className="flex items-center gap-1.5 px-3 py-1.5 bg-primary-foreground/15 hover:bg-primary-foreground/25 active:scale-[0.98] text-primary-foreground rounded-lg text-xs font-semibold backdrop-blur-sm transition-all cursor-pointer"
        >
          <RefreshCw className="h-3.5 w-3.5" />
          {t("refresh")}
        </button>
      </div>

      {/* Statistics Cards Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
        {/* Card 1: Total registrations */}
        <Card className="hover:shadow-lg hover:-translate-y-0.5 transition-all duration-300 bg-card border border-border/80 text-left">
          <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
            <CardTitle className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              {t("totalRegistrations")}
            </CardTitle>
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-primary border border-border">
              <Users className="h-5 w-5" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-extrabold text-foreground">{stats.total}</div>
            <p className="text-xs text-muted-foreground mt-1">{t("totalRegistrationsDesc")}</p>
          </CardContent>
        </Card>

        {/* Card 2: Suppliers */}
        <Card className="hover:shadow-lg hover:-translate-y-0.5 transition-all duration-300 bg-card border border-border/80 text-left">
          <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
            <CardTitle className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              {t("suppliers")}
            </CardTitle>
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-purple/10 text-purple border border-purple/20">
              <Building className="h-5 w-5" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-extrabold text-foreground">{stats.suppliers}</div>
            <p className="text-xs text-muted-foreground mt-1">{t("suppliersDesc")}</p>
          </CardContent>
        </Card>

        {/* Card 3: Buyers */}
        <Card className="hover:shadow-lg hover:-translate-y-0.5 transition-all duration-300 bg-card border border-border/80 text-left">
          <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
            <CardTitle className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              {t("buyers")}
            </CardTitle>
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-cyan/10 text-cyan border border-cyan/20">
              <UserCheck className="h-5 w-5" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-extrabold text-foreground">{stats.buyers}</div>
            <p className="text-xs text-muted-foreground mt-1">{t("buyersDesc")}</p>
          </CardContent>
        </Card>

        {/* Card 4: Pending Review */}
        <Card className="hover:shadow-lg hover:-translate-y-0.5 transition-all duration-300 bg-card border border-border/80 text-left">
          <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
            <CardTitle className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              {t("needReview")}
            </CardTitle>
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-warning/15 text-warning border border-warning/30">
              <Inbox className="h-5 w-5" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-extrabold text-foreground">{stats.pending}</div>
            <p className="text-xs text-muted-foreground mt-1">{t("needReviewDesc")}</p>
          </CardContent>
        </Card>
      </div>

      {/* Main Sections Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Left Column: Recent Activity */}
        <Card className="lg:col-span-8 border border-border/80 bg-card shadow-sm overflow-hidden flex flex-col justify-between">
          <div>
            <CardHeader className="border-b border-border bg-muted/10 p-5 flex flex-row items-center justify-between">
              <div className="space-y-1 text-left">
                <CardTitle className="text-base font-bold text-foreground">{t("recentRegistrations")}</CardTitle>
                <p className="text-xs text-muted-foreground">{t("recentRegistrationsDesc")}</p>
              </div>
              <Link
                href="/sysadmin/waiting-list"
                className="flex items-center gap-1 text-xs font-semibold text-primary hover:text-primary/80 hover:translate-x-0.5 transition-all cursor-pointer"
              >
                {t("viewAll")}
                <ArrowRight className="h-3.5 w-3.5" />
              </Link>
            </CardHeader>

            <CardContent className="p-0 divide-y divide-border">
              {recentEntries.length === 0 ? (
                <div className="py-12 text-center text-muted-foreground">
                  <Inbox className="h-10 w-10 mx-auto opacity-30 mb-2" />
                  <span className="text-sm">{t("noRecentRegistrations")}</span>
                </div>
              ) : (
                recentEntries.map((entry) => (
                  <div key={entry.id} className="p-4 flex items-center justify-between hover:bg-muted/10 transition-colors">
                    <div className="space-y-1 text-left">
                      <span className="font-bold text-foreground text-sm block">{entry.company_name}</span>
                      <span className="text-xs text-muted-foreground block">{entry.name} ({entry.email})</span>
                    </div>

                    <div className="flex items-center gap-3">
                      {getCompanyTypeBadge(entry.company_type)}
                      {getStatusBadge(entry.status)}
                    </div>
                  </div>
                ))
              )}
            </CardContent>
          </div>
        </Card>

        {/* Right Column: Quick Info / System Stats */}
        <Card className="lg:col-span-4 border border-border/80 bg-card shadow-sm overflow-hidden">
          <CardHeader className="border-b border-border bg-muted/10 p-5 text-left">
            <CardTitle className="text-base font-bold text-foreground">{t("adminInfo")}</CardTitle>
          </CardHeader>
          <CardContent className="p-6 space-y-6">
            {/* Status Panel */}
            <div className="flex items-center gap-3">
              <div className="h-12 w-12 rounded-lg bg-primary/10 text-primary flex items-center justify-center border border-border">
                <ShieldCheck className="h-6 w-6" />
              </div>
              <div className="text-left">
                <span className="text-sm font-bold block text-foreground">{t("securityStatus")}</span>
                <span className="text-xs text-primary font-semibold block mt-0.5">{t("securityStatusDesc")}</span>
              </div>
            </div>

            <div className="h-px bg-border" />

            {/* Quick Tips */}
            <div className="space-y-3 text-left">
              <span className="text-xs font-bold text-muted-foreground uppercase tracking-wider block">{t("featureShortcuts")}</span>
              <div className="space-y-2">
                <Link
                  href="/sysadmin/waiting-list"
                  className="flex items-center justify-between p-3 bg-muted/30 hover:bg-muted/70 border border-border rounded-lg text-sm text-foreground transition cursor-pointer"
                >
                  <span>{t("manageWaitlist")}</span>
                  <ArrowRight className="h-4 w-4 text-muted-foreground" />
                </Link>
              </div>
            </div>

            <div className="h-px bg-border" />

            {/* Platform Version */}
            <div className="flex items-center justify-between text-xs text-muted-foreground">
              <span className="flex items-center gap-1.5">
                <Calendar className="h-3.5 w-3.5" />
                Version: 1.0.0
              </span>
              <span>Built: June 2026</span>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
