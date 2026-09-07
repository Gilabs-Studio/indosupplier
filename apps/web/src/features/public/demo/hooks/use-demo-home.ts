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
import { useBuyerBookmarks } from "@/features/buyer/bookmarks/hooks/useBuyerBookmarks";

export const initialFilterState: DemoProductFilterState = {
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

/**
 * Hook for Section 1: Hero B2B Banner data
 */
export function useDemoHeroBanner(locale: string) {
  const isEn = locale === "en";

  const bannerQuery = useQuery<DemoHeroBanner>({
    queryKey: ["demo-hero-banner", locale],
    queryFn: () => demoService.getHeroBanner(),
    staleTime: 10 * 60 * 1000,
  });

  return {
    isEn,
    banner: bannerQuery.data,
    isLoading: bannerQuery.isLoading,
  };
}

/**
 * Hook for Section 2: Quick Actions Bar data
 */
export function useDemoQuickActions(locale: string) {
  const isEn = locale === "en";

  const quickActionsQuery = useQuery<DemoQuickAction[]>({
    queryKey: ["demo-quick-actions", locale],
    queryFn: () => demoService.getQuickActions(),
    staleTime: 10 * 60 * 1000,
  });

  return {
    isEn,
    quickActions: quickActionsQuery.data || [],
    isLoading: quickActionsQuery.isLoading,
  };
}

/**
 * Hook for Section 3: Popular Categories data
 */
export function useDemoPopularCategories(locale: string) {
  const isEn = locale === "en";

  const categoriesQuery = useQuery<DemoCategoryItem[]>({
    queryKey: ["demo-popular-categories", locale],
    queryFn: () => demoService.getPopularCategories(),
    staleTime: 10 * 60 * 1000,
  });

  return {
    isEn,
    popularCategories: categoriesQuery.data || [],
    isLoading: categoriesQuery.isLoading,
  };
}

/**
 * Hook for Section 4: Popular / Featured Products data and filtering
 */
export function useDemoFeaturedProducts(
  locale: string,
  initialFilters?: Partial<DemoProductFilterState>
) {
  const isEn = locale === "en";

  const [filters, setFilters] = useState<DemoProductFilterState>({
    ...initialFilterState,
    ...initialFilters,
  });
  const { bookmarks, toggleBookmarkOptimistic } = useBuyerBookmarks();
  const [cartFeedback, setCartFeedback] = useState<{ id: string; name: string } | null>(null);

  const productsQuery = useQuery<DemoProductItem[]>({
    queryKey: ["demo-featured-products", filters],
    queryFn: () => demoService.getProducts(filters),
    staleTime: 2 * 60 * 1000,
  });

  const bookmarkedProductIds = useMemo(() => {
    const ids = new Set<string>();
    for (const b of bookmarks) {
      if (b.supplierProductId) ids.add(b.supplierProductId);
      if (b.id) ids.add(b.id);
    }
    return ids;
  }, [bookmarks]);

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

  // Actions
  const toggleBookmark = useCallback((productId: string) => {
    const product = displayedProducts.find((p) => p.id === productId);
    if (product) {
      toggleBookmarkOptimistic({
        supplierProfileId: product.supplierId,
        supplierProductId: product.id,
        supplierSlug: product.supplierSlug,
        companyName: product.supplierCompanyName,
        productName: product.name,
        productPrice: product.price,
        productMinOrder: product.minOrder,
        productImage: product.photos?.[0] || "",
        category: product.categoryName || "General",
        location: product.supplierLocation || "Indonesia",
        rating: product.supplierRating,
        reviewCount: product.supplierReviewCount,
        isVerified: product.supplierVerified,
      });
    } else {
      toggleBookmarkOptimistic({
        supplierProfileId: productId,
        supplierProductId: productId,
      });
    }
  }, [displayedProducts, toggleBookmarkOptimistic]);

  const handleAddToCart = useCallback((product: DemoProductItem) => {
    setCartFeedback({ id: product.id, name: product.name });
    setTimeout(() => {
      setCartFeedback((current) => (current?.id === product.id ? null : current));
    }, 2500);
  }, []);

  return {
    isEn,
    filters,
    products: displayedProducts,
    isLoading: productsQuery.isLoading,
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

/**
 * Aggregator hook for backward compatibility
 */
export function useDemoHome(locale: string) {
  const hero = useDemoHeroBanner(locale);
  const actions = useDemoQuickActions(locale);
  const categories = useDemoPopularCategories(locale);
  const productsHook = useDemoFeaturedProducts(locale);

  return {
    isEn: hero.isEn,
    filters: productsHook.filters,
    banner: hero.banner,
    isBannerLoading: hero.isLoading,
    quickActions: actions.quickActions,
    isQuickActionsLoading: actions.isLoading,
    popularCategories: categories.popularCategories,
    isCategoriesLoading: categories.isLoading,
    products: productsHook.products,
    isProductsLoading: productsHook.isLoading,
    bookmarkedProductIds: productsHook.bookmarkedProductIds,
    cartFeedback: productsHook.cartFeedback,
    setLocation: productsHook.setLocation,
    setPriceRange: productsHook.setPriceRange,
    setMinOrder: productsHook.setMinOrder,
    togglePowerSupplier: productsHook.togglePowerSupplier,
    toggleVerifiedSupplier: productsHook.toggleVerifiedSupplier,
    toggleReadyStock: productsHook.toggleReadyStock,
    setSort: productsHook.setSort,
    resetFilters: productsHook.resetFilters,
    toggleBookmark: productsHook.toggleBookmark,
    handleAddToCart: productsHook.handleAddToCart,
  };
}
