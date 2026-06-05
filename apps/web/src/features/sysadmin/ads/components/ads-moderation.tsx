"use client";

import React, { useEffect, useState, useCallback } from "react";
import { adCampaignService } from "@/features/sysadmin/ads/services";
import type { AdCampaign } from "@/features/sysadmin/ads/types";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import {
  Trash2,
  RefreshCw,
  Inbox,
  ChevronLeft,
  ChevronRight,
  Loader2,
  SlidersHorizontal,
  Megaphone,
  Eye,
  TrendingUp,
  BarChart2,
  DollarSign
} from "lucide-react";
import { format } from "date-fns";

import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";

export default function AdsModeration() {
  const t = useTranslations("sysadminAds");
  const [campaigns, setCampaigns] = useState<AdCampaign[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const [statusFilter, setStatusFilter] = useState<string>("");
  const [isLoading, setIsLoading] = useState(true);
  const [selectedCampaign, setSelectedCampaign] = useState<AdCampaign | null>(null);

  const fetchCampaigns = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await adCampaignService.list({
        page,
        limit,
        status: statusFilter || undefined,
      });
      setCampaigns(data.items);
      setTotal(data.total);
    } catch {
      toast.error(t("errorLoad"));
    } finally {
      setIsLoading(false);
    }
  }, [page, limit, statusFilter, t]);

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchCampaigns();
    }, 0);
    return () => clearTimeout(timer);
  }, [fetchCampaigns]);

  const handleUpdateStatus = async (id: string, newStatus: AdCampaign["status"]) => {
    try {
      await adCampaignService.updateStatus(id, newStatus);
      toast.success(t("successStatus", { status: newStatus }));
      fetchCampaigns();
      if (selectedCampaign && selectedCampaign.id === id) {
        setSelectedCampaign(prev => prev ? { ...prev, status: newStatus } : null);
      }
    } catch {
      toast.error(t("errorStatus"));
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm(t("confirmDelete"))) return;
    try {
      await adCampaignService.delete(id);
      toast.success(t("successDelete"));
      fetchCampaigns();
    } catch {
      toast.error(t("errorDelete"));
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "approved":
        return <Badge className="bg-success/15 text-success border border-success/30 font-bold">{t("approved")}</Badge>;
      case "rejected":
        return <Badge variant="destructive" className="font-bold">{t("rejected")}</Badge>;
      case "revised":
        return <Badge className="bg-indigo-100 text-indigo-700 dark:bg-indigo-950/40 dark:text-indigo-400 border border-indigo-200/30 font-bold">{t("revised")}</Badge>;
      default:
        return <Badge className="bg-warning/15 text-warning border border-warning/30 font-bold">{t("pending")}</Badge>;
    }
  };

  const totalPages = Math.ceil(total / limit) || 1;

  // Calculate summary metrics locally
  const totalSpend = campaigns.reduce((acc, c) => acc + (c.status === "approved" ? c.budget : 0), 0);
  const totalClicks = campaigns.reduce((acc, c) => acc + c.clicks, 0);
  const totalImps = campaigns.reduce((acc, c) => acc + c.impressions, 0);
  const ctrRate = totalImps ? ((totalClicks / totalImps) * 100).toFixed(2) : "0";

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div className="text-left">
          <h1 className="text-2xl font-bold tracking-tight text-foreground">{t("title")}</h1>
          <p className="text-xs text-muted-foreground mt-0.5">{t("subtitle")}</p>
        </div>
        <Button
          onClick={fetchCampaigns}
          variant="outline"
          size="sm"
          className="gap-1.5 hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 hover:shadow-md border-border cursor-pointer"
        >
          <RefreshCw className={`h-3.5 w-3.5 ${isLoading ? "animate-spin" : ""}`} />
          {t("refresh")}
        </Button>
      </div>

      {/* Analytics widgets */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-6">
        <Card className="border border-border/85 bg-card">
          <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
            <CardTitle className="text-xs font-semibold text-muted-foreground uppercase">{t("activeSpend")}</CardTitle>
            <DollarSign className="h-4 w-4 text-primary" />
          </CardHeader>
          <CardContent className="text-left">
            <div className="text-2xl font-extrabold text-foreground">Rp {totalSpend.toLocaleString()}</div>
            <p className="text-[10px] text-muted-foreground mt-1">{t("activeSpendDesc")}</p>
          </CardContent>
        </Card>

        <Card className="border border-border/85 bg-card">
          <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
            <CardTitle className="text-xs font-semibold text-muted-foreground uppercase">{t("totalClicks")}</CardTitle>
            <TrendingUp className="h-4 w-4 text-primary" />
          </CardHeader>
          <CardContent className="text-left">
            <div className="text-2xl font-extrabold text-foreground">{totalClicks.toLocaleString()} Clicks</div>
            <p className="text-[10px] text-muted-foreground mt-1">{t("ctrRate", { rate: ctrRate })}</p>
          </CardContent>
        </Card>

        <Card className="border border-border/85 bg-card">
          <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
            <CardTitle className="text-xs font-semibold text-muted-foreground uppercase">{t("totalImpressions")}</CardTitle>
            <BarChart2 className="h-4 w-4 text-purple" />
          </CardHeader>
          <CardContent className="text-left">
            <div className="text-2xl font-extrabold text-foreground">{totalImps.toLocaleString()} Views</div>
            <p className="text-[10px] text-muted-foreground mt-1">{t("totalImpressionsDesc")}</p>
          </CardContent>
        </Card>
      </div>

      {/* Filter and Content Card */}
      <Card className="border border-border/80 shadow-sm overflow-hidden bg-card">
        {/* Filtering Header */}
        <div className="p-5 border-b border-border flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-muted/20">
          <div className="flex items-center gap-2">
            <SlidersHorizontal className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm font-semibold">{t("reviewQueue")}</span>
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
                <SelectItem value="approved" className="cursor-pointer">{t("approved")}</SelectItem>
                <SelectItem value="rejected" className="cursor-pointer">{t("rejected")}</SelectItem>
                <SelectItem value="revised" className="cursor-pointer">{t("revised")}</SelectItem>
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
        ) : campaigns.length === 0 ? (
          <div className="py-24 text-center space-y-2">
            <Inbox className="h-12 w-12 text-muted-foreground/30 mx-auto" />
            <h3 className="font-bold text-lg text-foreground">{t("noCampaigns")}</h3>
            <p className="text-muted-foreground text-sm max-w-sm mx-auto">
              {t("noCampaignsDesc")}
            </p>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow className="border-b border-border bg-muted/10">
                <TableHead className="pl-6 font-semibold">{t("supplier")}</TableHead>
                <TableHead className="font-semibold">{t("placement")}</TableHead>
                <TableHead className="font-semibold">{t("budget")}</TableHead>
                <TableHead className="font-semibold">{t("clicksImps")}</TableHead>
                <TableHead className="font-semibold">{t("status")}</TableHead>
                <TableHead className="font-semibold">{t("dateSubmitted")}</TableHead>
                <TableHead className="pr-6 text-right font-semibold">{t("actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {campaigns.map((camp) => (
                <TableRow key={camp.id} className="hover:bg-muted/10 border-b border-border transition-colors duration-200">
                  {/* Supplier */}
                  <TableCell className="pl-6 py-4">
                    <div className="text-left">
                      <span className="font-bold text-foreground block">{camp.supplierName}</span>
                      <span className="text-[10px] text-muted-foreground block mt-0.5">{camp.id}</span>
                    </div>
                  </TableCell>

                  {/* Placement */}
                  <TableCell className="py-4 text-left">
                    <span className="text-sm font-medium text-foreground">{camp.placement}</span>
                  </TableCell>

                  {/* Budget */}
                  <TableCell className="py-4 text-left font-medium">
                    Rp {camp.budget.toLocaleString()}
                  </TableCell>

                  {/* Performance */}
                  <TableCell className="py-4 text-left">
                    <div className="text-xs">
                      <span className="font-bold text-primary">{camp.clicks}</span>
                      <span className="text-muted-foreground mx-1">/</span>
                      <span className="text-muted-foreground">{camp.impressions}</span>
                    </div>
                  </TableCell>

                  {/* Status */}
                  <TableCell className="py-4 text-left">
                    {getStatusBadge(camp.status)}
                  </TableCell>

                  {/* Date */}
                  <TableCell className="py-4 text-xs text-muted-foreground text-left">
                    {format(new Date(camp.createdAt), "MMM d, yyyy")}
                  </TableCell>

                  {/* Actions */}
                  <TableCell className="pr-6 py-4 text-right">
                    <div className="inline-flex items-center justify-end gap-3">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => setSelectedCampaign(camp)}
                        className="h-8 w-8 text-muted-foreground hover:text-primary hover:bg-primary/10 cursor-pointer"
                        title={t("reviewAd")}
                      >
                        <Eye className="h-4 w-4" />
                      </Button>

                      <Select
                        value={camp.status}
                        onValueChange={(val) => handleUpdateStatus(camp.id, val as AdCampaign["status"])}
                      >
                        <SelectTrigger className="w-[120px] h-8 text-xs bg-background border-border cursor-pointer">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="pending" className="cursor-pointer">{t("pending")}</SelectItem>
                          <SelectItem value="approved" className="cursor-pointer">{t("approved")}</SelectItem>
                          <SelectItem value="rejected" className="cursor-pointer">{t("rejected")}</SelectItem>
                          <SelectItem value="revised" className="cursor-pointer">{t("revised")}</SelectItem>
                        </SelectContent>
                      </Select>

                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDelete(camp.id)}
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
        {campaigns.length > 0 && (
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

      {/* Inspect Campaign Modal */}
      <Dialog open={selectedCampaign !== null} onOpenChange={(open) => !open && setSelectedCampaign(null)}>
        <DialogContent className="max-w-md bg-card border-border text-foreground">
          <DialogHeader className="text-left">
            <div className="flex items-center gap-2 text-primary font-bold mb-2">
              <Megaphone className="h-5 w-5" />
              {t("detailTitle", { id: selectedCampaign?.id ?? "" })}
            </div>
            <DialogTitle className="text-lg font-bold text-foreground">
              {t("reviewHeading", { name: selectedCampaign?.supplierName ?? "" })}
            </DialogTitle>
            <DialogDescription className="text-xs text-muted-foreground">
              {selectedCampaign && t("submittedOn", { date: format(new Date(selectedCampaign.createdAt), "MMM d, yyyy") })}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 my-2 text-left">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <span className="text-xs font-semibold text-muted-foreground uppercase block mb-1">{t("placement")}</span>
                <div className="text-sm font-semibold text-foreground bg-muted/40 p-2 rounded-lg border border-border">
                  {selectedCampaign?.placement}
                </div>
              </div>
              <div>
                <span className="text-xs font-semibold text-muted-foreground uppercase block mb-1">{t("budget")}</span>
                <div className="text-sm font-bold text-foreground bg-muted/40 p-2 rounded-lg border border-border">
                  Rp {selectedCampaign?.budget.toLocaleString()}
                </div>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <span className="text-xs font-semibold text-muted-foreground uppercase block mb-1">{t("views")}</span>
                <div className="text-sm text-foreground bg-muted/20 p-2 rounded-lg border border-border/60">
                  {selectedCampaign?.impressions.toLocaleString()} views
                </div>
              </div>
              <div>
                <span className="text-xs font-semibold text-muted-foreground uppercase block mb-1">{t("clicks")}</span>
                <div className="text-sm text-primary font-semibold bg-muted/20 p-2 rounded-lg border border-border/60">
                  {selectedCampaign?.clicks.toLocaleString()} klik
                </div>
              </div>
            </div>

            <div>
              <span className="text-xs font-semibold text-muted-foreground uppercase block mb-1">{t("reviewActions")}</span>
              <div className="flex flex-wrap gap-2">
                {(["pending", "approved", "rejected", "revised"] as const).map((st) => (
                  <Button
                    key={st}
                    size="sm"
                    variant={selectedCampaign?.status === st ? "default" : "outline"}
                    className={
                      selectedCampaign?.status === st
                        ? "bg-primary hover:bg-primary/90 text-primary-foreground border-none cursor-pointer"
                        : "border-border cursor-pointer text-foreground"
                    }
                    onClick={() => selectedCampaign && handleUpdateStatus(selectedCampaign.id, st)}
                  >
                    <span className="capitalize">{st === "revised" ? t("reqRevision") : t(st)}</span>
                  </Button>
                ))}
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button onClick={() => setSelectedCampaign(null)} className="w-full bg-primary hover:bg-primary/90 text-primary-foreground font-semibold cursor-pointer">
              {t("close")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
