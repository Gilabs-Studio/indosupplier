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
}: ProductCardProps) {
  const formattedPrice = price ? formatPrice(price, currency) : "";
  const displayPrice = formattedPrice || priceLabel || "Hubungi Supplier";
  const displayOriginalPrice = originalPrice ? formatPrice(originalPrice, currency) : "";
  const hasDiscount = discountPercentage > 0;

  return (
    <Card id={id} className={`group overflow-hidden border border-border bg-card shadow-xs transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg hover:shadow-primary/10 ${className}`}>
      {/* Image & Badges overlay */}
      <div className="relative aspect-square w-full bg-muted/30 overflow-hidden">
        <Link href={href} className="block h-full w-full cursor-pointer">
          {image ? (
            <img
              src={image}
              alt={name}
              className="h-full w-full object-cover transition-transform duration-500 group-hover:scale-105"
            />
          ) : (
            <div className="flex h-full w-full items-center justify-center text-muted-foreground/50">
              <Package className="h-10 w-10 stroke-[1.5]" />
            </div>
          )}
        </Link>

        {/* Verification badge */}
        {isVerified && (
          <Badge className="absolute left-2 top-2 border-0 bg-success text-success-foreground text-[10px] font-bold rounded-sm px-1.5 py-0.5 shadow-xs">
            Verified
          </Badge>
        )}

        {/* Discount badge */}
        {hasDiscount && (
          <Badge className="absolute left-2 top-2 border-0 bg-destructive text-destructive-foreground font-extrabold text-[10px] rounded-sm px-1.5 py-0.5 shadow-xs">
            {discountPercentage}%
          </Badge>
        )}

        {/* Actions Overlay */}
        <div className="absolute right-2 top-2 flex flex-col gap-1.5 opacity-0 transition-opacity duration-300 group-hover:opacity-100">
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
              className="h-8 w-8 rounded-lg bg-card/90 text-foreground shadow-xs backdrop-blur-xs cursor-pointer hover:-translate-y-0.5 active:translate-y-0 active:scale-95 transition-all duration-200"
            >
              <Heart
                className={`h-4 w-4 transition-colors ${isBookmarked ? "text-destructive fill-destructive" : "text-foreground"}`}
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
              className="h-8 w-8 rounded-lg bg-card/90 text-foreground shadow-xs backdrop-blur-xs cursor-pointer hover:-translate-y-0.5 active:translate-y-0 active:scale-95 transition-all duration-200"
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
      <CardContent className="flex flex-col flex-1 p-3">
        {/* Title */}
        <Link href={href} className="block cursor-pointer flex-1">
          <h3 className="line-clamp-2 min-h-8 text-xs font-semibold leading-relaxed text-foreground group-hover:text-primary transition-colors duration-200">
            {name}
          </h3>
        </Link>

        {/* Price & Discount block */}
        <div className="mt-1.5 space-y-0.5">
          <p className="text-sm font-extrabold text-foreground">
            {displayPrice}
          </p>
          {hasDiscount && originalPrice && (
            <div className="flex items-center gap-1.5">
              <span className="text-[10px] font-bold text-destructive bg-destructive/5 px-1 py-0.2 rounded-xs">
                Hemat {discountPercentage}%
              </span>
              <span className="text-[10px] text-muted-foreground/70 line-through">
                {displayOriginalPrice}
              </span>
            </div>
          )}
        </div>

        {/* Rating and Reviews / Sold Count */}
        {(rating !== undefined || soldCountText) && (
          <div className="mt-2 flex items-center gap-1 text-[10px] text-muted-foreground">
            {rating !== undefined && (
              <div className="flex items-center gap-0.5 text-warning font-semibold">
                <Star className="h-3 w-3 fill-warning text-warning" />
                <span>{rating.toFixed(1)}</span>
              </div>
            )}
            {rating !== undefined && (reviewCount !== undefined || soldCountText) && (
              <span>•</span>
            )}
            {reviewCount !== undefined && (
              <span>{reviewCount} {ulasanLabel}</span>
            )}
            {soldCountText && (
              <span className="font-medium">{soldCountText}</span>
            )}
          </div>
        )}

        {/* Supplier Info */}
        {supplierName && (
          <div className="mt-2">
            {supplierHref ? (
              <Link
                href={supplierHref}
                className="flex cursor-pointer items-center gap-1.5 text-[10px] text-muted-foreground hover:text-primary transition-colors"
              >
                <Store className="h-3 w-3 shrink-0 text-primary/70" />
                <span className="truncate font-semibold">{supplierName}</span>
              </Link>
            ) : (
              <div className="flex items-center gap-1.5 text-[10px] text-muted-foreground">
                <Store className="h-3 w-3 shrink-0 text-primary/70" />
                <span className="truncate font-semibold">{supplierName}</span>
              </div>
            )}
          </div>
        )}

        {/* Tags (Category & MOQ) */}
        {(categoryName || minOrder) && (
          <div className="mt-2.5 flex flex-wrap gap-1">
            {categoryName && (
              <Badge variant="outline" className="border-border text-muted-foreground text-[9px] font-medium rounded-xs px-1.5 py-0">
                {categoryName}
              </Badge>
            )}
            {minOrder && (
              <Badge variant="outline" className="border-border text-muted-foreground text-[9px] font-medium rounded-xs px-1.5 py-0">
                {moqLabel} {minOrder}
              </Badge>
            )}
          </div>
        )}

        {/* Custom Actions Slot (Footer / bottom button) */}
        {customFooter && (
          <div className="mt-3.5 pt-3 border-t border-border/40">
            {customFooter}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
