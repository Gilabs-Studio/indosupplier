"use client";

import React, { useEffect, useState, useCallback } from "react";
import { waitingListService } from "@/features/sysadmin/waiting-list/services/waiting-list-service";
import type { WaitingListEntry } from "@/features/sysadmin/waiting-list/types";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import {
  Trash2,
  RefreshCw,
  Inbox,
  Mail,
  Phone,
  ChevronLeft,
  ChevronRight,
  Loader2,
  SlidersHorizontal
} from "lucide-react";
import { format } from "date-fns";

import { Card } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

export default function WaitingListTable() {
  const t = useTranslations("sysadminWaitingList");
  const [entries, setEntries] = useState<WaitingListEntry[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const [statusFilter, setStatusFilter] = useState<string>("");
  const [isLoading, setIsLoading] = useState(true);

  const fetchEntries = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await waitingListService.list({
        page,
        limit,
        status: statusFilter || undefined,
      });
      setEntries(data.items);
      setTotal(data.total);
    } catch {
      toast.error(t("errorLoad"));
    } finally {
      setIsLoading(false);
    }
  }, [page, limit, statusFilter, t]);

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchEntries();
    }, 0);
    return () => clearTimeout(timer);
  }, [fetchEntries]);

  const handleUpdateStatus = async (id: string, newStatus: string) => {
    try {
      await waitingListService.updateStatus(id, newStatus);
      toast.success(t("successStatus", { status: newStatus }));
      fetchEntries();
    } catch {
      toast.error(t("errorStatus"));
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm(t("confirmDelete"))) return;
    try {
      await waitingListService.delete(id);
      toast.success(t("successDelete"));
      fetchEntries();
    } catch {
      toast.error(t("errorDelete"));
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "approved":
        return <Badge className="bg-success/15 text-success border border-success/30 font-bold">{t("approved")}</Badge>;
      case "contacted":
        return <Badge className="bg-primary/10 text-primary border border-primary/20 font-bold">{t("contacted")}</Badge>;
      case "rejected":
        return <Badge variant="destructive" className="font-bold">{t("rejected")}</Badge>;
      default:
        return <Badge className="bg-warning/15 text-warning border border-warning/30 font-bold">{t("pending")}</Badge>;
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
          onClick={fetchEntries}
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
            <span className="text-sm font-semibold">{t("filterRegistrants")}</span>
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
                <SelectItem value="contacted" className="cursor-pointer">{t("contacted")}</SelectItem>
                <SelectItem value="approved" className="cursor-pointer">{t("approved")}</SelectItem>
                <SelectItem value="rejected" className="cursor-pointer">{t("rejected")}</SelectItem>
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
        ) : entries.length === 0 ? (
          <div className="py-24 text-center space-y-2">
            <Inbox className="h-12 w-12 text-muted-foreground/30 mx-auto" />
            <h3 className="font-bold text-lg text-foreground">{t("noRegistrations")}</h3>
            <p className="text-muted-foreground text-sm max-w-sm mx-auto">
              {t("noRegistrationsDesc")}
            </p>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow className="border-b border-border bg-muted/10">
                <TableHead className="pl-6 font-semibold">{t("companyName")}</TableHead>
                <TableHead className="font-semibold">{t("contact")}</TableHead>
                <TableHead className="font-semibold">{t("businessType")}</TableHead>
                <TableHead className="font-semibold">{t("status")}</TableHead>
                <TableHead className="font-semibold">{t("signupDate")}</TableHead>
                <TableHead className="pr-6 text-right font-semibold">{t("actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {entries.map((entry) => (
                <TableRow key={entry.id} className="hover:bg-muted/10 border-b border-border transition-colors duration-200">
                  {/* Name / Company */}
                  <TableCell className="pl-6 py-4">
                    <div className="text-left">
                      <span className="font-bold text-foreground block">{entry.company_name}</span>
                      <span className="text-xs text-muted-foreground block mt-0.5">{entry.name}</span>
                    </div>
                  </TableCell>

                  {/* Contact */}
                  <TableCell className="py-4">
                    <div className="space-y-1 text-left">
                      <span className="flex items-center gap-1.5 text-xs text-muted-foreground">
                        <Mail className="h-3.5 w-3.5 text-muted-foreground/60" />
                        {entry.email}
                      </span>
                      {entry.phone && (
                        <span className="flex items-center gap-1.5 text-xs text-muted-foreground">
                          <Phone className="h-3.5 w-3.5 text-muted-foreground/60" />
                          {entry.phone}
                        </span>
                      )}
                    </div>
                  </TableCell>

                  {/* Business Type */}
                  <TableCell className="py-4 text-left">
                    {getCompanyTypeBadge(entry.company_type)}
                  </TableCell>

                  {/* Status */}
                  <TableCell className="py-4 text-left">
                    {getStatusBadge(entry.status)}
                  </TableCell>

                  {/* Date */}
                  <TableCell className="py-4 text-xs text-muted-foreground text-left">
                    {format(new Date(entry.created_at), "MMM d, yyyy HH:mm")}
                  </TableCell>

                  {/* Actions */}
                  <TableCell className="pr-6 py-4 text-right">
                    <div className="inline-flex items-center justify-end gap-3">
                      <Select
                        value={entry.status}
                        onValueChange={(val) => handleUpdateStatus(entry.id, val)}
                      >
                        <SelectTrigger className="w-[120px] h-8 text-xs bg-background border-border cursor-pointer">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="pending" className="cursor-pointer">{t("pending")}</SelectItem>
                          <SelectItem value="contacted" className="cursor-pointer">{t("contacted")}</SelectItem>
                          <SelectItem value="approved" className="cursor-pointer">{t("approved")}</SelectItem>
                          <SelectItem value="rejected" className="cursor-pointer">{t("rejected")}</SelectItem>
                        </SelectContent>
                      </Select>

                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDelete(entry.id)}
                        className="h-8 w-8 text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors cursor-pointer"
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
        {entries.length > 0 && (
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
    </div>
  );
}
