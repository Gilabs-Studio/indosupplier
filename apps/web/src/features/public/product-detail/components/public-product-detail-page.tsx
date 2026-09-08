"use client";
/* eslint-disable @next/next/no-img-element */

import React from "react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";
import { formatPrice, resolveImageUrl } from "@/lib/utils";
import { PublicLayout } from "@/features/public/components/public-layout";
import { ProductCard } from "@/components/ui/product-card";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";
import { usePublicProductDetail } from "../hooks/use-public-product-detail";
import {
  GitCompareArrows,
  Heart,
  MapPin,
  MessageSquare,
  Minus,
  Package,
  Plus,
  Send,
  Share2,
  ShieldCheck,
  Star,
  Store,
  Truck,
} from "lucide-react";

interface PublicProductDetailPageProps {
  locale: string;
  id: string;
  detailBasePath?: "" | "/demo";
}


function formatDate(value: string) {
  if (!value) return "";
  return new Intl.DateTimeFormat("id-ID", {
    day: "numeric",
    month: "short",
    year: "numeric",
  }).format(new Date(value));
}

export function PublicProductDetailPage({ locale, id, detailBasePath = "" }: PublicProductDetailPageProps) {
  const [logoError, setLogoError] = React.useState(false);
  const t = useTranslations("public.productDetail");

  const {
    product,
    supplier,
    photos,
    variants,
    selectedVariant,
    setSelectedVariant,
    activePhoto,
    setActivePhoto,
    quantity,
    setQuantity,
    reviewFilter,
    setReviewFilter,
    visibleReviews,
    reviews,
    averageRating,
    distribution,
    subtotal,
    isBookmarked,
    isCompared,
    isProductBookmarked,
    isLoading,
    relatedProducts,
    isAdding,
    isComparing,
    toggleBookmark,
    toggleCompare,
    toggleProductBookmark,
    handleBuyerAction,
  } = usePublicProductDetail({ id, detailBasePath: detailBasePath as "" | "/demo" });

  const homeHref = detailBasePath || "/";

  if (isLoading) {
    return (
      <PublicLayout locale={locale}>
        <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
          <div className="grid grid-cols-1 gap-8 lg:grid-cols-[380px_1fr_280px]">
            <div className="aspect-square animate-pulse rounded-lg bg-muted" />
            <div className="space-y-4">
              <div className="h-8 w-3/4 animate-pulse rounded bg-muted" />
              <div className="h-10 w-48 animate-pulse rounded bg-muted" />
              <div className="h-44 animate-pulse rounded bg-muted" />
            </div>
            <div className="h-80 animate-pulse rounded-lg bg-muted" />
          </div>
        </div>
      </PublicLayout>
    );
  }

  if (!product || !supplier) {
    return (
      <PublicLayout locale={locale}>
        <div className="mx-auto max-w-7xl px-4 py-20 text-center sm:px-6 lg:px-8">
          <Package className="mx-auto h-12 w-12 text-muted-foreground" />
          <h1 className="mt-4 text-xl font-extrabold text-foreground">{t("productNotFound")}</h1>
          <Button asChild className="mt-6 cursor-pointer">
            <Link href={`${detailBasePath}/search`}>{t("backToSearch")}</Link>
          </Button>
        </div>
      </PublicLayout>
    );
  }

  const currentPhoto = photos[activePhoto];

  return (
    <PublicLayout locale={locale}>
      <main className="min-h-screen bg-background pb-16">
        <div className="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
          {/* Breadcrumbs */}
          <nav className="mb-6 flex flex-wrap items-center gap-2 text-xs font-medium text-muted-foreground">
            <Link href={homeHref} className="cursor-pointer text-muted-foreground transition-colors hover:text-primary">
              {t("home")}
            </Link>
            <span className="text-border">/</span>
            <Link
              href={`${detailBasePath}/search?query=${encodeURIComponent(product.categoryName || product.name)}`}
              className="cursor-pointer text-muted-foreground transition-colors hover:text-primary"
            >
              {product.categoryName || "Katalog"}
            </Link>
            <span className="text-border">/</span>
            <span className="line-clamp-1 max-w-[300px] text-foreground font-semibold">{product.name}</span>
          </nav>

          {/* Main Grid: Info, Gallery, and Actions */}
          <section className="grid grid-cols-1 gap-8 lg:grid-cols-[380px_minmax(0,1fr)_286px]">
            {/* Gallery Column */}
            <aside className="space-y-4 lg:sticky lg:top-24 lg:self-start">
              <div className="group relative aspect-square overflow-hidden rounded-lg border border-border bg-muted/30 shadow-xs transition-all duration-300 hover:shadow-md">
                {currentPhoto ? (
                  <img src={currentPhoto} alt={product.name} className="h-full w-full object-cover transition-transform duration-500 group-hover:scale-102" />
                ) : (
                  <div className="flex h-full items-center justify-center">
                    <Package className="h-12 w-12 text-muted-foreground/50" />
                  </div>
                )}
              </div>
              <div className="grid grid-cols-5 gap-2">
                {(photos.length ? photos : [""]).slice(0, 5).map((photo, index) => (
                  <button
                    key={photo || index}
                    type="button"
                    onClick={() => setActivePhoto(index)}
                    className={`aspect-square overflow-hidden rounded-lg border bg-card transition-all duration-300 cursor-pointer ${
                      activePhoto === index
                        ? "border-primary ring-1 ring-primary shadow-xs"
                        : "border-border hover:border-primary/50 hover:shadow-xs"
                    }`}
                    aria-label={`Foto product ${index + 1}`}
                  >
                    {photo ? (
                      <img src={photo} alt={`${product.name} ${index + 1}`} className="h-full w-full object-cover transition-transform duration-300 hover:scale-105" />
                    ) : (
                      <Package className="m-auto h-5 w-5 text-muted-foreground/40" />
                    )}
                  </button>
                ))}
              </div>
            </aside>

            {/* Core Info Column */}
            <section className="min-w-0 space-y-6">
              <div className="space-y-3 border-b border-border pb-5">
                <h1 className="text-xl font-bold tracking-tight text-foreground md:text-2xl lg:text-3xl leading-snug">
                  {product.name}
                </h1>
                <div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
                  <div className="flex items-center gap-1">
                    <Star className="h-3.5 w-3.5 fill-warning text-warning" />
                    <span className="font-bold text-foreground">{averageRating.toFixed(1)}</span>
                    <span>({supplier.reviewCount || reviews.length} {t("rating")})</span>
                  </div>
                  <span className="h-3 w-px bg-border" />
                  <span>{t("soldCount")}</span>
                  {supplier.isVerified && (
                    <>
                      <span className="h-3 w-px bg-border" />
                      <span className="inline-flex items-center gap-1 font-semibold text-primary">
                        <ShieldCheck className="h-3.5 w-3.5" />
                        {t("verifiedSupplier")}
                      </span>
                    </>
                  )}
                </div>
                <div className="mt-4 inline-block rounded-lg bg-primary/[0.03] px-4 py-2 border border-primary/10">
                  <p className="text-2xl font-black text-primary">{formatPrice(product.price, product.currency) || t("negotiable")}</p>
                </div>
              </div>

              {/* Variant Selector */}
              <div className="border-b border-border pb-5">
                <h2 className="text-sm font-bold text-foreground">
                  {t("selectVariant")} <span className="font-medium text-muted-foreground">{variants[selectedVariant]}</span>
                </h2>
                <div className="mt-3 flex flex-wrap gap-2">
                  {variants.map((variant, index) => (
                    <button
                      key={variant}
                      type="button"
                      onClick={() => setSelectedVariant(index)}
                      className={`rounded-lg border px-3.5 py-1.5 text-xs font-semibold transition-all duration-300 cursor-pointer hover:-translate-y-0.5 active:translate-y-0 ${
                        selectedVariant === index
                          ? "border-primary bg-primary/5 text-primary shadow-xs"
                          : "border-border bg-card text-muted-foreground hover:border-primary/50 hover:text-foreground"
                      }`}
                    >
                      {variant}
                    </button>
                  ))}
                </div>
              </div>

              {/* Product Tabs */}
              <Tabs defaultValue="detail" className="w-full">
                <TabsList className="w-full justify-start">
                  <TabsTrigger value="detail" className="cursor-pointer font-bold text-xs">{t("tabDetail")}</TabsTrigger>
                  <TabsTrigger value="spec" className="cursor-pointer font-bold text-xs">{t("tabSpec")}</TabsTrigger>
                  <TabsTrigger value="shipping" className="cursor-pointer font-bold text-xs">{t("tabShipping")}</TabsTrigger>
                </TabsList>
                <div className="min-h-[160px] text-xs leading-relaxed text-foreground/90 pt-4">
                  <TabsContent value="detail" className="outline-none">
                    <div className="space-y-4">
                      <div className="grid grid-cols-2 gap-y-2 max-w-sm rounded-lg bg-muted/30 p-3.5 border border-border text-xs">
                        <span className="text-muted-foreground">{t("conditionLabel")}</span>
                        <span className="font-medium">{t("conditionNew")}</span>
                        
                        <span className="text-muted-foreground">{t("moqLabel")}</span>
                        <span className="font-medium">{product.minOrder || t("negotiable")}</span>
                        
                        <span className="text-muted-foreground">{t("categoryLabel")}</span>
                        <span className="font-semibold text-primary">{product.categoryName || "-"}</span>
                        
                        <span className="text-muted-foreground">{t("capacityLabel")}</span>
                        <span className="font-medium">{product.capacityText || t("negotiable")}</span>
                      </div>
                      <p className="whitespace-pre-line text-sm leading-relaxed text-muted-foreground pt-1">
                        {product.description || t("emptyDescription")}
                      </p>
                    </div>
                  </TabsContent>
                  <TabsContent value="spec" className="outline-none">
                    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                      <InfoRow label={t("supplierName")} value={supplier.companyName} />
                      <InfoRow label={t("location")} value={product.supplierLocation || supplier.location || "-"} />
                      <InfoRow label={t("responseRate")} value={`${Math.round(supplier.responseRate || 0)}%`} />
                      <InfoRow label={t("responseTime")} value={supplier.responseTime || "-"} />
                    </div>
                  </TabsContent>
                  <TabsContent value="shipping" className="outline-none">
                    <div className="space-y-3.5">
                      <div className="flex gap-3 rounded-lg border border-border bg-card p-3">
                        <Truck className="h-4 w-4 text-primary shrink-0 mt-0.5" />
                        <div>
                          <p className="text-xs font-bold text-foreground">{t("shippingFrom")} {supplier.location || "Indonesia"}</p>
                          <p className="text-xs text-muted-foreground mt-0.5">{t("shippingDesc")}</p>
                        </div>
                      </div>
                      <div className="flex gap-3 rounded-lg border border-border bg-card p-3">
                        <ShieldCheck className="h-4 w-4 text-primary shrink-0 mt-0.5" />
                        <div>
                          <p className="text-xs font-bold text-foreground">{t("verifiedSupplier")}</p>
                          <p className="text-xs text-muted-foreground mt-0.5">{t("verifiedSupplierDesc")}</p>
                        </div>
                      </div>
                    </div>
                  </TabsContent>
                </div>
              </Tabs>
            </section>

            {/* Sticky Actions & Supplier Card Column */}
            <aside className="space-y-4 lg:sticky lg:top-24 lg:self-start">
              {/* RFQ / Order Card */}
              <div className="rounded-lg border border-border bg-card p-4 shadow-xs">
                <h2 className="text-sm font-bold text-foreground">{t("arrangeQty")}</h2>
                <p className="mt-2 text-xs font-semibold text-muted-foreground">{variants[selectedVariant]}</p>
                <div className="my-3.5 border-t border-border" />
                
                <div className="flex items-center justify-between gap-3">
                  <div className="flex h-8 items-center rounded-lg border border-border bg-muted/10">
                    <button
                      type="button"
                      onClick={() => setQuantity((value) => Math.max(1, value - 1))}
                      className="px-2.5 text-muted-foreground hover:text-foreground cursor-pointer transition-colors"
                      aria-label="Kurangi jumlah"
                    >
                      <Minus className="h-3.5 w-3.5" />
                    </button>
                    <span className="w-8 text-center text-xs font-bold">{quantity}</span>
                    <button
                      type="button"
                      onClick={() => setQuantity((value) => value + 1)}
                      className="px-2.5 text-primary hover:text-primary/80 cursor-pointer transition-colors"
                      aria-label="Tambah jumlah"
                    >
                      <Plus className="h-3.5 w-3.5" />
                    </button>
                  </div>
                  <p className="text-xs text-muted-foreground">{t("stock")} <span className="font-bold text-foreground">248</span></p>
                </div>

                <div className="mt-4 flex items-end justify-between gap-3">
                  <span className="text-xs text-muted-foreground">{t("subtotal")}</span>
                  <span className="text-lg font-black text-foreground">{formatPrice(subtotal, product.currency) || t("negotiable")}</span>
                </div>

                <div className="mt-4 space-y-2">
                  <Button
                    onClick={() => handleBuyerAction(t("authRequireRfq"))}
                    className="w-full cursor-pointer font-bold text-xs py-2 bg-primary text-primary-foreground hover:bg-primary/95 hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 hover:shadow-md hover:shadow-primary/20 rounded-lg"
                  >
                    <Send className="mr-1.5 h-3.5 w-3.5" />
                    {t("btnRfq")}
                  </Button>
                  <Button
                    onClick={() => handleBuyerAction(t("authRequireBuy"))}
                    variant="outline"
                    className="w-full cursor-pointer font-bold text-xs py-2 border-border text-foreground hover:bg-muted hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 hover:shadow-xs rounded-lg"
                  >
                    {t("btnBuy")}
                  </Button>
                </div>

                <div className="mt-3.5 grid grid-cols-3 gap-1 border-t border-border pt-3 text-[10px] font-bold text-muted-foreground">
                  <button
                    type="button"
                    onClick={() => handleBuyerAction(t("authRequireChat"))}
                    className="flex flex-col items-center justify-center gap-1 rounded-md py-1.5 hover:bg-muted hover:text-foreground cursor-pointer transition-colors"
                  >
                    <MessageSquare className="h-3.5 w-3.5" />
                    <span>{t("btnChat")}</span>
                  </button>
                  <button
                    type="button"
                    onClick={toggleBookmark}
                    disabled={isAdding}
                    className="flex flex-col items-center justify-center gap-1 rounded-md py-1.5 hover:bg-muted hover:text-foreground cursor-pointer transition-colors"
                  >
                    <Heart className={`h-3.5 w-3.5 ${isBookmarked ? "fill-rose-600 text-rose-600" : ""}`} />
                    <span>{t("btnWishlist")}</span>
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      navigator.clipboard.writeText(window.location.href);
                      toast.success(t("linkCopied"));
                    }}
                    className="flex flex-col items-center justify-center gap-1 rounded-md py-1.5 hover:bg-muted hover:text-foreground cursor-pointer transition-colors"
                  >
                    <Share2 className="h-3.5 w-3.5" />
                    <span>{t("btnShare")}</span>
                  </button>
                </div>
                
                <Button
                  onClick={toggleCompare}
                  disabled={isComparing}
                  variant="ghost"
                  className="mt-2.5 w-full cursor-pointer text-[10px] font-bold text-muted-foreground hover:bg-muted/50 hover:text-foreground rounded-lg h-7"
                >
                  <GitCompareArrows className="mr-1.5 h-3.5 w-3.5" />
                  {isCompared ? t("removeCompare") : t("addCompare")}
                </Button>
              </div>

              {/* Supplier Detail Sidebar Card */}
              <div className="rounded-lg border border-border bg-card p-4 transition-all duration-300 hover:shadow-md">
                <Link href={`${detailBasePath}/suppliers/${supplier.slug}`} className="flex cursor-pointer items-center gap-3 group">
                  <div 
                    className="flex h-10 w-10 shrink-0 items-center justify-center border border-border bg-muted transition-colors group-hover:bg-muted/90 overflow-hidden relative"
                    style={{ borderRadius: "9999px" }}
                  >
                    {supplier.logo && !logoError ? (
                      <img
                        src={resolveImageUrl(supplier.logo)}
                        alt={supplier.companyName}
                        className="h-full w-full object-cover"
                        style={{ borderRadius: "9999px" }}
                        onLoad={() => setLogoError(false)}
                        onError={() => setLogoError(true)}
                      />
                    ) : (
                      <div 
                        className="flex h-full w-full items-center justify-center bg-primary text-primary-foreground font-heading font-bold text-xs"
                        style={{ borderRadius: "9999px" }}
                      >
                        {supplier.companyName.slice(0, 2).toUpperCase()}
                      </div>
                    )}
                  </div>
                  <div className="min-w-0">
                    <p className="truncate text-xs font-bold text-foreground transition-colors group-hover:text-primary">
                      {supplier.companyName}
                    </p>
                    <div className="flex items-center gap-1 text-[10px] text-muted-foreground mt-0.5">
                      <MapPin className="h-3 w-3" />
                      <span className="truncate">{supplier.location || "Indonesia"}</span>
                    </div>
                  </div>
                </Link>
                
                <div className="mt-4 grid grid-cols-2 gap-2 text-center">
                  <InfoMetric label={t("rating")} value={supplier.rating?.toFixed(1) || "0.0"} />
                  <InfoMetric label={t("reviews")} value={`${supplier.reviewCount || reviews.length}`} />
                  <InfoMetric label={t("response")} value={`${Math.round(supplier.responseRate || 0)}%`} />
                  <InfoMetric label={t("time")} value={supplier.responseTime || "-"} />
                </div>
                
                <Button asChild variant="outline" className="mt-4 w-full cursor-pointer text-xs font-bold border-border hover:bg-muted hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 rounded-lg">
                  <Link href={`${detailBasePath}/suppliers/${supplier.slug}`}>
                    <Store className="mr-1.5 h-3.5 w-3.5" />
                    {t("detailSupplier")}
                  </Link>
                </Button>
              </div>
            </aside>
          </section>

          {/* Separator before Review Section */}
          <div className="my-10 border-t border-border" />

          {/* Reviews Section (Placed below main content grid) */}
          <section className="space-y-6">
            <div className="flex flex-col gap-1.5">
              <h2 className="text-base font-bold text-foreground font-heading tracking-tight md:text-lg">
                {t("buyerReviews", { count: supplier.reviewCount || reviews.length })}
              </h2>
              <p className="text-xs text-muted-foreground">{t("buyerReviewsDesc")}</p>
            </div>

            <div className="grid grid-cols-1 gap-8 lg:grid-cols-[280px_1fr]">
              {/* Left Column: Rating breakdown & Filters */}
              <div className="space-y-4">
                <div className="rounded-lg border border-border bg-card p-4.5 space-y-4.5">
                  <div className="flex items-end gap-2.5">
                    <Star className="h-7 w-7 fill-warning text-warning" />
                    <span className="text-3xl font-black text-foreground leading-none">{averageRating.toFixed(1)}</span>
                    <span className="text-xs text-muted-foreground pb-1">/ 5.0</span>
                  </div>
                  
                  <div className="space-y-1">
                    <p className="text-xs font-bold text-foreground">
                      {reviews.length ? `100% ${t("satisfactionRate")}` : t("noReviewsYet")}
                    </p>
                    <p className="text-[10px] text-muted-foreground">
                      {supplier.reviewCount || reviews.length} {t("reviewsBreakdown")}
                    </p>
                  </div>

                  <div className="space-y-2 pt-2 border-t border-border">
                    {distribution.map((item) => (
                      <div key={item.rating} className="flex items-center gap-2 text-[10px] text-muted-foreground">
                        <span className="w-2 text-right font-medium">{item.rating}</span>
                        <Star className="h-3 w-3 fill-warning text-warning shrink-0" />
                        <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-muted/60">
                          <div className="h-full rounded-full bg-primary transition-all duration-500" style={{ width: `${item.percent}%` }} />
                        </div>
                        <span className="w-6 text-right">({item.count})</span>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Filter Side Card */}
                <div className="rounded-lg border border-border bg-card overflow-hidden">
                  <div className="border-b border-border bg-muted/10 px-4 py-2.5">
                    <p className="text-[10px] font-bold text-foreground uppercase tracking-wider">{t("filterReviews")}</p>
                  </div>
                  <div className="divide-y divide-border">
                    {[
                      ["all", t("filterAll")],
                      ["media", t("filterMedia")],
                      ["high", t("filterHigh")],
                    ].map(([key, label]) => {
                      const isActive = reviewFilter === key;
                      return (
                        <button
                          key={key}
                          type="button"
                          onClick={() => setReviewFilter(key as "all" | "media" | "high")}
                          className={`flex w-full items-center justify-between px-4 py-2.5 text-xs font-medium cursor-pointer transition-all duration-300 hover:bg-muted/50 ${
                            isActive ? "bg-primary/[0.02] text-primary font-bold" : "text-foreground"
                          }`}
                        >
                          <span>{label}</span>
                          <span className={`h-1.5 w-1.5 rounded-full transition-transform ${isActive ? "bg-primary scale-120" : "bg-transparent"}`} />
                        </button>
                      );
                    })}
                  </div>
                </div>
              </div>

              {/* Right Column: Reviews List */}
              <div className="space-y-4">
                <div className="flex items-center justify-between border-b border-border pb-3">
                  <p className="text-xs text-muted-foreground font-semibold">
                    {t("showingReviews", { visibleCount: visibleReviews.length, totalCount: reviews.length })}
                  </p>
                  <Button variant="outline" size="sm" className="cursor-pointer text-xs font-semibold h-8 rounded-lg border-border hover:bg-muted">
                    {t("mostHelpful")}
                  </Button>
                </div>

                {visibleReviews.length === 0 ? (
                  <div className="rounded-lg border border-dashed border-border py-12 text-center text-xs text-muted-foreground bg-card">
                    {t("noReviewsFilter")}
                  </div>
                ) : (
                  <div className="space-y-4">
                    {visibleReviews.map((review) => (
                      <article
                        key={review.id}
                        className="rounded-lg border border-border bg-card p-4 transition-all duration-300 hover:shadow-xs hover:border-border/80"
                      >
                        <div className="flex items-start justify-between gap-3">
                          <div className="flex items-center gap-3">
                            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/5 text-xs font-bold text-primary uppercase">
                              {review.buyerName.slice(0, 2)}
                            </div>
                            <div>
                              <p className="text-xs font-bold text-foreground">{review.buyerName}</p>
                              <p className="text-[10px] text-muted-foreground mt-0.5">{formatDate(review.createdAt)}</p>
                            </div>
                          </div>
                          <div className="flex items-center gap-1 rounded-lg bg-muted/40 px-2 py-1 border border-border text-xs font-bold">
                            <Star className="h-3 w-3 fill-warning text-warning" />
                            <span>{review.rating}</span>
                          </div>
                        </div>
                        
                        <p className="mt-3 text-xs leading-relaxed text-foreground/90 pl-1">
                          {review.reviewText}
                        </p>
                        
                        {review.supplierReply && (
                          <div className="mt-3 rounded-lg border border-primary/10 bg-primary/[0.01] p-3 text-xs text-muted-foreground relative pl-8">
                            <span className="absolute left-3 top-3.5 h-1.5 w-1.5 rounded-full bg-primary/40" />
                            <p className="font-bold text-primary">{t("supplierReply")}</p>
                            <p className="mt-1 text-foreground/80 leading-relaxed font-medium">{review.supplierReply}</p>
                          </div>
                        )}
                      </article>
                    ))}
                  </div>
                )}
              </div>
            </div>
          </section>

          {/* Recommendations Section */}
          {relatedProducts.length > 0 && (
            <section className="mt-16 space-y-6">
              <div className="flex flex-col gap-1.5 border-b border-border pb-3">
                <h2 className="text-base font-bold text-foreground font-heading tracking-tight md:text-lg">
                  {t("recommendationsTitle")}
                </h2>
                <p className="text-xs text-muted-foreground">{t("recommendationsDesc")}</p>
              </div>
              <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
                {relatedProducts.map((item) => (
                  <ProductCard
                    key={item.id}
                    id={item.id}
                    name={item.name}
                    price={item.price}
                    currency={item.currency}
                    image={item.photos?.[0]}
                    href={`${detailBasePath}/products/${item.id}`}
                    rating={item.supplierRating}
                    reviewCount={item.supplierReviewCount}
                    supplierName={item.supplierCompanyName}
                    supplierHref={`${detailBasePath}/suppliers/${item.supplierSlug}`}
                    categoryName={item.categoryName || undefined}
                    minOrder={item.minOrder || undefined}
                    moqLabel={t("moqLabel")}
                    ulasanLabel={t("reviews")}
                    priceLabel={t("negotiable")}
                    showBookmarkOverlayButton={true}
                    isBookmarked={isProductBookmarked(item.id)}
                    onBookmark={() => toggleProductBookmark(item)}
                  />
                ))}
              </div>
            </section>
          )}
        </div>
      </main>
    </PublicLayout>
  );
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-border bg-muted/5 p-3 transition-colors hover:border-primary/25">
      <p className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider">{label}</p>
      <p className="mt-1 text-xs font-bold text-foreground">{value}</p>
    </div>
  );
}

function InfoMetric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-border bg-muted/10 p-2.5 transition-all duration-300 hover:border-primary/20 hover:bg-primary/[0.01]">
      <p className="text-sm font-black text-primary">{value}</p>
      <p className="text-[10px] font-medium text-muted-foreground mt-0.5">{label}</p>
    </div>
  );
}
