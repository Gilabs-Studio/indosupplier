"use client";

import React from "react";
import { useTranslations, useLocale } from "next-intl";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  RefreshCw,
  MessageSquareOff,
  ChevronLeft,
  ChevronRight,
  Sparkles,
} from "lucide-react";
import { useSupplierReviews } from "../hooks/useSupplierReviews";
import { ReviewStatsCards } from "./review-stats-cards";
import { ReviewFilters } from "./review-filters";
import { ReviewCard } from "./review-card";
import { ReviewReplyDialog } from "./review-reply-dialog";
import { ReviewsSkeleton } from "./reviews-skeleton";

export function SupplierReviewsList() {
  const t = useTranslations("supplier.reviews");
  const locale = useLocale();

  const {
    reviews,
    stats,
    pagination,
    isLoading,
    isFetching,
    refetch,
    filter,
    setFilter,
    resetFilter,
    activeReview,
    isReplyDialogOpen,
    openReplyDialog,
    closeReplyDialog,
    handleReplySubmit,
    isReplying,
  } = useSupplierReviews();

  const currentPage = pagination?.current_page ?? 1;
  const totalPages = pagination?.total_pages ?? 1;

  return (
    <div className="space-y-6 text-left pb-10">
      {/* 1. Header Section */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border/80 pb-5">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
              {t("title")}
            </h1>
            <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-bold bg-primary/10 text-primary border border-primary/20">
              <Sparkles className="h-3 w-3" />
              Live API
            </span>
          </div>
          <p className="text-xs sm:text-sm text-muted-foreground">
            {t("subtitle")}
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => refetch()}
            disabled={isFetching}
            className="text-xs h-9 cursor-pointer border-border hover:bg-muted/50 flex items-center gap-1.5"
          >
            <RefreshCw
              className={`h-3.5 w-3.5 ${isFetching ? "animate-spin text-primary" : ""}`}
            />
            <span>Refresh</span>
          </Button>
        </div>
      </div>

      {/* 2. Overview Stats Cards */}
      <ReviewStatsCards
        stats={stats}
        currentFilter={filter}
        onFilterChange={setFilter}
      />

      {/* 3. Search & Filter Bar */}
      <ReviewFilters
        filter={filter}
        stats={stats}
        onFilterChange={setFilter}
        onReset={resetFilter}
      />

      {/* 4. Reviews List */}
      <div className="space-y-4">
        {isLoading ? (
          <ReviewsSkeleton />
        ) : reviews.length === 0 ? (
          <Card className="border border-dashed border-border bg-card rounded-xl">
            <CardContent className="py-12 flex flex-col items-center justify-center text-center space-y-3">
              <div className="h-12 w-12 rounded-xl bg-muted/60 text-muted-foreground flex items-center justify-center border border-border">
                <MessageSquareOff className="h-6 w-6" />
              </div>
              <div className="space-y-1 max-w-sm">
                <h3 className="text-sm font-bold text-foreground">
                  {t("noReviewsFound")}
                </h3>
                <p className="text-xs text-muted-foreground">
                  {t("noReviewsFoundDesc")}
                </p>
              </div>
              {(filter.search || filter.rating !== undefined || filter.status !== "all") && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={resetFilter}
                  className="text-xs h-8 cursor-pointer mt-2"
                >
                  {t("clearFilterBtn")}
                </Button>
              )}
            </CardContent>
          </Card>
        ) : (
          reviews.map((rev) => (
            <ReviewCard
              key={rev.id}
              review={rev}
              locale={locale}
              onOpenReply={openReplyDialog}
            />
          ))
        )}
      </div>

      {/* 5. Pagination Controls */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between border-t border-border pt-4">
          <p className="text-xs text-muted-foreground">
            {t("pageOf", { current: currentPage, total: totalPages })}
          </p>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={currentPage <= 1 || isFetching}
              onClick={() => setFilter({ page: currentPage - 1 })}
              className="text-xs h-8 cursor-pointer flex items-center gap-1 border-border"
            >
              <ChevronLeft className="h-3.5 w-3.5" />
              <span>{t("previous")}</span>
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={currentPage >= totalPages || isFetching}
              onClick={() => setFilter({ page: currentPage + 1 })}
              className="text-xs h-8 cursor-pointer flex items-center gap-1 border-border"
            >
              <span>{t("next")}</span>
              <ChevronRight className="h-3.5 w-3.5" />
            </Button>
          </div>
        </div>
      )}

      {/* 6. Reply Modal Dialog */}
      <ReviewReplyDialog
        isOpen={isReplyDialogOpen}
        review={activeReview}
        onClose={closeReplyDialog}
        onSubmit={handleReplySubmit}
        isSubmitting={isReplying}
      />
    </div>
  );
}
