"use client";

import React, { useEffect, useState, useCallback } from "react";
import { abuseReportService } from "@/features/sysadmin/abuse-reports/services";
import type { AbuseReport } from "@/features/sysadmin/abuse-reports/types";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import {
  Trash2,
  RefreshCw,
  Inbox,
  Mail,
  ChevronLeft,
  ChevronRight,
  Loader2,
  SlidersHorizontal,
  Flag,
  AlertTriangle,
  Eye
} from "lucide-react";
import { format } from "date-fns";

import { Card } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";

export default function AbuseReportsList() {
  const t = useTranslations("sysadminAbuseReports");
  const [reports, setReports] = useState<AbuseReport[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const [statusFilter, setStatusFilter] = useState<string>("");
  const [isLoading, setIsLoading] = useState(true);
  const [selectedReport, setSelectedReport] = useState<AbuseReport | null>(null);

  const fetchReports = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await abuseReportService.list({
        page,
        limit,
        status: statusFilter || undefined,
      });
      setReports(data.items);
      setTotal(data.total);
    } catch {
      toast.error(t("errorLoad"));
    } finally {
      setIsLoading(false);
    }
  }, [page, limit, statusFilter, t]);

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchReports();
    }, 0);
    return () => clearTimeout(timer);
  }, [fetchReports]);

  const handleUpdateStatus = async (id: string, newStatus: AbuseReport["status"]) => {
    try {
      await abuseReportService.updateStatus(id, newStatus);
      toast.success(t("successStatus", { status: newStatus }));
      fetchReports();
      if (selectedReport && selectedReport.id === id) {
        setSelectedReport(prev => prev ? { ...prev, status: newStatus } : null);
      }
    } catch {
      toast.error(t("errorStatus"));
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm(t("confirmDelete"))) return;
    try {
      await abuseReportService.delete(id);
      toast.success(t("successDelete"));
      fetchReports();
    } catch {
      toast.error(t("errorDelete"));
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "dismissed":
        return <Badge className="bg-muted text-muted-foreground border border-border font-bold">{t("dismissed")}</Badge>;
      case "warned":
        return <Badge className="bg-warning/15 text-warning border border-warning/30 font-bold">{t("warned")}</Badge>;
      case "suspended":
        return <Badge className="bg-destructive/15 text-destructive border border-destructive/30 font-bold">{t("suspended")}</Badge>;
      default:
        return <Badge className="bg-amber-100 text-amber-700 dark:bg-amber-950/40 dark:text-amber-400 border border-amber-200/30 font-bold">{t("pending")}</Badge>;
    }
  };

  const totalPages = Math.ceil(total / limit) || 1;

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div className="text-left">
          <h1 className="text-2xl font-bold tracking-tight text-foreground">{t("title")}</h1>
          <p className="text-xs text-muted-foreground mt-0.5">{t("subtitle")}</p>
        </div>
        <Button
          onClick={fetchReports}
          variant="outline"
          size="sm"
          className="gap-1.5 hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 hover:shadow-md border-border cursor-pointer"
        >
          <RefreshCw className={`h-3.5 w-3.5 ${isLoading ? "animate-spin" : ""}`} />
          {t("refresh")}
        </Button>
      </div>

      {/* Filter and Content Card */}
      <Card className="border border-border/80 shadow-sm overflow-hidden bg-card">
        {/* Filtering Header */}
        <div className="p-5 border-b border-border flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-muted/20">
          <div className="flex items-center gap-2">
            <SlidersHorizontal className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm font-semibold">{t("filterReports")}</span>
          </div>

          <div className="flex items-center gap-3">
            <Select
              value={statusFilter || "all"}
              onValueChange={(val) => {
                setStatusFilter(val === "all" ? "" : val);
                setPage(1);
              }}
            >
              <SelectTrigger className="w-[180px] bg-background border-border cursor-pointer">
                <SelectValue placeholder={t("allStatuses")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all" className="cursor-pointer">{t("allStatuses")}</SelectItem>
                <SelectItem value="pending" className="cursor-pointer">{t("pending")}</SelectItem>
                <SelectItem value="dismissed" className="cursor-pointer">{t("dismissed")}</SelectItem>
                <SelectItem value="warned" className="cursor-pointer">{t("warned")}</SelectItem>
                <SelectItem value="suspended" className="cursor-pointer">{t("suspended")}</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        {/* Table Content */}
        {isLoading ? (
          <div className="py-24 text-center">
            <Loader2 className="h-8 w-8 animate-spin text-primary mx-auto mb-3" />
            <span className="text-sm font-semibold text-muted-foreground">{t("fetching")}</span>
          </div>
        ) : reports.length === 0 ? (
          <div className="py-24 text-center space-y-2">
            <Inbox className="h-12 w-12 text-muted-foreground/30 mx-auto" />
            <h3 className="font-bold text-lg text-foreground">{t("noReports")}</h3>
            <p className="text-muted-foreground text-sm max-w-sm mx-auto">
              {t("noReportsDesc")}
            </p>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow className="border-b border-border bg-muted/10">
                <TableHead className="pl-6 font-semibold">{t("reporter")}</TableHead>
                <TableHead className="font-semibold">{t("reportedEntity")}</TableHead>
                <TableHead className="font-semibold">{t("reason")}</TableHead>
                <TableHead className="font-semibold">{t("status")}</TableHead>
                <TableHead className="font-semibold">{t("reportDate")}</TableHead>
                <TableHead className="pr-6 text-right font-semibold">{t("actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {reports.map((report) => (
                <TableRow key={report.id} className="hover:bg-muted/10 border-b border-border transition-colors duration-200">
                  {/* Reporter */}
                  <TableCell className="pl-6 py-4">
                    <div className="text-left">
                      <span className="font-bold text-foreground block">{report.reporterName}</span>
                      <span className="text-xs text-muted-foreground flex items-center gap-1 mt-0.5">
                        <Mail className="h-3 w-3 shrink-0" />
                        {report.reporterEmail}
                      </span>
                    </div>
                  </TableCell>

                  {/* Reported Entity */}
                  <TableCell className="py-4 text-left">
                    <div>
                      <span className="font-bold text-foreground block">{report.reportedName}</span>
                      <span className="text-[10px] uppercase font-semibold text-muted-foreground block mt-0.5">
                        {report.reportedType}
                      </span>
                    </div>
                  </TableCell>

                  {/* Reason */}
                  <TableCell className="py-4 text-left">
                    <div className="flex items-center gap-1.5">
                      <AlertTriangle className="h-3.5 w-3.5 text-amber-500 shrink-0" />
                      <span className="text-sm font-medium text-foreground">{report.reason}</span>
                    </div>
                  </TableCell>

                  {/* Status */}
                  <TableCell className="py-4 text-left">
                    {getStatusBadge(report.status)}
                  </TableCell>

                  {/* Date */}
                  <TableCell className="py-4 text-xs text-muted-foreground text-left">
                    {format(new Date(report.createdAt), "MMM d, yyyy HH:mm")}
                  </TableCell>

                  {/* Actions */}
                  <TableCell className="pr-6 py-4 text-right">
                    <div className="inline-flex items-center justify-end gap-3">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => setSelectedReport(report)}
                        className="h-8 w-8 text-muted-foreground hover:text-primary hover:bg-primary/10 cursor-pointer"
                        title={t("viewDetail")}
                      >
                        <Eye className="h-4 w-4" />
                      </Button>

                      <Select
                        value={report.status}
                        onValueChange={(val) => handleUpdateStatus(report.id, val as AbuseReport["status"])}
                      >
                        <SelectTrigger className="w-[120px] h-8 text-xs bg-background border-border cursor-pointer">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="pending" className="cursor-pointer">{t("pending")}</SelectItem>
                          <SelectItem value="dismissed" className="cursor-pointer">{t("dismissed")}</SelectItem>
                          <SelectItem value="warned" className="cursor-pointer">{t("warned")}</SelectItem>
                          <SelectItem value="suspended" className="cursor-pointer">{t("suspended")}</SelectItem>
                        </SelectContent>
                      </Select>

                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDelete(report.id)}
                        className="h-8 w-8 text-muted-foreground hover:text-destructive hover:bg-destructive/10 cursor-pointer"
                        title={t("delete")}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}

        {/* Pagination Footer */}
        {reports.length > 0 && (
          <div className="p-4 border-t border-border flex items-center justify-between bg-muted/10">
            <span className="text-xs text-muted-foreground font-medium">
              {t("showingPage", { page, totalPages, total })}
            </span>

            <div className="inline-flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="h-8 w-8 p-0 border-border cursor-pointer"
              >
                <ChevronLeft className="h-4 w-4" />
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className="h-8 w-8 p-0 border-border cursor-pointer"
              >
                <ChevronRight className="h-4 w-4" />
              </Button>
            </div>
          </div>
        )}
      </Card>

      {/* Inspect Abuse Report Modal */}
      <Dialog open={selectedReport !== null} onOpenChange={(open) => !open && setSelectedReport(null)}>
        <DialogContent className="max-w-md bg-card border-border text-foreground">
          <DialogHeader className="text-left">
            <div className="flex items-center gap-2 text-primary font-bold mb-2">
              <Flag className="h-5 w-5" />
              {t("detailTitle", { id: selectedReport?.id ?? "" })}
            </div>
            <DialogTitle className="text-lg font-bold text-foreground">
              {t("reportedBy", { reportedName: selectedReport?.reportedName ?? "", reporterName: selectedReport?.reporterName ?? "" })}
            </DialogTitle>
            <DialogDescription className="text-xs text-muted-foreground">
              {selectedReport && t("submittedOn", { date: format(new Date(selectedReport.createdAt), "MMM d, yyyy HH:mm") })}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 my-2 text-left">
            <div>
              <span className="text-xs font-semibold text-muted-foreground uppercase block mb-1">{t("reasonLabel")}</span>
              <div className="text-sm font-semibold text-foreground bg-muted/40 p-2.5 rounded-lg border border-border">
                {selectedReport?.reason}
              </div>
            </div>
            
            <div>
              <span className="text-xs font-semibold text-muted-foreground uppercase block mb-1">{t("descLabel")}</span>
              <p className="text-sm text-foreground bg-muted/20 p-3 rounded-lg border border-border/60 leading-relaxed max-h-48 overflow-y-auto font-normal">
                {selectedReport?.description}
              </p>
            </div>

            <div>
              <span className="text-xs font-semibold text-muted-foreground uppercase block mb-1">{t("actionStatus")}</span>
              <div className="flex flex-wrap gap-2">
                {(["pending", "dismissed", "warned", "suspended"] as const).map((st) => (
                  <Button
                    key={st}
                    size="sm"
                    variant={selectedReport?.status === st ? "default" : "outline"}
                    className={
                      selectedReport?.status === st
                        ? "bg-primary hover:bg-primary/90 text-primary-foreground border-none cursor-pointer"
                        : "border-border cursor-pointer text-foreground"
                    }
                    onClick={() => selectedReport && handleUpdateStatus(selectedReport.id, st)}
                  >
                    <span className="capitalize">{t(st)}</span>
                  </Button>
                ))}
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button onClick={() => setSelectedReport(null)} className="w-full bg-primary hover:bg-primary/90 text-primary-foreground font-semibold cursor-pointer">
              {t("close")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
