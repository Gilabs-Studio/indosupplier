"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { useQuery } from "@tanstack/react-query";
import { Link, useRouter } from "@/i18n/routing";
import { PublicLayout } from "@/features/public/components/public-layout";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Field, FieldLabel, FieldGroup } from "@/components/ui/field";
import { toast } from "sonner";
import { searchService } from "@/features/public/search/services/search-service";
import { useBuyerBookmarks } from "@/features/buyer/bookmarks/hooks/useBuyerBookmarks";
import { useBuyerFollowing } from "@/features/buyer/following/hooks/useBuyerFollowing";
import { useBuyerCompare } from "@/features/buyer/compare/hooks/useBuyerCompare";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import { chatService } from "@/features/buyer/chat/services/chat.service";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { PublicSupplierProductCard } from "./public-supplier-product-card";
import { StageScrollLoader } from "@/components/ui/stage-scroll-loader";
import type { SupplierProductDto, PublicProductDto } from "@/features/public/search/types";
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
} from "lucide-react";

interface PublicSupplierProfilePageProps {
  locale: string;
  slug: string;
  detailBasePath?: "" | "/demo";
}

export function PublicSupplierProfilePage({ locale, slug, detailBasePath = "/demo" }: PublicSupplierProfilePageProps) {
  const tSup = useTranslations("public.supplier");
  const router = useRouter();
  const { isAuthenticated } = useAuthStore();
  const { bookmarks, addBookmark, deleteBookmark } = useBuyerBookmarks();
  const { isFollowingSupplier, followSupplier, unfollowSupplier, isMutating: isMutatingFollowing } = useBuyerFollowing();
  const { products: comparedProducts, addProduct, removeProduct } = useBuyerCompare();
  const [isOpeningChat, setIsOpeningChat] = useState(false);

  // Fetch Supplier profile by slug
  const { data: supplier, isLoading, isError } = useQuery({
    queryKey: ["public-supplier-profile", slug],
    queryFn: () => searchService.getSupplierBySlug(slug),
  });

  // State for products infinite loading
  const [productsList, setProductsList] = useState<PublicProductDto[]>([]);
  const [productsPage, setProductsPage] = useState(1);
  const [hasMoreProducts, setHasMoreProducts] = useState(true);
  const [isProductsLoading, setIsProductsLoading] = useState(false);
  const [isProductsLoadingMore, setIsProductsLoadingMore] = useState(false);

  // Mock RFQ form state
  const [subject, setSubject] = useState("");
  const [message, setMessage] = useState("");
  const [quantity, setQuantity] = useState("1");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isRfqOpen, setIsRfqOpen] = useState(false);

  const [activeTab, setActiveTab] = useState<"home" | "products" | "certifications" | "reviews">("home");

  const fetchProductsList = async (pageNumber: number, isLoadMore = false) => {
    if (!supplier) return;
    if (isLoadMore) {
      setIsProductsLoadingMore(true);
    } else {
      setIsProductsLoading(true);
    }

    try {
      const res = await searchService.searchProducts("", supplier.id, pageNumber, 12);
      if (isLoadMore) {
        setProductsList((prev) => [...prev, ...res]);
      } else {
        setProductsList(res);
      }
      setProductsPage(pageNumber);
      setHasMoreProducts(res.length === 12);
    } catch (error) {
      console.error("Error loading products:", error);
      toast.error("Gagal memuat produk supplier.");
    } finally {
      setIsProductsLoading(false);
      setIsProductsLoadingMore(false);
    }
  };

  const handleTabChange = (val: string) => {
    const tab = val as "home" | "products" | "certifications" | "reviews";
    setActiveTab(tab);
    if (tab === "products" && productsList.length === 0) {
      fetchProductsList(1, false);
    }
  };

  const handleLoadMoreProducts = () => {
    if (isProductsLoading || isProductsLoadingMore || !hasMoreProducts) return;
    fetchProductsList(productsPage + 1, true);
  };

  const mapSupplierProductToPublicProduct = (prod: SupplierProductDto) => {
    if (!supplier) {
      throw new Error("Supplier data is required to map products.");
    }
    return {
      id: prod.id,
      name: prod.name,
      description: prod.description || "",
      price: prod.price || 0,
      currency: prod.currency || "IDR",
      minOrder: prod.minOrder || "",
      capacityText: prod.capacityText || "",
      categoryName: prod.categoryName || "",
      photos: prod.photos || [],
      supplierId: supplier.id,
      supplierCompanyName: supplier.companyName,
      supplierSlug: supplier.slug,
      supplierLocation: supplier.location,
      supplierVerified: supplier.isVerified,
      supplierRating: supplier.rating,
      supplierReviewCount: supplier.reviewCount,
    };
  };

  const currentPath = `${detailBasePath}/suppliers/${slug}`;
  const sectionCardClass = "overflow-hidden rounded-lg border border-border/40 bg-card shadow-xs";
  const isFollowed = supplier ? isFollowingSupplier(supplier.id) : false;

  const requireAuth = (msg: string) => {
    if (isAuthenticated) return true;
    toast.error(msg);
    router.push(`/login?redirectTo=${encodeURIComponent(currentPath)}`);
    return false;
  };

  const getProductBookmarkId = (prodId: string) => {
    const found = bookmarks.find(
      (b) => b.type === "product" && b.supplierProductId === prodId
    );
    return found?.id || null;
  };

  const handleToggleProductBookmark = (prodId: string) => {
    if (!requireAuth("Silakan masuk terlebih dahulu untuk menyimpan produk.")) return;
    const bookmarkId = getProductBookmarkId(prodId);
    if (bookmarkId) {
      deleteBookmark(bookmarkId);
    } else {
      if (supplier) {
        addBookmark({ supplierProfileId: supplier.id, supplierProductId: prodId });
      }
    }
  };

  const handleToggleProductCompare = (prodId: string) => {
    if (!requireAuth("Silakan masuk terlebih dahulu untuk membandingkan produk.")) return;
    if (comparedProducts.some((p) => p.id === prodId)) {
      removeProduct(prodId);
    } else {
      addProduct(prodId);
    }
  };

  const handleToggleFollow = () => {
    if (!supplier) {
      requireAuth("Silakan masuk terlebih dahulu untuk mengikuti supplier.");
      return;
    }
    if (!requireAuth("Silakan masuk terlebih dahulu untuk mengikuti supplier.")) return;
    if (isFollowed) {
      unfollowSupplier(supplier.id);
      return;
    }
    followSupplier(supplier.id);
  };

  const handleOpenChat = async () => {
    if (!supplier) return;
    if (!requireAuth("Silakan masuk terlebih dahulu untuk chat dengan supplier.")) return;

    try {
      setIsOpeningChat(true);
      const room = await chatService.getOrCreateRoom(supplier.id);
      router.push(`/chat?roomId=${encodeURIComponent(room.id)}`);
    } catch (error) {
      console.error(error);
      toast.error("Gagal membuka chat supplier.");
    } finally {
      setIsOpeningChat(false);
    }
  };

  const handleOpenRfq = () => {
    if (!requireAuth("Silakan masuk terlebih dahulu untuk mengirim RFQ.")) return;
    setIsRfqOpen(true);
  };

  const handleSendRFQ = (e: React.FormEvent) => {
    e.preventDefault();
    if (!subject.trim() || !message.trim()) {
      toast.error("Please fill in all required fields.");
      return;
    }

    setIsSubmitting(true);
    setTimeout(() => {
      toast.success(tSup("quoteSuccess"));
      setSubject("");
      setMessage("");
      setQuantity("1");
      setIsSubmitting(false);
      setIsRfqOpen(false);
    }, 1000);
  };

  if (isLoading) {
    return (
      <PublicLayout locale={locale}>
        <div className="bg-muted/10 py-8">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 space-y-6">
            <div className="h-6 w-32 bg-muted-foreground/10 animate-pulse rounded-lg" />
            <div className="h-32 bg-card border border-border/50 rounded-lg animate-pulse" />
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
              <div className="lg:col-span-2 space-y-8">
                <div className="h-48 bg-card border border-border/50 rounded-lg animate-pulse" />
                <div className="h-64 bg-card border border-border/50 rounded-lg animate-pulse" />
              </div>
              <div className="h-64 bg-card border border-border/50 rounded-lg animate-pulse" />
            </div>
          </div>
        </div>
      </PublicLayout>
    );
  }

  if (isError || !supplier) {
    return (
      <PublicLayout locale={locale}>
        <div className="bg-muted/10 py-20 text-center">
          <h2 className="text-xl font-bold text-foreground">Supplier Tidak Ditemukan</h2>
          <p className="text-sm text-muted-foreground mt-2">
            Profil supplier yang Anda cari tidak aktif atau tidak terdaftar di sistem kami.
          </p>
          <Button asChild className="mt-6 cursor-pointer rounded-lg">
            <Link href={`${detailBasePath}/search`}>Kembali ke Pencarian</Link>
          </Button>
        </div>
      </PublicLayout>
    );
  }

  return (
    <PublicLayout locale={locale}>
      <div className="bg-muted/5 py-8 min-h-screen">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          {/* Back button */}
          <Link
            href={`${detailBasePath}/search`}
            className="inline-flex items-center gap-1 text-xs font-semibold text-muted-foreground hover:text-primary transition-all duration-300 mb-6 cursor-pointer"
          >
            <ChevronLeft className="h-3.5 w-3.5" />
            {tSup("backButton")}
          </Link>

          {/* Profile Header (Clean Tokopedia Shop Detail Inspired) */}
          <div className="bg-card border border-border/45 shadow-xs rounded-lg p-6 md:p-8 flex flex-col md:flex-row items-start md:items-center justify-between gap-6 transition-all duration-300 hover:shadow-md">
            <div className="flex flex-col md:flex-row items-start md:items-center gap-5 w-full md:w-auto">
              {/* Shop Logo Avatar */}
              <div className="h-18 w-18 rounded-full bg-primary text-primary-foreground font-heading font-bold text-2xl flex items-center justify-center shadow-xs shrink-0 border border-primary/10">
                {supplier.companyName.substring(0, 2).toUpperCase()}
              </div>
              <div className="space-y-2.5">
                <div className="flex flex-wrap items-center gap-2">
                  <h1 className="text-xl font-bold text-foreground font-heading tracking-tight leading-none md:text-2xl">
                    {supplier.companyName}
                  </h1>
                  {supplier.isVerified ? (
                    <Badge variant="outline" className="border-success text-success bg-success/5 font-semibold rounded-lg px-2.5 py-0.5 text-[10px]">
                      Terverifikasi
                    </Badge>
                  ) : (
                    <Badge variant="outline" className="border-border text-muted-foreground font-semibold rounded-lg px-2.5 py-0.5 text-[10px]">
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
                  <span className="w-1.5 h-1.5 rounded-full bg-border/80" />
                  <div className="flex items-center gap-1 font-semibold text-foreground">
                    <Star className="h-3.5 w-3.5 fill-warning text-warning" />
                    <span>{supplier.rating} ({supplier.reviewCount} ulasan)</span>
                  </div>
                </div>

                {/* Actions Row */}
                <div className="flex flex-wrap items-center gap-2 pt-1.5">
                  <Button
                    onClick={handleToggleFollow}
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
                    {isFollowed ? "Mengikuti" : "Follow"}
                  </Button>
                  <Button
                    onClick={handleOpenChat}
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
                      navigator.clipboard.writeText(window.location.href);
                      toast.success("Link profil supplier disalin ke clipboard!");
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
            <div className="flex flex-row md:flex-col items-center md:items-end justify-between md:justify-center border-t md:border-t-0 md:border-l border-border/50 pt-4 md:pt-0 md:pl-8 w-full md:w-auto gap-4">
              <div className="text-left md:text-right">
                <div className="flex items-center md:justify-end gap-1">
                  <Star className="h-4.5 w-4.5 fill-warning text-warning" />
                  <span className="text-lg font-black text-foreground">{supplier.rating?.toFixed(1) || "0.0"}</span>
                  <span className="text-[10px] text-muted-foreground">/ 5.0</span>
                </div>
                <p className="text-[9px] text-muted-foreground font-bold uppercase tracking-wider mt-0.5">Rating & Ulasan</p>
              </div>
              <div className="text-right">
                <p className="text-base font-black text-primary">50+ Terjual</p>
                <p className="text-[9px] text-muted-foreground font-bold uppercase tracking-wider mt-0.5">Transaksi Sukses</p>
              </div>
            </div>
          </div>

          {/* Navigation Tabs (Standard Workspace UI Component) */}
          <Tabs 
            value={activeTab} 
            onValueChange={handleTabChange} 
            className="mt-6 w-full"
          >
            <TabsList className="w-full justify-start">
              <TabsTrigger value="home" className="cursor-pointer font-bold">
                Beranda
              </TabsTrigger>
              <TabsTrigger value="products" className="cursor-pointer font-bold">
                Produk
              </TabsTrigger>
              <TabsTrigger value="certifications" className="cursor-pointer font-bold">
                Sertifikasi
              </TabsTrigger>
              <TabsTrigger value="reviews" className="cursor-pointer font-bold">
                Ulasan
              </TabsTrigger>
            </TabsList>
          </Tabs>

          {/* Render Tabs Contents */}
          <div className="mt-8">
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
                      <h2 className="text-base font-black text-foreground uppercase tracking-tight">Kemitraan Industri & Kontrak Kustom</h2>
                      <p className="text-xs text-muted-foreground leading-relaxed">
                        Kami melayani kontrak kustom jangka panjang, negosiasi MOQ khusus, serta opsi logistik terintegrasi untuk kebutuhan industri Anda.
                      </p>
                    </div>
                    <Button
                      onClick={handleOpenRfq}
                      size="sm"
                      className="shrink-0 cursor-pointer rounded-lg bg-primary hover:bg-primary/95 text-primary-foreground font-semibold hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300"
                    >
                      Kirim RFQ Kustom
                    </Button>
                  </div>

                  {/* Spotlight / Featured Products Row (No Card/No Border wrapper) */}
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <h3 className="text-sm font-bold font-heading text-foreground">Sesuai incaran kamu di toko ini</h3>
                      <button
                        type="button"
                        onClick={() => handleTabChange("products")}
                        className="text-xs font-bold text-primary hover:underline cursor-pointer transition-colors"
                      >
                        Lihat Semua
                      </button>
                    </div>
                    {!supplier.products || supplier.products.length === 0 ? (
                      <div className="text-center py-8 bg-muted/20 border border-dashed border-border/40 rounded-lg">
                        <Package className="mx-auto h-8 w-8 text-muted-foreground/35" />
                        <h3 className="mt-3 text-xs font-bold text-foreground">Tidak ada produk</h3>
                      </div>
                    ) : (
                      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
                        {supplier.products.slice(0, 3).map((product) => {
                          const publicProd = mapSupplierProductToPublicProduct(product);
                          return (
                            <PublicSupplierProductCard
                              key={product.id}
                              product={publicProd}
                              detailBasePath={detailBasePath}
                              isBookmarked={!!getProductBookmarkId(product.id)}
                              isCompared={comparedProducts.some((p) => p.id === product.id)}
                              onBookmark={() => handleToggleProductBookmark(product.id)}
                              onCompare={() => handleToggleProductCompare(product.id)}
                            />
                          );
                        })}
                      </div>
                    )}
                  </div>
                </div>

                {/* Sidebar Info (right) */}
                <div className="space-y-6">
                  {/* Business Specs */}
                  <Card className={sectionCardClass}>
                    <div className="border-b border-border/30 px-4 py-3 bg-transparent">
                      <h3 className="text-xs font-bold font-heading uppercase tracking-wider text-muted-foreground/90">{tSup("businessInfo")}</h3>
                    </div>
                    <CardContent className="p-0">
                      <div className="divide-y divide-border/40 text-xs">
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
                            {supplier.employeeCount || tSup("na")} Orang
                          </span>
                        </div>
                      </div>
                    </CardContent>
                  </Card>

                  {/* Contact Info */}
                  <Card className={sectionCardClass}>
                    <div className="border-b border-border/30 px-4 py-3 bg-transparent">
                      <h3 className="text-xs font-bold font-heading uppercase tracking-wider text-muted-foreground/90">{tSup("contactInfo")}</h3>
                    </div>
                    <CardContent className="space-y-3.5 p-4 text-xs text-muted-foreground">
                      <div className="flex items-center gap-3">
                        <Mail className="h-3.5 w-3.5 text-primary/70 shrink-0" />
                        <span className="font-medium text-foreground">{supplier.email || tSup("na")}</span>
                      </div>
                      <div className="flex items-center gap-3">
                        <Phone className="h-3.5 w-3.5 text-primary/70 shrink-0" />
                        <span className="font-medium text-foreground">{supplier.phone || tSup("na")}</span>
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
                    <div className="border-b border-border/30 px-4 py-3 bg-transparent">
                      <h3 className="text-xs font-bold font-heading uppercase tracking-wider text-muted-foreground/90">{tSup("requestQuote")}</h3>
                    </div>
                    <CardContent className="p-4 space-y-4">
                      <p className="text-xs text-muted-foreground leading-relaxed">
                        Butuh penawaran harga khusus, negosiasi kuantitas besar, atau penyesuaian produk? Kirim permintaan penawaran (RFQ) langsung ke supplier ini.
                      </p>
                      <Button
                        onClick={handleOpenRfq}
                        className="w-full text-xs font-bold cursor-pointer rounded-lg bg-primary hover:bg-primary/95 text-primary-foreground transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 shadow-xs"
                      >
                        Minta Penawaran
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
                  <h2 className="text-lg font-bold text-foreground">Semua Produk</h2>
                  <p className="text-xs text-muted-foreground">Katalog lengkap produk manufaktur dan suplai.</p>
                </div>

                {isProductsLoading && productsList.length === 0 ? (
                  <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4 animate-pulse">
                    {Array.from({ length: 12 }).map((_, i) => (
                      <div key={i} className="flex flex-col justify-between overflow-hidden rounded-lg border border-border/40 bg-card p-4 h-[320px]">
                        <div className="aspect-square w-full bg-muted-foreground/10 rounded-md" />
                        <div className="h-4 w-3/4 bg-muted-foreground/10 rounded-md mt-3" />
                        <div className="h-4 w-1/2 bg-muted-foreground/10 rounded-md mt-1.5" />
                        <div className="h-8 w-full bg-muted-foreground/10 rounded-md mt-4" />
                      </div>
                    ))}
                  </div>
                ) : productsList.length === 0 ? (
                  <div className="text-center py-16 bg-card border border-border/40 rounded-lg shadow-xs">
                    <Package className="mx-auto h-12 w-12 text-muted-foreground/35 stroke-[1.5]" />
                    <h3 className="mt-4 text-sm font-bold text-foreground">Tidak ada produk</h3>
                    <p className="text-xs text-muted-foreground mt-1">Supplier belum mengunggah produk untuk saat ini.</p>
                  </div>
                ) : (
                  <>
                    <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4 animate-fade-in">
                      {productsList.map((product) => (
                        <PublicSupplierProductCard
                          key={product.id}
                          product={product}
                          detailBasePath={detailBasePath}
                          isBookmarked={!!getProductBookmarkId(product.id)}
                          isCompared={comparedProducts.some((p) => p.id === product.id)}
                          onBookmark={() => handleToggleProductBookmark(product.id)}
                          onCompare={() => handleToggleProductCompare(product.id)}
                        />
                      ))}
                    </div>
                    <StageScrollLoader
                      onLoadMore={handleLoadMoreProducts}
                      hasMore={hasMoreProducts}
                      isLoading={isProductsLoadingMore}
                    />
                  </>
                )}
              </div>
            )}

            {/* CERTIFICATIONS TAB */}
            {activeTab === "certifications" && (
              <div className="space-y-6">
                <div className="flex flex-col gap-1.5">
                  <h2 className="text-lg font-bold text-foreground">Sertifikasi & Lisensi Perusahaan</h2>
                  <p className="text-xs text-muted-foreground">Daftar sertifikat resmi terverifikasi tim penilai Indosupplier.</p>
                </div>

                {!supplier.certificationList || supplier.certificationList.length === 0 ? (
                  <div className="text-center py-16 bg-card border border-dashed border-border/40 rounded-lg shadow-xs">
                    <Award className="mx-auto h-12 w-12 text-muted-foreground/35 stroke-[1.5]" />
                    <h3 className="mt-4 text-sm font-bold text-foreground">{tSup("emptyCertsTitle")}</h3>
                    <p className="mt-1 text-xs text-muted-foreground">{tSup("emptyCertsDesc")}</p>
                  </div>
                ) : (
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    {supplier.certificationList.map((cert) => (
                      <div
                        key={cert.id}
                        className="flex items-center justify-between p-4.5 border border-border/40 rounded-lg bg-card transition-all duration-300 hover:border-primary/20 hover:shadow-xs"
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
                        <Badge variant="outline" className="border-success text-success bg-success/5 font-semibold rounded-lg px-2.5 py-0.5 text-[10px]">
                          Terverifikasi
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
                  <p className="text-xs text-muted-foreground">Feedback asli dari transaksi terverifikasi.</p>
                </div>

                {/* Reviews List */}
                <div className="space-y-4">
                  {!supplier.reviews || supplier.reviews.length === 0 ? (
                    <div className="rounded-lg border border-dashed border-border py-12 text-center text-xs text-muted-foreground bg-card">
                      Belum ada ulasan pembeli untuk supplier ini.
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
                          <div className="flex items-center gap-1 rounded-lg bg-muted/40 px-2 py-1 border border-border/30 text-xs font-bold">
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
                            <p className="font-bold text-primary">Balasan dari Supplier:</p>
                            <p className="mt-1 text-foreground/80 leading-relaxed font-medium">{review.supplierReply}</p>
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
                <DialogTitle>{tSup("requestQuote")}</DialogTitle>
                <DialogDescription>
                  Lengkapi detail di bawah ini untuk mengirim permintaan penawaran kepada {supplier.companyName}.
                </DialogDescription>
              </DialogHeader>
              <form onSubmit={handleSendRFQ} className="space-y-4 mt-2">
                <FieldGroup className="space-y-3.5">
                  <Field className="space-y-1.5">
                    <FieldLabel className="text-xs font-bold text-foreground">{tSup("quoteSubject")}</FieldLabel>
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
                    <FieldLabel className="text-xs font-bold text-foreground">{tSup("quoteQuantity")}</FieldLabel>
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
                    <FieldLabel className="text-xs font-bold text-foreground">{tSup("quoteMessage")}</FieldLabel>
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
                    Batal
                  </Button>
                  <Button
                    type="submit"
                    disabled={isSubmitting}
                    className="bg-primary text-primary-foreground hover:bg-primary/95 font-bold cursor-pointer rounded-lg text-xs h-9 hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 hover:shadow-md hover:shadow-primary/20"
                  >
                    {isSubmitting ? "Sending..." : tSup("btnSendQuote")}
                  </Button>
                </div>
              </form>
            </DialogContent>
          </Dialog>
        </div>
      </div>
    </PublicLayout>
  );
}
