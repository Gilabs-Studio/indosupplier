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
import { usePublicProductDetail } from "../hooks/use-public-product-detail";
import {
  Check,
  ChevronDown,
  FileText,
  GitCompareArrows,
  Heart,
  Info,
  Loader2,
  Mail,
  MapPin,
  MessageCircle,
  Package,
  Share2,
  ShieldCheck,
  Star,
  Store,
  Truck,
} from "lucide-react";
import { ProductRfqModal } from "./product-rfq-modal";
import { ProductDirectBuyModal } from "./product-direct-buy-modal";

interface PublicProductDetailPageProps {
  locale: string;
  id: string;
  detailBasePath?: "" | "/demo";
}

function formatDate(value: string, locale: string = "id") {
  if (!value) return "";
  return new Intl.DateTimeFormat(locale === "en" ? "en-US" : "id-ID", {
    day: "numeric",
    month: "short",
    year: "numeric",
  }).format(new Date(value));
}

export function PublicProductDetailPage({ locale, id, detailBasePath = "" }: PublicProductDetailPageProps) {
  const [logoError, setLogoError] = React.useState(false);
  const [rfqInitialNotes, setRfqInitialNotes] = React.useState("");
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
    isRfqModalOpen,
    setIsRfqModalOpen,
    isDirectBuyModalOpen,
    setIsDirectBuyModalOpen,
    isCopied,
    openRfqModal,
    handleShare,
    reviewState,
    productRating,
    productReviewCount,
  } = usePublicProductDetail({ id, detailBasePath: detailBasePath as "" | "/demo" });

  const handleOpenRfqModal = () => {
    setRfqInitialNotes("");
    openRfqModal();
  };

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
      <main className="min-h-screen bg-background pb-28">
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
              {product.categoryName || t("catalogDefault")}
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
                    aria-label={t("photoAriaLabel", { index: index + 1 })}
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
                <h1 className="text-xl font-extrabold tracking-tight text-foreground md:text-2xl lg:text-3xl leading-snug font-heading">
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
                <div className="mt-4 flex flex-col gap-3">
                  <div className="inline-block rounded-lg bg-primary/[0.03] px-4 py-2 border border-primary/10 self-start">
                    <p className="text-3xl font-extrabold text-foreground">{formatPrice(product.price, product.currency) || t("negotiable")}</p>
                  </div>

                  {/* Informational Warning Note */}
                  <div className="flex items-start gap-2.5 rounded-lg border border-border bg-muted/40 p-3 text-xs text-foreground">
                    <Info className="h-4 w-4 shrink-0 text-primary mt-0.5" />
                    <p className="leading-relaxed font-medium text-muted-foreground">
                      {t("priceDisclaimer")}
                    </p>
                  </div>

                  {/* Key Specifications Card */}
                  <div className="rounded-lg border border-border bg-card p-4">
                    <div className="grid grid-cols-2 gap-y-2.5 text-xs">
                      <span className="text-muted-foreground">{t("stock")}</span>
                      <span className="font-semibold text-foreground">{product.capacityText || t("stockAvailable")}</span>

                      <span className="text-muted-foreground">{t("specOriginCountry")}</span>
                      <span className="font-semibold text-foreground">{t("specOriginCountryValue")}</span>

                      <span className="text-muted-foreground">{t("specShippedFrom")}</span>
                      <span className="font-semibold text-foreground">{supplier.location || product.supplierLocation || t("specOriginCountryValue")}</span>

                      <span className="text-muted-foreground">{t("categoryLabel")}</span>
                      <span className="font-semibold text-primary">{product.categoryName || "-"}</span>

                      <span className="text-muted-foreground">{t("specLastUpdate")}</span>
                      <span className="font-semibold text-foreground">{formatDate(product.updatedAt || new Date().toISOString(), locale)}</span>
                    </div>
                  </div>

                  {/* Primary CTA: Request Quotation Button */}
                  <Button
                    onClick={handleOpenRfqModal}
                    className="w-full cursor-pointer font-semibold text-sm h-11 bg-primary text-primary-foreground hover:bg-primary/90 hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 shadow-xs hover:shadow-primary/30 rounded-lg flex items-center justify-center gap-2"
                  >
                    <FileText className="h-4 w-4" />
                    <span>{t("btnRequestQuote")}</span>
                  </Button>
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

            {/* Right Sidebar: Supplier Profile, Direct Channels & RFQ Action */}
            <aside className="space-y-4 lg:sticky lg:top-24 lg:self-start">
              <div className="rounded-lg border border-border bg-card p-5 shadow-xs space-y-4">
                {/* Supplier Header */}
                <div className="flex items-center justify-between border-b border-border pb-3">
                  <h2 className="text-xs font-bold uppercase tracking-wider text-muted-foreground font-heading">
                    {t("supplierCardTitle")}
                  </h2>
                  <div className="flex items-center gap-1.5 text-[11px] font-medium text-emerald-600 dark:text-emerald-400">
                    <span className="h-2 w-2 rounded-full bg-emerald-500 animate-pulse" />
                    <span>{t("onlineStatus")}</span>
                  </div>
                </div>

                {/* Supplier Identity */}
                <Link
                  href={`${detailBasePath}/suppliers/${supplier.slug}`}
                  className="flex cursor-pointer items-center gap-3 group"
                >
                  <div
                    className="flex h-11 w-11 shrink-0 items-center justify-center border border-border bg-muted transition-colors group-hover:bg-muted/80 overflow-hidden relative rounded-lg"
                  >
                    {supplier.logo && !logoError ? (
                      <img
                        src={resolveImageUrl(supplier.logo)}
                        alt={supplier.companyName}
                        className="h-full w-full object-cover rounded-lg"
                        onLoad={() => setLogoError(false)}
                        onError={() => setLogoError(true)}
                      />
                    ) : (
                      <div
                        className="flex h-full w-full items-center justify-center bg-primary/10 text-primary font-heading font-bold text-xs rounded-lg"
                      >
                        {supplier.companyName.slice(0, 2).toUpperCase()}
                      </div>
                    )}
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-xs font-bold text-foreground transition-colors group-hover:text-primary font-heading">
                      {supplier.companyName}
                    </p>
                    <p className="text-[11px] text-muted-foreground mt-0.5">
                      {supplier.businessType || t("businessEntity")}
                    </p>
                  </div>
                </Link>

                {/* Badges */}
                <div className="flex flex-wrap gap-1.5">
                  {supplier.isVerified && (
                    <span className="inline-flex items-center gap-1 rounded bg-primary/10 px-2 py-0.5 text-[10px] font-semibold text-primary border border-primary/20">
                      <ShieldCheck className="h-3 w-3" />
                      {t("badgeVerified")}
                    </span>
                  )}
                  <span className="inline-flex items-center gap-1 rounded bg-muted/60 px-2 py-0.5 text-[10px] font-semibold text-muted-foreground border border-border">
                    {t("badgeEstablished", { year: supplier.establishedYear || 2021 })}
                  </span>
                </div>

                {/* Tax Status & Location */}
                <div className="space-y-1.5 text-xs border-t border-border pt-3">
                  <div className="flex items-center justify-between text-[11px]">
                    <span className="text-muted-foreground">{t("taxStatusLabel")}</span>
                    <span className="font-semibold text-foreground">
                      {supplier.taxStatus === "non_pkp" ? t("taxStatusNonPKP") : t("taxStatusPKP")}
                    </span>
                  </div>
                  <div className="flex items-start gap-1.5 text-[11px] text-muted-foreground pt-0.5">
                    <MapPin className="h-3.5 w-3.5 shrink-0 text-primary mt-0.5" />
                    <span className="leading-snug">{supplier.address || supplier.location || t("specOriginCountryValue")}</span>
                  </div>
                </div>

                {/* Direct Channels (WhatsApp & Email only - Minimalist & Balanced) */}
                <div className="grid grid-cols-2 gap-2.5 pt-3 border-t border-border">
                  {/* WhatsApp */}
                  {(() => {
                    const hasWhatsApp = Boolean(supplier.whatsApp);
                    const raw = supplier.whatsApp || "";
                    const digits = raw.replace(/\D/g, "");
                    const waNum = digits.startsWith("0") ? "62" + digits.slice(1) : digits;
                    const waText = encodeURIComponent(
                      t("waChatTemplate", { supplier: supplier.companyName, product: product.name })
                    );
                    const waLink = `https://wa.me/${waNum}?text=${waText}`;

                    return (
                      <Button
                        asChild={hasWhatsApp}
                        disabled={!hasWhatsApp}
                        variant="outline"
                        className={`w-full text-xs font-semibold border-border rounded-lg h-10 px-3 flex items-center justify-center gap-2 transition-all duration-300 ${
                          hasWhatsApp
                            ? "cursor-pointer hover:bg-muted text-foreground hover:-translate-y-0.5 active:translate-y-0"
                            : "opacity-50 cursor-not-allowed text-muted-foreground"
                        }`}
                        title={!hasWhatsApp ? t("whatsappNotSet") : undefined}
                      >
                        {hasWhatsApp ? (
                          <a href={waLink} target="_blank" rel="noopener noreferrer">
                            <MessageCircle className="h-4 w-4 text-emerald-600 dark:text-emerald-400 shrink-0" />
                            <span className="whitespace-nowrap">{t("btnWhatsapp")}</span>
                          </a>
                        ) : (
                          <span>
                            <MessageCircle className="h-4 w-4 shrink-0" />
                            <span className="whitespace-nowrap">{t("btnWhatsapp")}</span>
                          </span>
                        )}
                      </Button>
                    );
                  })()}

                  {/* Email */}
                  <Button
                    asChild={Boolean(supplier.email)}
                    disabled={!supplier.email}
                    variant="outline"
                    className={`w-full text-xs font-semibold border-border rounded-lg h-10 px-3 flex items-center justify-center gap-2 transition-all duration-300 ${
                      supplier.email
                        ? "cursor-pointer hover:bg-muted text-foreground hover:-translate-y-0.5 active:translate-y-0"
                        : "opacity-50 cursor-not-allowed text-muted-foreground"
                    }`}
                    title={!supplier.email ? t("emailNotSet") : undefined}
                  >
                    {supplier.email ? (
                      <a
                        href={`mailto:${supplier.email}?subject=${encodeURIComponent(
                          t("emailSubjectTemplate", { product: product.name })
                        )}&body=${encodeURIComponent(
                          t("emailBodyTemplate", { supplier: supplier.companyName, product: product.name })
                        )}`}
                      >
                        <Mail className="h-4 w-4 text-primary shrink-0" />
                        <span className="whitespace-nowrap">{t("btnMessage")}</span>
                      </a>
                    ) : (
                      <span>
                        <Mail className="h-4 w-4 shrink-0" />
                        <span className="whitespace-nowrap">{t("btnMessage")}</span>
                      </span>
                    )}
                  </Button>
                </div>

                {/* Toolbar: Wishlist, Compare, Share, Store Detail */}
                <div className="border-t border-border pt-3">
                  <div className="flex items-center justify-around py-1 text-xs text-muted-foreground">
                    <button
                      type="button"
                      onClick={toggleBookmark}
                      disabled={isAdding}
                      className="flex items-center gap-1.5 hover:text-foreground cursor-pointer transition-colors px-2 py-1 rounded"
                    >
                      <Heart className={`h-3.5 w-3.5 ${isBookmarked ? "fill-destructive text-destructive" : ""}`} />
                      <span className="text-[11px] font-medium whitespace-nowrap">{t("btnWishlist")}</span>
                    </button>
                    <span className="h-3 w-px bg-border" />
                    <button
                      type="button"
                      onClick={toggleCompare}
                      disabled={isComparing}
                      className="flex items-center gap-1.5 hover:text-foreground cursor-pointer transition-colors px-2 py-1 rounded"
                    >
                      <GitCompareArrows className="h-3.5 w-3.5" />
                      <span className="text-[11px] font-medium whitespace-nowrap">
                        {isCompared ? t("removeCompare") : t("addCompare")}
                      </span>
                    </button>
                    <span className="h-3 w-px bg-border" />
                    <button
                      type="button"
                      onClick={handleShare}
                      className="flex items-center gap-1.5 hover:text-foreground cursor-pointer transition-colors px-2 py-1 rounded"
                    >
                      {isCopied ? <Check className="h-3.5 w-3.5 text-success" /> : <Share2 className="h-3.5 w-3.5" />}
                      <span className="text-[11px] font-medium whitespace-nowrap">{isCopied ? t("copied") : t("btnShare")}</span>
                    </button>
                  </div>
                  <div className="border-t border-border mt-3 pt-3">
                    <Button
                      asChild
                      variant="ghost"
                      className="w-full cursor-pointer text-xs font-semibold hover:bg-muted rounded-lg h-9 hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300"
                    >
                      <Link href={`${detailBasePath}/suppliers/${supplier.slug}`}>
                        <Store className="mr-1.5 h-3.5 w-3.5" />
                        {t("detailSupplier")}
                      </Link>
                    </Button>
                  </div>
                </div>
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

          {/* Modals */}
          <ProductRfqModal
            isOpen={isRfqModalOpen}
            onClose={() => setIsRfqModalOpen(false)}
            product={product}
            supplier={supplier}
            initialQuantity={quantity}
            selectedVariantText={variants[selectedVariant] || t("rfqDefaultVariant")}
            initialNotes={rfqInitialNotes}
          />
          <ProductDirectBuyModal
            isOpen={isDirectBuyModalOpen}
            onClose={() => setIsDirectBuyModalOpen(false)}
            product={product}
            supplier={supplier}
            quantity={quantity}
            selectedVariantText={variants[selectedVariant] || t("rfqDefaultVariant")}
          />
        </div>

        {/* Bottom Sticky Bar */}
        <div className="fixed bottom-0 inset-x-0 z-30 border-t border-border bg-card/95 backdrop-blur-md px-4 py-2.5 shadow-lg sm:px-6 lg:px-8">
          <div className="mx-auto flex max-w-7xl items-center justify-between gap-4">
            <div className="flex items-center gap-3 min-w-0">
              <div className="h-10 w-10 shrink-0 overflow-hidden rounded-md border border-border bg-muted flex items-center justify-center">
                {photos[0] ? (
                  <img src={photos[0]} alt="" className="h-full w-full object-cover" />
                ) : (
                  <Package className="h-5 w-5 text-muted-foreground" />
                )}
              </div>
              <div className="min-w-0">
                <p className="truncate text-xs font-bold text-foreground max-w-[200px] sm:max-w-md">
                  {product.name}
                </p>
                <p className="text-[11px] text-muted-foreground">
                  {t("stickyTotalPrice")}{" "}
                  <span className="font-extrabold text-primary">
                    {formatPrice(product.price, product.currency) || t("stickyRequestQuote")}
                  </span>
                </p>
              </div>
            </div>
            <Button
              onClick={handleOpenRfqModal}
              className="cursor-pointer font-semibold text-xs h-9 px-5 bg-primary text-primary-foreground hover:bg-primary/90 hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 rounded-lg shrink-0 flex items-center gap-1.5 shadow-xs hover:shadow-primary/30"
            >
              <FileText className="h-3.5 w-3.5" />
              <span>{t("btnRequestQuote")}</span>
            </Button>
          </div>
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
