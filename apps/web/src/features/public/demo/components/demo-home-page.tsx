"use client";

import React from "react";
import { PublicLayout } from "@/features/public/components/public-layout";
import { ChevronDown, PackageOpen } from "lucide-react";
import { useDemoHome } from "../hooks/use-demo-home";
import { HeroBanner } from "./hero-banner";
import { QuickActionsBar } from "./quick-actions-bar";
import { PopularCategoriesSection } from "./popular-categories-section";
import { MarketplaceProductCard } from "./marketplace-product-card";

interface DemoHomePageProps {
  locale: string;
}

export function DemoHomePage({ locale }: Readonly<DemoHomePageProps>) {
  const {
    isEn,
    filters,
    quickActions,
    popularCategories,
    products,
    isProductsLoading,
    bookmarkedProductIds,
    setSort,
    toggleBookmark,
  } = useDemoHome(locale);

  return (
    <PublicLayout locale={locale}>
      <div className="min-h-screen bg-background pb-16 font-sans">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 pt-6 space-y-8">
          {/* 1. Hero B2B Banner (Direct Image matching user specification) */}
          <HeroBanner />

          {/* 2. Quick Actions Bar (Minimalist Vector Images) */}
          <QuickActionsBar actions={quickActions} isEn={isEn} />

          {/* 3. Popular Categories (8 Cards with realistic single-object images) */}
          <PopularCategoriesSection categories={popularCategories} isEn={isEn} />

          {/* 4. Featured Products Section (Full-Width Tokopedia/Shopee Grid) */}
          <section className="w-full space-y-4 pt-2">
            {/* Header: Title + Sort */}
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <h2 className="text-lg sm:text-xl font-bold tracking-tight text-foreground">
                {isEn ? "Featured Products" : "Produk Unggulan"}
              </h2>

              {/* Sort selector */}
              <div className="flex items-center gap-2">
                <span className="text-xs font-semibold text-muted-foreground">
                  {isEn ? "Sort by:" : "Urutkan:"}
                </span>
                <div className="relative">
                  <select
                    value={filters.sort}
                    onChange={(e) => setSort(e.target.value as typeof filters.sort)}
                    className="appearance-none rounded-lg border border-border bg-card px-3 py-1.5 pr-8 text-xs font-bold text-foreground outline-hidden focus:border-primary focus:ring-1 focus:ring-primary cursor-pointer shadow-2xs"
                  >
                    <option value="terlaris">{isEn ? "Best Selling" : "Terlaris"}</option>
                    <option value="price_asc">{isEn ? "Lowest Price" : "Harga Terendah"}</option>
                    <option value="price_desc">{isEn ? "Highest Price" : "Harga Tertinggi"}</option>
                    <option value="rating">{isEn ? "Highest Rating" : "Rating Tertinggi"}</option>
                    <option value="newest">{isEn ? "Newest" : "Terbaru"}</option>
                  </select>
                  <ChevronDown className="pointer-events-none absolute right-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
                </div>
              </div>
            </div>

            {/* Products Grid (Clean Full-Width Tokopedia/Shopee 6-Column Layout) */}
            <div className="w-full">
              {isProductsLoading ? (
                <div className="grid grid-cols-2 gap-3 sm:gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
                  {Array.from({ length: 12 }).map((_, idx) => (
                    <div key={idx} className="h-72 animate-pulse rounded-xl bg-muted/40" />
                  ))}
                </div>
              ) : products.length === 0 ? (
                <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border bg-card p-12 text-center">
                  <PackageOpen className="h-12 w-12 text-muted-foreground/40 mb-3" />
                  <p className="text-sm font-bold text-foreground">
                    {isEn ? "No products found" : "Produk tidak ditemukan"}
                  </p>
                </div>
              ) : (
                <div className="grid grid-cols-2 gap-3 sm:gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
                  {products.map((product) => (
                    <MarketplaceProductCard
                      key={product.id}
                      product={product}
                      isEn={isEn}
                      isBookmarked={bookmarkedProductIds.has(product.id)}
                      onToggleBookmark={toggleBookmark}
                    />
                  ))}
                </div>
              )}
            </div>
          </section>
        </div>
      </div>
    </PublicLayout>
  );
}
