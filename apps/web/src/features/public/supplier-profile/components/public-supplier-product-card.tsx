"use client";
/* eslint-disable @next/next/no-img-element */

import React from "react";
import { Link } from "@/i18n/routing";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Star, Package, Heart, Check, GitCompareArrows } from "lucide-react";
import type { PublicProductDto } from "@/features/public/search/types";

interface PublicSupplierProductCardProps {
  product: PublicProductDto;
  detailBasePath: "" | "/demo";
  isBookmarked: boolean;
  isCompared: boolean;
  onBookmark: () => void;
  onCompare: () => void;
}

function formatPrice(price: number, currency = "IDR") {
  if (!price) return "Hubungi Supplier";
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency,
    minimumFractionDigits: 0,
  }).format(price);
}

export function PublicSupplierProductCard({
  product,
  detailBasePath,
  isBookmarked,
  isCompared,
  onBookmark,
  onCompare,
}: PublicSupplierProductCardProps) {
  const image = product.photos?.[0];
  
  // Calculate mock discount & slashed price for realistic premium look (just like the reference image)
  const hasDiscount = (product.price || 0) > 0 && (product.id.charCodeAt(0) % 2 === 0);
  const discountPercentage = hasDiscount ? (product.id.charCodeAt(1) % 3 === 0 ? 30 : 50) : 0;
  const originalPrice = hasDiscount ? Math.round((product.price || 0) * (100 / (100 - discountPercentage))) : 0;
  
  // Mock sold counts to align with the Tokopedia UI reference
  const mockSoldCount = product.id.charCodeAt(2) % 2 === 0 
    ? `${(product.id.charCodeAt(3) % 9) + 1}rb+ terjual`
    : `${(product.id.charCodeAt(3) % 80) + 10}+ terjual`;

  return (
    <div className="group relative flex flex-col justify-between overflow-hidden rounded-lg border border-border/40 bg-card p-0 transition-all duration-300 hover:-translate-y-1 hover:border-primary/35 hover:shadow-lg hover:shadow-primary/5">
      {/* Product Image Area */}
      <div className="relative aspect-square w-full bg-muted/30 overflow-hidden">
        <Link 
          href={`${detailBasePath}/products/${product.id}`} 
          className="block h-full w-full cursor-pointer"
        >
          {image ? (
            <img
              src={image}
              alt={product.name}
              className="h-full w-full object-cover transition-transform duration-500 group-hover:scale-103"
            />
          ) : (
            <div className="flex h-full w-full items-center justify-center text-muted-foreground/45">
              <Package className="h-10 w-10 stroke-[1.5]" />
            </div>
          )}
        </Link>

        {/* Promo Discount Tag */}
        {hasDiscount && (
          <Badge className="absolute left-2.5 top-2.5 border-0 bg-rose-500 font-extrabold text-[10px] text-white px-1.5 py-0.5 rounded-sm">
            {discountPercentage}%
          </Badge>
        )}

        {/* Action Overlay Buttons (top right) */}
        <div className="absolute right-2.5 top-2.5 flex flex-col gap-1.5 opacity-0 transition-opacity duration-300 group-hover:opacity-100">
          <Button
            type="button"
            variant="secondary"
            size="icon"
            onClick={(e) => {
              e.preventDefault();
              e.stopPropagation();
              onBookmark();
            }}
            className="h-7 w-7 rounded-lg bg-card/95 text-foreground shadow-sm hover:text-rose-600 backdrop-blur-xs cursor-pointer hover:-translate-y-0.5 active:translate-y-0 transition-all duration-200"
            title="Simpan ke Bookmark"
          >
            <Heart className={`h-3.5 w-3.5 ${isBookmarked ? "fill-rose-500 text-rose-500" : ""}`} />
          </Button>
        </div>
      </div>

      {/* Product Details Info */}
      <div className="flex flex-col flex-1 p-3">
        <Link 
          href={`${detailBasePath}/products/${product.id}`} 
          className="block cursor-pointer flex-1"
        >
          <h3 className="line-clamp-2 min-h-8 text-xs font-medium leading-relaxed text-foreground group-hover:text-primary transition-colors duration-200">
            {product.name}
          </h3>
        </Link>

        {/* Price Section */}
        <div className="mt-1.5 space-y-0.5">
          <p className="text-sm font-black text-foreground">
            {formatPrice(product.price, product.currency)}
          </p>
          {hasDiscount && (
            <div className="flex items-center gap-1.5">
              <span className="text-[10px] font-bold text-rose-500/80 bg-rose-500/5 px-1 py-0.2 rounded-xs">
                Hemat {discountPercentage}%
              </span>
              <span className="text-[10px] text-muted-foreground/75 line-through">
                {formatPrice(originalPrice, product.currency)}
              </span>
            </div>
          )}
        </div>

        {/* Rating and Sold metrics */}
        <div className="mt-2 flex items-center gap-1 text-[10px] text-muted-foreground">
          <div className="flex items-center gap-0.5 text-warning font-semibold">
            <Star className="h-3 w-3 fill-warning text-warning" />
            <span>{product.supplierRating?.toFixed(1) || "0.0"}</span>
          </div>
          <span>•</span>
          <span className="font-medium">{mockSoldCount}</span>
        </div>

        {/* B2B Specs Tag */}
        <div className="mt-2.5 flex flex-wrap gap-1">
          {product.categoryName && (
            <Badge variant="outline" className="border-border/60 text-muted-foreground text-[9px] font-medium rounded-sm px-1.5 py-0">
              {product.categoryName}
            </Badge>
          )}
          {product.minOrder && (
            <Badge variant="outline" className="border-border/60 text-muted-foreground text-[9px] font-medium rounded-sm px-1.5 py-0">
              Min. {product.minOrder}
            </Badge>
          )}
        </div>

        {/* Action CTA Button (+ Bandingkan) */}
        <div className="mt-3.5">
          <Button
            type="button"
            variant="outline"
            onClick={(e) => {
              e.preventDefault();
              e.stopPropagation();
              onCompare();
            }}
            className={`w-full py-1.5 text-[11px] font-bold border rounded-lg cursor-pointer transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-md ${
              isCompared
                ? "bg-success/5 border-success text-success hover:bg-success/10 hover:shadow-success/10"
                : "border-primary/45 text-primary bg-primary/0 hover:bg-primary hover:text-primary-foreground hover:shadow-primary/10"
            }`}
          >
            {isCompared ? (
              <>
                <Check className="h-3 w-3 mr-1 shrink-0" />
                Ditambahkan
              </>
            ) : (
              <>
                <GitCompareArrows className="h-3 w-3 mr-1 shrink-0" />
                + Bandingkan
              </>
            )}
          </Button>
        </div>
      </div>
    </div>
  );
}
