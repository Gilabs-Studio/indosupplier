import React from "react";

export function HeroBannerSkeleton() {
  return (
    <section className="relative w-full overflow-hidden rounded-xl shadow-xs transition-all duration-300">
      <div className="relative aspect-[894/192] w-full overflow-hidden rounded-xl bg-muted/40 animate-pulse flex items-center justify-center">
        <div className="h-6 sm:h-8 w-48 sm:w-64 rounded-md bg-muted/60" />
      </div>
    </section>
  );
}
