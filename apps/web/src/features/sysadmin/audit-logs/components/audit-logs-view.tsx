"use client";

import React, { useEffect, useState, useCallback } from "react";
import { auditLogService } from "@/features/sysadmin/audit-logs/services";
import type { AuditLog } from "@/features/sysadmin/audit-logs/types";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import {
  RefreshCw,
  Inbox,
  ChevronLeft,
  ChevronRight,
  Loader2,
  ScrollText,
  Search,
  Eye
} from "lucide-react";
import { format } from "date-fns";

import { Card } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";

export default function AuditLogsView() {
  const t = useTranslations("sysadminAuditLogs");
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const [actionFilter, setActionFilter] = useState<string>("");
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [isLoading, setIsLoading] = useState(true);
  const [selectedLog, setSelectedLog] = useState<AuditLog | null>(null);

  const fetchLogs = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await auditLogService.list({
        page,
        limit,
        action: actionFilter || undefined,
        search: searchQuery || undefined
      });
      setLogs(data.items);
      setTotal(data.total);
    } catch {
      toast.error(t("errorLoad"));
    } finally {
      setIsLoading(false);
    }
  }, [page, limit, actionFilter, searchQuery, t]);

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchLogs();
    }, 0);
    return () => clearTimeout(timer);
  }, [fetchLogs]);

  const totalPages = Math.ceil(total / limit) || 1;

  const getActionColor = (action: string) => {
    if (action.includes("suspend") || action.includes("ban") || action.includes("delete")) {
      return "bg-destructive/15 text-destructive border-destructive/30";
    }
    if (action.includes("approve") || action.includes("create")) {
      return "bg-success/15 text-success border-success/30";
    }
    return "bg-primary/10 text-primary border-primary/20";
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div className="text-left">
          <h1 className="text-2xl font-bold tracking-tight text-foreground">{t("title")}</h1>
          <p className="text-xs text-muted-foreground mt-0.5">{t("subtitle")}</p>
        </div>
        <Button
          onClick={fetchLogs}
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
          <div className="flex items-center gap-2 max-w-sm w-full relative">
            <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
            <input
              type="text"
              placeholder={t("searchPlaceholder")}
              value={searchQuery}
              onChange={(e) => {
                setSearchQuery(e.target.value);
                setPage(1);
              }}
              className="w-full pl-9 pr-4 py-2 text-xs bg-background border border-border rounded-lg placeholder-muted-foreground focus:outline-none focus:border-primary"
            />
          </div>

          <div className="flex items-center gap-3">
            <Select
              value={actionFilter || "all"}
              onValueChange={(val) => {
                setActionFilter(val === "all" ? "" : val);
                setPage(1);
              }}
            >
              <SelectTrigger className="w-[180px] bg-background border-border cursor-pointer">
                <SelectValue placeholder={t("allActions")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all" className="cursor-pointer">{t("allActions")}</SelectItem>
                <SelectItem value="user.suspend" className="cursor-pointer">user.suspend</SelectItem>
                <SelectItem value="waitlist.auto_approve" className="cursor-pointer">waitlist.auto_approve</SelectItem>
                <SelectItem value="ads.approve" className="cursor-pointer">ads.approve</SelectItem>
                <SelectItem value="faq.create" className="cursor-pointer">faq.create</SelectItem>
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
        ) : logs.length === 0 ? (
          <div className="py-24 text-center space-y-2">
            <Inbox className="h-12 w-12 text-muted-foreground/30 mx-auto" />
            <h3 className="font-bold text-lg text-foreground">{t("noLogs")}</h3>
            <p className="text-muted-foreground text-sm max-w-sm mx-auto">
              {t("noLogsDesc")}
            </p>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow className="border-b border-border bg-muted/10">
                <TableHead className="pl-6 font-semibold">{t("logId")}</TableHead>
                <TableHead className="font-semibold">{t("actor")}</TableHead>
                <TableHead className="font-semibold">{t("action")}</TableHead>
                <TableHead className="font-semibold">{t("target")}</TableHead>
                <TableHead className="font-semibold">{t("dateTime")}</TableHead>
                <TableHead className="pr-6 text-right font-semibold">{t("details")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {logs.map((log) => (
                <TableRow key={log.id} className="hover:bg-muted/10 border-b border-border transition-colors duration-200">
                  {/* Log ID */}
                  <TableCell className="pl-6 py-4 font-mono text-xs font-semibold text-left">
                    {log.id}
                  </TableCell>

                  {/* Actor */}
                  <TableCell className="py-4">
                    <div className="text-left">
                      <span className="font-bold text-foreground block">{log.actorName}</span>
                      <span className="text-xs text-muted-foreground block">{log.actorEmail}</span>
                    </div>
                  </TableCell>

                  {/* Action */}
                  <TableCell className="py-4 text-left">
                    <Badge className={`font-mono text-[10px] capitalize border ${getActionColor(log.action)}`}>
                      {log.action}
                    </Badge>
                  </TableCell>

                  {/* Target */}
                  <TableCell className="py-4 text-left font-medium text-sm text-foreground">
                    {log.target}
                  </TableCell>

                  {/* Date */}
                  <TableCell className="py-4 text-xs text-muted-foreground text-left">
                    {format(new Date(log.createdAt), "yyyy-MM-dd HH:mm:ss")}
                  </TableCell>

                  {/* Actions */}
                  <TableCell className="pr-6 py-4 text-right">
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => setSelectedLog(log)}
                      className="h-8 w-8 text-muted-foreground hover:text-primary hover:bg-primary/10 cursor-pointer"
                      title={t("inspect")}
                    >
                      <Eye className="h-4 w-4" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}

        {/* Pagination Footer */}
        {logs.length > 0 && (
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

      {/* Inspect Log Details Modal */}
      <Dialog open={selectedLog !== null} onOpenChange={(open) => !open && setSelectedLog(null)}>
        <DialogContent className="max-w-md bg-card border-border text-foreground">
          <DialogHeader className="text-left">
            <div className="flex items-center gap-2 text-primary font-bold mb-2">
              <ScrollText className="h-5 w-5" />
              {t("detailTitle", { id: selectedLog?.id ?? "" })}
            </div>
            <DialogTitle className="text-lg font-bold text-foreground">
              {t("actionLabel", { action: selectedLog?.action ?? "" })}
            </DialogTitle>
            <DialogDescription className="text-xs text-muted-foreground">
              {selectedLog && t("recordedBy", { name: selectedLog.actorName, date: format(new Date(selectedLog.createdAt), "yyyy-MM-dd HH:mm:ss") })}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 my-2 text-left">
            <div>
              <span className="text-xs font-semibold text-muted-foreground uppercase block mb-1">{t("targetChange")}</span>
              <div className="text-sm font-semibold text-foreground bg-muted/40 p-2.5 rounded-lg border border-border">
                {selectedLog?.target}
              </div>
            </div>

            <div>
              <span className="text-xs font-semibold text-muted-foreground uppercase block mb-1">{t("metadataJson")}</span>
              <pre className="text-xs text-foreground bg-muted/20 p-3 rounded-lg border border-border/60 overflow-x-auto leading-relaxed font-mono font-normal">
                {selectedLog && JSON.stringify(JSON.parse(selectedLog.metadata), null, 2)}
              </pre>
            </div>
          </div>

          <DialogFooter>
            <Button onClick={() => setSelectedLog(null)} className="w-full bg-primary hover:bg-primary/90 text-primary-foreground font-semibold cursor-pointer">
              {t("close")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
