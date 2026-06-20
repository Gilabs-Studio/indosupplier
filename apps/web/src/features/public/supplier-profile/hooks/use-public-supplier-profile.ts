"use client";

import React, { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { useRouter } from "@/i18n/routing";
import { toast } from "sonner";
import { searchService } from "@/features/public/search/services/search-service";
import { useBuyerBookmarks } from "@/features/buyer/bookmarks/hooks/useBuyerBookmarks";
import { useBuyerFollowing } from "@/features/buyer/following/hooks/useBuyerFollowing";
import { useBuyerCompare } from "@/features/buyer/compare/hooks/useBuyerCompare";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import { chatService } from "@/features/buyer/chat/services/chat.service";
import type { SupplierProductDto, PublicProductDto } from "@/features/public/search/types";

interface UsePublicSupplierProfileProps {
  slug: string;
  detailBasePath: "" | "/demo";
}

export type ProfileTab = "home" | "products" | "certifications" | "reviews";

export function usePublicSupplierProfile({ slug, detailBasePath }: UsePublicSupplierProfileProps) {
  const t = useTranslations("public.supplier");
  const router = useRouter();
  const { isAuthenticated } = useAuthStore();
  const { bookmarks, addBookmark, deleteBookmark } = useBuyerBookmarks();
  const { isFollowingSupplier, followSupplier, unfollowSupplier, isMutating: isMutatingFollowing } = useBuyerFollowing();
  const { products: comparedProducts, addProduct, removeProduct } = useBuyerCompare();
  
  const [isOpeningChat, setIsOpeningChat] = useState(false);
  const [activeTab, setActiveTab] = useState<ProfileTab>("home");

  // Fetch Supplier profile by slug
  const { data: supplier, isLoading, isError } = useQuery({
    queryKey: ["public-supplier-profile", slug],
    queryFn: () => searchService.getSupplierBySlug(slug),
  });

  // State for products infinite loading
  const [productsList, setProductsList] = useState<PublicProductDto[]>([]);
  const [productsPage, setProductsPage] = useState(1);
  const [hasMoreProducts, setHasMoreProducts] = useState(true);
  const [isProductsLoading, setIsProductsLoading] = useState(false);
  const [isProductsLoadingMore, setIsProductsLoadingMore] = useState(false);

  // RFQ form state
  const [subject, setSubject] = useState("");
  const [message, setMessage] = useState("");
  const [quantity, setQuantity] = useState("1");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isRfqOpen, setIsRfqOpen] = useState(false);

  const fetchProductsList = async (pageNumber: number, isLoadMore = false) => {
    if (!supplier) return;
    if (isLoadMore) {
      setIsProductsLoadingMore(true);
    } else {
      setIsProductsLoading(true);
    }

    try {
      const res = await searchService.searchProducts("", supplier.id, pageNumber, 12);
      if (isLoadMore) {
        setProductsList((prev) => [...prev, ...res]);
      } else {
        setProductsList(res);
      }
      setProductsPage(pageNumber);
      setHasMoreProducts(res.length === 12);
    } catch (error) {
      console.error("Error loading products:", error);
      toast.error(t("failLoadProducts"));
    } finally {
      setIsProductsLoading(false);
      setIsProductsLoadingMore(false);
    }
  };

  const handleTabChange = (val: string) => {
    const tab = val as ProfileTab;
    setActiveTab(tab);
    if (tab === "products" && productsList.length === 0) {
      fetchProductsList(1, false);
    }
  };

  const handleLoadMoreProducts = () => {
    if (isProductsLoading || isProductsLoadingMore || !hasMoreProducts) return;
    fetchProductsList(productsPage + 1, true);
  };

  const mapSupplierProductToPublicProduct = (prod: SupplierProductDto) => {
    if (!supplier) {
      throw new Error("Supplier data is required to map products.");
    }
    return {
      id: prod.id,
      name: prod.name,
      description: prod.description || "",
      price: prod.price || 0,
      currency: prod.currency || "IDR",
      minOrder: prod.minOrder || "",
      capacityText: prod.capacityText || "",
      categoryName: prod.categoryName || "",
      photos: prod.photos || [],
      supplierId: supplier.id,
      supplierCompanyName: supplier.companyName,
      supplierSlug: supplier.slug,
      supplierLocation: supplier.location,
      supplierVerified: supplier.isVerified,
      supplierRating: supplier.rating,
      supplierReviewCount: supplier.reviewCount,
    };
  };

  const currentPath = `${detailBasePath}/suppliers/${slug}`;

  const requireAuth = (msg: string) => {
    if (isAuthenticated) return true;
    toast.error(msg);
    router.push(`/login?redirectTo=${encodeURIComponent(currentPath)}`);
    return false;
  };

  const getProductBookmarkId = (prodId: string) => {
    const found = bookmarks.find(
      (b) => b.type === "product" && b.supplierProductId === prodId
    );
    return found?.id || null;
  };

  const handleToggleProductBookmark = (prodId: string) => {
    if (!requireAuth(t("authRequireBookmark"))) return;
    const bookmarkId = getProductBookmarkId(prodId);
    if (bookmarkId) {
      deleteBookmark(bookmarkId);
    } else {
      if (supplier) {
        addBookmark({ supplierProfileId: supplier.id, supplierProductId: prodId });
      }
    }
  };

  const handleToggleProductCompare = (prodId: string) => {
    if (!requireAuth(t("authRequireCompareProduct"))) return;
    if (comparedProducts.some((p) => p.id === prodId)) {
      removeProduct(prodId);
    } else {
      addProduct(prodId);
    }
  };

  const handleToggleFollow = () => {
    if (!supplier) {
      requireAuth(t("authRequireFollow"));
      return;
    }
    if (!requireAuth(t("authRequireFollow"))) return;
    if (isFollowingSupplier(supplier.id)) {
      unfollowSupplier(supplier.id);
      return;
    }
    followSupplier(supplier.id);
  };

  const handleOpenChat = async () => {
    if (!supplier) return;
    if (!requireAuth(t("authRequireChat"))) return;

    try {
      setIsOpeningChat(true);
      const room = await chatService.getOrCreateRoom(supplier.id);
      router.push(`/chat?roomId=${encodeURIComponent(room.id)}`);
    } catch (error) {
      console.error(error);
      toast.error(t("failOpenChat"));
    } finally {
      setIsOpeningChat(false);
    }
  };

  const handleOpenRfq = () => {
    if (!requireAuth(t("authRequireRfq"))) return;
    setIsRfqOpen(true);
  };

  const handleSendRFQ = (e: React.FormEvent) => {
    e.preventDefault();
    if (!subject.trim() || !message.trim()) {
      toast.error("Please fill in all required fields.");
      return;
    }

    setIsSubmitting(true);
    setTimeout(() => {
      toast.success(t("quoteSuccess"));
      setSubject("");
      setMessage("");
      setQuantity("1");
      setIsSubmitting(false);
      setIsRfqOpen(false);
    }, 1000);
  };

  return {
    supplier,
    isLoading,
    isError,
    activeTab,
    productsList,
    isProductsLoading,
    isProductsLoadingMore,
    hasMoreProducts,
    isMutatingFollowing,
    isOpeningChat,
    isFollowed: supplier ? isFollowingSupplier(supplier.id) : false,
    comparedProducts,
    bookmarks,
    subject,
    setSubject,
    message,
    setMessage,
    quantity,
    setQuantity,
    isSubmitting,
    isRfqOpen,
    setIsRfqOpen,
    handleTabChange,
    handleLoadMoreProducts,
    handleToggleProductBookmark,
    handleToggleProductCompare,
    handleToggleFollow,
    handleOpenChat,
    handleOpenRfq,
    handleSendRFQ,
    getProductBookmarkId,
    mapSupplierProductToPublicProduct,
  };
}
