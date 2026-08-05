"use client";
/* eslint-disable @next/next/no-img-element */

import React, { useRef } from "react";
import { useTranslations } from "next-intl";
import { PublicLayout } from "@/features/public/components/public-layout";
import { Link } from "@/i18n/routing";
import { formatPrice } from "@/lib/utils";
import { AiSearchInput } from "@/features/public/components/ai-search-input";
import { useDemoHome, type EnhancedDemoProduct } from "../hooks/use-demo-home";
import { contentTypeLabels, formatContentDate } from "@/features/content/utils/content-format";
import {
  ArrowUp,
  CheckCircle,
  ChevronLeft,
  ChevronRight,
  MoreHorizontal,
  Package,
  ShieldCheck,
  Star,
  Tag,
} from "lucide-react";

interface DemoHomePageProps {
  locale: string;
}

export function DemoHomePage({ locale }: DemoHomePageProps) {
  const t = useTranslations("public.demoHome");
  const scrollContainerRef = useRef<HTMLDivElement>(null);

  const {
    categoryPills,
    activeCategoryTab,
    setActiveCategoryTab,
    newsTypes,
    activeNewsType,
    setActiveNewsType,
    newsArticles,
    popularProducts,
    isNewsLoading,
    isProductsLoading,
    showBackToTop,
    scrollToTop,
  } = useDemoHome(locale);

  const scrollPillBar = (direction: "left" | "right") => {
    if (scrollContainerRef.current) {
      const scrollAmount = direction === "left" ? -240 : 240;
      scrollContainerRef.current.scrollBy({ left: scrollAmount, behavior: "smooth" });
    }
  };

  return (
    <PublicLayout locale={locale}>
      <div className="min-h-screen bg-background pb-16 font-sans">
        {/* Hero Section */}
        <section
          className="relative flex w-full flex-col items-center overflow-visible bg-cover bg-center bg-no-repeat px-4 pb-12 pt-10 text-center"
          style={{
            backgroundImage: "url('https://images.unsplash.com/photo-1586528116311-ad8dd3c8310d?auto=format&fit=crop&w=1600&q=80')",
          }}
        >
          <div className="absolute inset-0 bg-background/90 backdrop-blur-[0.5px]" />
          <div className="relative z-10 mx-auto flex max-w-4xl flex-col items-center space-y-4">
            <h1 className="font-heading text-2xl font-extrabold leading-tight tracking-tight text-foreground sm:text-3xl md:text-4xl">
              {locale === "en" ? "Discover Verified Indonesian Suppliers" : "Temukan Supplier Terverifikasi Indonesia"}
            </h1>
            <p className="max-w-2xl text-xs leading-relaxed text-muted-foreground sm:text-sm">
              {locale === "en"
                ? "Access a curated directory of Indonesian manufacturers, exporters, and raw material providers."
                : "Akses direktori terkurasi dari produsen, eksportir, dan penyedia bahan baku terbaik Indonesia."}
            </p>
            <div className="w-full max-w-2xl pt-1">
              <AiSearchInput />
            </div>
            <div className="flex flex-wrap justify-center gap-x-6 gap-y-2 pt-1 text-[11px] font-medium text-muted-foreground">
              {[
                locale === "en" ? "Verified Suppliers" : "Supplier Terverifikasi",
                locale === "en" ? "Secure Trading" : "Perdagangan Aman",
                locale === "en" ? "Direct Export" : "Ekspor Langsung",
                locale === "en" ? "Certifications Checked" : "Sertifikasi Dicek",
              ].map((item) => (
                <span key={item} className="flex items-center gap-1.5">
                  <CheckCircle className="h-3.5 w-3.5 shrink-0 text-primary" />
                  {item}
                </span>
              ))}
            </div>
          </div>
        </section>

        {/* Sticky Filter Bar (Scrollbar-Free) */}
        <div className="sticky top-0 z-40 border-b border-border/40 bg-background/95 backdrop-blur-md py-2 shadow-2xs">
          <div className="mx-auto flex max-w-7xl items-center px-4 sm:px-6 lg:px-8">
            <button
              onClick={() => scrollPillBar("left")}
              className="mr-1.5 hidden h-7 w-7 shrink-0 cursor-pointer items-center justify-center rounded-full text-muted-foreground transition-colors hover:bg-muted hover:text-foreground md:flex"
              aria-label="Scroll left"
            >
              <ChevronLeft className="h-4 w-4" />
            </button>

            {/* Scrollable Container with Hidden Scrollbars */}
            <div
              ref={scrollContainerRef}
              className="flex flex-1 items-center gap-2 overflow-x-auto py-1 scroll-smooth whitespace-nowrap scrollbar-none [&::-webkit-scrollbar]:hidden [-ms-overflow-style:none] [scrollbar-width:none]"
            >
              {categoryPills.map((pill) => {
                const isActive = activeCategoryTab === pill.id;
                const label = locale === "en" ? pill.labelEn : pill.labelId;
                return (
                  <button
                    key={pill.id}
                    onClick={() => setActiveCategoryTab(pill.id)}
                    className={`flex cursor-pointer items-center gap-1.5 rounded-full px-3.5 py-1 text-xs transition-colors ${
                      isActive
                        ? "bg-primary text-primary-foreground font-semibold shadow-xs"
                        : "bg-muted/50 text-muted-foreground hover:bg-muted hover:text-foreground border border-border/30"
                    }`}
                  >
                    {pill.badge && (
                      <span
                        className={`rounded-xs px-1 text-[9px] font-black uppercase tracking-wider ${
                          isActive ? "bg-primary-foreground text-primary" : "bg-destructive text-destructive-foreground"
                        }`}
                      >
                        {pill.badge}
                      </span>
                    )}
                    <span>{label}</span>
                  </button>
                );
              })}
            </div>

            <button
              onClick={() => scrollPillBar("right")}
              className="ml-1.5 hidden h-7 w-7 shrink-0 cursor-pointer items-center justify-center rounded-full text-muted-foreground transition-colors hover:bg-muted hover:text-foreground md:flex"
              aria-label="Scroll right"
            >
              <ChevronRight className="h-4 w-4" />
            </button>
          </div>
        </div>

        {/* Compact Popular Products Grid (Tokopedia Style 6-Column Marketplace Grid) */}
        <section className="mx-auto mt-6 max-w-7xl space-y-3.5 px-4 sm:px-6 lg:px-8">
          <SectionHeader title={t("produkTitle")} href="/demo/search" action={t("lihatSemuaProduk")} />

          {isProductsLoading ? (
            <CardGridSkeleton />
          ) : (
            <div className="grid grid-cols-2 gap-2.5 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
              {popularProducts.map((product) => (
                <CompactProductCard key={product.id} product={product} locale={locale} />
              ))}
              {popularProducts.length === 0 && <EmptySection label="Belum ada produk aktif pada kategori ini." />}
            </div>
          )}
        </section>

        {/* News & Articles Section - Standardized 16:9 Image Ratios */}
        <section id="berita" className="mx-auto mt-10 max-w-7xl space-y-3.5 px-4 sm:px-6 lg:px-8">
          <SectionHeader title={t("beritaTitle")} href="/demo/help" action={t("bacaSemuaBerita")} />
          <div className="flex flex-wrap gap-1.5">
            {newsTypes.map((type) => (
              <button
                key={type}
                onClick={() => setActiveNewsType(type)}
                className={`cursor-pointer rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                  activeNewsType === type
                    ? "bg-primary text-primary-foreground font-semibold"
                    : "bg-muted/60 text-muted-foreground hover:bg-muted hover:text-foreground"
                }`}
              >
                {contentTypeLabels[type][locale === "en" ? "en" : "id"]}
              </button>
            ))}
          </div>

          {isNewsLoading ? (
            <CardGridSkeleton columns="md:grid-cols-4" />
          ) : (
            <div className="grid grid-cols-1 gap-3.5 sm:grid-cols-2 lg:grid-cols-4">
              {newsArticles.map((article) => (
                <Link
                  key={article.id}
                  href={`/demo/help`}
                  className="group flex cursor-pointer flex-col justify-between rounded-lg bg-card p-0 transition-colors duration-150 hover:bg-card/80"
                >
                  <div>
                    {/* Fixed 16:9 Aspect Ratio for News Thumbnails */}
                    <div className="relative aspect-[16/9] w-full overflow-hidden rounded-t-lg bg-muted/40">
                      {article.imageUrl ? (
                        <img src={article.imageUrl} alt={article.title} className="h-full w-full object-cover transition-opacity duration-200 group-hover:opacity-90" />
                      ) : (
                        <div className="flex h-full items-center justify-center">
                          <Package className="h-7 w-7 text-muted-foreground/40" />
                        </div>
                      )}
                    </div>
                    <div className="space-y-1 p-3">
                      <h3 className="line-clamp-2 text-xs font-bold leading-snug text-foreground transition-colors group-hover:text-primary">
                        {article.title}
                      </h3>
                      <p className="line-clamp-2 text-[11px] leading-relaxed text-muted-foreground">{article.excerpt}</p>
                    </div>
                  </div>
                  <div className="flex items-center px-3 pb-3 pt-0 text-[10px] text-muted-foreground">
                    <span className="mr-1 font-medium text-primary">{article.authorName || "IndoSupplier"}</span>
                    <span>• {formatContentDate(article, locale)}</span>
                  </div>
                </Link>
              ))}
              {newsArticles.length === 0 && <EmptySection label="Belum ada konten pada kategori ini." />}
            </div>
          )}
        </section>

        {/* Minimal Floating Back to Top Button */}
        <button
          onClick={scrollToTop}
          className={`fixed bottom-6 right-6 z-50 flex h-9 w-9 cursor-pointer items-center justify-center rounded-full bg-primary text-primary-foreground shadow-md transition-all duration-200 hover:bg-primary/90 ${
            showBackToTop ? "pointer-events-auto opacity-100" : "pointer-events-none opacity-0"
          }`}
          aria-label={t("kembaliKeAtas")}
        >
          <ArrowUp className="h-4 w-4 stroke-[2.5]" />
        </button>
      </div>
    </PublicLayout>
  );
}

function SectionHeader({ title, href, action }: { title: string; href: string; action: string }) {
  return (
    <div className="flex items-center justify-between pb-1">
      <h2 className="font-heading text-sm font-bold text-foreground sm:text-base">{title}</h2>
      <Link href={href} className="cursor-pointer text-xs font-semibold text-primary transition-colors hover:text-primary/80">
        {action}
      </Link>
    </div>
  );
}

{/* Compact Marketplace Borderless Product Card (Tokopedia 6-Column Style) */}
function CompactProductCard({ product, locale }: { product: EnhancedDemoProduct; locale: string }) {
  const formattedPrice = product.price ? formatPrice(product.price, product.currency) : undefined;
  const displayPrice = formattedPrice || (locale === "en" ? "Contact Supplier" : "Hubungi Supplier");
  const displayOriginalPrice = product.originalPrice ? formatPrice(product.originalPrice, product.currency) : undefined;

  return (
    <Link
      href={`/demo/products/${product.id}`}
      className="group relative flex cursor-pointer flex-col justify-between rounded-lg bg-card p-0 transition-colors duration-150"
    >
      <div>
        {/* Square Aspect Image Container */}
        <div className="relative aspect-square w-full overflow-hidden rounded-t-lg bg-muted/30">
          {product.photos?.[0] ? (
            <img
              src={product.photos[0]}
              alt={product.name}
              className="h-full w-full object-cover transition-opacity duration-200 group-hover:opacity-95"
            />
          ) : (
            <div className="flex h-full w-full items-center justify-center">
              <Package className="h-7 w-7 text-muted-foreground/30" />
            </div>
          )}

          {/* Top Left Discount Badge */}
          {product.discountPercentage && product.discountPercentage > 0 && (
            <div className="absolute left-1.5 top-1.5 z-10 rounded-xs bg-destructive px-1.5 py-0.5 text-[9px] font-black text-destructive-foreground">
              {product.discountPercentage}%
            </div>
          )}

          {/* Bottom Left Promo Badge Banner */}
          {product.badgeOverlay && (
            <div className="absolute bottom-1.5 left-1.5 z-10 flex items-center gap-1 rounded-xs bg-primary/95 px-1 py-0.5 text-[8px] font-bold text-primary-foreground shadow-xs">
              <Tag className="h-2 w-2" />
              <span>{product.badgeOverlay}</span>
            </div>
          )}
        </div>

        {/* Card Body Information - Compact Spacing */}
        <div className="space-y-1 p-2 pt-1.5 pb-1.5">
          {/* Title - Line clamp 2 */}
          <h3 className="line-clamp-2 min-h-[2rem] text-[11px] font-normal leading-snug text-foreground transition-colors duration-150 group-hover:text-primary sm:text-xs">
            {product.name}
          </h3>

          {/* Price & Original Price */}
          <div className="flex items-baseline gap-1 pt-0.5">
            <span className="text-xs font-black tracking-tight text-destructive sm:text-sm">
              {displayPrice}
            </span>
            {product.originalPrice && (
              <span className="text-[9px] text-muted-foreground/60 line-through">
                {displayOriginalPrice}
              </span>
            )}
          </div>

          {/* Hemat Bonus Tag */}
          {product.promoBadge && (
            <div>
              <span className="inline-block rounded-xs bg-destructive/10 px-1 py-0.5 text-[9px] font-semibold text-destructive">
                {product.promoBadge}
              </span>
            </div>
          )}

          {/* Rating & Sales Count */}
          <div className="flex items-center gap-1 pt-0.5 text-[9px] text-muted-foreground sm:text-[10px]">
            <div className="flex items-center gap-0.5 font-semibold text-warning">
              <Star className="h-2.5 w-2.5 fill-warning text-warning shrink-0" />
              <span>{(product.supplierRating || 5.0).toFixed(1)}</span>
            </div>
            <span>•</span>
            <span className="truncate">{product.salesCountText || "100+ terjual"}</span>
          </div>

          {/* Verified Seller Tag & Location */}
          <div className="flex items-center justify-between pt-0.5 text-[9px] text-muted-foreground sm:text-[10px]">
            <div className="flex items-center gap-1 truncate font-medium">
              {product.supplierVerified && (
                <ShieldCheck className="h-2.5 w-2.5 shrink-0 text-success" />
              )}
              <span className="truncate">{product.locationTag || product.supplierCompanyName}</span>
            </div>

            <button
              type="button"
              onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
              }}
              aria-label="More options"
              className="cursor-pointer p-0.5 text-muted-foreground/40 transition-colors hover:text-foreground"
            >
              <MoreHorizontal className="h-3 w-3" />
            </button>
          </div>
        </div>
      </div>
    </Link>
  );
}

function CardGridSkeleton({ columns = "lg:grid-cols-6" }: { columns?: string }) {
  return (
    <div className={`grid grid-cols-2 gap-2.5 sm:grid-cols-3 ${columns}`}>
      {Array.from({ length: 6 }).map((_, index) => (
        <div key={index} className="h-52 animate-pulse rounded-lg bg-muted/40" />
      ))}
    </div>
  );
}

function EmptySection({ label }: { label: string }) {
  return (
    <div className="col-span-full rounded-lg bg-card py-6 text-center text-xs text-muted-foreground">
      {label}
    </div>
  );
}
