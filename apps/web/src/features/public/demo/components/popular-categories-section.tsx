"use client";

import React from "react";
import Image from "next/image";
import { Link } from "@/i18n/routing";
import { ChevronRight } from "lucide-react";
import type { DemoCategoryItem } from "../types/demo.types";

interface PopularCategoriesSectionProps {
  categories: DemoCategoryItem[];
  isEn: boolean;
}

export function PopularCategoriesSection({ categories, isEn }: Readonly<PopularCategoriesSectionProps>) {
  return (
    <section className="w-full space-y-3.5">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h2 className="text-base sm:text-lg font-black tracking-tight text-foreground">
          {isEn ? "Popular Categories" : "Kategori Populer"}
        </h2>
        <Link
          href="/demo/search"
          className="inline-flex items-center gap-0.5 text-xs font-bold text-primary hover:text-primary/80 transition-colors cursor-pointer"
        >
          <span>{isEn ? "View All" : "Lihat Semua"}</span>
          <ChevronRight className="h-3.5 w-3.5" />
        </Link>
      </div>

      {/* 8 Categories Grid */}
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4 lg:grid-cols-8">
        {categories.map((cat) => {
          const name = isEn ? cat.nameEn : cat.nameId;

          return (
            <Link
              key={cat.id}
              href={cat.href}
              className="group flex cursor-pointer flex-col items-center justify-between rounded-xl border border-border bg-card p-2.5 shadow-2xs transition-all duration-200 hover:-translate-y-0.5 hover:border-primary/50 hover:shadow-md text-center"
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
      </div>
    </section>
  );
}
