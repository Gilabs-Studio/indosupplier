"use client";

import { useState, useEffect, useCallback, useRef, useMemo } from "react";
import { searchService } from "@/features/public/search/services/search-service";
import type {
  ProductReviewItem,
  ProductReviewSummary,
} from "@/features/public/search/types";

interface UseProductReviewsProps {
  productId: string;
}

const DEFAULT_SUMMARY: ProductReviewSummary = {
  averageRating: 0,
  totalReviews: 0,
  ratingBreakdown: { "1": 0, "2": 0, "3": 0, "4": 0, "5": 0 },
  positivePercent: 0,
};

const BATCH_SIZE = 5;

export function useProductReviews({ productId }: UseProductReviewsProps) {
  const [selectedRating, setSelectedRating] = useState<number | null>(null);
  const [summary, setSummary] = useState<ProductReviewSummary>(DEFAULT_SUMMARY);
  const [reviews, setReviews] = useState<ProductReviewItem[]>([]);
  const [currentPage, setCurrentPage] = useState(1);
  const [hasMore, setHasMore] = useState(false);
  const [totalItems, setTotalItems] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);

  // Avoid stale closures with ref
  const currentReqRef = useRef(0);

  // Initial fetch or filter change
  const fetchReviews = useCallback(
    async (ratingFilter: number | null) => {
      if (!productId) return;
      const reqId = ++currentReqRef.current;
      setIsLoading(true);

      try {
        const res = await searchService.getProductReviews(productId, {
          page: 1,
          limit: BATCH_SIZE,
          rating: ratingFilter ?? undefined,
        });

        if (reqId !== currentReqRef.current) return;

        setSummary(res.summary);
        setReviews(res.reviews);
        setCurrentPage(res.pagination.currentPage);
        setHasMore(res.pagination.hasMore);
        setTotalItems(res.pagination.totalItems);
      } catch (err) {
        console.error("Failed to load product reviews:", err);
      } finally {
        if (reqId === currentReqRef.current) {
          setIsLoading(false);
        }
      }
    },
    [productId]
  );

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    fetchReviews(selectedRating);
  }, [fetchReviews, selectedRating]);

  // Button-driven load more (strictly 5 per batch, no automatic infinite scroll)
  const loadMore = useCallback(async () => {
    if (!productId || isLoadingMore || !hasMore) return;
    setIsLoadingMore(true);

    const nextPage = currentPage + 1;
    try {
      const res = await searchService.getProductReviews(productId, {
        page: nextPage,
        limit: BATCH_SIZE,
        rating: selectedRating ?? undefined,
      });

      setReviews((prev) => {
        const existingIds = new Set(prev.map((r) => r.id));
        const newItems = res.reviews.filter((r) => !existingIds.has(r.id));
        return [...prev, ...newItems];
      });
      setCurrentPage(res.pagination.currentPage);
      setHasMore(res.pagination.hasMore);
      setTotalItems(res.pagination.totalItems);
    } catch (err) {
      console.error("Failed to load more reviews:", err);
    } finally {
      setIsLoadingMore(false);
    }
  }, [productId, isLoadingMore, hasMore, currentPage, selectedRating]);

  const [sortBy, setSortBy] = useState<"newest" | "highest" | "lowest">("newest");

  const sortedReviews = useMemo(() => {
    const list = [...reviews];
    if (sortBy === "highest") {
      return list.sort((a, b) => b.rating - a.rating);
    }
    if (sortBy === "lowest") {
      return list.sort((a, b) => a.rating - b.rating);
    }
    return list.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
  }, [reviews, sortBy]);

  const handleSelectRating = useCallback((rating: number | null) => {
    setSelectedRating((prev) => (prev === rating ? null : rating));
  }, []);

  return {
    summary,
    reviews: sortedReviews,
    selectedRating,
    setSelectedRating: handleSelectRating,
    sortBy,
    setSortBy,
    currentPage,
    hasMore,
    totalItems,
    isLoading,
    isLoadingMore,
    loadMore,
    isEmpty: summary.totalReviews === 0,
    isFilteredEmpty: reviews.length === 0 && !isLoading,
  };
}
