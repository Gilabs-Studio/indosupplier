"use client";
/* eslint-disable @next/next/no-img-element */

import React from "react";
import { Link } from "@/i18n/routing";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Check,
  GitCompareArrows,
  Heart,
  Package,
  Star,
  Store,
} from "lucide-react";
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

function formatPrice(price: number, currency = "IDR") {
  if (!price) return "Hubungi Supplier";
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency,
    minimumFractionDigits: 0,
  }).format(price);
}

export function PublicProductCard({
  product,
  detailBasePath,
  isAuthenticated,
  isBookmarked,
  isCompared,
  onBookmark,
  onCompare,
}: PublicProductCardProps) {
  const image = product.photos?.[0];

  return (
    <Card className="group overflow-hidden border border-border bg-card shadow-xs transition-all duration-300 hover:-translate-y-0.5 hover:shadow-md">
      <div className="relative aspect-square bg-muted">
        <Link href={`${detailBasePath}/products/${product.id}`} className="block h-full cursor-pointer">
          {image ? (
            <img
              src={image}
              alt={product.name}
              className="h-full w-full object-cover transition-transform duration-500 group-hover:scale-105"
            />
          ) : (
            <div className="flex h-full items-center justify-center text-muted-foreground">
              <Package className="h-10 w-10" />
            </div>
          )}
        </Link>
        {product.supplierVerified && (
          <Badge className="absolute left-2 top-2 border-0 bg-success text-success-foreground text-[10px] font-bold">
            Verified
          </Badge>
        )}
        <div className="absolute right-2 top-2 flex gap-1">
          <Button
            type="button"
            variant="secondary"
            size="icon"
            onClick={(e) => {
              e.preventDefault();
              e.stopPropagation();
              onBookmark();
            }}
            className="h-8 w-8 bg-card/90 text-foreground shadow-xs backdrop-blur cursor-pointer hover:-translate-y-0.5 active:translate-y-0"
            title={isAuthenticated ? "Simpan produk" : "Masuk untuk menyimpan"}
          >
            <Heart className={`h-4 w-4 ${isBookmarked ? "fill-rose-600 text-rose-600" : ""}`} />
          </Button>
          <Button
            type="button"
            variant="secondary"
            size="icon"
            onClick={(e) => {
              e.preventDefault();
              e.stopPropagation();
              onCompare();
            }}
            className="h-8 w-8 bg-card/90 text-foreground shadow-xs backdrop-blur cursor-pointer hover:-translate-y-0.5 active:translate-y-0"
            title={isAuthenticated ? "Bandingkan produk" : "Masuk untuk membandingkan"}
          >
            {isCompared ? <Check className="h-4 w-4 text-success" /> : <GitCompareArrows className="h-4 w-4" />}
          </Button>
        </div>
      </div>

      <CardContent className="space-y-2 p-3">
        <Link href={`${detailBasePath}/products/${product.id}`} className="block cursor-pointer">
          <h3 className="line-clamp-2 min-h-10 text-sm font-semibold leading-5 text-foreground group-hover:text-primary transition-colors">
            {product.name}
          </h3>
        </Link>
        <p className="text-base font-extrabold text-foreground">{formatPrice(product.price, product.currency)}</p>
        <div className="flex items-center gap-1 text-xs text-muted-foreground">
          <Star className="h-3.5 w-3.5 fill-warning text-warning" />
          <span>{product.supplierRating?.toFixed(1) || "0.0"}</span>
          <span>•</span>
          <span>{product.supplierReviewCount || 0} ulasan</span>
        </div>
        <Link
          href={`${detailBasePath}/suppliers/${product.supplierSlug}`}
          className="flex cursor-pointer items-center gap-1.5 text-xs text-muted-foreground hover:text-primary transition-colors"
        >
          <Store className="h-3.5 w-3.5" />
          <span className="truncate font-medium">{product.supplierCompanyName}</span>
        </Link>
        <div className="flex flex-wrap gap-1 text-[11px] text-muted-foreground">
          {product.categoryName && <span className="rounded border border-border px-1.5 py-0.5">{product.categoryName}</span>}
          {product.minOrder && <span className="rounded border border-border px-1.5 py-0.5">MOQ {product.minOrder}</span>}
        </div>
      </CardContent>
    </Card>
  );
}
