"use client";

import React, { useRef, useState, useEffect, useCallback } from "react";
import Image from "next/image";
import { Link } from "@/i18n/routing";
import { ChevronRight, ChevronLeft } from "lucide-react";
import type { DemoCategoryItem } from "../types/demo.types";

interface PopularCategoriesSectionProps {
  categories: DemoCategoryItem[];
  isEn: boolean;
}

export function PopularCategoriesSection({
  categories,
  isEn,
}: Readonly<PopularCategoriesSectionProps>) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const sentinelRef = useRef<HTMLDivElement>(null);

  // Buffer of items for continuous horizontal infinite loading without extra server hits
  const [items, setItems] = useState<DemoCategoryItem[]>(categories);
  const [canScrollLeft, setCanScrollLeft] = useState(false);
  const [canScrollRight, setCanScrollRight] = useState(true);

  // Sync initial items when categories prop changes
  useEffect(() => {
    if (categories.length > 0) {
      setItems(categories);
    }
  }, [categories]);

  // Check scroll bounds
  const updateScrollBounds = useCallback(() => {
    const el = scrollRef.current;
    if (!el) return;
    setCanScrollLeft(el.scrollLeft > 10);
    setCanScrollRight(el.scrollLeft + el.clientWidth < el.scrollWidth - 10);
  }, []);

  useEffect(() => {
    const el = scrollRef.current;
    if (!el) return;
    updateScrollBounds();
    el.addEventListener("scroll", updateScrollBounds, { passive: true });
    window.addEventListener("resize", updateScrollBounds);
    return () => {
      el.removeEventListener("scroll", updateScrollBounds);
      window.removeEventListener("resize", updateScrollBounds);
    };
  }, [updateScrollBounds, items]);

  // Optimized client-side infinite loading:
  // Appends next cycle when reaching end without stressing backend network
  useEffect(() => {
    if (!sentinelRef.current || categories.length === 0) return;

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting) {
          setItems((prev) => {
            // Cap at 40 items to ensure ultra-low DOM footprint
            if (prev.length >= 40) return prev;
            return [...prev, ...categories];
          });
        }
      },
      {
        root: scrollRef.current,
        threshold: 0.1,
      }
    );

    observer.observe(sentinelRef.current);
    return () => observer.disconnect();
  }, [categories]);

  const handleScrollLeft = () => {
    if (!scrollRef.current) return;
    const distance = scrollRef.current.clientWidth * 0.7;
    scrollRef.current.scrollBy({ left: -distance, behavior: "smooth" });
  };

  const handleScrollRight = () => {
    if (!scrollRef.current) return;
    const distance = scrollRef.current.clientWidth * 0.7;
    scrollRef.current.scrollBy({ left: distance, behavior: "smooth" });
  };

  if (categories.length === 0) {
    return null;
  }

  return (
    <section className="w-full space-y-3.5">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h2 className="text-base sm:text-lg font-black tracking-tight text-foreground">
          {isEn ? "Popular Categories" : "Kategori Populer"}
        </h2>
        <Link
          href="/categories"
          className="inline-flex items-center gap-0.5 text-xs font-bold text-primary hover:text-primary/80 transition-colors cursor-pointer"
        >
          <span>{isEn ? "View All" : "Lihat Semua"}</span>
          <ChevronRight className="h-3.5 w-3.5" />
        </Link>
      </div>

      {/* 1 Row Carousel Scroller */}
      <div className="relative group/carousel">
        {/* Scroll Left Button */}
        {canScrollLeft && (
          <button
            type="button"
            onClick={handleScrollLeft}
            className="absolute -left-3 top-1/2 -translate-y-1/2 z-20 hidden sm:flex h-8 w-8 items-center justify-center rounded-full bg-card/95 border border-border shadow-md hover:bg-muted text-foreground transition-all cursor-pointer hover:scale-105 active:scale-95"
            aria-label="Scroll left"
          >
            <ChevronLeft className="h-4 w-4" />
          </button>
        )}

        {/* Single Row Horizontal Scroller Container */}
        <div
          ref={scrollRef}
          className="flex flex-row overflow-x-auto gap-3 pb-2 pt-1 no-scrollbar scroll-smooth snap-x snap-mandatory"
        >
          {items.map((cat, idx) => {
            const name = isEn ? cat.nameEn : cat.nameId;
            return (
              <Link
                key={`${cat.id || cat.slug}-${idx}`}
                href={cat.href}
                className="group shrink-0 w-[125px] sm:w-[145px] snap-start flex cursor-pointer flex-col items-center justify-between rounded-xl border border-border bg-card p-2.5 shadow-2xs transition-all duration-200 hover:-translate-y-0.5 hover:border-primary/50 hover:shadow-md text-center"
              >
                {/* Clean Image Area */}
                <div className="relative aspect-square w-full overflow-hidden rounded-lg bg-muted/20 p-1 flex items-center justify-center">
                  <Image
                    src={cat.image}
                    alt={name}
                    width={140}
                    height={140}
                    className="h-full w-full object-contain transition-transform duration-200 group-hover:scale-105"
                  />
                </div>

                {/* Title */}
                <span className="mt-2 text-[11px] sm:text-xs font-semibold text-foreground line-clamp-2 leading-tight group-hover:text-primary transition-colors min-h-[1.75rem] flex items-center justify-center">
                  {name}
                </span>
              </Link>
            );
          })}

          {/* Infinite Scroll Sentinel */}
          <div ref={sentinelRef} className="w-4 shrink-0 pointer-events-none" />
        </div>

        {/* Scroll Right Button */}
        {canScrollRight && (
          <button
            type="button"
            onClick={handleScrollRight}
            className="absolute -right-3 top-1/2 -translate-y-1/2 z-20 hidden sm:flex h-8 w-8 items-center justify-center rounded-full bg-card/95 border border-border shadow-md hover:bg-muted text-foreground transition-all cursor-pointer hover:scale-105 active:scale-95"
            aria-label="Scroll right"
          >
            <ChevronRight className="h-4 w-4" />
          </button>
        )}
      </div>
    </section>
  );
}
