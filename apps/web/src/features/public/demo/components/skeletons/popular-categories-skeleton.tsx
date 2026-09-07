import React from "react";

export function PopularCategoriesSkeleton() {
  return (
    <section className="w-full space-y-3.5">
      {/* Header Skeleton */}
      <div className="flex items-center justify-between">
        <div className="h-5 w-36 rounded-md bg-muted/40 animate-pulse" />
        <div className="h-3.5 w-16 rounded-xs bg-muted/30 animate-pulse" />
      </div>

      {/* 8 Categories Grid Skeleton */}
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4 lg:grid-cols-8">
        {Array.from({ length: 8 }).map((_, idx) => (
          <div
            key={idx}
            className="flex flex-col items-center justify-between rounded-xl border border-border bg-card p-2.5 shadow-2xs animate-pulse text-center"
          >
            {/* Aspect Square Image Placeholder */}
            <div className="aspect-square w-full rounded-lg bg-muted/20 p-2 flex items-center justify-center">
              <div className="h-12 w-12 rounded-md bg-muted/40" />
            </div>

            {/* Category Title Placeholder */}
            <div className="mt-2.5 h-3 w-4/5 rounded-xs bg-muted/40 min-h-[1.75rem] flex items-center justify-center" />
          </div>
        ))}
      </div>
    </section>
  );
}
