"use client";

import React, { useState } from "react";
import Image from "next/image";
import { Link } from "@/i18n/routing";
import { Crown, Heart, Star, CheckCircle2, ShoppingCart, Check } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { formatPrice } from "@/lib/utils";
import type { DemoProductItem } from "../types/demo.types";

interface MarketplaceProductCardProps {
  product: DemoProductItem;
  isEn: boolean;
  isBookmarked: boolean;
  isJustAddedToCart: boolean;
  onToggleBookmark: (id: string) => void;
  onAddToCart: (product: DemoProductItem) => void;
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
  isJustAddedToCart,
  onToggleBookmark,
  onAddToCart,
}: Readonly<MarketplaceProductCardProps>) {
  const formattedPrice = formatPrice(product.price, product.currency);
  const unit = product.unit || "unit";
  const reviewCountFormatted =
    product.supplierReviewCount >= 1000
      ? `${(product.supplierReviewCount / 1000).toFixed(1).replace(".", ",")}rb`
      : `${product.supplierReviewCount}`;

  const fallback = getSmartFallbackImage(product.name, product.categorySlug);
  const preferredPhoto =
    product.photos && product.photos.length > 0 && !product.photos[0].includes("unsplash.com")
      ? product.photos[0]
      : fallback;
  const [hasError, setHasError] = useState(false);
  const imgSrc = hasError ? fallback : preferredPhoto;

  return (
    <div className="group flex flex-col justify-between rounded-xl border border-border bg-card overflow-hidden shadow-2xs transition-all duration-200 hover:-translate-y-0.5 hover:border-primary/50 hover:shadow-md">
      {/* Top Image Container */}
      <div className="relative aspect-square w-full overflow-hidden bg-muted/20 p-2">
        <Link href={`/demo/products/${product.id}`} className="block h-full w-full cursor-pointer relative">
          <Image
            src={imgSrc}
            alt={product.name}
            fill
            sizes="(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 25vw"
            onError={() => {
              setHasError(true);
            }}
            className="object-contain p-1 transition-transform duration-200 group-hover:scale-105"
          />
        </Link>

        {/* Top-Left Power Supplier Badge */}
        {product.isPowerSupplier && (
          <Badge variant="power" className="absolute left-2.5 top-2.5 z-10 gap-1 text-[9px] px-1.5 py-0.5">
            <Crown className="h-3 w-3 fill-warning text-warning" />
            <span>Power Supplier</span>
          </Badge>
        )}

        {/* Top-Right Heart Wishlist Toggle */}
        <button
          type="button"
          onClick={(e) => {
            e.preventDefault();
            e.stopPropagation();
            onToggleBookmark(product.id);
          }}
          aria-label="Wishlist toggle"
          className="absolute right-2.5 top-2.5 z-10 flex h-7 w-7 cursor-pointer items-center justify-center rounded-full bg-background/80 text-muted-foreground backdrop-blur-xs transition-colors hover:text-destructive hover:bg-background"
        >
          <Heart
            className={`h-4 w-4 transition-colors ${
              isBookmarked ? "fill-destructive text-destructive" : "text-muted-foreground"
            }`}
          />
        </button>
      </div>

      {/* Card Information Body */}
      <div className="flex flex-1 flex-col justify-between p-3 space-y-2">
        <div className="space-y-1.5">
          {/* Title */}
          <Link href={`/demo/products/${product.id}`} className="block cursor-pointer">
            <h3 className="line-clamp-2 text-xs sm:text-sm font-bold leading-snug text-foreground group-hover:text-primary transition-colors min-h-[2.5rem]">
              {product.name}
            </h3>
          </Link>

          {/* Supplier Name + Verified Badge */}
          <div className="flex items-center gap-1 text-[11px] text-muted-foreground">
            <span className="truncate font-medium">{product.supplierCompanyName}</span>
            {product.supplierVerified && (
              <CheckCircle2 className="h-3.5 w-3.5 shrink-0 text-success fill-success/20" />
            )}
          </div>

          {/* Rating + Reviews */}
          <div className="flex items-center gap-1 text-[11px] text-muted-foreground">
            <Star className="h-3.5 w-3.5 fill-warning text-warning shrink-0" />
            <span className="font-bold text-foreground">
              {product.supplierRating.toFixed(1)}
            </span>
            <span className="text-muted-foreground/80">({reviewCountFormatted})</span>
          </div>

          {/* Price & Unit */}
          <div className="pt-0.5">
            <span className="text-sm sm:text-base font-extrabold text-foreground">
              {formattedPrice}
            </span>
            <span className="text-[11px] text-muted-foreground ml-1">
              / {unit}
            </span>
          </div>

          {/* Min Order */}
          <p className="text-[11px] text-muted-foreground">
            {isEn ? "Min. Order: " : "Min. Pembelian: "}
            <span className="font-semibold text-foreground">{product.minOrder}</span>
          </p>

          {/* Tag Badges using Badge Component */}
          <div className="flex flex-wrap gap-1 pt-1">
            <Badge variant="tag-success">
              Ready Stock
            </Badge>
            {product.tags
              ?.filter((t) => !t.toLowerCase().includes("ready stock"))
              .map((tag) => (
                <Badge key={tag} variant="tag-primary">
                  {tag}
                </Badge>
              ))}
          </div>
        </div>

        {/* Bottom Cart Button */}
        <div className="pt-2">
          <button
            type="button"
            onClick={() => onAddToCart(product)}
            className={`w-full flex cursor-pointer items-center justify-center gap-1.5 rounded-lg border py-2 text-xs font-bold transition-all duration-200 active:scale-98 ${
              isJustAddedToCart
                ? "bg-primary text-primary-foreground border-primary"
                : "border-primary text-primary hover:bg-primary hover:text-primary-foreground"
            }`}
          >
            {isJustAddedToCart ? (
              <>
                <Check className="h-3.5 w-3.5 stroke-[2.5]" />
                <span>{isEn ? "Added to Cart" : "Ditambahkan"}</span>
              </>
            ) : (
              <>
                <ShoppingCart className="h-3.5 w-3.5" />
                <span>{isEn ? "Add to Cart" : "Tambah ke Keranjang"}</span>
              </>
            )}
          </button>
        </div>
      </div>
    </div>
  );
}
