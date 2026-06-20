"use client";
/* eslint-disable @next/next/no-img-element */

import React from "react";
import { Link } from "@/i18n/routing";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Package, Star, Store, Heart, Check, GitCompareArrows } from "lucide-react";
import { formatPrice } from "@/lib/utils";

interface ProductCardProps {
  id: string;
  name: string;
  price: number;
  currency?: string;
  image?: string;
  
  // Navigation
  href: string;
  
  // Badges & Overlay Flags
  isVerified?: boolean;
  discountPercentage?: number;
  originalPrice?: number;
  
  // Action Overlay
  customOverlayButton?: React.ReactNode;
  showBookmarkOverlayButton?: boolean;
  showCompareOverlayButton?: boolean;
  isBookmarked?: boolean;
  isCompared?: boolean;
  onBookmark?: () => void;
  onCompare?: () => void;
  
  // Details
  rating?: number;
  reviewCount?: number;
  soldCountText?: string;
  
  // Supplier
  supplierName?: string;
  supplierHref?: string;
  
  // Tags
  categoryName?: string;
  minOrder?: string;
  
  // Subtitles / Labels (i18n support)
  priceLabel?: string;
  moqLabel?: string;
  ulasanLabel?: string;
  
  // Custom Bottom Footer Slot
  customFooter?: React.ReactNode;
  
  className?: string;
}



export function ProductCard({
  id,
  name,
  price,
  currency = "IDR",
  image,
  href,
  isVerified = false,
  discountPercentage = 0,
  originalPrice,
  customOverlayButton,
  showBookmarkOverlayButton = false,
  showCompareOverlayButton = false,
  isBookmarked = false,
  isCompared = false,
  onBookmark,
  onCompare,
  rating,
  reviewCount,
  soldCountText,
  supplierName,
  supplierHref,
  categoryName,
  minOrder,
  priceLabel,
  moqLabel = "MOQ",
  ulasanLabel = "ulasan",
  customFooter,
  className = "",
}: Readonly<ProductCardProps>) {
  const formattedPrice = price ? formatPrice(price, currency) : "";
  const displayPrice = formattedPrice || priceLabel || "Hubungi Supplier";
  const displayOriginalPrice = originalPrice ? formatPrice(originalPrice, currency) : "";
  const hasDiscount = discountPercentage > 0;
  const hasOverlayActions =
    Boolean(customOverlayButton) || (showBookmarkOverlayButton && onBookmark) || (showCompareOverlayButton && onCompare);

  return (
    <Card
      id={id}
      className={`group flex flex-col overflow-hidden border-0 bg-transparent shadow-none ${className}`}
    >
      {/* Image Container with light background and rounded corners */}
      <div className="relative aspect-square w-full bg-muted/40 rounded-lg overflow-hidden transition-colors duration-200 group-hover:bg-muted/50">
        <Link href={href} className="block h-full w-full cursor-pointer">
          {image ? (
            <img
              src={image}
              alt={name}
              className="h-full w-full object-cover"
            />
          ) : (
            <div className="flex h-full w-full items-center justify-center text-muted-foreground/50">
              <Package className="h-10 w-10 stroke-[1.5]" />
            </div>
          )}
        </Link>

        {/* Verification badge on the bottom-left corner */}
        {isVerified && (
          <div className="absolute bottom-2 left-2 bg-success text-success-foreground font-extrabold text-[9px] md:text-[10px] rounded-xs px-1.5 py-0.5 shadow-xs z-10 uppercase tracking-wider">
            Verified
          </div>
        )}

        {/* Discount tag on the top-left corner */}
        {hasDiscount && (
          <div className="absolute left-0 top-0 bg-destructive text-destructive-foreground font-extrabold text-[10px] md:text-[11px] rounded-br-lg px-2.5 py-1 shadow-xs z-10">
            {discountPercentage}%
          </div>
        )}

        {/* Actions Overlay */}
        <div
          className={`absolute right-2 top-2 flex flex-col gap-1.5 transition-opacity duration-200 ${
            hasOverlayActions
              ? "opacity-100 md:opacity-0 md:group-hover:opacity-100"
              : "opacity-0 pointer-events-none"
          }`}
        >
          {customOverlayButton}
          {showBookmarkOverlayButton && onBookmark && (
            <Button
              type="button"
              variant="secondary"
              size="icon"
              onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                onBookmark();
              }}
              className="h-8 w-8 rounded-lg border border-border/60 bg-background/95 text-foreground shadow-xs backdrop-blur-xs transition-colors duration-200 hover:bg-background cursor-pointer"
            >
              <Heart
                className={`h-4 w-4 transition-colors ${
                  isBookmarked ? "text-destructive fill-destructive" : "text-foreground"
                }`}
              />
            </Button>
          )}

          {showCompareOverlayButton && onCompare && (
            <Button
              type="button"
              variant="secondary"
              size="icon"
              onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                onCompare();
              }}
              className="h-8 w-8 rounded-lg border border-border/60 bg-background/95 text-foreground shadow-xs backdrop-blur-xs transition-colors duration-200 hover:bg-background cursor-pointer"
            >
              {isCompared ? (
                <Check className="h-4 w-4 text-success" />
              ) : (
                <GitCompareArrows className="h-4 w-4 text-foreground" />
              )}
            </Button>
          )}
        </div>
      </div>

      {/* Info Content Area */}
      <CardContent className="flex flex-col flex-1 px-0 pt-2.5 pb-1">
        {/* Title */}
        <Link href={href} className="block cursor-pointer flex-1 mb-1">
          <h3 className="line-clamp-2 min-h-8 text-xs md:text-sm font-medium leading-relaxed text-foreground group-hover:text-primary transition-colors duration-200">
            {name}
          </h3>
        </Link>

        {/* Price & Discount block */}
        <div className="flex items-baseline flex-wrap gap-1.5 mt-1">
          <span className="text-sm md:text-base font-extrabold text-foreground leading-none">
            {displayPrice}
          </span>
          {hasDiscount && originalPrice && (
            <>
              <span className="text-[10px] md:text-xs text-muted-foreground/60 line-through leading-none">
                {displayOriginalPrice}
              </span>
              <span className="text-[10px] md:text-xs font-bold text-destructive leading-none">
                {discountPercentage}%
              </span>
            </>
          )}
        </div>

        {/* Promo / Discount Label */}
        {hasDiscount && (
          <p className="mt-1 text-[10px] md:text-xs font-semibold text-primary">
            Hemat s.d {discountPercentage}%
          </p>
        )}

        {/* Rating and Reviews / Sold Count */}
        {(rating !== undefined || soldCountText) && (
          <div className="mt-1.5 flex items-center flex-wrap gap-1 text-[10px] md:text-xs text-muted-foreground">
            {rating !== undefined && (
              <div className="flex items-center gap-0.5 text-warning font-semibold">
                <Star className="h-3.5 w-3.5 fill-warning text-warning" />
                <span>{rating.toFixed(1)}</span>
              </div>
            )}
            {rating !== undefined && (reviewCount !== undefined || soldCountText) && (
              <span className="text-muted-foreground/40">•</span>
            )}
            {reviewCount !== undefined && (
              <span>
                {reviewCount} {ulasanLabel}
              </span>
            )}
            {((rating !== undefined || reviewCount !== undefined) && soldCountText) && (
              <span className="text-muted-foreground/40">•</span>
            )}
            {soldCountText && <span className="font-medium">{soldCountText}</span>}
          </div>
        )}

        {/* Supplier Info */}
        {supplierName && (
          <div className="mt-1.5">
            {supplierHref ? (
              <Link
                href={supplierHref}
                className="flex cursor-pointer items-center gap-1 text-[10px] md:text-xs text-muted-foreground hover:text-primary transition-colors"
              >
                <Store className="h-3 w-3 shrink-0 text-primary/70" />
                <span className="truncate font-semibold">{supplierName}</span>
              </Link>
            ) : (
              <div className="flex items-center gap-1 text-[10px] md:text-xs text-muted-foreground">
                <Store className="h-3 w-3 shrink-0 text-primary/70" />
                <span className="truncate font-semibold">{supplierName}</span>
              </div>
            )}
          </div>
        )}

        {/* Tags (Category & MOQ) */}
        {(categoryName || minOrder) && (
          <div className="mt-2 flex flex-wrap gap-1">
            {categoryName && (
              <Badge
                variant="outline"
                className="border-border/60 text-muted-foreground text-[9px] font-medium rounded-xs px-1.5 py-0 bg-muted/10"
              >
                {categoryName}
              </Badge>
            )}
            {minOrder && (
              <Badge
                variant="outline"
                className="border-border/60 text-muted-foreground text-[9px] font-medium rounded-xs px-1.5 py-0 bg-muted/10"
              >
                {moqLabel} {minOrder}
              </Badge>
            )}
          </div>
        )}

        {/* Custom Actions Slot (Footer / bottom button) */}
        {customFooter && (
          <div className="mt-2.5 pt-2.5 border-t border-border/40">
            {customFooter}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
