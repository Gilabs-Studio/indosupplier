"use client";

import React from "react";
import { ProductCard } from "@/components/ui/product-card";
import { useTranslations } from "next-intl";
import type { PublicProductDto } from "@/features/public/search/types";

interface PublicProductCardProps {
  product: PublicProductDto;
  detailBasePath: "" | "/demo";
  isAuthenticated: boolean;
  isBookmarked: boolean;
  isCompared: boolean;
  onBookmark: () => void;
  onCompare: () => void;
}

export function PublicProductCard({
  product,
  detailBasePath,
  isBookmarked,
  isCompared,
  onBookmark,
  onCompare,
}: PublicProductCardProps) {
  const t = useTranslations("public.search");

  return (
    <ProductCard
      id={product.id}
      name={product.name}
      price={product.price}
      currency={product.currency}
      image={product.photos?.[0]}
      href={`${detailBasePath}/products/${product.id}`}
      isVerified={product.supplierVerified}
      showBookmarkOverlayButton={true}
      showCompareOverlayButton={true}
      isBookmarked={isBookmarked}
      isCompared={isCompared}
      onBookmark={onBookmark}
      onCompare={onCompare}
      rating={product.supplierRating}
      reviewCount={product.supplierReviewCount}
      supplierName={product.supplierCompanyName}
      supplierHref={`${detailBasePath}/suppliers/${product.supplierSlug}`}
      categoryName={product.categoryName || undefined}
      minOrder={product.minOrder || undefined}
      moqLabel={t("moq")}
      ulasanLabel={t("ulasan")}
      priceLabel={t("contactSupplier")}
    />
  );
}
