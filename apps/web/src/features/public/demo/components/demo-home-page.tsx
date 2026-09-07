"use client";

import React from "react";
import { PublicLayout } from "@/features/public/components/public-layout";
import { ChevronDown, PackageOpen, CheckCircle2 } from "lucide-react";
import { useDemoHome } from "../hooks/use-demo-home";
import { HeroBanner } from "./hero-banner";
import { QuickActionsBar } from "./quick-actions-bar";
import { PopularCategoriesSection } from "./popular-categories-section";
import { ProductFilterSidebar } from "./product-filter-sidebar";
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
  } = useDemoHome(locale);

  return (
    <PublicLayout locale={locale}>
      <div className="min-h-screen bg-background pb-16 font-sans">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 pt-6 space-y-8">
          {/* 1. Hero B2B Banner (Direct Image matching user specification) */}
          <HeroBanner />

          {/* 2. Quick Actions Bar (4 Cards with theme tokens) */}
          <QuickActionsBar actions={quickActions} isEn={isEn} />

          {/* 3. Popular Categories (8 Cards with realistic single-object images) */}
          <PopularCategoriesSection categories={popularCategories} isEn={isEn} />

          {/* 4. Featured Products Section */}
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

            {/* Layout: Left Filter Sidebar + Right Product Grid */}
            <div className="flex flex-col gap-6 lg:flex-row items-start">
              {/* Left Filter Sidebar */}
              <div className="w-full lg:w-64 shrink-0">
                <ProductFilterSidebar
                  filters={filters}
                  isEn={isEn}
                  onLocationChange={setLocation}
                  onPriceChange={setPriceRange}
                  onMinOrderChange={setMinOrder}
                  onTogglePowerSupplier={togglePowerSupplier}
                  onToggleVerifiedSupplier={toggleVerifiedSupplier}
                  onToggleReadyStock={toggleReadyStock}
                  onReset={resetFilters}
                />
              </div>

              {/* Right Products Grid */}
              <div className="flex-1 w-full">
                {isProductsLoading ? (
                  <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
                    {Array.from({ length: 8 }).map((_, idx) => (
                      <div key={idx} className="h-80 animate-pulse rounded-xl bg-muted/40" />
                    ))}
                  </div>
                ) : products.length === 0 ? (
                  <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border bg-card p-12 text-center">
                    <PackageOpen className="h-12 w-12 text-muted-foreground/40 mb-3" />
                    <p className="text-sm font-bold text-foreground">
                      {isEn ? "No products found" : "Produk tidak ditemukan"}
                    </p>
                    <p className="text-xs text-muted-foreground mt-1 max-w-sm">
                      {isEn
                        ? "Try adjusting your filters or search keywords to find what you are looking for."
                        : "Coba sesuaikan filter atau kata kunci pencarian Anda untuk menemukan produk yang sesuai."}
                    </p>
                    <button
                      onClick={resetFilters}
                      className="mt-4 rounded-lg bg-primary px-4 py-2 text-xs font-bold text-primary-foreground hover:bg-primary/90 cursor-pointer transition-colors"
                    >
                      {isEn ? "Reset All Filters" : "Reset Semua Filter"}
                    </button>
                  </div>
                ) : (
                  <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
                    {products.map((product) => (
                      <MarketplaceProductCard
                        key={product.id}
                        product={product}
                        isEn={isEn}
                        isBookmarked={bookmarkedProductIds.has(product.id)}
                        isJustAddedToCart={cartFeedback?.id === product.id}
                        onToggleBookmark={toggleBookmark}
                        onAddToCart={handleAddToCart}
                      />
                    ))}
                  </div>
                )}
              </div>
            </div>
          </section>
        </div>

        {/* Interactive Toast Notification on Cart Add */}
        {cartFeedback && (
          <div className="fixed bottom-6 right-6 z-50 flex items-center gap-2.5 rounded-xl border border-primary/30 bg-card p-4 shadow-xl text-foreground animate-in fade-in slide-in-from-bottom-2">
            <CheckCircle2 className="h-5 w-5 text-success shrink-0" />
            <div className="text-xs">
              <p className="font-bold">{isEn ? "Added to Cart" : "Berhasil Ditambahkan"}</p>
              <p className="text-muted-foreground line-clamp-1 max-w-xs">{cartFeedback.name}</p>
            </div>
          </div>
        )}
      </div>
    </PublicLayout>
  );
}
