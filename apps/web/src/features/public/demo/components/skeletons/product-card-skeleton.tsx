import React from "react";
import { Package, Star } from "lucide-react";

export function ProductCardSkeleton() {
  return (
    <div className="group flex flex-col justify-between rounded-xl border border-border bg-card overflow-hidden shadow-2xs animate-pulse">
      {/* 1. Full-Bleed Image Container Skeleton */}
      <div className="relative aspect-square w-full overflow-hidden bg-muted/30 flex items-center justify-center">
        <Package className="h-9 w-9 text-muted-foreground/20 stroke-[1.5]" />
        
        {/* Bookmark circle outline */}
        <div className="absolute right-2 top-2 h-7 w-7 rounded-full bg-background/60 backdrop-blur-xs" />
      </div>

      {/* 2. Card Information Body Skeleton */}
      <div className="flex flex-1 flex-col justify-between p-2.5 sm:p-3 space-y-2">
        <div className="space-y-1.5">
          {/* Title: 2 lines matching line-clamp-2 */}
          <div className="space-y-1.5 min-h-[2.25rem]">
            <div className="h-3.5 w-11/12 rounded-xs bg-muted/50" />
            <div className="h-3.5 w-3/5 rounded-xs bg-muted/40" />
          </div>

          {/* Price & Unit */}
          <div className="pt-1 flex items-baseline gap-1.5">
            <div className="h-4.5 w-24 rounded-xs bg-muted/60" />
            <div className="h-3 w-8 rounded-xs bg-muted/30" />
          </div>

          {/* Ready Stock Badge */}
          <div className="pt-0.5">
            <div className="h-4 w-16 rounded-xs bg-success/20 border border-success/30" />
          </div>

          {/* Rating & Terjual */}
          <div className="flex items-center gap-1.5 pt-0.5">
            <Star className="h-3.5 w-3.5 fill-warning/30 text-warning/30 shrink-0" />
            <div className="h-3 w-6 rounded-xs bg-muted/50" />
            <span className="text-[11px] text-muted-foreground/30">•</span>
            <div className="h-3 w-16 rounded-xs bg-muted/30" />
          </div>

          {/* Supplier Name & Location */}
          <div className="h-5 flex items-center gap-1.5 pt-1">
            <div className="h-3 w-3 rounded-full bg-muted/40 shrink-0" />
            <div className="h-3 w-28 rounded-xs bg-muted/40" />
          </div>
        </div>
      </div>
    </div>
  );
}

interface PopularProductsSkeletonProps {
  count?: number;
}

export function PopularProductsSkeleton({ count = 12 }: Readonly<PopularProductsSkeletonProps>) {
  return (
    <div className="grid grid-cols-2 gap-3 sm:gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
      {Array.from({ length: count }).map((_, idx) => (
        <ProductCardSkeleton key={idx} />
      ))}
    </div>
  );
}
