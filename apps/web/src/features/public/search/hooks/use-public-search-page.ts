"use client";

import React, { useMemo, useState } from "react";
import { useSearchParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { useRouter } from "@/i18n/routing";
import { useSupplierSearch } from "./use-supplier-search";
import { searchService } from "../services/search-service";
import type { PublicProductDto, PublicSupplierDto } from "../types";
import { useBuyerBookmarks } from "@/features/buyer/bookmarks/hooks/useBuyerBookmarks";
import { useBuyerCompare } from "@/features/buyer/compare/hooks/useBuyerCompare";
import { useBuyerFollowing } from "@/features/buyer/following/hooks/useBuyerFollowing";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import { toast } from "sonner";

export type SearchTab = "products" | "suppliers";

interface UsePublicSearchPageProps {
  detailBasePath: "" | "/demo";
}

export function usePublicSearchPage({ detailBasePath }: UsePublicSearchPageProps) {
  const t = useTranslations("public.search");
  const router = useRouter();
  const searchParams = useSearchParams();
  const initialQuery = searchParams.get("query") || searchParams.get("q") || "";
  const initialCategory = searchParams.get("category") || "";
  const initialRegion = searchParams.get("region") || "";
  const initialVerifiedOnly = searchParams.get("verified") === "true";

  const {
    params,
    suppliers,
    categories,
    isLoading,
    setQuery,
    setCategory,
    setRegion,
    setVerifiedOnly,
    resetFilters,
  } = useSupplierSearch({
    query: initialQuery,
    category: initialCategory,
    region: initialRegion,
    verifiedOnly: initialVerifiedOnly,
  });

  const [activeTab, setActiveTab] = useState<SearchTab>("products");
  const [searchInput, setSearchInput] = useState(initialQuery);
  const [showMobileFilters, setShowMobileFilters] = useState(false);
  const { isAuthenticated } = useAuthStore();
  const { bookmarks, addBookmark, deleteBookmark } = useBuyerBookmarks();
  const { isFollowingSupplier, followSupplier, unfollowSupplier } = useBuyerFollowing();
  
  const {
    suppliers: comparedSuppliers,
    products: comparedProducts,
    addSupplier,
    removeSupplier,
    addProduct,
    removeProduct,
  } = useBuyerCompare();

  const { data: products = [], isLoading: isProductsLoading } = useQuery({
    queryKey: ["public", "products", params.query],
    queryFn: () => searchService.searchProducts(params.query || ""),
    placeholderData: (previousData) => previousData,
  });

  const filteredProducts = useMemo(() => {
    return products.filter((product) => {
      if (params.category && product.categoryName) {
        const selected = categories.find((category) => category.id === params.category);
        if (selected && selected.name !== product.categoryName) return false;
      }
      if (params.region && !product.supplierLocation?.toLowerCase().includes(params.region.toLowerCase())) return false;
      if (params.verifiedOnly && !product.supplierVerified) return false;
      return true;
    });
  }, [categories, params.category, params.region, params.verifiedOnly, products]);

  const requireAuth = (message: string) => {
    if (!isAuthenticated) {
      toast.error(message);
      const redirectTarget = `${detailBasePath || ""}/search${searchParams.toString() ? `?${searchParams.toString()}` : ""}`;
      router.push(`/login?redirectTo=${encodeURIComponent(redirectTarget)}`);
      return false;
    }
    return true;
  };

  const toggleSupplierFollowing = (supplier: PublicSupplierDto) => {
    if (!requireAuth(t("authRequireFollow"))) return;
    if (isFollowingSupplier(supplier.id)) {
      unfollowSupplier(supplier.id);
      return;
    }
    followSupplier(supplier.id);
  };

  const toggleProductBookmark = (product: PublicProductDto) => {
    if (!requireAuth(t("authRequireBookmark"))) return;
    const bookmark = bookmarks.find((item) => item.type === "product" && item.supplierProductId === product.id);
    if (bookmark) {
      deleteBookmark(bookmark.id);
    } else {
      addBookmark({ supplierProfileId: product.supplierId, supplierProductId: product.id });
    }
  };

  const toggleSupplierCompare = (supplier: PublicSupplierDto) => {
    if (!requireAuth(t("authRequireCompareSupplier"))) return;
    if (comparedSuppliers.some((item) => item.id === supplier.id)) {
      removeSupplier(supplier.id);
    } else {
      addSupplier(supplier.id);
    }
  };

  const toggleProductCompare = (product: PublicProductDto) => {
    if (!requireAuth(t("authRequireCompareProduct"))) return;
    if (comparedProducts.some((item) => item.id === product.id)) {
      removeProduct(product.id);
    } else {
      addProduct(product.id);
    }
  };

  const isProductBookmarked = (productId: string) => bookmarks.some((item) => item.type === "product" && item.supplierProductId === productId);

  const resultCount = activeTab === "products" ? filteredProducts.length : suppliers.length;

  const handleSearchSubmit = (event: React.FormEvent) => {
    event.preventDefault();
    setQuery(searchInput.trim());
  };

  return {
    params,
    suppliers,
    categories,
    isLoading: activeTab === "products" ? isProductsLoading : isLoading,
    activeTab,
    setActiveTab,
    searchInput,
    setSearchInput,
    showMobileFilters,
    setShowMobileFilters,
    filteredProducts,
    comparedProducts,
    comparedSuppliers,
    isFollowingSupplier,
    isProductBookmarked,
    resultCount,
    handleSearchSubmit,
    toggleSupplierFollowing,
    toggleProductBookmark,
    toggleSupplierCompare,
    toggleProductCompare,
    setCategory,
    setRegion,
    setVerifiedOnly,
    resetFilters,
  };
}
