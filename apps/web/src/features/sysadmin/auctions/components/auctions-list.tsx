"use client";

import React, { useEffect, useState, useCallback } from "react";
import { auctionSessionService } from "@/features/sysadmin/auctions/services";
import type { AuctionSession } from "@/features/sysadmin/auctions/types";
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
  Gavel,
  Plus,
  Calendar
} from "lucide-react";
import { format } from "date-fns";

import { Card } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Field, FieldLabel, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";

export default function AuctionsList() {
  const t = useTranslations("sysadminAuctions");
  const [auctions, setAuctions] = useState<AuctionSession[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const [statusFilter, setStatusFilter] = useState<string>("");
  const [isLoading, setIsLoading] = useState(true);
  
  // Creation modal state
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [newCategory, setNewCategory] = useState("");
  const [newSlots, setNewSlots] = useState(2);
  const [newMinBid, setNewMinBid] = useState(1500000);
  const [newStartDate, setNewStartDate] = useState("2026-06-05");
  const [newEndDate, setNewEndDate] = useState("2026-06-20");

  const fetchAuctions = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await auctionSessionService.list({
        page,
        limit,
        status: statusFilter || undefined,
      });
      setAuctions(data.items);
      setTotal(data.total);
    } catch {
      toast.error(t("errorLoad"));
    } finally {
      setIsLoading(false);
    }
  }, [page, limit, statusFilter, t]);

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchAuctions();
    }, 0);
    return () => clearTimeout(timer);
  }, [fetchAuctions]);

  const handleUpdateStatus = async (id: string, newStatus: AuctionSession["status"]) => {
    try {
      await auctionSessionService.updateStatus(id, newStatus);
      toast.success(t("successStatus", { status: newStatus }));
      fetchAuctions();
    } catch {
      toast.error(t("errorStatus"));
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm(t("confirmDelete"))) return;
    try {
      await auctionSessionService.delete(id);
      toast.success(t("successDelete"));
      fetchAuctions();
    } catch {
      toast.error(t("errorDelete"));
    }
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newCategory.trim()) {
      toast.error(t("categoryRequired"));
      return;
    }
    try {
      await auctionSessionService.create({
        category: newCategory,
        slots: Number(newSlots),
        minBid: Number(newMinBid),
        status: "draft",
        startDate: new Date(newStartDate).toISOString(),
        endDate: new Date(newEndDate).toISOString()
      });
      toast.success(t("successCreate"));
      setIsCreateOpen(false);
      setNewCategory("");
      fetchAuctions();
    } catch {
      toast.error(t("errorCreate"));
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "open":
        return <Badge className="bg-success/15 text-success border border-success/30 font-bold">{t("open")}</Badge>;
      case "closed":
        return <Badge className="bg-muted text-muted-foreground border border-border font-bold">{t("closed")}</Badge>;
      default:
        return <Badge className="bg-warning/15 text-warning border border-warning/30 font-bold">{t("draft")}</Badge>;
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
        <div className="flex items-center gap-3">
          <Button
            onClick={() => setIsCreateOpen(true)}
            className="bg-primary hover:bg-primary/90 text-primary-foreground font-bold text-xs uppercase tracking-wider py-5 px-4 rounded-lg flex items-center gap-1.5 hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 hover:shadow-md cursor-pointer"
          >
            <Plus className="h-4 w-4" />
            {t("newSession")}
          </Button>
          <Button
            onClick={fetchAuctions}
            variant="outline"
            size="sm"
            className="gap-1.5 hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 hover:shadow-md border-border cursor-pointer"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${isLoading ? "animate-spin" : ""}`} />
            {t("refresh")}
          </Button>
        </div>
      </div>

      {/* Filter and Content Card */}
      <Card className="border border-border/80 shadow-sm overflow-hidden bg-card">
        {/* Filtering Header */}
        <div className="p-5 border-b border-border flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-muted/20">
          <div className="flex items-center gap-2">
            <SlidersHorizontal className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm font-semibold">{t("filterSessions")}</span>
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
                <SelectItem value="draft" className="cursor-pointer">{t("draft")}</SelectItem>
                <SelectItem value="open" className="cursor-pointer">{t("open")}</SelectItem>
                <SelectItem value="closed" className="cursor-pointer">{t("closed")}</SelectItem>
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
        ) : auctions.length === 0 ? (
          <div className="py-24 text-center space-y-2">
            <Inbox className="h-12 w-12 text-muted-foreground/30 mx-auto" />
            <h3 className="font-bold text-lg text-foreground">{t("noAuctions")}</h3>
            <p className="text-muted-foreground text-sm max-w-sm mx-auto">
              {t("noAuctionsDesc")}
            </p>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow className="border-b border-border bg-muted/10">
                <TableHead className="pl-6 font-semibold">{t("category")}</TableHead>
                <TableHead className="font-semibold">{t("slotQuota")}</TableHead>
                <TableHead className="font-semibold">{t("minBid")}</TableHead>
                <TableHead className="font-semibold">{t("bidsCount")}</TableHead>
                <TableHead className="font-semibold">{t("highestBid")}</TableHead>
                <TableHead className="font-semibold">{t("status")}</TableHead>
                <TableHead className="font-semibold">{t("sessionSchedule")}</TableHead>
                <TableHead className="pr-6 text-right font-semibold">{t("actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {auctions.map((auc) => (
                <TableRow key={auc.id} className="hover:bg-muted/10 border-b border-border transition-colors duration-200">
                  {/* Category */}
                  <TableCell className="pl-6 py-4">
                    <div className="text-left">
                      <span className="font-bold text-foreground block">{auc.category}</span>
                      <span className="text-[10px] text-muted-foreground block mt-0.5">{auc.id}</span>
                    </div>
                  </TableCell>

                  {/* Slots */}
                  <TableCell className="py-4 text-left font-medium">
                    {auc.slots} {t("slotUnit")}
                  </TableCell>

                  {/* Min Bid */}
                  <TableCell className="py-4 text-left font-medium">
                    Rp {auc.minBid.toLocaleString()}
                  </TableCell>

                  {/* Total Bids */}
                  <TableCell className="py-4 text-left font-bold text-primary">
                    {auc.bidsCount} {t("bidUnit")}
                  </TableCell>

                  {/* Highest Bid */}
                  <TableCell className="py-4 text-left font-bold text-foreground">
                    {auc.highestBid > 0 ? `Rp ${auc.highestBid.toLocaleString()}` : "-"}
                  </TableCell>

                  {/* Status */}
                  <TableCell className="py-4 text-left">
                    {getStatusBadge(auc.status)}
                  </TableCell>

                  {/* Date Schedule */}
                  <TableCell className="py-4 text-xs text-muted-foreground text-left font-normal">
                    <div className="space-y-0.5">
                      <span className="flex items-center gap-1">
                        <Calendar className="h-3 w-3 shrink-0" />
                        {t("startDate", { date: format(new Date(auc.startDate), "MMM d, yyyy") })}
                      </span>
                      <span className="flex items-center gap-1">
                        <Calendar className="h-3 w-3 shrink-0" />
                        {t("endDate", { date: format(new Date(auc.endDate), "MMM d, yyyy") })}
                      </span>
                    </div>
                  </TableCell>

                  {/* Actions */}
                  <TableCell className="pr-6 py-4 text-right">
                    <div className="inline-flex items-center justify-end gap-3">
                      <Select
                        value={auc.status}
                        onValueChange={(val) => handleUpdateStatus(auc.id, val as AuctionSession["status"])}
                      >
                        <SelectTrigger className="w-[120px] h-8 text-xs bg-background border-border cursor-pointer">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="draft" className="cursor-pointer">{t("draft")}</SelectItem>
                          <SelectItem value="open" className="cursor-pointer">{t("open")}</SelectItem>
                          <SelectItem value="closed" className="cursor-pointer">{t("closed")}</SelectItem>
                        </SelectContent>
                      </Select>

                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDelete(auc.id)}
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
        {auctions.length > 0 && (
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

      {/* Create Auction Session Dialog */}
      <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
        <DialogContent className="max-w-md bg-card border-border text-foreground">
          <DialogHeader className="text-left">
            <div className="flex items-center gap-2 text-primary font-bold mb-2">
              <Gavel className="h-5 w-5" />
              {t("dialogTitle")}
            </div>
            <DialogTitle className="text-lg font-bold text-foreground">
              {t("dialogTitle")}
            </DialogTitle>
            <DialogDescription className="text-xs text-muted-foreground">
              {t("dialogSubtitle")}
            </DialogDescription>
          </DialogHeader>

          <form onSubmit={handleCreate} className="space-y-4 my-2 text-left">
            <FieldGroup className="space-y-3">
              <Field>
                <FieldLabel className="text-xs font-semibold uppercase text-muted-foreground">{t("formCategory")}</FieldLabel>
                <Input
                  type="text"
                  placeholder="e.g. Agriculture, Furniture"
                  value={newCategory}
                  onChange={(e) => setNewCategory(e.target.value)}
                  className="bg-background border-border"
                />
              </Field>

              <div className="grid grid-cols-2 gap-4">
                <Field>
                  <FieldLabel className="text-xs font-semibold uppercase text-muted-foreground">{t("formSlots")}</FieldLabel>
                  <Input
                    type="number"
                    value={newSlots}
                    onChange={(e) => setNewSlots(Number(e.target.value))}
                    min={1}
                    className="bg-background border-border"
                  />
                </Field>
                <Field>
                  <FieldLabel className="text-xs font-semibold uppercase text-muted-foreground">{t("formMinBid")}</FieldLabel>
                  <Input
                    type="number"
                    value={newMinBid}
                    onChange={(e) => setNewMinBid(Number(e.target.value))}
                    min={100000}
                    className="bg-background border-border"
                  />
                </Field>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <Field>
                  <FieldLabel className="text-xs font-semibold uppercase text-muted-foreground">{t("formStartDate")}</FieldLabel>
                  <Input
                    type="date"
                    value={newStartDate}
                    onChange={(e) => setNewStartDate(e.target.value)}
                    className="bg-background border-border"
                  />
                </Field>
                <Field>
                  <FieldLabel className="text-xs font-semibold uppercase text-muted-foreground">{t("formEndDate")}</FieldLabel>
                  <Input
                    type="date"
                    value={newEndDate}
                    onChange={(e) => setNewEndDate(e.target.value)}
                    className="bg-background border-border"
                  />
                </Field>
              </div>
            </FieldGroup>

            <DialogFooter className="pt-2">
              <Button type="button" variant="outline" onClick={() => setIsCreateOpen(false)} className="border-border cursor-pointer text-foreground">
                {t("batal")}
              </Button>
              <Button type="submit" className="bg-primary hover:bg-primary/90 text-primary-foreground font-bold cursor-pointer">
                {t("simpan")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
