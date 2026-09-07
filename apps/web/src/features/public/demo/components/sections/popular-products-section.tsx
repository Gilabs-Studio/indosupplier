"use client";

import React from "react";
import { ChevronDown, PackageOpen } from "lucide-react";
import { useDemoFeaturedProducts } from "../../hooks/use-demo-home";
import { MarketplaceProductCard } from "../marketplace-product-card";
import { PopularProductsSkeleton } from "../skeletons/product-card-skeleton";

interface PopularProductsSectionProps {
  locale: string;
}

export function PopularProductsSection({ locale }: Readonly<PopularProductsSectionProps>) {
  const {
    isEn,
    filters,
    products,
    isLoading,
    bookmarkedProductIds,
    setSort,
    toggleBookmark,
  } = useDemoFeaturedProducts(locale);

  return (
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

      {/* Products Grid / Skeletons */}
      <div className="w-full">
        {isLoading ? (
          <PopularProductsSkeleton count={12} />
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
  );
}
