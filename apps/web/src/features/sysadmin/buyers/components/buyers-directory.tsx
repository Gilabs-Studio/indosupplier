"use client";

import React, { useEffect, useState, useCallback } from "react";
import { buyerService } from "@/features/sysadmin/buyers/services";
import type { Buyer } from "@/features/sysadmin/buyers/types";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import {
  Trash2,
  RefreshCw,
  Inbox,
  ChevronLeft,
  ChevronRight,
  Loader2,
  Search,
  Award
} from "lucide-react";
import { format } from "date-fns";

import { Card } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

export default function BuyersDirectory() {
  const t = useTranslations("sysadminBuyers");
  const [buyers, setBuyers] = useState<Buyer[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const [statusFilter, setStatusFilter] = useState<string>("");
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [isLoading, setIsLoading] = useState(true);

  const fetchBuyers = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await buyerService.list({
        page,
        limit,
        status: statusFilter || undefined,
        search: searchQuery || undefined
      });
      setBuyers(data.items);
      setTotal(data.total);
    } catch {
      toast.error(t("errorLoad"));
    } finally {
      setIsLoading(false);
    }
  }, [page, limit, statusFilter, searchQuery, t]);

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchBuyers();
    }, 0);
    return () => clearTimeout(timer);
  }, [fetchBuyers]);

  const handleUpdateStatus = async (id: string, newStatus: Buyer["status"]) => {
    try {
      await buyerService.updateStatus(id, newStatus);
      toast.success(t("successStatus", { status: newStatus }));
      fetchBuyers();
    } catch {
      toast.error(t("errorStatus"));
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm(t("confirmDelete"))) return;
    try {
      await buyerService.delete(id);
      toast.success(t("successDelete"));
      fetchBuyers();
    } catch {
      toast.error(t("errorDelete"));
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "active":
        return <Badge className="bg-success/15 text-success border border-success/30 font-bold">{t("active")}</Badge>;
      case "suspended":
        return <Badge variant="destructive" className="font-bold">{t("suspended")}</Badge>;
      default:
        return <Badge className="bg-warning/15 text-warning border border-warning/30 font-bold">{t("review")}</Badge>;
    }
  };

  const getScoreBadge = (score: number) => {
    if (score >= 80) {
      return (
        <Badge className="bg-success/10 text-success border border-success/20 font-bold flex items-center gap-1 w-fit">
          <Award className="h-3 w-3" />
          {score} / 100 (A)
        </Badge>
      );
    }
    if (score >= 50) {
      return (
        <Badge className="bg-warning/10 text-warning border border-warning/20 font-bold flex items-center gap-1 w-fit">
          <Award className="h-3 w-3" />
          {score} / 100 (B)
        </Badge>
      );
    }
    return (
      <Badge className="bg-destructive/10 text-destructive border border-destructive/20 font-bold flex items-center gap-1 w-fit">
        <Award className="h-3 w-3" />
        {score} / 100 (C)
      </Badge>
    );
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
          onClick={fetchBuyers}
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
                <SelectItem value="active" className="cursor-pointer">{t("active")}</SelectItem>
                <SelectItem value="review" className="cursor-pointer">{t("review")}</SelectItem>
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
        ) : buyers.length === 0 ? (
          <div className="py-24 text-center space-y-2">
            <Inbox className="h-12 w-12 text-muted-foreground/30 mx-auto" />
            <h3 className="font-bold text-lg text-foreground">{t("noBuyers")}</h3>
            <p className="text-muted-foreground text-sm max-w-sm mx-auto">
              {t("noBuyersDesc")}
            </p>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow className="border-b border-border bg-muted/10">
                <TableHead className="pl-6 font-semibold">{t("buyerCompany")}</TableHead>
                <TableHead className="font-semibold">{t("country")}</TableHead>
                <TableHead className="font-semibold">{t("totalRfq")}</TableHead>
                <TableHead className="font-semibold">{t("leadQualityScore")}</TableHead>
                <TableHead className="font-semibold">{t("status")}</TableHead>
                <TableHead className="font-semibold">{t("signupDate")}</TableHead>
                <TableHead className="pr-6 text-right font-semibold">{t("actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {buyers.map((buyer) => (
                <TableRow key={buyer.id} className="hover:bg-muted/10 border-b border-border transition-colors duration-200">
                  {/* Name / Company */}
                  <TableCell className="pl-6 py-4">
                    <div className="text-left">
                      <span className="font-bold text-foreground block">{buyer.companyName}</span>
                      <span className="text-xs text-muted-foreground block mt-0.5">{buyer.name} ({buyer.email})</span>
                    </div>
                  </TableCell>

                  {/* Country */}
                  <TableCell className="py-4 text-left font-medium text-foreground">
                    {buyer.country}
                  </TableCell>

                  {/* RFQ Count */}
                  <TableCell className="py-4 text-left font-bold text-foreground">
                    {buyer.rfqCount} RFQ
                  </TableCell>

                  {/* Quality Score */}
                  <TableCell className="py-4 text-left">
                    {getScoreBadge(buyer.leadQualityScore)}
                  </TableCell>

                  {/* Status */}
                  <TableCell className="py-4 text-left">
                    {getStatusBadge(buyer.status)}
                  </TableCell>

                  {/* Date */}
                  <TableCell className="py-4 text-xs text-muted-foreground text-left">
                    {format(new Date(buyer.createdAt), "MMM d, yyyy")}
                  </TableCell>

                  {/* Actions */}
                  <TableCell className="pr-6 py-4 text-right">
                    <div className="inline-flex items-center justify-end gap-3">
                      <Select
                        value={buyer.status}
                        onValueChange={(val) => handleUpdateStatus(buyer.id, val as Buyer["status"])}
                      >
                        <SelectTrigger className="w-[120px] h-8 text-xs bg-background border-border cursor-pointer">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="active" className="cursor-pointer">{t("active")}</SelectItem>
                          <SelectItem value="review" className="cursor-pointer">{t("review")}</SelectItem>
                          <SelectItem value="suspended" className="cursor-pointer">{t("suspended")}</SelectItem>
                        </SelectContent>
                      </Select>

                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDelete(buyer.id)}
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
        {buyers.length > 0 && (
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
