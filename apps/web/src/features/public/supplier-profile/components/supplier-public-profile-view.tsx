"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";
import { resolveImageUrl } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Field, FieldLabel, FieldGroup } from "@/components/ui/field";
import { toast } from "sonner";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { ProductCard } from "@/components/ui/product-card";
import { StageScrollLoader } from "@/components/ui/stage-scroll-loader";
import type { PublicSupplierDto, PublicProductDto, SupplierProductDto } from "@/features/public/search/types";
import {
  Building,
  Calendar,
  Users,
  MapPin,
  Mail,
  Phone,
  Globe,
  Award,
  Package,
  ChevronLeft,
  MessageSquare,
  Share2,
  Star,
  UserPlus,
  Check,
  GitCompareArrows,
} from "lucide-react";

export type ProfileTab = "home" | "products" | "certifications" | "reviews";

export interface SupplierPublicProfileViewProps {
  supplier: PublicSupplierDto;
  detailBasePath?: "" | "/demo";
  isPreviewMode?: boolean;
  backHref?: string;
  // Tab state & handlers
  activeTab?: ProfileTab;
  onTabChange?: (tab: ProfileTab) => void;
  // Products
  productsList?: PublicProductDto[];
  isProductsLoading?: boolean;
  isProductsLoadingMore?: boolean;
  hasMoreProducts?: boolean;
  onLoadMoreProducts?: () => void;
  // Interactions
  isFollowed?: boolean;
  isMutatingFollowing?: boolean;
  onToggleFollow?: () => void;
  isOpeningChat?: boolean;
  onOpenChat?: () => void;
  onToggleProductBookmark?: (id: string) => void;
  isProductBookmarked?: (id: string) => boolean;
  onToggleProductCompare?: (id: string) => void;
  isProductCompared?: (id: string) => boolean;
  // RFQ Modal
  isRfqOpen?: boolean;
  onOpenRfq?: () => void;
  onCloseRfq?: () => void;
  onSendRfq?: (e: React.FormEvent) => void;
  subject?: string;
  onSubjectChange?: (val: string) => void;
  quantity?: string;
  onQuantityChange?: (val: string) => void;
  message?: string;
  onMessageChange?: (val: string) => void;
  isSubmittingRfq?: boolean;
}

export function SupplierPublicProfileView({
  supplier,
  detailBasePath = "/demo",
  isPreviewMode = false,
  backHref,
  activeTab: controlledActiveTab,
  onTabChange: controlledOnTabChange,
  productsList: controlledProductsList,
  isProductsLoading = false,
  isProductsLoadingMore = false,
  hasMoreProducts = false,
  onLoadMoreProducts,
  isFollowed = false,
  isMutatingFollowing = false,
  onToggleFollow,
  isOpeningChat = false,
  onOpenChat,
  onToggleProductBookmark,
  isProductBookmarked,
  onToggleProductCompare,
  isProductCompared,
  isRfqOpen: controlledIsRfqOpen,
  onOpenRfq: controlledOnOpenRfq,
  onCloseRfq: controlledOnCloseRfq,
  onSendRfq: controlledOnSendRfq,
  subject: controlledSubject,
  onSubjectChange: controlledOnSubjectChange,
  quantity: controlledQuantity,
  onQuantityChange: controlledOnQuantityChange,
  message: controlledMessage,
  onMessageChange: controlledOnMessageChange,
  isSubmittingRfq = false,
}: SupplierPublicProfileViewProps) {
  const [internalActiveTab, setInternalActiveTab] = useState<ProfileTab>("home");
  const [internalIsRfqOpen, setInternalIsRfqOpen] = useState(false);
  const [internalSubject, setInternalSubject] = useState("");
  const [internalQuantity, setInternalQuantity] = useState("1");
  const [internalMessage, setInternalMessage] = useState("");
  const [logoError, setLogoError] = useState(false);

  const tSup = useTranslations("public.supplier");

  const activeTab = controlledActiveTab ?? internalActiveTab;
  const handleTabChange = (tab: string) => {
    const nextTab = tab as ProfileTab;
    if (controlledOnTabChange) {
      controlledOnTabChange(nextTab);
    } else {
      setInternalActiveTab(nextTab);
    }
  };

  const isRfqOpen = controlledIsRfqOpen ?? internalIsRfqOpen;
  const setIsRfqOpen = (open: boolean) => {
    if (open) {
      if (controlledOnOpenRfq) controlledOnOpenRfq();
      else setInternalIsRfqOpen(true);
    } else {
      if (controlledOnCloseRfq) controlledOnCloseRfq();
      else setInternalIsRfqOpen(false);
    }
  };

  const subject = controlledSubject ?? internalSubject;
  const setSubject = controlledOnSubjectChange ?? setInternalSubject;
  const quantity = controlledQuantity ?? internalQuantity;
  const setQuantity = controlledOnQuantityChange ?? setInternalQuantity;
  const message = controlledMessage ?? internalMessage;
  const setMessage = controlledOnMessageChange ?? setInternalMessage;

  const handleToggleFollowAction = () => {
    if (isPreviewMode) {
      toast.info("Mode Pratinjau: Aksi 'Mengikuti' disimulasikan untuk pembeli");
      return;
    }
    onToggleFollow?.();
  };

  const handleOpenChatAction = () => {
    if (isPreviewMode) {
      toast.info("Mode Pratinjau: Fitur Chat Penjual disimulasikan untuk pembeli");
      return;
    }
    onOpenChat?.();
  };

  const handleOpenRfqAction = () => {
    if (isPreviewMode) {
      toast.info("Mode Pratinjau: Fitur Permintaan Penawaran (RFQ) disimulasikan untuk pembeli");
      return;
    }
    if (controlledOnOpenRfq) controlledOnOpenRfq();
    else setInternalIsRfqOpen(true);
  };

  const handleSendRFQAction = (e: React.FormEvent) => {
    e.preventDefault();
    if (isPreviewMode) {
      toast.info("Mode Pratinjau: Permintaan Penawaran disimulasikan");
      setIsRfqOpen(false);
      return;
    }
    controlledOnSendRfq?.(e);
  };

  // Map local products to PublicProductDto
  const mapSupplierProductToPublicProduct = (p: SupplierProductDto): PublicProductDto => {
    return {
      id: p.id,
      name: p.name,
      description: p.description || "",
      price: p.price || 0,
      currency: p.currency || "IDR",
      minOrder: p.minOrder || "1",
      capacityText: p.capacityText || "",
      categoryName: p.categoryName || "",
      photos: p.photos || [],
      supplierId: supplier.id,
      supplierCompanyName: supplier.companyName,
      supplierSlug: supplier.slug,
      supplierLocation: supplier.location,
      supplierVerified: supplier.isVerified,
      supplierRating: supplier.rating,
      supplierReviewCount: supplier.reviewCount,
    };
  };

  const productsList =
    controlledProductsList ??
    (supplier.products ? supplier.products.map(mapSupplierProductToPublicProduct) : []);

  const sectionCardClass = "overflow-hidden rounded-lg border border-border bg-card shadow-xs";

  const renderProfileProductCard = (product: PublicProductDto) => {
    const hasDiscount = (product.price || 0) > 0 && product.id.charCodeAt(0) % 2 === 0;
    const discountPercentage = hasDiscount ? (product.id.charCodeAt(1) % 3 === 0 ? 30 : 50) : 0;
    const originalPrice = hasDiscount
      ? Math.round((product.price || 0) * (100 / (100 - discountPercentage)))
      : 0;
    const mockSoldCount =
      product.id.charCodeAt(2) % 2 === 0
        ? `${(product.id.charCodeAt(3) % 9) + 1}rb+ terjual`
        : `${(product.id.charCodeAt(3) % 80) + 10}+ terjual`;

    const isCompared = isProductCompared?.(product.id) ?? false;

    return (
      <ProductCard
        key={product.id}
        id={product.id}
        name={product.name}
        price={product.price}
        currency={product.currency}
        image={product.photos?.[0]}
        href={isPreviewMode ? "#" : `${detailBasePath}/products/${product.id}`}
        discountPercentage={discountPercentage || undefined}
        originalPrice={originalPrice || undefined}
        rating={product.supplierRating || supplier.rating}
        soldCountText={mockSoldCount}
        categoryName={product.categoryName || undefined}
        minOrder={product.minOrder || undefined}
        moqLabel={tSup("moqLabel")}
        priceLabel={tSup("na")}
        showBookmarkOverlayButton={!isPreviewMode}
        isBookmarked={isProductBookmarked?.(product.id) ?? false}
        onBookmark={() => onToggleProductBookmark?.(product.id)}
        customFooter={
          <Button
            type="button"
            variant="outline"
            onClick={(e) => {
              e.preventDefault();
              e.stopPropagation();
              if (isPreviewMode) {
                toast.info("Mode Pratinjau: Komparasi disimulasikan");
                return;
              }
              onToggleProductCompare?.(product.id);
            }}
            className={`w-full py-1.5 text-[11px] font-bold border rounded-lg cursor-pointer transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-md ${
              isCompared
                ? "bg-success/5 border-success text-success hover:bg-success/10 hover:shadow-success/10"
                : "border-primary/45 text-primary bg-primary/0 hover:bg-primary hover:text-primary-foreground hover:shadow-primary/10"
            }`}
          >
            {isCompared ? (
              <>
                <Check className="h-3 w-3 mr-1 shrink-0" />
                {tSup("addedLabel")}
              </>
            ) : (
              <>
                <GitCompareArrows className="h-3 w-3 mr-1 shrink-0" />
                {tSup("compareBtn")}
              </>
            )}
          </Button>
        }
      />
    );
  };

  return (
    <div className="space-y-6 text-left">
      {/* Back button (only shown when not in preview mode or when explicit backHref provided) */}
      {!isPreviewMode && (
        <Link
          href={backHref || `${detailBasePath}/search`}
          className="inline-flex items-center gap-1 text-xs font-semibold text-muted-foreground hover:text-primary transition-all duration-300 mb-2 cursor-pointer"
        >
          <ChevronLeft className="h-3.5 w-3.5" />
          {tSup("backButton")}
        </Link>
      )}

      {/* Profile Header (Clean Tokopedia Shop Detail Inspired) */}
      <div className="bg-card border border-border shadow-xs rounded-lg p-6 md:p-8 flex flex-col md:flex-row items-start md:items-center justify-between gap-6 transition-all duration-300 hover:shadow-md">
        <div className="flex flex-col md:flex-row items-start md:items-center gap-5 w-full md:w-auto">
          {/* Shop Logo Avatar */}
          <div
            className="h-18 w-18 border border-border bg-muted flex items-center justify-center shadow-xs shrink-0 overflow-hidden relative"
            style={{ borderRadius: "9999px" }}
          >
            {supplier.logo && !logoError ? (
              // eslint-disable-next-line @next/next/no-img-element
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
                className="flex h-full w-full items-center justify-center bg-primary text-primary-foreground font-heading font-bold text-2xl"
                style={{ borderRadius: "9999px" }}
              >
                {supplier.companyName ? supplier.companyName.substring(0, 2).toUpperCase() : "PT"}
              </div>
            )}
          </div>
          <div className="space-y-2.5">
            <div className="flex flex-wrap items-center gap-2">
              <h1 className="text-xl font-bold text-foreground font-heading tracking-tight leading-none md:text-2xl">
                {supplier.companyName}
              </h1>
              {supplier.isVerified ? (
                <Badge
                  variant="outline"
                  className="border-success text-success bg-success/5 font-semibold rounded-lg px-2.5 py-0.5 text-[10px]"
                >
                  {tSup("verifiedSupplier")}
                </Badge>
              ) : (
                <Badge
                  variant="outline"
                  className="border-border text-muted-foreground font-semibold rounded-lg px-2.5 py-0.5 text-[10px]"
                >
                  {tSup("notVerified")}
                </Badge>
              )}
            </div>
            {supplier.description && (
              <p className="text-xs text-muted-foreground leading-relaxed max-w-2xl whitespace-pre-line">
                {supplier.description}
              </p>
            )}
            <div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground font-medium">
              <div className="flex items-center gap-1">
                <MapPin className="h-3.5 w-3.5 text-primary/70" />
                <span>{supplier.location}</span>
              </div>
              <span className="w-1.5 h-1.5 rounded-full bg-border" />
              <div className="flex items-center gap-1 font-semibold text-foreground">
                <Star className="h-3.5 w-3.5 fill-warning text-warning" />
                <span>
                  {supplier.rating?.toFixed(1) || "0.0"} ({supplier.reviewCount || 0}{" "}
                  {tSup("tabReviews")})
                </span>
              </div>
            </div>

            {/* Actions Row */}
            <div className="flex flex-wrap items-center gap-2 pt-1.5">
              <Button
                onClick={handleToggleFollowAction}
                variant={isFollowed ? "outline" : "default"}
                size="sm"
                className={`cursor-pointer rounded-lg text-xs transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 ${
                  isFollowed
                    ? "border-primary/25 bg-primary/5 text-primary hover:bg-primary/10"
                    : "bg-primary text-primary-foreground hover:bg-primary/95 hover:shadow-md hover:shadow-primary/20"
                }`}
                disabled={isMutatingFollowing}
              >
                <UserPlus className="h-3.5 w-3.5 mr-1" />
                {isFollowed ? tSup("following") : tSup("follow")}
              </Button>
              <Button
                onClick={handleOpenChatAction}
                variant="outline"
                size="sm"
                className="cursor-pointer rounded-lg text-xs hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300"
                disabled={isOpeningChat}
              >
                <MessageSquare className="h-3.5 w-3.5 mr-1" />
                Chat Penjual
              </Button>
              <Button
                onClick={() => {
                  if (typeof window !== "undefined") {
                    navigator.clipboard.writeText(window.location.href);
                    toast.success(tSup("linkCopied"));
                  }
                }}
                variant="outline"
                size="sm"
                className="cursor-pointer rounded-lg text-xs hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300"
              >
                <Share2 className="h-3.5 w-3.5 mr-1" />
                Share
              </Button>
            </div>
          </div>
        </div>

        {/* Shop Metrics Block */}
        <div className="flex flex-row md:flex-col items-center md:items-end justify-between md:justify-center border-t md:border-t-0 md:border-l border-border pt-4 md:pt-0 md:pl-8 w-full md:w-auto gap-4">
          <div className="text-left md:text-right">
            <div className="flex items-center md:justify-end gap-1">
              <Star className="h-4.5 w-4.5 fill-warning text-warning" />
              <span className="text-lg font-black text-foreground">
                {supplier.rating?.toFixed(1) || "0.0"}
              </span>
              <span className="text-[10px] text-muted-foreground">/ 5.0</span>
            </div>
            <p className="text-[9px] text-muted-foreground font-bold uppercase tracking-wider mt-0.5">
              {tSup("ratingReviews")}
            </p>
          </div>
          <div className="text-right">
            <p className="text-base font-black text-primary">50+ Terjual</p>
            <p className="text-[9px] text-muted-foreground font-bold uppercase tracking-wider mt-0.5">
              {tSup("successfulTransactions")}
            </p>
          </div>
        </div>
      </div>

      {/* Navigation Tabs (Standard Workspace UI Component) */}
      <Tabs value={activeTab} onValueChange={handleTabChange} className="w-full">
        <TabsList className="w-full justify-start">
          <TabsTrigger value="home" className="cursor-pointer font-bold">
            {tSup("tabHome")}
          </TabsTrigger>
          <TabsTrigger value="products" className="cursor-pointer font-bold">
            {tSup("tabProducts")}
          </TabsTrigger>
          <TabsTrigger value="certifications" className="cursor-pointer font-bold">
            {tSup("tabCertifications")}
          </TabsTrigger>
          <TabsTrigger value="reviews" className="cursor-pointer font-bold">
            {tSup("tabReviews")}
          </TabsTrigger>
        </TabsList>
      </Tabs>

      {/* Render Tabs Contents */}
      <div className="mt-2">
        {/* BERANDA TAB (2-Column layout with sidebar specs) */}
        {activeTab === "home" && (
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Main Content (left) */}
            <div className="lg:col-span-2 space-y-8">
              {/* B2B Promo Banner */}
              <div className="relative overflow-hidden rounded-lg border border-primary/10 bg-linear-to-r from-primary/10 via-primary/[0.02] to-cyan-500/5 p-6 shadow-xs flex flex-col md:flex-row items-center justify-between gap-6">
                <div className="absolute top-2.5 right-2.5 flex items-center gap-1 rounded bg-primary/20 px-2 py-0.5 text-[9px] font-bold text-primary uppercase tracking-wider">
                  Mitra B2B
                </div>
                <div className="space-y-2 max-w-md">
                  <h2 className="text-base font-black text-foreground uppercase tracking-tight">
                    Kemitraan Industri & Kontrak Kustom
                  </h2>
                  <p className="text-xs text-muted-foreground leading-relaxed">
                    Kami melayani kontrak kustom jangka panjang, negosiasi MOQ khusus, serta opsi
                    logistik terintegrated untuk kebutuhan industri Anda.
                  </p>
                </div>
                <Button
                  onClick={handleOpenRfqAction}
                  size="sm"
                  className="shrink-0 cursor-pointer rounded-lg bg-primary hover:bg-primary/95 text-primary-foreground font-semibold hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300"
                >
                  Kirim RFQ Kustom
                </Button>
              </div>

              {/* Spotlight / Featured Products Row */}
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <h3 className="text-sm font-bold font-heading text-foreground">
                    Sesuai incaran kamu di toko ini
                  </h3>
                  <button
                    type="button"
                    onClick={() => handleTabChange("products")}
                    className="text-xs font-bold text-primary hover:underline cursor-pointer transition-colors"
                  >
                    Lihat Semua
                  </button>
                </div>
                {productsList.length === 0 ? (
                  <div className="text-center py-8 bg-muted/20 border border-dashed border-border rounded-lg">
                    <Package className="mx-auto h-8 w-8 text-muted-foreground/35" />
                    <h3 className="mt-3 text-xs font-bold text-foreground">
                      {tSup("emptyProductsTitle")}
                    </h3>
                  </div>
                ) : (
                  <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
                    {productsList.slice(0, 3).map((product) => renderProfileProductCard(product))}
                  </div>
                )}
              </div>
            </div>

            {/* Sidebar Info (right) */}
            <div className="space-y-6">
              {/* Business Specs */}
              <Card className={sectionCardClass}>
                <div className="border-b border-border px-4 py-3 bg-transparent">
                  <h3 className="text-xs font-bold font-heading uppercase tracking-wider text-muted-foreground/90">
                    {tSup("businessInfo")}
                  </h3>
                </div>
                <CardContent className="p-0">
                  <div className="divide-y divide-border text-xs">
                    <div className="flex justify-between px-4 py-3">
                      <span className="text-muted-foreground font-medium flex items-center gap-2">
                        <Building className="h-3.5 w-3.5 text-primary/70 shrink-0" />
                        {tSup("businessType")}
                      </span>
                      <span className="font-bold text-foreground">
                        {supplier.businessType || tSup("na")}
                      </span>
                    </div>
                    <div className="flex justify-between px-4 py-3">
                      <span className="text-muted-foreground font-medium flex items-center gap-2">
                        <Calendar className="h-3.5 w-3.5 text-primary/70 shrink-0" />
                        {tSup("established")}
                      </span>
                      <span className="font-bold text-foreground">
                        {supplier.establishedYear || tSup("na")}
                      </span>
                    </div>
                    <div className="flex justify-between px-4 py-3">
                      <span className="text-muted-foreground font-medium flex items-center gap-2">
                        <Users className="h-3.5 w-3.5 text-primary/70 shrink-0" />
                        {tSup("employees")}
                      </span>
                      <span className="font-bold text-foreground">
                        {supplier.employeeCount ? `${supplier.employeeCount} Orang` : tSup("na")}
                      </span>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* Contact Info */}
              <Card className={sectionCardClass}>
                <div className="border-b border-border px-4 py-3 bg-transparent">
                  <h3 className="text-xs font-bold font-heading uppercase tracking-wider text-muted-foreground/90">
                    {tSup("contactInfo")}
                  </h3>
                </div>
                <CardContent className="space-y-3.5 p-4 text-xs text-muted-foreground">
                  <div className="flex items-center gap-3">
                    <Mail className="h-3.5 w-3.5 text-primary/70 shrink-0" />
                    <span className="font-medium text-foreground">
                      {supplier.email || tSup("na")}
                    </span>
                  </div>
                  <div className="flex items-center gap-3">
                    <Phone className="h-3.5 w-3.5 text-primary/70 shrink-0" />
                    <span className="font-medium text-foreground">
                      {supplier.phone || tSup("na")}
                    </span>
                  </div>
                  <div className="flex items-center gap-3">
                    <Globe className="h-3.5 w-3.5 text-primary/70 shrink-0" />
                    {supplier.website ? (
                      <a
                        href={supplier.website}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-primary hover:underline font-bold transition-all"
                      >
                        {supplier.website}
                      </a>
                    ) : (
                      <span className="font-medium">{tSup("na")}</span>
                    )}
                  </div>
                </CardContent>
              </Card>

              {/* Ask for Quotation card */}
              <Card className={sectionCardClass}>
                <div className="border-b border-border px-4 py-3 bg-transparent">
                  <h3 className="text-xs font-bold font-heading uppercase tracking-wider text-muted-foreground/90">
                    {tSup("requestQuote")}
                  </h3>
                </div>
                <CardContent className="p-4 space-y-4">
                  <p className="text-xs text-muted-foreground leading-relaxed">
                    Butuh penawaran harga khusus, negosiasi kuantitas besar, atau penyesuaian produk?
                    Kirim permintaan penawaran (RFQ) langsung ke supplier ini.
                  </p>
                  <Button
                    onClick={handleOpenRfqAction}
                    className="w-full text-xs font-bold cursor-pointer rounded-lg bg-primary hover:bg-primary/95 text-primary-foreground transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 shadow-xs"
                  >
                    {tSup("quoteTitle")}
                  </Button>
                </CardContent>
              </Card>
            </div>
          </div>
        )}

        {/* PRODUK TAB (Full width grid) */}
        {activeTab === "products" && (
          <div className="space-y-6">
            <div className="flex flex-col gap-1.5">
              <h2 className="text-lg font-bold text-foreground">{tSup("products")}</h2>
              <p className="text-xs text-muted-foreground">
                Katalog lengkap produk manufaktur dan suplai.
              </p>
            </div>

            {isProductsLoading && productsList.length === 0 ? (
              <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4 animate-pulse">
                {Array.from({ length: 12 }).map((_, i) => (
                  <div
                    key={i}
                    className="flex flex-col justify-between overflow-hidden rounded-lg border border-border bg-card p-4 h-[320px]"
                  >
                    <div className="aspect-square w-full bg-muted-foreground/10 rounded-md" />
                    <div className="h-4 w-3/4 bg-muted-foreground/10 rounded-md mt-3" />
                    <div className="h-4 w-1/2 bg-muted-foreground/10 rounded-md mt-1.5" />
                    <div className="h-8 w-full bg-muted-foreground/10 rounded-md mt-4" />
                  </div>
                ))}
              </div>
            ) : productsList.length === 0 ? (
              <div className="text-center py-16 bg-card border border-border rounded-lg shadow-xs">
                <Package className="mx-auto h-12 w-12 text-muted-foreground/35 stroke-[1.5]" />
                <h3 className="mt-4 text-sm font-bold text-foreground">
                  {tSup("emptyProductsTitle")}
                </h3>
                <p className="text-xs text-muted-foreground mt-1">{tSup("emptyProductsDesc")}</p>
              </div>
            ) : (
              <>
                <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4 animate-fade-in">
                  {productsList.map((product) => renderProfileProductCard(product))}
                </div>
                {onLoadMoreProducts && (
                  <StageScrollLoader
                    onLoadMore={onLoadMoreProducts}
                    hasMore={hasMoreProducts}
                    isLoading={isProductsLoadingMore}
                  />
                )}
              </>
            )}
          </div>
        )}

        {/* CERTIFICATIONS TAB */}
        {activeTab === "certifications" && (
          <div className="space-y-6">
            <div className="flex flex-col gap-1.5">
              <h2 className="text-lg font-bold text-foreground">{tSup("certifications")}</h2>
              <p className="text-xs text-muted-foreground">
                Daftar sertifikat resmi terverifikasi tim penilai IndoSupplier.
              </p>
            </div>

            {!supplier.certificationList || supplier.certificationList.length === 0 ? (
              <div className="text-center py-16 bg-card border border-dashed border-border rounded-lg shadow-xs">
                <Award className="mx-auto h-12 w-12 text-muted-foreground/35 stroke-[1.5]" />
                <h3 className="mt-4 text-sm font-bold text-foreground">
                  {tSup("emptyCertsTitle")}
                </h3>
                <p className="mt-1 text-xs text-muted-foreground">{tSup("emptyCertsDesc")}</p>
              </div>
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {supplier.certificationList.map((cert) => (
                  <div
                    key={cert.id}
                    className="flex items-center justify-between p-4.5 border border-border rounded-lg bg-card transition-all duration-300 hover:border-primary/20 hover:shadow-xs"
                  >
                    <div className="flex items-center gap-3">
                      <div className="h-10 w-10 bg-primary/5 text-primary flex items-center justify-center rounded-lg border border-primary/10 shrink-0">
                        <Award className="h-5 w-5" />
                      </div>
                      <div>
                        <p className="text-sm font-bold text-foreground">{cert.name}</p>
                        <p className="text-xs text-muted-foreground mt-0.5">
                          Diterbitkan oleh: {cert.institution} {cert.year ? `(${cert.year})` : ""}
                        </p>
                      </div>
                    </div>
                    <Badge
                      variant="outline"
                      className="border-success text-success bg-success/5 font-semibold rounded-lg px-2.5 py-0.5 text-[10px]"
                    >
                      {tSup("verifiedSupplier")}
                    </Badge>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* REVIEWS TAB */}
        {activeTab === "reviews" && (
          <div className="space-y-6">
            <div className="flex flex-col gap-1.5">
              <h2 className="text-lg font-bold text-foreground">Ulasan Pembeli</h2>
              <p className="text-xs text-muted-foreground">
                Feedback asli dari transaksi terverifikasi.
              </p>
            </div>

            {/* Reviews List */}
            <div className="space-y-4">
              {!supplier.reviews || supplier.reviews.length === 0 ? (
                <div className="rounded-lg border border-dashed border-border py-12 text-center text-xs text-muted-foreground bg-card">
                  {tSup("emptyReviews")}
                </div>
              ) : (
                supplier.reviews.map((review) => (
                  <article
                    key={review.id}
                    className="rounded-lg border border-border bg-card p-5 transition-all duration-300 hover:shadow-xs hover:border-border/80"
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div className="flex items-center gap-3">
                        <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/5 text-xs font-bold text-primary uppercase">
                          {review.buyerName.slice(0, 2)}
                        </div>
                        <div>
                          <p className="text-sm font-bold text-foreground">{review.buyerName}</p>
                          <p className="text-[10px] text-muted-foreground mt-0.5">
                            {new Intl.DateTimeFormat("id-ID", {
                              day: "numeric",
                              month: "short",
                              year: "numeric",
                            }).format(new Date(review.createdAt))}
                          </p>
                        </div>
                      </div>
                      <div className="flex items-center gap-1 rounded-lg bg-muted/40 px-2 py-1 border border-border text-xs font-bold">
                        <Star className="h-3.5 w-3.5 fill-warning text-warning" />
                        <span>{review.rating}</span>
                      </div>
                    </div>

                    <p className="mt-3 text-xs leading-relaxed text-foreground/90 pl-1">
                      {review.reviewText}
                    </p>

                    {review.supplierReply && (
                      <div className="relative mt-3 rounded-lg border border-primary/10 bg-primary/1 p-3.5 pl-8 text-xs text-muted-foreground">
                        <span className="absolute left-3 top-4 h-1.5 w-1.5 rounded-full bg-primary/40" />
                        <p className="font-bold text-primary">{tSup("supplierReply")}</p>
                        <p className="mt-1 text-foreground/80 leading-relaxed font-medium">
                          {review.supplierReply}
                        </p>
                      </div>
                    )}
                  </article>
                ))
              )}
            </div>
          </div>
        )}
      </div>

      {/* Quotation Dialog Form */}
      <Dialog open={isRfqOpen} onOpenChange={setIsRfqOpen}>
        <DialogContent size="lg" className="sm:max-w-lg rounded-lg border-border bg-card">
          <DialogHeader>
            <DialogTitle>{tSup("quoteTitle")}</DialogTitle>
            <DialogDescription>
              Lengkapi detail di bawah ini untuk mengirim permintaan penawaran kepada{" "}
              {supplier.companyName}.
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleSendRFQAction} className="space-y-4 mt-2">
            <FieldGroup className="space-y-3.5">
              <Field className="space-y-1.5">
                <FieldLabel className="text-xs font-bold text-foreground">
                  {tSup("quoteSubject")}
                </FieldLabel>
                <Input
                  type="text"
                  placeholder="e.g. Bulk Procurement for Teak Wood chairs"
                  value={subject}
                  onChange={(e) => setSubject(e.target.value)}
                  required
                  className="bg-card border-border text-xs focus-visible:border-primary focus-visible:ring-1 focus-visible:ring-primary rounded-lg h-9"
                />
              </Field>

              <Field className="space-y-1.5">
                <FieldLabel className="text-xs font-bold text-foreground">
                  {tSup("quoteQuantity")}
                </FieldLabel>
                <Input
                  type="number"
                  min="1"
                  value={quantity}
                  onChange={(e) => setQuantity(e.target.value)}
                  required
                  className="bg-card border-border text-xs focus-visible:border-primary focus-visible:ring-1 focus-visible:ring-primary rounded-lg h-9"
                />
              </Field>

              <Field className="space-y-1.5">
                <FieldLabel className="text-xs font-bold text-foreground">
                  {tSup("quoteMessage")}
                </FieldLabel>
                <Textarea
                  rows={4}
                  placeholder={tSup("quoteMessage")}
                  value={message}
                  onChange={(e) => setMessage(e.target.value)}
                  required
                  className="bg-card border-border text-xs focus-visible:border-primary focus-visible:ring-1 focus-visible:ring-primary rounded-lg resize-none p-3"
                />
              </Field>
            </FieldGroup>
            <div className="flex justify-end gap-2 pt-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setIsRfqOpen(false)}
                className="rounded-lg text-xs font-bold px-4 h-9 cursor-pointer"
              >
                {tSup("quoteCancel")}
              </Button>
              <Button
                type="submit"
                disabled={isSubmittingRfq}
                className="bg-primary text-primary-foreground hover:bg-primary/95 font-bold cursor-pointer rounded-lg text-xs h-9 hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 hover:shadow-md hover:shadow-primary/20"
              >
                {isSubmittingRfq ? tSup("quoteSending") : tSup("btnSendQuote")}
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
