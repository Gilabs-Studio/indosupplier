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
  ChevronDown,
  GitCompareArrows,
  Heart,
  Loader2,
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

function RatingStars({ rating, size = "md" }: { rating: number; size?: "sm" | "md" | "lg" }) {
  const sizeClasses = {
    sm: "h-3.5 w-3.5",
    md: "h-4 w-4",
    lg: "h-5 w-5",
  };
  const iconSize = sizeClasses[size];

  return (
    <div className="flex items-center gap-0.5">
      {[1, 2, 3, 4, 5].map((starIndex) => {
        const isFilled = rating >= starIndex;
        const isHalf = !isFilled && rating >= starIndex - 0.5;
        return (
          <Star
            key={starIndex}
            className={`${iconSize} transition-colors ${
              isFilled
                ? "fill-warning text-warning"
                : isHalf
                ? "fill-warning/50 text-warning"
                : "text-muted-foreground/30"
            }`}
          />
        );
      })}
    </div>
  );
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
    reviewState,
    productRating,
    productReviewCount,
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
          <section id="detail" className="grid grid-cols-1 gap-8 lg:grid-cols-[380px_minmax(0,1fr)_286px] scroll-mt-24">
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
                    key={`photo-thumb-${index}`}
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
                  <div className="flex items-center gap-1.5">
                    <Star className="h-4 w-4 fill-warning text-warning" />
                    {productReviewCount > 0 ? (
                      <>
                        <span className="font-extrabold text-foreground">{productRating.toFixed(1)}</span>
                        <span>({productReviewCount} {t("rating")})</span>
                      </>
                    ) : (
                      <span className="font-medium text-muted-foreground">{t("noReviewsYet")}</span>
                    )}
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
                    <Heart className={`h-3.5 w-3.5 ${isBookmarked ? "fill-destructive text-destructive" : ""}`} />
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

          {/* Section Navigation Tabs (Minimalist sticky anchor bar) */}
          <div className="sticky top-16 z-20 -mx-4 my-8 border-y border-border bg-background/95 px-4 backdrop-blur-md sm:-mx-6 sm:px-6 lg:-mx-8 lg:px-8">
            <div className="mx-auto flex max-w-7xl items-center justify-between gap-4">
              <span className="hidden truncate text-xs font-bold text-foreground md:inline-block max-w-[320px]">
                {product.name}
              </span>
              <div className="flex items-center gap-6 text-xs font-bold">
                <a
                  href="#detail"
                  className="py-3 border-b-2 border-transparent text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
                >
                  {t("navDetail")}
                </a>
                <a
                  href="#reviews"
                  className="py-3 border-b-2 border-primary text-primary transition-colors cursor-pointer flex items-center gap-1.5"
                >
                  <span>{t("navReviews")}</span>
                  {reviewState.summary.totalReviews > 0 && (
                    <span className="text-[11px] font-semibold opacity-80">
                      ({reviewState.summary.totalReviews})
                    </span>
                  )}
                </a>
                {relatedProducts.length > 0 && (
                  <a
                    href="#recommendations"
                    className="py-3 border-b-2 border-transparent text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
                  >
                    {t("navRecommendations")}
                  </a>
                )}
              </div>
            </div>
          </div>

          {/* Reviews Section (Minimalism Modern Tokopedia-Style) */}
          <section id="reviews" className="space-y-6 scroll-mt-24">
            <h2 className="text-sm font-extrabold uppercase tracking-wider text-foreground">
              {t("buyerReviewsTitle")}
            </h2>

            {/* Top Rating Summary Card */}
            <div className="rounded-xl border border-border bg-card p-6 shadow-2xs">
              <div className="grid grid-cols-1 md:grid-cols-[240px_1fr] gap-8 items-center">
                {/* Left: Big Score & Satisfaction */}
                <div className="space-y-2 border-b md:border-b-0 md:border-r border-border pb-6 md:pb-0 md:pr-6">
                  <div className="flex items-center gap-2.5">
                    <Star className="h-8 w-8 fill-warning text-warning shrink-0" />
                    <span className="text-4xl font-black text-foreground leading-none">
                      {reviewState.summary.averageRating > 0 ? reviewState.summary.averageRating.toFixed(1) : "0.0"}
                    </span>
                    <span className="text-xs font-semibold text-muted-foreground self-end pb-1">/ 5.0</span>
                  </div>
                  <p className="text-xs font-bold text-foreground">
                    {reviewState.summary.totalReviews > 0
                      ? `${reviewState.summary.positivePercent}% ${t("satisfactionRate")}`
                      : t("noReviewsYet")}
                  </p>
                  <p className="text-[11px] text-muted-foreground">
                    {reviewState.summary.totalReviews} rating • {reviewState.summary.totalReviews} {t("reviews")}
                  </p>
                </div>

                {/* Right: 2 Sub-columns of Rating Bars */}
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-2.5">
                  {/* Left sub-column: 5, 4, 3 */}
                  <div className="space-y-2">
                    {[5, 4, 3].map((star) => {
                      const count = reviewState.summary.ratingBreakdown[String(star)] || 0;
                      const total = reviewState.summary.totalReviews || 1;
                      const percent = reviewState.summary.totalReviews > 0 ? Math.round((count / total) * 100) : 0;
                      const isSelected = reviewState.selectedRating === star;

                      return (
                        <button
                          key={star}
                          type="button"
                          onClick={() => reviewState.setSelectedRating(star)}
                          className={`w-full flex items-center gap-2 text-xs cursor-pointer group py-0.5 rounded transition-opacity ${
                            isSelected ? "font-bold" : "opacity-90 hover:opacity-100"
                          }`}
                        >
                          <div className="flex items-center gap-1 w-6 shrink-0">
                            <Star className="h-3 w-3 fill-warning text-warning shrink-0" />
                            <span className="text-xs text-foreground font-semibold">{star}</span>
                          </div>
                          <div className="h-1.5 flex-1 rounded-full bg-muted overflow-hidden">
                            <div
                              className="h-full rounded-full bg-primary transition-all duration-500"
                              style={{ width: `${percent}%` }}
                            />
                          </div>
                          <span className="w-12 text-right text-[11px] text-muted-foreground">
                            ({count})
                          </span>
                        </button>
                      );
                    })}
                  </div>

                  {/* Right sub-column: 2, 1 */}
                  <div className="space-y-2">
                    {[2, 1].map((star) => {
                      const count = reviewState.summary.ratingBreakdown[String(star)] || 0;
                      const total = reviewState.summary.totalReviews || 1;
                      const percent = reviewState.summary.totalReviews > 0 ? Math.round((count / total) * 100) : 0;
                      const isSelected = reviewState.selectedRating === star;

                      return (
                        <button
                          key={star}
                          type="button"
                          onClick={() => reviewState.setSelectedRating(star)}
                          className={`w-full flex items-center gap-2 text-xs cursor-pointer group py-0.5 rounded transition-opacity ${
                            isSelected ? "font-bold" : "opacity-90 hover:opacity-100"
                          }`}
                        >
                          <div className="flex items-center gap-1 w-6 shrink-0">
                            <Star className="h-3 w-3 fill-warning text-warning shrink-0" />
                            <span className="text-xs text-foreground font-semibold">{star}</span>
                          </div>
                          <div className="h-1.5 flex-1 rounded-full bg-muted overflow-hidden">
                            <div
                              className="h-full rounded-full bg-primary transition-all duration-500"
                              style={{ width: `${percent}%` }}
                            />
                          </div>
                          <span className="w-12 text-right text-[11px] text-muted-foreground">
                            ({count})
                          </span>
                        </button>
                      );
                    })}
                  </div>
                </div>
              </div>
            </div>

            {/* Lower Area: Filter Sidebar & Reviews List */}
            <div className="grid grid-cols-1 md:grid-cols-[220px_1fr] gap-8 pt-2">
              {/* Left Column: Filter Ulasan Sidebar */}
              <aside className="space-y-4">
                <div className="rounded-xl border border-border bg-card p-4 space-y-3 shadow-2xs">
                  <h3 className="text-xs font-extrabold uppercase tracking-wider text-foreground">
                    {t("filterReviewsTitle")}
                  </h3>

                  <div className="space-y-2.5 pt-2 border-t border-border">
                    <p className="text-[11px] font-bold text-muted-foreground uppercase tracking-wider">
                      {t("ratingSection")}
                    </p>
                    <div className="space-y-2">
                      {/* All */}
                      <label className="flex items-center gap-2.5 text-xs cursor-pointer select-none group">
                        <input
                          type="checkbox"
                          checked={reviewState.selectedRating === null}
                          onChange={() => reviewState.setSelectedRating(null)}
                          className="h-4 w-4 rounded border-border text-primary cursor-pointer accent-primary"
                        />
                        <span className={`text-xs ${reviewState.selectedRating === null ? "font-bold text-foreground" : "text-muted-foreground group-hover:text-foreground"}`}>
                          {t("allRatings")}
                        </span>
                        <span className="ml-auto text-[11px] text-muted-foreground">
                          ({reviewState.summary.totalReviews})
                        </span>
                      </label>

                      {/* Stars 5 to 1 */}
                      {[5, 4, 3, 2, 1].map((star) => {
                        const count = reviewState.summary.ratingBreakdown[String(star)] || 0;
                        const isChecked = reviewState.selectedRating === star;

                        return (
                          <label key={star} className="flex items-center gap-2.5 text-xs cursor-pointer select-none group">
                            <input
                              type="checkbox"
                              checked={isChecked}
                              onChange={() => reviewState.setSelectedRating(star)}
                              className="h-4 w-4 rounded border-border text-primary cursor-pointer accent-primary"
                            />
                            <div className="flex items-center gap-1">
                              <Star className="h-3.5 w-3.5 fill-warning text-warning" />
                              <span className={`text-xs ${isChecked ? "font-bold text-foreground" : "text-muted-foreground group-hover:text-foreground"}`}>
                                {star}
                              </span>
                            </div>
                            <span className="ml-auto text-[11px] text-muted-foreground">
                              ({count})
                            </span>
                          </label>
                        );
                      })}
                    </div>
                  </div>
                </div>
              </aside>

              {/* Right Column: Ulasan Pilihan List */}
              <div className="space-y-4 min-w-0">
                {/* Header: Title + Sort + Counter */}
                <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border pb-3">
                  <div>
                    <h3 className="text-xs font-extrabold uppercase tracking-wider text-foreground">
                      {t("featuredReviewsTitle")}
                    </h3>
                    <p className="text-[11px] text-muted-foreground mt-0.5">
                      {t("showingReviews", {
                        visibleCount: reviewState.reviews.length,
                        totalCount: reviewState.totalItems,
                      })}
                    </p>
                  </div>

                  {/* Sort Dropdown */}
                  <div className="flex items-center gap-2 text-xs">
                    <span className="text-muted-foreground font-medium">{t("sortByLabel")}</span>
                    <select
                      value={reviewState.sortBy}
                      onChange={(e) => reviewState.setSortBy(e.target.value as "newest" | "highest" | "lowest")}
                      className="rounded-lg border border-border bg-background px-2.5 py-1 text-xs font-semibold text-foreground cursor-pointer focus:outline-hidden focus:ring-1 focus:ring-primary"
                    >
                      <option value="newest">{t("sortNewest")}</option>
                      <option value="highest">{t("sortHighest")}</option>
                      <option value="lowest">{t("sortLowest")}</option>
                    </select>
                  </div>
                </div>

                {/* Minimalist Reviews List (No heavy containers, clean divider lines) */}
                {reviewState.isEmpty ? (
                  <div className="rounded-xl border border-dashed border-border p-12 text-center">
                    <Star className="mx-auto h-8 w-8 text-muted-foreground/30 stroke-1" />
                    <h4 className="mt-3 text-sm font-bold text-foreground">
                      {t("emptyReviewsTitle")}
                    </h4>
                    <p className="mt-1 text-xs text-muted-foreground max-w-sm mx-auto">
                      {t("emptyReviewsDesc")}
                    </p>
                  </div>
                ) : reviewState.isFilteredEmpty ? (
                  <div className="rounded-xl border border-dashed border-border p-8 text-center space-y-3">
                    <p className="text-xs text-muted-foreground">{t("noReviewsFilter")}</p>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => reviewState.setSelectedRating(null)}
                      className="cursor-pointer text-xs font-bold rounded-lg"
                    >
                      {t("allRatings")}
                    </Button>
                  </div>
                ) : (
                  <div className="divide-y divide-border">
                    {reviewState.reviews.map((review) => (
                      <article key={review.id} className="py-5 first:pt-2 space-y-2.5">
                        {/* Top Line: Stars + Date */}
                        <div className="flex items-center gap-2">
                          <div className="flex items-center gap-0.5">
                            {[1, 2, 3, 4, 5].map((s) => (
                              <Star
                                key={s}
                                className={`h-3.5 w-3.5 ${
                                  review.rating >= s
                                    ? "fill-warning text-warning"
                                    : "text-muted-foreground/20"
                                }`}
                              />
                            ))}
                          </div>
                          <span className="text-[11px] text-muted-foreground">
                            {formatDate(review.createdAt)}
                          </span>
                        </div>

                        {/* User Line: Avatar initial + Name + Optional Company */}
                        <div className="flex items-center gap-2">
                          <div className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-muted text-[10px] font-bold text-foreground uppercase">
                            {review.buyerName.slice(0, 1)}
                          </div>
                          <span className="text-xs font-bold text-foreground">
                            {review.buyerName}
                          </span>
                          {review.buyerCompany && review.buyerCompany !== review.buyerName && (
                            <span className="text-[10px] text-muted-foreground">
                              • {review.buyerCompany}
                            </span>
                          )}
                        </div>

                        {/* Review Content */}
                        <p className="text-xs leading-relaxed text-foreground/90">
                          {review.reviewText}
                        </p>

                        {/* Supplier Reply (Modern conversational thread style) */}
                        {review.supplierReply && (
                          <div className="mt-3 relative pl-4 sm:pl-5 before:absolute before:left-1 before:top-2 before:bottom-2 before:w-0.5 before:bg-border before:rounded-full">
                            <div className="rounded-lg border border-border/80 bg-muted/30 p-3 text-xs space-y-1.5 transition-colors hover:bg-muted/50">
                              <div className="flex flex-wrap items-center justify-between gap-2">
                                <div className="flex items-center gap-1.5">
                                  <div className="flex h-4 w-4 items-center justify-center rounded-full bg-primary/10 text-primary">
                                    <Store className="h-2.5 w-2.5" />
                                  </div>
                                  <span className="font-bold text-foreground text-xs">{supplier.companyName || t("supplierReply")}</span>
                                  <span className="rounded bg-primary/10 px-1.5 py-0.5 text-[9px] font-bold text-primary">
                                    {t("supplierReply")}
                                  </span>
                                </div>
                                {review.supplierRepliedAt && (
                                  <span className="text-[10px] text-muted-foreground font-medium">
                                    {formatDate(review.supplierRepliedAt)}
                                  </span>
                                )}
                              </div>
                              <p className="text-xs text-foreground/85 leading-relaxed pl-5.5 font-normal">
                                {review.supplierReply}
                              </p>
                            </div>
                          </div>
                        )}
                      </article>
                    ))}

                    {/* Button-driven Load More */}
                    {reviewState.hasMore && (
                      <div className="pt-6 text-center">
                        <Button
                          type="button"
                          onClick={reviewState.loadMore}
                          disabled={reviewState.isLoadingMore}
                          variant="outline"
                          className="cursor-pointer font-bold border-border hover:bg-muted rounded-lg px-6 py-2 text-xs shadow-2xs"
                        >
                          {reviewState.isLoadingMore ? (
                            <div className="flex items-center gap-2">
                              <Loader2 className="h-3.5 w-3.5 animate-spin text-primary" />
                              <span>{t("loadingMore")}</span>
                            </div>
                          ) : (
                            <div className="flex items-center gap-1.5">
                              <span>{t("loadMoreReviews")}</span>
                              <ChevronDown className="h-3.5 w-3.5" />
                            </div>
                          )}
                        </Button>
                      </div>
                    )}

                    {!reviewState.hasMore && reviewState.reviews.length > 5 && (
                      <p className="text-center text-[11px] text-muted-foreground font-medium pt-6">
                        {t("allReviewsLoaded")}
                      </p>
                    )}
                  </div>
                )}
              </div>
            </div>
          </section>

          {/* Recommendations Section */}
          {relatedProducts.length > 0 && (
            <section id="recommendations" className="mt-16 space-y-6 scroll-mt-24">
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
