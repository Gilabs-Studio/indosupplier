"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Search, Star, X, RotateCcw } from "lucide-react";
import type { SupplierReviewsFilter, SupplierReviewStats } from "../types/reviews.types";

interface ReviewFiltersProps {
  filter: SupplierReviewsFilter;
  stats?: SupplierReviewStats;
  onFilterChange: (partial: Partial<SupplierReviewsFilter>) => void;
  onReset: () => void;
}

export function ReviewFilters({
  filter,
  stats,
  onFilterChange,
  onReset,
}: ReviewFiltersProps) {
  const t = useTranslations("supplier.reviews");
  const [searchTerm, setSearchTerm] = useState(filter.search || "");

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(e.target.value);
  };

  const handleSearchKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      onFilterChange({ search: searchTerm.trim() });
    }
  };

  const handleSearchBlur = () => {
    if (searchTerm.trim() !== (filter.search || "")) {
      onFilterChange({ search: searchTerm.trim() });
    }
  };

  const handleClear = () => {
    setSearchTerm("");
    onFilterChange({ search: "" });
  };

  const handleFullReset = () => {
    setSearchTerm("");
    onReset();
  };

  const hasActiveFilters =
    Boolean(filter.search) ||
    filter.status !== "all" ||
    filter.rating !== undefined;

  const statusOptions: Array<{ id: "all" | "unreplied" | "replied"; label: string; count?: number }> = [
    { id: "all", label: t("allStatus"), count: stats?.total_reviews },
    { id: "unreplied", label: t("needReply"), count: stats?.unreplied_count },
    { id: "replied", label: t("replied"), count: stats?.replied_count },
  ];

  const ratingOptions = [undefined, 5, 4, 3, 2, 1];

  return (
    <div className="space-y-4 bg-card border border-border p-4 rounded-xl shadow-xs">
      <div className="flex flex-col md:flex-row gap-3 items-stretch md:items-center justify-between">
        {/* Search input */}
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
          <Input
            value={searchTerm}
            onChange={handleSearchChange}
            onKeyDown={handleSearchKeyDown}
            onBlur={handleSearchBlur}
            placeholder={t("searchPlaceholder")}
            className="pl-9 pr-8 h-9 text-xs rounded-lg bg-background border-border"
          />
          {searchTerm && (
            <button
              type="button"
              onClick={handleClear}
              className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground cursor-pointer p-0.5"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          )}
        </div>

        {/* Reset button */}
        {hasActiveFilters && (
          <Button
            variant="ghost"
            size="sm"
            onClick={handleFullReset}
            className="h-9 px-3 text-xs text-muted-foreground hover:text-foreground cursor-pointer flex items-center gap-1.5 self-start md:self-auto"
          >
            <RotateCcw className="h-3.5 w-3.5" />
            {t("resetFilters")}
          </Button>
        )}
      </div>

      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-t border-border pt-3">
        {/* Status Pills */}
        <div className="flex items-center gap-1.5 flex-wrap">
          {statusOptions.map((opt) => {
            const isActive = filter.status === opt.id || (!filter.status && opt.id === "all");
            return (
              <button
                key={opt.id}
                onClick={() => onFilterChange({ status: opt.id })}
                className={`inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium cursor-pointer transition-all ${
                  isActive
                    ? "bg-primary text-primary-foreground shadow-xs font-semibold"
                    : "bg-muted/50 text-muted-foreground hover:bg-muted hover:text-foreground"
                }`}
              >
                <span>{opt.label}</span>
                {opt.count !== undefined && (
                  <span
                    className={`text-[10px] px-1.5 py-0.2 rounded-full font-bold ${
                      isActive
                        ? "bg-primary-foreground/20 text-primary-foreground"
                        : "bg-background text-muted-foreground border border-border"
                    }`}
                  >
                    {opt.count}
                  </span>
                )}
              </button>
            );
          })}
        </div>

        {/* Rating filter pills */}
        <div className="flex items-center gap-1 flex-wrap">
          <span className="text-[11px] font-medium text-muted-foreground mr-1">
            {t("filterRating")}:
          </span>
          {ratingOptions.map((r) => {
            const isActive = filter.rating === r;
            const count = r ? stats?.rating_breakdown?.[r] : undefined;
            return (
              <button
                key={r ?? "all-ratings"}
                onClick={() => onFilterChange({ rating: r })}
                className={`inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs font-medium cursor-pointer transition-all ${
                  isActive
                    ? "bg-amber-500/15 border border-amber-500/40 text-amber-600 dark:text-amber-400 font-semibold"
                    : "bg-background border border-border text-muted-foreground hover:bg-muted/50 hover:text-foreground"
                }`}
              >
                {r ? (
                  <>
                    <span>{r}</span>
                    <Star className="h-3 w-3 fill-amber-400 text-amber-400" />
                    {count !== undefined && (
                      <span className="text-[10px] text-muted-foreground ml-0.5">
                        ({count})
                      </span>
                    )}
                  </>
                ) : (
                  <span>{t("allRatings")}</span>
                )}
              </button>
            );
          })}
        </div>
      </div>
    </div>
  );
}
