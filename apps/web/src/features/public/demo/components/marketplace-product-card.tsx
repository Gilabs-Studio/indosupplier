"use client";

import React from "react";
import { ProductCard } from "@/components/ui/product-card";
import type { DemoProductItem } from "../types/demo.types";

interface MarketplaceProductCardProps {
  product: DemoProductItem;
  isEn: boolean;
  isBookmarked: boolean;
  onToggleBookmark: (id: string) => void;
}

function getSmartFallbackImage(name: string, categorySlug?: string): string {
  const lower = (name + " " + (categorySlug || "")).toLowerCase();
  if (lower.includes("kertas") || lower.includes("hvs") || lower.includes("atk") || lower.includes("pen")) {
    return "/images/products/prod-hvs.png";
  }
  if (lower.includes("masker") || lower.includes("medis")) {
    return "/images/products/prod-masker.png";
  }
  if (lower.includes("helm") || lower.includes("helmet") || lower.includes("safety") || lower.includes("k3")) {
    return "/images/products/prod-helmet.png";
  }
  if (lower.includes("laptop") || lower.includes("elektronik") || lower.includes("computer") || lower.includes("pc")) {
    return "/images/products/prod-laptop.png";
  }
  if (lower.includes("pompa") || lower.includes("pump") || lower.includes("mesin") || lower.includes("machinery")) {
    return "/images/products/prod-water-pump.png";
  }
  if (lower.includes("kursi") || lower.includes("chair") || lower.includes("furniture") || lower.includes("meja")) {
    return "/images/products/prod-office-chair.png";
  }
  if (lower.includes("karton") || lower.includes("box") || lower.includes("packaging") || lower.includes("kardus")) {
    return "/images/products/prod-carton-boxes.png";
  }
  if (
    lower.includes("sand") ||
    lower.includes("garnet") ||
    lower.includes("bentonite") ||
    lower.includes("clay") ||
    lower.includes("carbon") ||
    lower.includes("powder")
  ) {
    return "/images/products/prod-mineral-powder.png";
  }
  if (
    lower.includes("sugar") ||
    lower.includes("gula") ||
    lower.includes("kopi") ||
    lower.includes("coffee") ||
    lower.includes("makanan") ||
    lower.includes("food")
  ) {
    return "/images/categories/cat-makanan-minuman.png";
  }
  if (
    lower.includes("steel") ||
    lower.includes("baja") ||
    lower.includes("rebar") ||
    lower.includes("plate") ||
    lower.includes("coil") ||
    lower.includes("yarn") ||
    lower.includes("denim") ||
    lower.includes("ginger") ||
    lower.includes("bahan")
  ) {
    return "/images/categories/cat-bahan-baku.png";
  }
  return "/images/products/prod-hvs.png";
}

export function MarketplaceProductCard({
  product,
  isEn,
  isBookmarked,
  onToggleBookmark,
}: Readonly<MarketplaceProductCardProps>) {
  const fallback = getSmartFallbackImage(product.name, product.categorySlug);

  return (
    <ProductCard
      id={product.id}
      name={product.name}
      price={product.price}
      currency={product.currency}
      unit={product.unit || "unit"}
      image={product.photos?.[0]}
      fallbackImage={fallback}
      href={`/demo/products/${product.id}`}
      isReadyStock={true}
      isVerified={product.supplierVerified}
      rating={product.supplierRating}
      reviewCount={product.supplierReviewCount}
      ulasanLabel={isEn ? "terjual" : "terjual"}
      supplierName={product.supplierCompanyName}
      supplierLocation={product.supplierLocation}
      isBookmarked={isBookmarked}
      onBookmark={onToggleBookmark}
    />
  );
}
