"use client";

import React, { useEffect, useState, useCallback } from "react";
import { reviewService } from "@/features/sysadmin/reviews/services";
import type { Review } from "@/features/sysadmin/reviews/types";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import {
  Trash2,
  RefreshCw,
  Inbox,
  CheckCircle,
  AlertTriangle,
  Star,
  MessageSquare,
  Search,
  Filter,
  Loader2
} from "lucide-react";

import { Card } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Dialog, DialogContent, DialogHeader, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Field, FieldLabel, FieldGroup } from "@/components/ui/field";
import { Textarea } from "@/components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Input } from "@/components/ui/input";

export default function ReviewsModeration() {
  const t = useTranslations("sysadminReviews");
  const [reviews, setReviews] = useState<Review[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  // Search & Filter state
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedRating, setSelectedRating] = useState<string>("all");
  const [selectedStatus, setSelectedStatus] = useState<string>("all");

  // Reply Modal State
  const [isReplyOpen, setIsReplyOpen] = useState(false);
  const [activeReview, setActiveReview] = useState<Review | null>(null);
  const [replyText, setReplyText] = useState("");

  const fetchReviews = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await reviewService.list();
      setReviews(data);
    } catch {
      toast.error(t("subtitle"));
    } finally {
      setIsLoading(false);
    }
  }, [t]);

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchReviews();
    }, 0);
    return () => clearTimeout(timer);
  }, [fetchReviews]);

  const handleUpdateStatus = async (id: string, newStatus: Review["status"]) => {
    try {
      await reviewService.update(id, { status: newStatus });
      toast.success(t("successStatus"));
      fetchReviews();
    } catch {
      toast.error("Error");
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm(t("delete") + "?")) return;
    try {
      await reviewService.delete(id);
      toast.success(t("successStatus"));
      fetchReviews();
    } catch {
      toast.error("Error");
    }
  };

  const handleOpenReply = (review: Review) => {
    setActiveReview(review);
    setReplyText(review.reply || "");
    setIsReplyOpen(true);
  };

  const handleSaveReply = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!activeReview) return;

    try {
      await reviewService.update(activeReview.id, { reply: replyText });
      toast.success(t("successSave"));
      setIsReplyOpen(false);
      fetchReviews();
    } catch {
      toast.error("Error");
    }
  };

  // Metrics calculations
  const totalCount = reviews.length;
  const avgRating = totalCount > 0 ? (reviews.reduce((sum, r) => sum + r.rating, 0) / totalCount).toFixed(1) : "0.0";
  const pendingCount = reviews.filter(r => r.status === "pending").length;
  const flaggedCount = reviews.filter(r => r.status === "flagged").length;

  // Filters logic
  const filteredReviews = reviews.filter(r => {
    const matchesSearch = r.buyerName.toLowerCase().includes(searchQuery.toLowerCase()) ||
                          r.supplierName.toLowerCase().includes(searchQuery.toLowerCase()) ||
                          (r.productName && r.productName.toLowerCase().includes(searchQuery.toLowerCase())) ||
                          r.content.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesRating = selectedRating === "all" || r.rating.toString() === selectedRating;
    const matchesStatus = selectedStatus === "all" || r.status === selectedStatus;
    return matchesSearch && matchesRating && matchesStatus;
  });

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div className="text-left">
          <h1 className="text-2xl font-bold tracking-tight text-foreground">{t("title")}</h1>
          <p className="text-xs text-muted-foreground mt-0.5">{t("subtitle")}</p>
        </div>
        <Button
          onClick={fetchReviews}
          variant="outline"
          size="sm"
          className="gap-1.5 hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 hover:shadow-md border-border cursor-pointer"
        >
          <RefreshCw className={`h-3.5 w-3.5 ${isLoading ? "animate-spin" : ""}`} />
          {t("refresh")}
        </Button>
      </div>

      {/* Metrics Section */}
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4">
        <Card className="p-4 border border-border bg-card hover:shadow-md transition-shadow text-left">
          <span className="text-[10px] uppercase font-bold text-muted-foreground tracking-wider">{t("totalReviews")}</span>
          <h2 className="text-2xl font-extrabold text-foreground mt-1">{totalCount}</h2>
        </Card>

        <Card className="p-4 border border-border bg-card hover:shadow-md transition-shadow text-left">
          <span className="text-[10px] uppercase font-bold text-muted-foreground tracking-wider">{t("averageRating")}</span>
          <div className="flex items-center gap-1.5 mt-1">
            <h2 className="text-2xl font-extrabold text-foreground">{avgRating}</h2>
            <div className="flex text-amber-400">
              <Star className="h-4 w-4 fill-current" />
            </div>
          </div>
        </Card>

        <Card className="p-4 border border-border bg-card hover:shadow-md transition-shadow text-left">
          <span className="text-[10px] uppercase font-bold text-muted-foreground tracking-wider">{t("pendingReview")}</span>
          <h2 className="text-2xl font-extrabold text-warning mt-1">{pendingCount}</h2>
        </Card>

        <Card className="p-4 border border-border bg-card hover:shadow-md transition-shadow text-left">
          <span className="text-[10px] uppercase font-bold text-muted-foreground tracking-wider">{t("flaggedReview")}</span>
          <h2 className="text-2xl font-extrabold text-destructive mt-1">{flaggedCount}</h2>
        </Card>
      </div>

      {/* Filter Bar */}
      <Card className="p-4 border border-border bg-card flex flex-col md:flex-row gap-4 items-center justify-between">
        <div className="relative w-full md:w-80">
          <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder={t("searchPlaceholder")}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9 bg-background border-border text-sm"
          />
        </div>
        <div className="flex flex-wrap w-full md:w-auto items-center gap-4">
          {/* Status filter */}
          <div className="flex items-center gap-2">
            <Filter className="h-3.5 w-3.5 text-muted-foreground" />
            <span className="text-xs text-muted-foreground">{t("status")}:</span>
          </div>
          <Select value={selectedStatus} onValueChange={setSelectedStatus}>
            <SelectTrigger className="w-[140px] bg-background border-border text-xs h-9 cursor-pointer">
              <SelectValue placeholder={t("all")} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all" className="cursor-pointer">{t("all")}</SelectItem>
              <SelectItem value="approved" className="cursor-pointer">Approved</SelectItem>
              <SelectItem value="pending" className="cursor-pointer">Pending</SelectItem>
              <SelectItem value="flagged" className="cursor-pointer">Flagged</SelectItem>
              <SelectItem value="spam" className="cursor-pointer">Spam</SelectItem>
            </SelectContent>
          </Select>

          {/* Rating filter */}
          <div className="flex items-center gap-2">
            <Star className="h-3.5 w-3.5 text-muted-foreground" />
            <span className="text-xs text-muted-foreground">{t("rating")}:</span>
          </div>
          <Select value={selectedRating} onValueChange={setSelectedRating}>
            <SelectTrigger className="w-[120px] bg-background border-border text-xs h-9 cursor-pointer">
              <SelectValue placeholder={t("all")} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all" className="cursor-pointer">{t("all")}</SelectItem>
              <SelectItem value="5" className="cursor-pointer">5 Bintang</SelectItem>
              <SelectItem value="4" className="cursor-pointer">4 Bintang</SelectItem>
              <SelectItem value="3" className="cursor-pointer">3 Bintang</SelectItem>
              <SelectItem value="2" className="cursor-pointer">2 Bintang</SelectItem>
              <SelectItem value="1" className="cursor-pointer">1 Bintang</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </Card>

      {/* Content Table Card */}
      <Card className="border border-border/80 shadow-sm overflow-hidden bg-card">
        {isLoading ? (
          <div className="py-24 text-center">
            <Loader2 className="h-8 w-8 animate-spin text-primary mx-auto mb-3" />
            <span className="text-sm font-semibold text-muted-foreground">Loading...</span>
          </div>
        ) : filteredReviews.length === 0 ? (
          <div className="py-24 text-center space-y-2">
            <Inbox className="h-12 w-12 text-muted-foreground/30 mx-auto" />
            <h3 className="font-bold text-lg text-foreground">Empty</h3>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow className="border-b border-border bg-muted/10">
                <TableHead className="pl-6 font-semibold">{t("buyerSupplier")}</TableHead>
                <TableHead className="font-semibold">{t("ratingContent")}</TableHead>
                <TableHead className="font-semibold">{t("reply")}</TableHead>
                <TableHead className="font-semibold">{t("status")}</TableHead>
                <TableHead className="pr-6 text-right font-semibold">{t("actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredReviews.map((rev) => (
                <TableRow key={rev.id} className="hover:bg-muted/10 border-b border-border transition-colors duration-200">
                  <TableCell className="pl-6 py-4">
                    <div className="text-left space-y-1">
                      <div className="font-bold text-foreground">{rev.buyerName}</div>
                      <div className="text-xs text-muted-foreground">
                        ke: <span className="font-bold text-primary">{rev.supplierName}</span>
                      </div>
                      {rev.productName && (
                        <div className="text-[10px] text-muted-foreground italic truncate max-w-[200px]">
                          {rev.productName}
                        </div>
                      )}
                    </div>
                  </TableCell>

                  <TableCell className="py-4 max-w-md">
                    <div className="text-left space-y-1.5">
                      <div className="flex text-amber-400">
                        {Array.from({ length: 5 }).map((_, i) => (
                          <Star
                            key={i}
                            className={`h-3.5 w-3.5 ${
                              i < rev.rating ? "fill-current" : "text-slate-200"
                            }`}
                          />
                        ))}
                      </div>
                      <p className="text-xs font-normal text-foreground whitespace-pre-wrap leading-relaxed line-clamp-2">
                        &quot;{rev.content}&quot;
                      </p>
                    </div>
                  </TableCell>

                  <TableCell className="py-4 max-w-xs text-left">
                    {rev.reply ? (
                      <div className="text-xs text-muted-foreground line-clamp-2 italic bg-muted/45 p-2 rounded border border-border/50">
                        &quot;{rev.reply}&quot;
                      </div>
                    ) : (
                      <span className="text-[10px] text-muted-foreground/60 italic">(Belum ditanggapi)</span>
                    )}
                  </TableCell>

                  <TableCell className="py-4 text-left">
                    <Badge className={`capitalize font-bold text-[10px] px-2 py-0.5 ${
                      rev.status === "approved"
                        ? "bg-success/15 text-success border-success/30"
                        : rev.status === "pending"
                        ? "bg-warning/15 text-warning border-warning/30"
                        : rev.status === "flagged"
                        ? "bg-destructive/15 text-destructive border-destructive/30"
                        : "bg-muted text-muted-foreground border-border"
                    }`}>
                      {rev.status}
                    </Badge>
                  </TableCell>

                  <TableCell className="pr-6 py-4 text-right">
                    <div className="flex items-center justify-end gap-1.5">
                      {rev.status !== "approved" && (
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleUpdateStatus(rev.id, "approved")}
                          className="h-8 w-8 text-success hover:text-success/90 hover:bg-success/10 cursor-pointer"
                          title={t("approve")}
                        >
                          <CheckCircle className="h-4 w-4" />
                        </Button>
                      )}
                      
                      {rev.status !== "flagged" && (
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleUpdateStatus(rev.id, "flagged")}
                          className="h-8 w-8 text-warning hover:text-warning/90 hover:bg-warning/10 cursor-pointer"
                          title={t("flag")}
                        >
                          <AlertTriangle className="h-4 w-4" />
                        </Button>
                      )}

                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleOpenReply(rev)}
                        className="h-8 w-8 text-primary hover:text-primary/90 hover:bg-primary/10 cursor-pointer"
                        title={t("submitReply")}
                      >
                        <MessageSquare className="h-4 w-4" />
                      </Button>

                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDelete(rev.id)}
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
      </Card>

      {/* Reply Dialog */}
      <Dialog open={isReplyOpen} onOpenChange={setIsReplyOpen}>
        <DialogContent className="max-w-md bg-card border-border text-foreground">
          <DialogHeader className="text-left">
            <div className="flex items-center gap-2 text-primary font-bold mb-1">
              <MessageSquare className="h-5 w-5" />
              {t("replyTitle")}
            </div>
            <DialogDescription className="text-xs text-muted-foreground">
              {t("replyPlaceholder")}
            </DialogDescription>
          </DialogHeader>

          <div className="my-2 p-3 bg-muted/40 rounded border border-border text-left">
            <span className="text-[10px] font-bold text-muted-foreground uppercase">Pengulas: {activeReview?.buyerName}</span>
            <p className="text-xs text-foreground mt-1 font-normal leading-relaxed">
              &quot;{activeReview?.content}&quot;
            </p>
          </div>

          <form onSubmit={handleSaveReply} className="space-y-4 my-2 text-left">
            <FieldGroup>
              <Field>
                <FieldLabel className="text-xs font-semibold uppercase text-muted-foreground">{t("reply")}</FieldLabel>
                <Textarea
                  placeholder={t("replyPlaceholder")}
                  value={replyText}
                  onChange={(e) => setReplyText(e.target.value)}
                  className="bg-background border-border min-h-[100px]"
                />
              </Field>
            </FieldGroup>

            <DialogFooter className="pt-2">
              <Button type="button" variant="outline" onClick={() => setIsReplyOpen(false)} className="border-border cursor-pointer text-foreground">
                {t("cancel")}
              </Button>
              <Button type="submit" className="bg-primary hover:bg-primary/90 text-primary-foreground font-bold cursor-pointer hover:-translate-y-0.5 active:translate-y-0 transition-transform">
                {t("save")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
