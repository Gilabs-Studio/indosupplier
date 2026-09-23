"use client";

import React, { useState, useMemo } from "react";
import Image from "next/image";
import { useTranslations } from "next-intl";
import { useQuery } from "@tanstack/react-query";
import { Link } from "@/i18n/routing";
import { PublicLayout } from "@/features/public/components/public-layout";
import { searchService } from "@/features/public/search/services/search-service";
import { getCategoryThumbnail } from "../utils/category-image";
import { Search, ArrowUpRight, ArrowLeft } from "lucide-react";

interface PublicCategoryPageProps {
  locale: string;
}

export function PublicCategoryPage({ locale }: Readonly<PublicCategoryPageProps>) {
  const tCat = useTranslations("public.categories");
  const [searchQuery, setSearchQuery] = useState("");

  const { data: categories = [], isLoading } = useQuery({
    queryKey: ["public", "categories", "full"],
    queryFn: () => searchService.getCategories(),
  });

  const filteredCategories = useMemo(() => {
    if (!searchQuery.trim()) return categories;
    const q = searchQuery.toLowerCase().trim();
    return categories.filter(
      (cat) =>
        cat.name.toLowerCase().includes(q) ||
        cat.description?.toLowerCase().includes(q) ||
        cat.slug.toLowerCase().includes(q)
    );
  }, [categories, searchQuery]);

  return (
    <PublicLayout locale={locale}>
      <main className="min-h-screen bg-background text-foreground">
        {/* Minimalist Corporate Header */}
        <section className="border-b border-border/80 bg-card/40 py-10 md:py-14">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 space-y-3">
            <div className="flex items-center justify-between gap-4">
              <span className="font-mono text-[11px] font-medium tracking-widest text-muted-foreground uppercase">
                {tCat("directoryTag")}
              </span>
              {categories.length > 0 && (
                <span className="font-mono text-[11px] text-muted-foreground">
                  {tCat("totalIndexed", { count: categories.length })}
                </span>
              )}
            </div>

            <h1 className="text-2xl sm:text-3xl font-semibold tracking-tight text-foreground font-sans">
              {tCat("title")}
            </h1>

            <p className="text-xs sm:text-sm text-muted-foreground max-w-2xl leading-relaxed">
              {tCat("subtitle")}
            </p>

            {/* Corporate Hairline Search Input */}
            <div className="pt-2 max-w-md">
              <div className="relative flex items-center">
                <Search className="absolute left-3.5 h-3.5 w-3.5 text-muted-foreground pointer-events-none" />
                <input
                  type="text"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder={tCat("searchPlaceholder")}
                  className="w-full h-9 pl-9 pr-4 bg-muted/20 border border-border/80 rounded-md text-xs sm:text-sm text-foreground placeholder:text-muted-foreground/70 focus:bg-background focus:border-foreground/40 outline-hidden transition-colors"
                />
              </div>
            </div>
          </div>
        </section>

        {/* Minimalist Grid Section */}
        <section className="py-8 md:py-12">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            {isLoading ? (
              <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                {Array.from({ length: 8 }).map((_, idx) => (
                  <div
                    key={idx}
                    className="h-64 rounded-md border border-border/60 bg-muted/20 animate-pulse"
                  />
                ))}
              </div>
            ) : filteredCategories.length === 0 ? (
              <div className="text-center py-16 border border-border/60 rounded-md bg-card/50 max-w-md mx-auto p-6 space-y-2">
                <p className="text-xs font-mono text-muted-foreground uppercase tracking-wider">
                  404 // NOT_FOUND
                </p>
                <h3 className="text-sm font-semibold text-foreground">{tCat("empty")}</h3>
                <p className="text-xs text-muted-foreground">
                  Periksa ejaan kata kunci atau hapus filter pencarian.
                </p>
              </div>
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                {filteredCategories.map((cat, idx) => {
                  const thumbnail = getCategoryThumbnail(
                    cat.slug,
                    cat.icon_url || cat.iconUrl,
                    cat.name
                  );
                  const productCount = cat.product_count ?? cat.productCount ?? 0;
                  const supplierCount = cat.supplier_count ?? cat.supplierCount ?? 0;

                  return (
                    <Link
                      key={cat.id || cat.slug}
                      href={`/search?category=${encodeURIComponent(cat.slug)}`}
                      className="group flex flex-col justify-between rounded-lg border border-border/70 bg-card p-4 transition-all duration-200 hover:border-foreground/40 hover:bg-muted/5 cursor-pointer"
                    >
                      <div>
                        {/* Technical Index Header */}
                        <div className="flex items-center justify-between pb-3 text-[10px] font-mono text-muted-foreground">
                          <span>SEC // {String(idx + 1).padStart(2, "0")}</span>
                          <ArrowUpRight className="h-3.5 w-3.5 opacity-40 group-hover:opacity-100 group-hover:translate-x-0.5 group-hover:-translate-y-0.5 transition-all text-foreground" />
                        </div>

                        {/* Clean Product/Category Artwork Display */}
                        <div className="relative aspect-16/10 w-full overflow-hidden rounded-md bg-muted/15 border border-border/40 p-2.5 flex items-center justify-center mb-3">
                          <Image
                            src={thumbnail}
                            alt={cat.name}
                            width={180}
                            height={115}
                            className="h-full w-full object-contain transition-transform duration-200 group-hover:scale-105"
                            onError={(e) => {
                              const target = e.currentTarget;
                              if (target.src !== "/images/categories/cat-bahan-baku.webp") {
                                target.src = "/images/categories/cat-bahan-baku.webp";
                              }
                            }}
                          />
                        </div>

                        {/* Information Section */}
                        <div className="space-y-1">
                          <h2 className="text-sm font-semibold text-foreground group-hover:text-primary transition-colors tracking-tight line-clamp-1">
                            {cat.name}
                          </h2>
                          <p className="text-[11px] text-muted-foreground line-clamp-2 leading-relaxed font-normal">
                            {cat.description || "Klasifikasi pengadaan terstandarisasi untuk kebutuhan B2B industri."}
                          </p>
                        </div>
                      </div>

                      {/* Clean Metric Row (No badges, purely minimalist typography) */}
                      <div className="pt-3 mt-4 border-t border-border/40 flex items-center justify-between text-[11px] font-mono text-muted-foreground">
                        <span>
                          {productCount} {tCat("productCountSuffix")}
                        </span>
                        {supplierCount > 0 && (
                          <span>· {supplierCount} {tCat("countSuffix")}</span>
                        )}
                      </div>
                    </Link>
                  );
                })}
              </div>
            )}
          </div>
        </section>
      </main>
    </PublicLayout>
  );
}
