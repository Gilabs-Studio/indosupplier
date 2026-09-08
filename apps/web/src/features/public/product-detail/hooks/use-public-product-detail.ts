"use client";

import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { useRouter } from "@/i18n/routing";
import { searchService } from "@/features/public/search/services/search-service";
import type { PublicProductDto, PublicReviewDto } from "@/features/public/search/types";
import { useBuyerBookmarks } from "@/features/buyer/bookmarks/hooks/useBuyerBookmarks";
import { useBuyerCompare } from "@/features/buyer/compare/hooks/useBuyerCompare";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import { toast } from "sonner";

interface UsePublicProductDetailProps {
  id: string;
  detailBasePath: "" | "/demo";
}

function getVariants(minOrder: string, capacityText: string) {
  const base = [
    minOrder || "MOQ Nego",
    capacityText || "Kapasitas Nego",
    "Sampel",
    "Kontrak Bulanan",
  ];
  return Array.from(new Set(base.filter(Boolean))).slice(0, 6);
}

function ratingDistribution(reviews: PublicReviewDto[]) {
  const distribution = [5, 4, 3, 2, 1].map((rating) => ({
    rating,
    count: reviews.filter((review) => review.rating === rating).length,
  }));
  const total = reviews.length || 1;
  return distribution.map((item) => ({
    ...item,
    percent: Math.round((item.count / total) * 100),
  }));
}

export function usePublicProductDetail({ id, detailBasePath }: UsePublicProductDetailProps) {
  const t = useTranslations("public.productDetail");
  const router = useRouter();
  const [activePhoto, setActivePhoto] = useState(0);
  const [quantity, setQuantity] = useState(1);
  const [reviewFilter, setReviewFilter] = useState<"all" | "media" | "high">("all");
  const { isAuthenticated } = useAuthStore();
  const { bookmarks, addBookmark, deleteBookmark, isAdding, isDeleting } = useBuyerBookmarks();
  const { products: comparedProducts, addProduct, removeProduct, isAddingProduct, isRemovingProduct } = useBuyerCompare();

  const { data, isLoading } = useQuery({
    queryKey: ["public-product-detail", id],
    queryFn: () => searchService.getProductById(id),
    staleTime: 60_000,
  });

  const product = data?.product;
  const supplier = data?.supplier;
  const photos = product?.photos?.length ? product.photos : [];
  const variants = useMemo(
    () => getVariants(product?.minOrder ?? "", product?.capacityText ?? ""),
    [product?.minOrder, product?.capacityText],
  );
  const [selectedVariant, setSelectedVariant] = useState(0);
  const reviews = data?.reviews ?? [];
  
  const visibleReviews = reviews.filter((review) => {
    if (reviewFilter === "high") return review.rating >= 5;
    return true;
  });
  
  const averageRating = reviews.length
    ? reviews.reduce((sum, review) => sum + review.rating, 0) / reviews.length
    : supplier?.rating ?? 0;
  
  const distribution = ratingDistribution(reviews);
  const subtotal = product ? product.price * quantity : 0;
  const currentInternalPath = `${detailBasePath}/products/${id}`;
  const isBookmarked = product ? bookmarks.some((item) => item.type === "product" && item.supplierProductId === product.id) : false;
  const isCompared = product ? comparedProducts.some((item) => item.id === product.id) : false;

  const isProductBookmarked = (prodId: string) =>
    bookmarks.some((item) => item.type === "product" && (item.supplierProductId === prodId || item.id === prodId));

  const isProductCompared = (prodId: string) =>
    comparedProducts.some((item) => item.id === prodId);

  const redirectToLogin = (message: string) => {
    toast.error(message);
    router.push(`/login?redirectTo=${encodeURIComponent(currentInternalPath)}`);
  };

  const requireAuth = (message: string) => {
    if (!isAuthenticated) {
      redirectToLogin(message);
      return false;
    }
    return true;
  };

  const toggleBookmark = () => {
    if (!product) return;
    if (!requireAuth(t("authRequireBookmark"))) return;
    const bookmark = bookmarks.find((item) => item.type === "product" && item.supplierProductId === product.id);
    if (bookmark) {
      deleteBookmark(bookmark.id);
    } else {
      addBookmark({ supplierProfileId: product.supplierId, supplierProductId: product.id });
    }
  };

  const toggleCompare = () => {
    if (!product) return;
    if (!requireAuth(t("authRequireCompare"))) return;
    if (isCompared) {
      removeProduct(product.id);
    } else {
      addProduct(product.id);
    }
  };

  const toggleProductBookmark = (prod: PublicProductDto) => {
    if (!requireAuth(t("authRequireBookmark"))) return;
    const bookmark = bookmarks.find((item) => item.type === "product" && item.supplierProductId === prod.id);
    if (bookmark) {
      deleteBookmark(bookmark.id);
    } else {
      addBookmark({ supplierProfileId: prod.supplierId, supplierProductId: prod.id });
    }
  };

  const toggleProductCompare = (prod: PublicProductDto) => {
    if (!requireAuth(t("authRequireCompare"))) return;
    if (comparedProducts.some((item) => item.id === prod.id)) {
      removeProduct(prod.id);
    } else {
      addProduct(prod.id);
    }
  };

  const handleBuyerAction = (message: string) => {
    if (!requireAuth(message)) return;
    toast.success(t("actionReady"));
  };

  return {
    product,
    supplier,
    photos,
    variants,
    selectedVariant,
    setSelectedVariant,
    activePhoto,
    setActivePhoto,
    quantity,
    setQuantity,
    reviewFilter,
    setReviewFilter,
    visibleReviews,
    reviews,
    averageRating,
    distribution,
    subtotal,
    isBookmarked,
    isCompared,
    isProductBookmarked,
    isProductCompared,
    isLoading,
    relatedProducts: data?.relatedProducts ?? [],
    isAdding: isAdding || isDeleting,
    isComparing: isAddingProduct || isRemovingProduct,
    toggleBookmark,
    toggleCompare,
    toggleProductBookmark,
    toggleProductCompare,
    handleBuyerAction,
  };
}
