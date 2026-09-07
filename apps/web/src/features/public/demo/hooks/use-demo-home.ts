"use client";

import { useState, useMemo, useCallback } from "react";
import { useQuery } from "@tanstack/react-query";
import { demoService } from "../services/demo.service";
import type {
  DemoHeroBanner,
  DemoQuickAction,
  DemoCategoryItem,
  DemoProductItem,
  DemoProductFilterState,
} from "../types/demo.types";

const initialFilterState: DemoProductFilterState = {
  searchQuery: "",
  location: "all",
  minPrice: undefined,
  maxPrice: undefined,
  minOrder: "all",
  isPowerSupplier: false,
  isVerifiedSupplier: false,
  isReadyStock: false,
  sort: "terlaris",
  page: 1,
};

export function useDemoHome(locale: string) {
  const isEn = locale === "en";

  const [filters, setFilters] = useState<DemoProductFilterState>(initialFilterState);
  const [bookmarkedProductIds, setBookmarkedProductIds] = useState<Set<string>>(new Set());
  const [cartFeedback, setCartFeedback] = useState<{ id: string; name: string } | null>(null);

  // Queries
  const bannerQuery = useQuery<DemoHeroBanner>({
    queryKey: ["demo-hero-banner", locale],
    queryFn: () => demoService.getHeroBanner(),
    staleTime: 10 * 60 * 1000,
  });

  const quickActionsQuery = useQuery<DemoQuickAction[]>({
    queryKey: ["demo-quick-actions", locale],
    queryFn: () => demoService.getQuickActions(),
    staleTime: 10 * 60 * 1000,
  });

  const categoriesQuery = useQuery<DemoCategoryItem[]>({
    queryKey: ["demo-popular-categories", locale],
    queryFn: () => demoService.getPopularCategories(),
    staleTime: 10 * 60 * 1000,
  });

  const productsQuery = useQuery<DemoProductItem[]>({
    queryKey: ["demo-featured-products", filters],
    queryFn: () => demoService.getProducts(filters),
    staleTime: 2 * 60 * 1000,
  });

  // Filter setters
  const setLocation = useCallback((location: string) => {
    setFilters((prev) => ({ ...prev, location, page: 1 }));
  }, []);

  const setPriceRange = useCallback((minPrice: number | undefined, maxPrice: number | undefined) => {
    setFilters((prev) => ({ ...prev, minPrice, maxPrice, page: 1 }));
  }, []);

  const setMinOrder = useCallback((minOrder: string) => {
    setFilters((prev) => ({ ...prev, minOrder, page: 1 }));
  }, []);

  const togglePowerSupplier = useCallback(() => {
    setFilters((prev) => ({ ...prev, isPowerSupplier: !prev.isPowerSupplier, page: 1 }));
  }, []);

  const toggleVerifiedSupplier = useCallback(() => {
    setFilters((prev) => ({ ...prev, isVerifiedSupplier: !prev.isVerifiedSupplier, page: 1 }));
  }, []);

  const toggleReadyStock = useCallback(() => {
    setFilters((prev) => ({ ...prev, isReadyStock: !prev.isReadyStock, page: 1 }));
  }, []);

  const setSort = useCallback((sort: DemoProductFilterState["sort"]) => {
    setFilters((prev) => ({ ...prev, sort, page: 1 }));
  }, []);

  const resetFilters = useCallback(() => {
    setFilters(initialFilterState);
  }, []);

  // Actions
  const toggleBookmark = useCallback((productId: string) => {
    setBookmarkedProductIds((prev) => {
      const next = new Set(prev);
      if (next.has(productId)) {
        next.delete(productId);
      } else {
        next.add(productId);
      }
      return next;
    });
  }, []);

  const handleAddToCart = useCallback((product: DemoProductItem) => {
    setCartFeedback({ id: product.id, name: product.name });
    setTimeout(() => {
      setCartFeedback((current) => (current?.id === product.id ? null : current));
    }, 2500);
  }, []);

  // Filtered/Sorted list fallback client-side if API is static
  const displayedProducts = useMemo(() => {
    let list = [...(productsQuery.data || [])];

    if (filters.location && filters.location !== "all" && filters.location !== "Semua Lokasi") {
      list = list.filter((p) =>
        p.supplierLocation.toLowerCase().includes(filters.location.toLowerCase())
      );
    }
    if (filters.minPrice !== undefined && filters.minPrice > 0) {
      list = list.filter((p) => p.price >= (filters.minPrice ?? 0));
    }
    if (filters.maxPrice !== undefined && filters.maxPrice > 0) {
      list = list.filter((p) => p.price <= (filters.maxPrice ?? Infinity));
    }
    if (filters.isPowerSupplier) {
      list = list.filter((p) => p.isPowerSupplier);
    }
    if (filters.isVerifiedSupplier) {
      list = list.filter((p) => p.supplierVerified);
    }
    if (filters.isReadyStock) {
      list = list.filter((p) => p.tags?.some((t) => t.toLowerCase().includes("ready stock")));
    }

    if (filters.sort === "price_asc") {
      list.sort((a, b) => a.price - b.price);
    } else if (filters.sort === "price_desc") {
      list.sort((a, b) => b.price - a.price);
    } else if (filters.sort === "rating") {
      list.sort((a, b) => b.supplierRating - a.supplierRating);
    }

    return list;
  }, [productsQuery.data, filters]);

  return {
    isEn,
    filters,
    banner: bannerQuery.data,
    isBannerLoading: bannerQuery.isLoading,
    quickActions: quickActionsQuery.data || [],
    isQuickActionsLoading: quickActionsQuery.isLoading,
    popularCategories: categoriesQuery.data || [],
    isCategoriesLoading: categoriesQuery.isLoading,
    products: displayedProducts,
    isProductsLoading: productsQuery.isLoading,
    bookmarkedProductIds,
    cartFeedback,
    setLocation,
    setPriceRange,
    setMinOrder,
    togglePowerSupplier,
    toggleVerifiedSupplier,
    toggleReadyStock,
    setSort,
    resetFilters,
    toggleBookmark,
    handleAddToCart,
  };
}
