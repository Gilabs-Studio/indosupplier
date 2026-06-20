"use client";
/* eslint-disable @next/next/no-img-element */

import React from "react";
import Image from "next/image";
import { useTranslations } from "next-intl";
import { PublicLayout } from "@/features/public/components/public-layout";
import { Link } from "@/i18n/routing";
import { formatPrice } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter } from "@/components/ui/card";
import { AiSearchInput } from "@/features/public/components/ai-search-input";
import { useDemoHome } from "../hooks/use-demo-home";
import { contentTypeLabels, formatContentDate, formatViewCount } from "@/features/content/utils/content-format";
import {
  ArrowUp,
  CheckCircle,
  ChevronRight,
  GitCompareArrows,
  Package,
  Play,
  ShieldCheck,
  Star,
  Store,
} from "lucide-react";

interface DemoHomePageProps {
  locale: string;
}



export function DemoHomePage({ locale }: DemoHomePageProps) {
  const t = useTranslations("public.demoHome");
  const {
    newsTypes,
    activeNewsType,
    setActiveNewsType,
    newsArticles,
    videos,
    popularProducts,
    comparisonCards,
    isNewsLoading,
    isVideosLoading,
    isProductsLoading,
    isSuppliersLoading,
    showBackToTop,
    scrollToTop,
  } = useDemoHome(locale);

  return (
    <PublicLayout locale={locale}>
      <div className="min-h-screen bg-background pb-16 font-sans">
        <section
          className="relative flex w-full flex-col items-center overflow-visible bg-cover bg-center bg-no-repeat px-4 pb-20 pt-14 text-center"
          style={{
            backgroundImage: "url('https://images.unsplash.com/photo-1586528116311-ad8dd3c8310d?auto=format&fit=crop&w=1600&q=80')",
          }}
        >
          <div className="absolute inset-0 bg-background/90 backdrop-blur-[0.5px]" />
          <div className="relative z-10 mx-auto flex max-w-4xl flex-col items-center space-y-5">
            <h1 className="font-heading text-2xl font-extrabold leading-tight tracking-tight text-foreground sm:text-3xl md:text-4xl">
              {locale === "en" ? "Discover Verified Indonesian Suppliers" : "Temukan Supplier Terverifikasi Indonesia"}
            </h1>
            <p className="max-w-2xl text-xs leading-relaxed text-muted-foreground sm:text-sm">
              {locale === "en"
                ? "Access a curated directory of Indonesian manufacturers, exporters, and raw material providers."
                : "Akses direktori terkurasi dari produsen, eksportir, dan penyedia bahan baku terbaik Indonesia."}
            </p>
            <div className="w-full max-w-2xl pt-2">
              <AiSearchInput locale={locale} />
            </div>
            <div className="flex flex-wrap justify-center gap-x-6 gap-y-2 pt-2 text-[11px] font-medium text-muted-foreground">
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

        <div className="relative z-20 mx-auto -mt-12 max-w-7xl px-4 sm:px-6 lg:px-8">
          <Card className="border border-border bg-card text-card-foreground shadow-md">
            <CardContent className="space-y-6 p-6 sm:p-8">
              <h2 className="font-heading text-xl font-bold tracking-tight text-foreground sm:text-2xl">
                {t("cariYangTerbaik")}
              </h2>
              <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-6">
                {[
                  { label: t("supplierBaru"), sub: t("subSupplierBaru"), image: "/images/categories/cat-new-supplier.png", link: "/demo/search" },
                  { label: t("supplierPremium"), sub: t("subSupplierPremium"), image: "/images/categories/cat-premium-supplier.png", link: "/demo/search?verified=true" },
                  { label: t("komoditasTani"), sub: t("subKomoditasTani"), image: "/images/categories/cat-agriculture.png", link: "/demo/search?query=kopi" },
                  { label: t("bahanBaku"), sub: t("subBahanBaku"), image: "/images/categories/cat-raw-material.png", link: "/demo/search?query=bahan%20baku" },
                  { label: t("beritaB2B"), sub: t("subBeritaB2B"), image: "/images/categories/cat-news.png", link: "#berita" },
                  { label: t("bandingkan"), sub: t("subBandingkan"), image: "/images/categories/cat-compare.png", link: "/compare" },
                ].map((item) => (
                  <Link
                    key={item.label}
                    href={item.link}
                    className="group flex cursor-pointer flex-col items-center justify-between rounded-lg border border-border bg-card p-4 text-center transition-all duration-300 hover:-translate-y-0.5 hover:border-primary hover:shadow-lg hover:shadow-primary/5 active:translate-y-0"
                  >
                    <div className="relative mb-2 flex h-16 w-16 items-center justify-center transition-transform duration-300 group-hover:scale-105">
                      <Image src={item.image} alt={item.label} width={56} height={56} className="object-contain" />
                    </div>
                    <div className="space-y-1">
                      <span className="block text-xs font-bold leading-tight text-foreground transition-colors group-hover:text-primary sm:text-sm">
                        {item.label}
                      </span>
                      <span className="block text-[10px] leading-none text-muted-foreground">{item.sub}</span>
                    </div>
                  </Link>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>

        <section id="berita" className="mx-auto mt-16 max-w-7xl space-y-6 px-4 sm:px-6 lg:px-8">
          <SectionHeader title={t("beritaTitle")} href="/demo/help" action={t("bacaSemuaBerita")} />
          <div className="flex flex-wrap gap-2">
            {newsTypes.map((type) => (
              <button
                key={type}
                onClick={() => setActiveNewsType(type)}
                className={`cursor-pointer rounded border px-4 py-1.5 text-xs font-semibold transition-all ${
                  activeNewsType === type
                    ? "border-primary/20 bg-primary/10 text-primary shadow-xs"
                    : "border-border bg-card text-muted-foreground hover:bg-muted"
                }`}
              >
                {contentTypeLabels[type][locale === "en" ? "en" : "id"]}
              </button>
            ))}
          </div>

          {isNewsLoading ? (
            <CardGridSkeleton />
          ) : (
            <div className="relative">
              <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
                {newsArticles.map((article) => (
                  <Card key={article.id} className="group flex cursor-pointer flex-col justify-between overflow-hidden rounded-lg border border-border bg-card shadow-xs transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg">
                    <div>
                      <div className="relative h-[145px] overflow-hidden bg-muted">
                        {article.imageUrl ? (
                          <img src={article.imageUrl} alt={article.title} className="h-full w-full object-cover transition-transform duration-500 group-hover:scale-105" />
                        ) : (
                          <div className="flex h-full items-center justify-center">
                            <Package className="h-8 w-8 text-muted-foreground" />
                          </div>
                        )}
                      </div>
                      <CardContent className="space-y-2 p-4">
                        <h3 className="line-clamp-2 text-xs font-bold leading-snug text-foreground transition-colors group-hover:text-primary sm:text-sm">
                          {article.title}
                        </h3>
                        <p className="line-clamp-2 text-[11px] leading-relaxed text-muted-foreground">{article.excerpt}</p>
                      </CardContent>
                    </div>
                    <CardFooter className="flex items-center border-t border-border p-4 pt-0 text-[10px] text-muted-foreground">
                      <span className="mr-1.5 font-medium text-primary">{article.authorName || "IndoSupplier"}</span>
                      <span>• {formatContentDate(article, locale)}</span>
                    </CardFooter>
                  </Card>
                ))}
                {newsArticles.length === 0 && <EmptySection label="Belum ada konten pada kategori ini." />}
              </div>
              {newsArticles.length > 0 && <CarouselArrow />}
            </div>
          )}
        </section>

        <section className="mx-auto mt-16 max-w-7xl space-y-6 px-4 sm:px-6 lg:px-8">
          <SectionHeader title={t("produkTitle")} href="/demo/search" action={t("lihatSemuaProduk")} />
          {isProductsLoading ? (
            <CardGridSkeleton />
          ) : (
            <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
              {popularProducts.map((product) => (
                <Link key={product.id} href={`/demo/products/${product.id}`} className="group flex cursor-pointer flex-col justify-between overflow-hidden rounded-lg border border-border bg-card shadow-xs transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg">
                  <div>
                    <div className="h-[155px] overflow-hidden bg-muted">
                      {product.photos?.[0] ? (
                        <img src={product.photos[0]} alt={product.name} className="h-full w-full object-cover transition-transform duration-500 group-hover:scale-105" />
                      ) : (
                        <div className="flex h-full items-center justify-center">
                          <Package className="h-8 w-8 text-muted-foreground" />
                        </div>
                      )}
                    </div>
                    <CardContent className="space-y-2 p-4">
                      <h3 className="line-clamp-2 text-xs font-bold leading-snug text-foreground transition-colors group-hover:text-primary sm:text-sm">
                        {product.name}
                      </h3>
                      <div className="space-y-1 text-[11px] text-muted-foreground">
                        <p className="truncate font-semibold text-foreground/80">{product.supplierCompanyName}</p>
                        <p>MOQ: <span className="font-bold text-foreground">{product.minOrder || "Nego"}</span></p>
                      </div>
                    </CardContent>
                  </div>
                  <CardFooter className="flex items-center justify-between border-t border-border p-4 pt-0 text-xs">
                    <span className="font-bold text-primary">
                      {formatPrice(product.price, product.currency) || (locale === "en" ? "Contact Supplier" : "Hubungi Supplier")}
                    </span>
                    <span className="rounded-sm border border-border px-1.5 py-0.5 text-[10px] text-muted-foreground">RFQ</span>
                  </CardFooter>
                </Link>
              ))}
              {popularProducts.length === 0 && <EmptySection label="Belum ada produk aktif." />}
            </div>
          )}
        </section>

        <section className="mx-auto mt-16 max-w-7xl space-y-6 px-4 sm:px-6 lg:px-8">
          <SectionHeader title={t("komparasiTitle")} href="/compare" action={t("bandingkanLebihBanyak")} />
          {isSuppliersLoading ? (
            <CardGridSkeleton columns="md:grid-cols-3" />
          ) : (
            <div className="grid grid-cols-1 gap-8 md:grid-cols-3">
              {comparisonCards.map((card) => (
                <Card key={card.id} className="flex flex-col justify-between overflow-hidden rounded-lg border border-border bg-card shadow-xs transition-all duration-300 hover:shadow-lg">
                  <CardContent className="space-y-4 p-5">
                    <h3 className="border-b border-border pb-2 text-center text-xs font-bold text-foreground sm:text-sm">
                      {card.title}
                    </h3>
                    <div className="relative flex items-center justify-between pt-2 text-center">
                      <SupplierMini supplier={card.first} />
                      <div className="absolute left-1/2 top-1/2 z-10 flex h-7 w-7 -translate-x-1/2 -translate-y-1/2 items-center justify-center rounded border border-border bg-muted text-[9px] font-black text-muted-foreground shadow-xs">
                        VS
                      </div>
                      <SupplierMini supplier={card.second} />
                    </div>
                    <div className="space-y-2 border-t border-border pt-2 text-[11px]">
                      {[
                        ["Kategori", card.first.keyProducts?.[0] || "-", card.second.keyProducts?.[0] || "-"],
                        ["Respons", `${Math.round(card.first.responseRate || 0)}%`, `${Math.round(card.second.responseRate || 0)}%`],
                        ["Rating", card.first.rating?.toFixed(1) || "0.0", card.second.rating?.toFixed(1) || "0.0"],
                      ].map(([label, left, right]) => (
                        <div key={label} className="flex items-center justify-between border-b border-dotted border-border py-1 last:border-0">
                          <span className="w-[32%] truncate text-left text-muted-foreground">{left}</span>
                          <span className="w-[36%] text-center text-[9px] font-bold uppercase tracking-wider text-muted-foreground">{label}</span>
                          <span className="w-[32%] truncate text-right text-muted-foreground">{right}</span>
                        </div>
                      ))}
                    </div>
                  </CardContent>
                  <div className="p-5 pt-0">
                    <Button asChild variant="outline" className="w-full cursor-pointer text-xs font-semibold">
                      <Link href="/compare">
                        <GitCompareArrows className="mr-2 h-4 w-4" />
                        {t("bandingkan")}
                      </Link>
                    </Button>
                  </div>
                </Card>
              ))}
              {comparisonCards.length === 0 && <EmptySection label="Butuh minimal dua supplier aktif untuk komparasi." />}
            </div>
          )}
        </section>

        <section className="mx-auto mt-16 max-w-7xl space-y-6 px-4 sm:px-6 lg:px-8">
          <SectionHeader title={t("videoTitle")} href="/demo/help" action={t("lihatSemuaVideo")} />
          {isVideosLoading ? (
            <CardGridSkeleton />
          ) : (
            <div className="relative">
              <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
                {videos.map((video) => (
                  <a key={video.id} href={video.videoUrl || "#"} className="group relative flex cursor-pointer flex-col gap-2.5">
                    <div className="relative h-[145px] overflow-hidden rounded-lg border border-border bg-muted shadow-xs">
                      {video.imageUrl ? (
                        <img src={video.imageUrl} alt={video.title} className="h-full w-full object-cover opacity-90 transition-all duration-500 group-hover:scale-105 group-hover:opacity-80" />
                      ) : (
                        <div className="flex h-full items-center justify-center">
                          <Play className="h-8 w-8 text-muted-foreground" />
                        </div>
                      )}
                      <div className="absolute inset-0 flex items-center justify-center">
                        <div className="flex h-10 w-10 items-center justify-center rounded bg-card/95 text-foreground shadow-lg transition-all duration-300 group-hover:scale-105 group-hover:bg-primary group-hover:text-primary-foreground">
                          <Play className="ml-0.5 h-4.5 w-4.5 fill-current" />
                        </div>
                      </div>
                      {video.duration && (
                        <span className="absolute bottom-2 right-2 rounded bg-neutral-950/70 px-1 py-0.5 text-[9px] font-bold text-neutral-50">
                          {video.duration}
                        </span>
                      )}
                    </div>
                    <div className="space-y-1">
                      <h3 className="line-clamp-2 text-xs font-bold leading-snug text-foreground transition-colors group-hover:text-primary">
                        {video.title}
                      </h3>
                      <p className="text-[10px] text-muted-foreground">{formatViewCount(video.viewCount, locale)}</p>
                    </div>
                  </a>
                ))}
                {videos.length === 0 && <EmptySection label="Belum ada video aktif." />}
              </div>
              {videos.length > 0 && <CarouselArrow topClass="top-1/3" />}
            </div>
          )}
        </section>

        <section className="mt-16 border-t border-border bg-muted py-12 text-center text-muted-foreground">
          <div className="mx-auto max-w-7xl space-y-6 px-4 sm:px-6 lg:px-8">
            <h2 className="text-[11px] font-bold uppercase tracking-widest text-primary">{t("trustTitle")}</h2>
            <div className="flex flex-wrap justify-center gap-6 text-xs font-semibold text-foreground sm:gap-12 sm:text-sm">
              {[t("nibVerified"), t("factoryInspection"), t("exportCompliant"), t("secureChat")].map((item) => (
                <span key={item} className="flex items-center gap-2">
                  <CheckCircle className="h-4 w-4 shrink-0 text-success" />
                  {item}
                </span>
              ))}
            </div>
          </div>
        </section>

        <button
          onClick={scrollToTop}
          className={`fixed bottom-6 right-6 z-50 flex h-11 w-11 cursor-pointer items-center justify-center rounded bg-cyan text-cyan-foreground shadow-lg transition-all duration-300 hover:-translate-y-0.5 hover:scale-105 hover:shadow-xl active:scale-95 ${
            showBackToTop ? "pointer-events-auto scale-100 opacity-100" : "pointer-events-none scale-75 opacity-0"
          }`}
          aria-label={t("kembaliKeAtas")}
        >
          <ArrowUp className="h-5 w-5 stroke-[2.5]" />
        </button>
      </div>
    </PublicLayout>
  );
}

function SectionHeader({ title, href, action }: { title: string; href: string; action: string }) {
  return (
    <div className="flex items-end justify-between border-b border-border pb-3">
      <h2 className="font-heading text-lg font-bold text-foreground">{title}</h2>
      <Link href={href} className="cursor-pointer text-xs font-bold uppercase tracking-wider text-primary transition-colors hover:text-primary/80">
        {action}
      </Link>
    </div>
  );
}

function SupplierMini({ supplier }: { supplier: { companyName: string; slug: string; rating: number; isVerified: boolean } }) {
  return (
    <Link href={`/demo/suppliers/${supplier.slug}`} className="w-[42%] cursor-pointer space-y-1">
      <div className="mx-auto flex h-10 w-10 items-center justify-center rounded bg-primary/10 text-xs font-bold text-primary shadow-xs">
        {supplier.companyName.slice(0, 2).toUpperCase()}
      </div>
      <p className="truncate text-[11px] font-bold text-foreground">{supplier.companyName}</p>
      <div className="flex items-center justify-center gap-1 text-[10px] font-semibold text-warning">
        {supplier.isVerified ? <ShieldCheck className="h-3 w-3 text-success" /> : <Store className="h-3 w-3" />}
        <Star className="h-3 w-3 fill-current" />
        <span>{supplier.rating?.toFixed(1) || "0.0"}</span>
      </div>
    </Link>
  );
}

function CardGridSkeleton({ columns = "lg:grid-cols-4" }: { columns?: string }) {
  return (
    <div className={`grid grid-cols-1 gap-6 sm:grid-cols-2 ${columns}`}>
      {Array.from({ length: 4 }).map((_, index) => (
        <div key={index} className="h-64 animate-pulse rounded-lg border border-border bg-muted" />
      ))}
    </div>
  );
}

function EmptySection({ label }: { label: string }) {
  return (
    <div className="col-span-full rounded-lg border border-dashed border-border bg-card py-10 text-center text-xs text-muted-foreground">
      {label}
    </div>
  );
}

function CarouselArrow({ topClass = "top-1/2" }: { topClass?: string }) {
  return (
    <button className={`absolute -right-3 ${topClass} z-10 hidden h-8 w-8 -translate-y-1/2 cursor-pointer items-center justify-center rounded border border-border bg-card text-foreground shadow-md transition-all hover:bg-muted hover:shadow-lg lg:flex`}>
      <ChevronRight className="h-4.5 w-4.5" />
    </button>
  );
}
