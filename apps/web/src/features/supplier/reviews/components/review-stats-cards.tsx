"use client";

import React from "react";
import { useTranslations } from "next-intl";
import { Card, CardContent } from "@/components/ui/card";
import { MessageSquare, Star, Clock, CheckCircle2 } from "lucide-react";
import type { SupplierReviewStats, SupplierReviewsFilter } from "../types/reviews.types";

interface ReviewStatsCardsProps {
  stats?: SupplierReviewStats;
  currentFilter: SupplierReviewsFilter;
  onFilterChange: (filter: Partial<SupplierReviewsFilter>) => void;
}

export function ReviewStatsCards({
  stats,
  currentFilter,
  onFilterChange,
}: ReviewStatsCardsProps) {
  const t = useTranslations("supplier.reviews");

  const totalReviews = stats?.total_reviews ?? 0;
  const avgRating = stats?.average_rating ? stats.average_rating.toFixed(1) : "0.0";
  const unrepliedCount = stats?.unreplied_count ?? 0;
  const repliedCount = stats?.replied_count ?? 0;

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      {/* 1. Total Reviews Card */}
      <Card
        onClick={() => onFilterChange({ status: "all", rating: undefined })}
        className={`border transition-all duration-300 rounded-xl cursor-pointer hover:-translate-y-0.5 active:translate-y-0 ${
          currentFilter.status === "all" && currentFilter.rating === undefined
            ? "border-primary shadow-sm bg-primary/5"
            : "border-border hover:border-border/80 bg-card hover:shadow-sm"
        }`}
      >
        <CardContent className="p-5 flex items-center justify-between">
          <div className="space-y-1">
            <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              {t("totalReviews")}
            </p>
            <p className="text-2xl font-bold tracking-tight text-foreground font-heading">
              {totalReviews}
            </p>
          </div>
          <div className="h-10 w-10 rounded-xl bg-primary/10 text-primary flex items-center justify-center border border-primary/20 shrink-0">
            <MessageSquare className="h-5 w-5" />
          </div>
        </CardContent>
      </Card>

      {/* 2. Average Rating Card */}
      <Card className="border border-border rounded-xl bg-card">
        <CardContent className="p-5 flex items-center justify-between">
          <div className="space-y-1">
            <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              {t("avgRating")}
            </p>
            <div className="flex items-center gap-2">
              <span className="text-2xl font-bold tracking-tight text-foreground font-heading">
                {avgRating}
              </span>
              <div className="flex items-center gap-0.5">
                {[1, 2, 3, 4, 5].map((star) => (
                  <Star
                    key={star}
                    className={`h-3.5 w-3.5 ${
                      Number(avgRating) >= star
                        ? "fill-amber-400 text-amber-400"
                        : Number(avgRating) >= star - 0.5
                        ? "fill-amber-400/50 text-amber-400"
                        : "text-muted-foreground/30"
                    }`}
                  />
                ))}
              </div>
            </div>
          </div>
          <div className="h-10 w-10 rounded-xl bg-amber-500/10 text-amber-500 flex items-center justify-center border border-amber-500/20 shrink-0">
            <Star className="h-5 w-5 fill-amber-500" />
          </div>
        </CardContent>
      </Card>

      {/* 3. Needs Reply Card */}
      <Card
        onClick={() => onFilterChange({ status: "unreplied" })}
        className={`border transition-all duration-300 rounded-xl cursor-pointer hover:-translate-y-0.5 active:translate-y-0 ${
          currentFilter.status === "unreplied"
            ? "border-amber-500 shadow-sm bg-amber-500/5"
            : "border-border hover:border-border/80 bg-card hover:shadow-sm"
        }`}
      >
        <CardContent className="p-5 flex items-center justify-between">
          <div className="space-y-1">
            <div className="flex items-center gap-1.5">
              <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                {t("unrepliedReviews")}
              </p>
              {unrepliedCount > 0 && (
                <span className="relative flex h-2 w-2">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-amber-400 opacity-75"></span>
                  <span className="relative inline-flex rounded-full h-2 w-2 bg-amber-500"></span>
                </span>
              )}
            </div>
            <p className="text-2xl font-bold tracking-tight text-amber-600 dark:text-amber-400 font-heading">
              {unrepliedCount}
            </p>
          </div>
          <div className="h-10 w-10 rounded-xl bg-amber-500/10 text-amber-600 dark:text-amber-400 flex items-center justify-center border border-amber-500/20 shrink-0">
            <Clock className="h-5 w-5" />
          </div>
        </CardContent>
      </Card>

      {/* 4. Replied Reviews Card */}
      <Card
        onClick={() => onFilterChange({ status: "replied" })}
        className={`border transition-all duration-300 rounded-xl cursor-pointer hover:-translate-y-0.5 active:translate-y-0 ${
          currentFilter.status === "replied"
            ? "border-emerald-500 shadow-sm bg-emerald-500/5"
            : "border-border hover:border-border/80 bg-card hover:shadow-sm"
        }`}
      >
        <CardContent className="p-5 flex items-center justify-between">
          <div className="space-y-1">
            <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              {t("repliedReviews")}
            </p>
            <p className="text-2xl font-bold tracking-tight text-emerald-600 dark:text-emerald-400 font-heading">
              {repliedCount}
            </p>
          </div>
          <div className="h-10 w-10 rounded-xl bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 flex items-center justify-center border border-emerald-500/20 shrink-0">
            <CheckCircle2 className="h-5 w-5" />
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
