"use client";

import React, { useState } from "react";
import Image from "next/image";
import { Link } from "@/i18n/routing";
import { Badge } from "@/components/ui/badge";
import { Star, Heart, CheckCircle2, MapPin, Package, Check, GitCompareArrows } from "lucide-react";
import { formatPrice } from "@/lib/utils";

export interface ProductCardProps {
  id: string;
  name: string;
  price: number;
  currency?: string;
  unit?: string;
  image?: string;
  fallbackImage?: string;
  href: string;

  // Badges
  isReadyStock?: boolean;
  isVerified?: boolean;
  discountPercentage?: number;
  originalPrice?: number;

  // Overlay / Bookmark Actions
  showBookmarkOverlayButton?: boolean;
  isBookmarked?: boolean;
  onBookmark?: (id: string) => void;
  bookmarkAriaLabel?: string;

  // Compare Actions
  showCompareOverlayButton?: boolean;
  showCompareCheckbox?: boolean;
  isCompared?: boolean;
  onCompare?: (id: string) => void;
  compareLabel?: string;
  compareCheckboxLabel?: string;
  customOverlayButton?: React.ReactNode;

  // Details
  rating?: number;
  reviewCount?: number;
  soldCountText?: string;
  ulasanLabel?: string;

  // Supplier & Location Flip (Tokopedia Style)
  supplierName?: string;
  supplierLocation?: string;
  supplierHref?: string;

  // Tags & Labels (backward-compatibility)
  categoryName?: string;
  minOrder?: string;
  moqLabel?: string;
  priceLabel?: string;

  // Custom Footer slot
  customFooter?: React.ReactNode;
  className?: string;
}

export function ProductCard({
  id,
  name,
  price,
  currency = "IDR",
  unit,
  image,
  fallbackImage = "/images/products/prod-hvs.webp",
  href,
  isReadyStock = true,
  isVerified = false,
  discountPercentage = 0,
  originalPrice,
  showBookmarkOverlayButton = true,
  isBookmarked = false,
  onBookmark,
  bookmarkAriaLabel,
  showCompareOverlayButton = false,
  showCompareCheckbox = false,
  isCompared = false,
  onCompare,
  compareLabel,
  compareCheckboxLabel,
  customOverlayButton,
  rating = 4.9,
  reviewCount,
  soldCountText,
  ulasanLabel = "terjual",
  supplierName,
  supplierLocation,
  priceLabel,
  customFooter,
  className = "",
}: Readonly<ProductCardProps>) {
  const [hasError, setHasError] = useState(false);
  const [isHovered, setIsHovered] = useState(false);

  const formattedPrice = price ? formatPrice(price, currency) : "";
  const displayPrice = formattedPrice || priceLabel || "Hubungi Supplier";
  const displayOriginalPrice = originalPrice ? formatPrice(originalPrice, currency) : "";
  const hasDiscount = discountPercentage > 0;

  const reviewCountFormatted =
    reviewCount !== undefined
      ? reviewCount >= 1000
        ? `${(reviewCount / 1000).toFixed(1).replace(".", ",")}rb`
        : `${reviewCount}`
      : undefined;

  const sanitizedImage =
    image && image.startsWith("/images/") && image.endsWith(".png")
      ? image.replace(/\.png$/, ".webp")
      : image;

  const finalImgSrc = hasError || !sanitizedImage || sanitizedImage.includes("unsplash.com") ? fallbackImage : sanitizedImage;

  return (
    <div
      id={id}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      className={`group flex flex-col justify-between rounded-xl border border-border bg-card overflow-hidden shadow-2xs transition-shadow duration-200 hover:shadow-md cursor-pointer ${className}`}
    >
      {/* 1. Full-Bleed Image Container (Zero Padding, edge-to-edge, stable no excessive zoom) */}
      <div className="relative aspect-square w-full overflow-hidden bg-muted/20">
        <Link href={href} className="block h-full w-full cursor-pointer relative">
          {finalImgSrc ? (
            <Image
              src={finalImgSrc}
              alt={name}
              fill
              sizes="(max-width: 640px) 50vw, (max-width: 1024px) 33vw, 16vw"
              onError={() => setHasError(true)}
              className="object-cover w-full h-full"
            />
          ) : (
            <div className="flex h-full w-full items-center justify-center text-muted-foreground/40">
              <Package className="h-10 w-10 stroke-[1.5]" />
            </div>
          )}
        </Link>

        {/* Discount tag on top-left if present */}
        {hasDiscount && (
          <div className="absolute left-0 top-0 bg-destructive text-white font-extrabold text-[10px] px-2 py-0.5 rounded-br-lg shadow-xs z-10">
            {discountPercentage}%
          </div>
        )}

        {/* Top-Right Action Overlays */}
        {(showBookmarkOverlayButton || showCompareOverlayButton || customOverlayButton) && (
          <div className="absolute right-2 top-2 z-10 flex flex-col items-center gap-1.5">
            {customOverlayButton}

            {/* Heart Wishlist Toggle */}
            {showBookmarkOverlayButton && (
              <button
                type="button"
                onClick={(e) => {
                  e.preventDefault();
                  e.stopPropagation();
                  onBookmark?.(id);
                }}
                aria-label={bookmarkAriaLabel || (isBookmarked ? "Disimpan" : "Simpan")}
                title={bookmarkAriaLabel || (isBookmarked ? "Disimpan" : "Simpan")}
                className={`flex h-7 w-7 cursor-pointer items-center justify-center rounded-full backdrop-blur-xs transition-all duration-200 hover:scale-110 active:scale-95 shadow-xs ${
                  isBookmarked
                    ? "bg-background text-destructive ring-1 ring-destructive/20 shadow-destructive/10"
                    : "bg-background/80 text-muted-foreground hover:text-destructive hover:bg-background"
                }`}
              >
                <Heart
                  className={`h-4 w-4 transition-colors ${
                    isBookmarked ? "fill-destructive text-destructive" : "text-muted-foreground"
                  }`}
                />
              </button>
            )}

            {/* Compare Overlay Button */}
            {showCompareOverlayButton && (
              <button
                type="button"
                onClick={(e) => {
                  e.preventDefault();
                  e.stopPropagation();
                  onCompare?.(id);
                }}
                aria-label={compareLabel || (isCompared ? "Dalam perbandingan" : "Bandingkan")}
                title={compareLabel || (isCompared ? "Dalam perbandingan" : "Bandingkan")}
                className={`flex h-7 w-7 cursor-pointer items-center justify-center rounded-full backdrop-blur-xs transition-all duration-200 hover:scale-110 active:scale-95 shadow-xs ${
                  isCompared
                    ? "bg-primary text-primary-foreground shadow-primary/20 hover:bg-primary/90"
                    : "bg-background/80 text-muted-foreground hover:text-primary hover:bg-background"
                }`}
              >
                {isCompared ? (
                  <Check className="h-3.5 w-3.5 stroke-[2.5]" />
                ) : (
                  <GitCompareArrows className="h-3.5 w-3.5" />
                )}
              </button>
            )}
          </div>
        )}
      </div>

      {/* 2. Card Information Body */}
      <div className="flex flex-1 flex-col justify-between p-2.5 sm:p-3 space-y-1.5">
        <div className="space-y-1">
          {/* Title (2 lines clamp) */}
          <Link href={href} className="block cursor-pointer">
            <h3 className="line-clamp-2 text-xs sm:text-sm font-semibold leading-snug text-foreground group-hover:text-primary transition-colors min-h-[2.25rem]">
              {name}
            </h3>
          </Link>

          {/* Price & Unit & Optional Discount */}
          <div className="pt-0.5">
            <span className="text-sm sm:text-base font-bold text-foreground">
              {displayPrice}
            </span>
            {unit && (
              <span className="text-[11px] text-muted-foreground ml-1">
                / {unit}
              </span>
            )}
            {hasDiscount && originalPrice && (
              <div className="flex items-center gap-1.5 text-[10px] text-muted-foreground/60">
                <span className="line-through">{displayOriginalPrice}</span>
                <span className="font-bold text-destructive">{discountPercentage}%</span>
              </div>
            )}
          </div>

          {/* Solid Ready Stock Badge from badge.tsx */}
          {isReadyStock && (
            <div className="pt-0.5">
              <Badge variant="success" className="text-[10px] font-bold px-1.5 py-0.5 rounded-xs shadow-none">
                Ready Stock
              </Badge>
            </div>
          )}

          {/* Rating & Terjual count (Tokopedia Reference) */}
          {(rating !== undefined || soldCountText || reviewCountFormatted) && (
            <div className="flex items-center gap-1 text-[11px] text-muted-foreground pt-0.5">
              {rating !== undefined && (
                <>
                  <Star className="h-3.5 w-3.5 fill-warning text-warning shrink-0" />
                  <span className="font-semibold text-foreground">{rating.toFixed(1)}</span>
                </>
              )}
              {soldCountText ? (
                <>
                  <span>•</span>
                  <span>{soldCountText}</span>
                </>
              ) : reviewCountFormatted ? (
                <>
                  <span>•</span>
                  <span>{reviewCountFormatted} {ulasanLabel}</span>
                </>
              ) : null}
            </div>
          )}

          {/* Tokopedia Motion Rolling Text: Supplier Name <-> Supplier Location */}
          {(supplierName || supplierLocation) && (
            <div className="relative h-5 overflow-hidden text-[11px] text-muted-foreground">
              <div
                className="flex flex-col transition-transform duration-300 ease-out"
                style={{
                  transform:
                    isHovered && supplierName && supplierLocation
                      ? "translateY(-50%)"
                      : "translateY(0%)",
                }}
              >
                {/* Row 1: Supplier Name (Default) */}
                <div className="h-5 flex items-center gap-1 shrink-0">
                  {supplierName ? (
                    <>
                      {isVerified && (
                        <CheckCircle2 className="h-3 w-3 shrink-0 text-success fill-success/20" />
                      )}
                      <span className="truncate font-medium">{supplierName}</span>
                    </>
                  ) : (
                    <>
                      <MapPin className="h-3 w-3 shrink-0 text-muted-foreground" />
                      <span className="truncate font-medium">{supplierLocation}</span>
                    </>
                  )}
                </div>

                {/* Row 2: Location (Rolls in from bottom on hover) */}
                {supplierName && supplierLocation && (
                  <div className="h-5 flex items-center gap-1 shrink-0">
                    <MapPin className="h-3 w-3 shrink-0 text-muted-foreground" />
                    <span className="truncate font-medium">{supplierLocation}</span>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>

        {/* Optional Compare Checkbox & Custom Footer */}
        {(showCompareCheckbox || customFooter) && (
          <div className="pt-2 border-t border-border/40">
            {showCompareCheckbox && customFooter ? (
              <div className="flex items-center justify-between gap-2">
                <label
                  onClick={(e) => e.stopPropagation()}
                  className="flex cursor-pointer items-center gap-2 text-xs font-semibold text-foreground select-none hover:text-primary transition-colors min-w-0"
                >
                  <input
                    type="checkbox"
                    checked={isCompared}
                    onChange={(e) => {
                      e.stopPropagation();
                      onCompare?.(id);
                    }}
                    className="h-4 w-4 cursor-pointer rounded border-border text-primary focus:ring-primary shrink-0"
                  />
                  <span className="truncate">{compareCheckboxLabel || "Pilih untuk bandingkan"}</span>
                </label>
                {customFooter}
              </div>
            ) : showCompareCheckbox ? (
              <label
                onClick={(e) => e.stopPropagation()}
                className="flex cursor-pointer items-center gap-2 text-xs font-semibold text-foreground select-none hover:text-primary transition-colors"
              >
                <input
                  type="checkbox"
                  checked={isCompared}
                  onChange={(e) => {
                    e.stopPropagation();
                    onCompare?.(id);
                  }}
                  className="h-4 w-4 cursor-pointer rounded border-border text-primary focus:ring-primary shrink-0"
                />
                <span className="truncate">{compareCheckboxLabel || "Pilih untuk bandingkan"}</span>
              </label>
            ) : (
              customFooter
            )}
          </div>
        )}
      </div>
    </div>
  );
}
