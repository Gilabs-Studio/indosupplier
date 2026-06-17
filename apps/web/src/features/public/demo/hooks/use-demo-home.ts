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

const newsTypes: ContentType[] = ["news", "feature", "tips", "editorial_review"];

export function useDemoHome(locale: string) {
  const contentLocale = locale === "en" ? "en" : "id";
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

  const popularProducts = useMemo<PublicProductDto[]>(() => {
    return [...(productsQuery.data || [])]
      .sort((a, b) => (b.supplierRating || 0) - (a.supplierRating || 0))
      .slice(0, 4);
  }, [productsQuery.data]);

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
    newsTypes,
    activeNewsType,
    setActiveNewsType,
    newsArticles: newsQuery.data?.items || ([] as ContentArticle[]),
    videos: videosQuery.data?.items || ([] as ContentArticle[]),
    popularProducts,
    comparisonCards,
    isNewsLoading: newsQuery.isLoading,
    isVideosLoading: videosQuery.isLoading,
    isProductsLoading: productsQuery.isLoading,
    isSuppliersLoading: suppliersQuery.isLoading,
    showBackToTop,
    scrollToTop,
  };
}
