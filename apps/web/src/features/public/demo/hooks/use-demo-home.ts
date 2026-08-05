"use client";

import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { contentService } from "@/features/content/services/content.service";
import type { ContentArticle, ContentType } from "@/features/content/types/content.types";
import { searchService } from "@/features/public/search/services/search-service";
import type { PublicProductDto, PublicSupplierDto } from "@/features/public/search/types";

export interface DemoComparisonCard {
  id: string;
  title: string;
  first: PublicSupplierDto;
  second: PublicSupplierDto;
}

export interface DemoCategoryPill {
  id: string;
  labelId: string;
  labelEn: string;
  badge?: string;
  isPopular?: boolean;
}

export interface EnhancedDemoProduct extends PublicProductDto {
  discountPercentage?: number;
  originalPrice?: number;
  promoBadge?: string;
  salesCountText?: string;
  locationTag?: string;
  badgeOverlay?: string;
}

const newsTypes: ContentType[] = ["news", "feature", "tips", "editorial_review"];

export const categoryPills: DemoCategoryPill[] = [
  { id: "all", labelId: "Untuk Anda", labelEn: "For You", isPopular: true },
  { id: "promo", labelId: "Promo Guncang B2B", labelEn: "B2B Mega Promo", badge: "8.8" },
  { id: "verified", labelId: "Supplier Terverifikasi", labelEn: "Verified Suppliers" },
  { id: "tani", labelId: "Komoditas Tani", labelEn: "Agriculture" },
  { id: "bahan_baku", labelId: "Bahan Baku Industri", labelEn: "Raw Materials" },
  { id: "elektronik", labelId: "Elektronik & Mesin", labelEn: "Electronics & Machinery" },
  { id: "fashion", labelId: "Tekstil & Fashion", labelEn: "Textiles & Fashion" },
  { id: "otomotif", labelId: "Otomotif & Sparepart", labelEn: "Automotive" },
];

export function useDemoHome(locale: string) {
  const contentLocale = locale === "en" ? "en" : "id";
  const [activeCategoryTab, setActiveCategoryTab] = useState<string>("all");
  const [activeNewsType, setActiveNewsType] = useState<ContentType>("news");
  const [showBackToTop, setShowBackToTop] = useState<boolean>(false);

  useEffect(() => {
    const handleScroll = () => {
      setShowBackToTop(window.scrollY > 400);
    };

    window.addEventListener("scroll", handleScroll);
    return () => {
      window.removeEventListener("scroll", handleScroll);
    };
  }, []);

  const newsQuery = useQuery({
    queryKey: ["demo-home-news", contentLocale, activeNewsType],
    queryFn: () => contentService.listPublic({ type: activeNewsType, locale: contentLocale, perPage: 4 }),
    staleTime: 5 * 60 * 1000,
  });

  const videosQuery = useQuery({
    queryKey: ["demo-home-videos", contentLocale],
    queryFn: () => contentService.listPublic({ type: "video", locale: contentLocale, perPage: 4 }),
    staleTime: 5 * 60 * 1000,
  });

  const productsQuery = useQuery({
    queryKey: ["demo-home-products"],
    queryFn: () => searchService.searchProducts(""),
    staleTime: 5 * 60 * 1000,
  });

  const suppliersQuery = useQuery({
    queryKey: ["demo-home-suppliers"],
    queryFn: () => searchService.search({}),
    staleTime: 5 * 60 * 1000,
  });

  const popularProducts = useMemo<EnhancedDemoProduct[]>(() => {
    const raw = [...(productsQuery.data || [])];
    const sorted = raw.sort((a, b) => (b.supplierRating || 0) - (a.supplierRating || 0));

    // Map each product to add Tokopedia-style marketplace metadata matching screenshot
    return sorted.map((prod, idx) => {
      const discounts = [36, 50, 41, 30, 24, 65, 88];
      const salesTexts = ["100+ terjual", "50rb+ terjual", "10rb+ terjual", "4rb+ terjual", "3rb+ terjual"];
      const locations = ["Kab. Tangerang", "Kota Bandung", "Kota Bekasi", "Jakarta Barat", "Kab. Bogor"];
      const badgeOverlays = ["Promo 8.8", "Bonus Cashback", "Bebas Ongkir", "Terlaris"];

      const disc = discounts[idx % discounts.length];
      const origPrice = prod.price ? Math.round(prod.price * (1 + disc / 100)) : undefined;

      return {
        ...prod,
        discountPercentage: disc,
        originalPrice: origPrice,
        promoBadge: `Hemat s.d ${disc}% Pakai Bonus`,
        salesCountText: salesTexts[idx % salesTexts.length],
        locationTag: prod.supplierLocation || locations[idx % locations.length],
        badgeOverlay: badgeOverlays[idx % badgeOverlays.length],
      };
    }).slice(0, 12); // show up to 12 products for 6-column grid symmetry
  }, [productsQuery.data]);

  const filteredProducts = useMemo<EnhancedDemoProduct[]>(() => {
    if (activeCategoryTab === "all") return popularProducts;
    if (activeCategoryTab === "verified") return popularProducts.filter((p) => p.supplierVerified);
    if (activeCategoryTab === "tani") return popularProducts.filter((p) => (p.categoryName || "").toLowerCase().includes("tani") || (p.name || "").toLowerCase().includes("kopi"));
    if (activeCategoryTab === "bahan_baku") return popularProducts.filter((p) => (p.categoryName || "").toLowerCase().includes("bahan") || (p.description || "").toLowerCase().includes("baku"));
    return popularProducts;
  }, [activeCategoryTab, popularProducts]);

  const comparisonCards = useMemo<DemoComparisonCard[]>(() => {
    const suppliers = suppliersQuery.data || [];
    const pairs: DemoComparisonCard[] = [];
    for (let index = 0; index + 1 < suppliers.length && pairs.length < 3; index += 2) {
      const first = suppliers[index];
      const second = suppliers[index + 1];
      pairs.push({
        id: `${first.id}-${second.id}`,
        title: `${first.keyProducts?.[0] || first.businessType || "Supplier"} vs ${second.keyProducts?.[0] || second.businessType || "Supplier"}`,
        first,
        second,
      });
    }
    return pairs;
  }, [suppliersQuery.data]);

  const scrollToTop = () => {
    window.scrollTo({
      top: 0,
      behavior: "smooth",
    });
  };

  return {
    categoryPills,
    activeCategoryTab,
    setActiveCategoryTab,
    newsTypes,
    activeNewsType,
    setActiveNewsType,
    newsArticles: newsQuery.data?.items || ([] as ContentArticle[]),
    videos: videosQuery.data?.items || ([] as ContentArticle[]),
    popularProducts: filteredProducts,
    allPopularProductsCount: popularProducts.length,
    comparisonCards,
    isNewsLoading: newsQuery.isLoading,
    isVideosLoading: videosQuery.isLoading,
    isProductsLoading: productsQuery.isLoading,
    isSuppliersLoading: suppliersQuery.isLoading,
    showBackToTop,
    scrollToTop,
  };
}

