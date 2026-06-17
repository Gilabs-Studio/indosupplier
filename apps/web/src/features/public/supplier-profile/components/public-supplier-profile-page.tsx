"use client";
/* eslint-disable @next/next/no-img-element */

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { useQuery } from "@tanstack/react-query";
import { Link } from "@/i18n/routing";
import { PublicLayout } from "@/features/public/components/public-layout";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Field, FieldLabel, FieldGroup, FieldDescription, FieldError } from "@/components/ui/field";
import { toast } from "sonner";
import { searchService } from "@/features/public/search/services/search-service";
import { useBuyerBookmarks } from "@/features/buyer/bookmarks/hooks/useBuyerBookmarks";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
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
  Heart,
  Star,
} from "lucide-react";

interface PublicSupplierProfilePageProps {
  locale: string;
  slug: string;
  detailBasePath?: "" | "/demo";
}

export function PublicSupplierProfilePage({ locale, slug, detailBasePath = "/demo" }: PublicSupplierProfilePageProps) {
  const tSup = useTranslations("public.supplier");
  const { isAuthenticated } = useAuthStore();
  const { bookmarks, addBookmark, deleteBookmark } = useBuyerBookmarks();

  // Fetch Supplier profile by slug
  const { data: supplier, isLoading, isError } = useQuery({
    queryKey: ["public-supplier-profile", slug],
    queryFn: () => searchService.getSupplierBySlug(slug),
  });

  // Mock RFQ form state
  const [subject, setSubject] = useState("");
  const [message, setMessage] = useState("");
  const [quantity, setQuantity] = useState("1");
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Check if supplier is bookmarked
  const isSupplierBookmarked = bookmarks.some(
    (b) => b.type === "supplier" && b.supplierProfileId === supplier?.id
  );

  const getProductBookmarkId = (prodId: string) => {
    const found = bookmarks.find(
      (b) => b.type === "product" && b.supplierProductId === prodId
    );
    return found?.id || null;
  };

  const handleToggleSupplierBookmark = () => {
    if (!isAuthenticated) {
      toast.error("Silakan masuk terlebih dahulu untuk menyimpan supplier.");
      return;
    }
    if (isSupplierBookmarked) {
      const found = bookmarks.find(
        (b) => b.type === "supplier" && b.supplierProfileId === supplier?.id
      );
      if (found) deleteBookmark(found.id);
    } else {
      if (supplier) {
        addBookmark({ supplierProfileId: supplier.id });
      }
    }
  };

  const handleToggleProductBookmark = (prodId: string) => {
    if (!isAuthenticated) {
      toast.error("Silakan masuk terlebih dahulu untuk menyimpan produk.");
      return;
    }
    const bookmarkId = getProductBookmarkId(prodId);
    if (bookmarkId) {
      deleteBookmark(bookmarkId);
    } else {
      if (supplier) {
        addBookmark({ supplierProfileId: supplier.id, supplierProductId: prodId });
      }
    }
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
    }, 1000);
  };

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
    }).format(price);
  };

  if (isLoading) {
    return (
      <PublicLayout locale={locale}>
        <div className="bg-muted py-8">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 space-y-6">
            <div className="h-6 w-32 bg-muted-foreground/10 animate-pulse rounded-lg" />
            <div className="h-32 bg-card border border-border rounded-lg animate-pulse" />
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
              <div className="lg:col-span-2 space-y-8">
                <div className="h-48 bg-card border border-border rounded-lg animate-pulse" />
                <div className="h-64 bg-card border border-border rounded-lg animate-pulse" />
              </div>
              <div className="h-64 bg-card border border-border rounded-lg animate-pulse" />
            </div>
          </div>
        </div>
      </PublicLayout>
    );
  }

  if (isError || !supplier) {
    return (
      <PublicLayout locale={locale}>
        <div className="bg-muted py-20 text-center">
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
      <div className="bg-muted/30 py-8 min-h-screen">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          {/* Back button */}
          <Link
            href={`${detailBasePath}/search`}
            className="inline-flex items-center gap-1 text-xs font-semibold text-muted-foreground hover:text-primary transition-all duration-300 mb-6 cursor-pointer"
          >
            <ChevronLeft className="h-3.5 w-3.5" />
            {tSup("backButton")}
          </Link>

          {/* Profile Header */}
          <div className="bg-card border border-border/80 shadow-xs rounded-lg overflow-hidden p-6 md:p-8 flex flex-col md:flex-row items-start md:items-center justify-between gap-6 transition-all duration-300 hover:shadow-md">
            <div className="flex items-center gap-5">
              <div className="h-14 w-14 rounded-lg bg-primary text-primary-foreground font-heading font-bold text-xl flex items-center justify-center shadow-sm shrink-0">
                {supplier.companyName.substring(0, 2).toUpperCase()}
              </div>
              <div className="space-y-1">
                <div className="flex flex-wrap items-center gap-2">
                  <h1 className="text-xl font-bold text-foreground font-heading tracking-tight leading-none md:text-2xl">
                    {supplier.companyName}
                  </h1>
                  {supplier.isVerified ? (
                    <Badge variant="outline" className="border-success text-success bg-success/5 font-semibold rounded-lg px-2 py-0.5 text-[10px]">
                      Terverifikasi
                    </Badge>
                  ) : (
                    <Badge variant="outline" className="border-border text-muted-foreground font-semibold rounded-lg px-2 py-0.5 text-[10px]">
                      {tSup("notVerified")}
                    </Badge>
                  )}
                </div>
                <div className="flex items-center gap-3 text-xs text-muted-foreground font-medium">
                  <div className="flex items-center gap-1">
                    <MapPin className="h-3.5 w-3.5 text-primary/70" />
                    <span>{supplier.location}</span>
                  </div>
                  <span className="w-1 h-1 rounded-full bg-border" />
                  <div className="flex items-center gap-1 font-semibold text-foreground">
                    <Star className="h-3.5 w-3.5 fill-warning text-warning" />
                    <span>{supplier.rating} ({supplier.reviewCount} ulasan)</span>
                  </div>
                </div>
              </div>
            </div>

            {/* Favorite Supplier Button */}
            <Button
              onClick={handleToggleSupplierBookmark}
              variant="outline"
              className={`cursor-pointer rounded-lg font-bold text-xs py-2 px-4 border-border transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 ${
                isSupplierBookmarked
                  ? "bg-rose-50 text-rose-600 border-rose-200 hover:bg-rose-100"
                  : "hover:bg-muted"
              }`}
            >
              <Heart className={`mr-1.5 h-3.5 w-3.5 ${isSupplierBookmarked ? "fill-rose-600 text-rose-600" : ""}`} />
              {isSupplierBookmarked ? "Tersimpan" : "Simpan Supplier"}
            </Button>
          </div>

          {/* Content Layout */}
          <div className="mt-8 grid grid-cols-1 lg:grid-cols-3 gap-8">
            {/* Left/Main Column: Overview, Catalog, Certs */}
            <div className="lg:col-span-2 space-y-8">
              {/* Company Overview */}
              <Card className="border border-border/80 shadow-xs rounded-lg overflow-hidden bg-card transition-all duration-300 hover:shadow-xs">
                <CardHeader className="border-b border-border/60 py-3.5 px-5">
                  <CardTitle className="text-sm font-bold font-heading text-foreground">{tSup("overview")}</CardTitle>
                </CardHeader>
                <CardContent className="p-5 text-xs text-muted-foreground leading-relaxed">
                  <p className="whitespace-pre-line text-sm leading-relaxed">{supplier.description || tSup("emptyOverview")}</p>
                </CardContent>
              </Card>

              {/* Products Catalog */}
              <Card className="border border-border/80 shadow-xs rounded-lg overflow-hidden bg-card transition-all duration-300 hover:shadow-xs">
                <CardHeader className="border-b border-border/60 py-3.5 px-5">
                  <CardTitle className="text-sm font-bold font-heading text-foreground">{tSup("products")}</CardTitle>
                </CardHeader>
                <CardContent className="p-5">
                  {!supplier.products || supplier.products.length === 0 ? (
                    <div className="text-center py-10 bg-muted/40 border border-dashed border-border rounded-lg">
                      <Package className="mx-auto h-8 w-8 text-muted-foreground/40" />
                      <h3 className="mt-3 text-xs font-bold text-foreground">{tSup("emptyProductsTitle")}</h3>
                      <p className="mt-1 text-[10px] text-muted-foreground">{tSup("emptyProductsDesc")}</p>
                    </div>
                  ) : (
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                      {supplier.products.map((product) => {
                        const isProductBookmarked = !!getProductBookmarkId(product.id);
                        return (
                          <div
                            key={product.id}
                            id={`product-${product.id}`}
                            className="group border border-border/80 rounded-lg overflow-hidden bg-card flex flex-col justify-between hover:shadow-md hover:border-primary/20 transition-all duration-300 relative scroll-mt-20"
                          >
                            <div>
                              {/* Product Image */}
                              <div className="aspect-video bg-muted/30 border-b border-border/60 flex items-center justify-center relative overflow-hidden">
                                {product.photos && product.photos.length > 0 ? (
                                  <img
                                    src={product.photos[0]}
                                    alt={product.name}
                                    className="object-cover w-full h-full transition-transform duration-500 group-hover:scale-103"
                                  />
                                ) : (
                                  <Package className="h-6 w-6 text-muted-foreground opacity-30" />
                                )}

                                {/* Favorite Product Toggle */}
                                <Button
                                  onClick={() => handleToggleProductBookmark(product.id)}
                                  variant="ghost"
                                  size="icon"
                                  className="absolute right-2 top-2 h-7 w-7 text-foreground/75 bg-card/85 backdrop-blur-xs hover:text-rose-600 hover:bg-rose-50 rounded-lg cursor-pointer transition-colors shadow-xs"
                                >
                                  <Heart
                                    className={`h-3.5 w-3.5 ${
                                      isProductBookmarked ? "fill-rose-600 text-rose-600" : ""
                                    }`}
                                  />
                                </Button>
                              </div>

                              <div className="p-3.5 space-y-1.5">
                                <h3 className="font-bold text-foreground text-xs leading-snug line-clamp-2 transition-colors group-hover:text-primary">
                                  {product.name}
                                </h3>
                                <p className="text-[10px] text-muted-foreground line-clamp-2 leading-relaxed">
                                  {product.description || "Tidak ada deskripsi tambahan."}
                                </p>
                              </div>
                            </div>

                            <div className="p-3.5 border-t border-border/60 bg-muted/5 flex items-center justify-between gap-4">
                              <div className="space-y-0.5">
                                <p className="text-[9px] font-bold text-muted-foreground uppercase tracking-wider">Mulai Dari</p>
                                <p className="text-xs font-black text-primary">
                                  {product.price ? formatPrice(product.price) : "Hubungi Kami"}
                                </p>
                              </div>
                              <Button
                                asChild
                                size="sm"
                                className="text-[10px] h-7 px-3 bg-primary text-primary-foreground hover:bg-primary/95 font-bold cursor-pointer rounded-lg transition-transform hover:-translate-y-0.5"
                              >
                                <Link href={`${detailBasePath}/products/${product.id}`}>Detail</Link>
                              </Button>
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  )}
                </CardContent>
              </Card>

              {/* Certifications */}
              <Card className="border border-border/80 shadow-xs rounded-lg overflow-hidden bg-card transition-all duration-300 hover:shadow-xs">
                <CardHeader className="border-b border-border/60 py-3.5 px-5">
                  <CardTitle className="text-sm font-bold font-heading text-foreground">{tSup("certifications")}</CardTitle>
                </CardHeader>
                <CardContent className="p-5">
                  {!supplier.certificationList || supplier.certificationList.length === 0 ? (
                    <div className="text-center py-10 bg-muted/40 border border-dashed border-border rounded-lg">
                      <Award className="mx-auto h-8 w-8 text-muted-foreground/40" />
                      <h3 className="mt-3 text-xs font-bold text-foreground">{tSup("emptyCertsTitle")}</h3>
                      <p className="mt-1 text-[10px] text-muted-foreground">{tSup("emptyCertsDesc")}</p>
                    </div>
                  ) : (
                    <div className="space-y-3">
                      {supplier.certificationList.map((cert) => (
                        <div
                          key={cert.id}
                          className="flex items-center justify-between p-3.5 border border-border/60 rounded-lg bg-muted/5 transition-colors hover:border-primary/20"
                        >
                          <div className="flex items-center gap-3">
                            <div className="h-8 w-8 bg-primary/10 text-primary flex items-center justify-center rounded-lg border border-primary/20 shrink-0">
                              <Award className="h-4 w-4" />
                            </div>
                            <div>
                              <p className="text-xs font-bold text-foreground">{cert.name}</p>
                              <p className="text-[10px] text-muted-foreground mt-0.5">
                                Diterbitkan oleh: {cert.institution} {cert.year ? `(${cert.year})` : ""}
                              </p>
                            </div>
                          </div>
                          <Badge variant="outline" className="border-success text-success bg-success/5 font-semibold rounded-lg px-2 py-0.5 text-[9px]">
                            Terverifikasi
                          </Badge>
                        </div>
                      ))}
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>

            {/* Right Column: Business Details & Contact RFQ */}
            <div className="space-y-8">
              {/* Business Overview Stats */}
              <Card className="border border-border/80 shadow-xs rounded-lg overflow-hidden bg-card transition-all duration-300 hover:shadow-xs">
                <CardHeader className="border-b border-border/60 py-3.5 px-5">
                  <CardTitle className="text-sm font-bold font-heading text-foreground">{tSup("businessInfo")}</CardTitle>
                </CardHeader>
                <CardContent className="p-0">
                  <div className="divide-y divide-border/60 text-xs">
                    <div className="flex justify-between px-5 py-3.5">
                      <span className="text-muted-foreground font-medium flex items-center gap-2">
                        <Building className="h-3.5 w-3.5 text-primary/70 shrink-0" />
                        {tSup("businessType")}
                      </span>
                      <span className="font-bold text-foreground">
                        {supplier.businessType || tSup("na")}
                      </span>
                    </div>
                    <div className="flex justify-between px-5 py-3.5">
                      <span className="text-muted-foreground font-medium flex items-center gap-2">
                        <Calendar className="h-3.5 w-3.5 text-primary/70 shrink-0" />
                        {tSup("established")}
                      </span>
                      <span className="font-bold text-foreground">
                        {supplier.establishedYear || tSup("na")}
                      </span>
                    </div>
                    <div className="flex justify-between px-5 py-3.5">
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

              {/* Contact Information */}
              <Card className="border border-border/80 shadow-xs rounded-lg overflow-hidden bg-card transition-all duration-300 hover:shadow-xs">
                <CardHeader className="border-b border-border/60 py-3.5 px-5">
                  <CardTitle className="text-sm font-bold font-heading text-foreground">{tSup("contactInfo")}</CardTitle>
                </CardHeader>
                <CardContent className="space-y-3.5 p-5 text-xs text-muted-foreground">
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

              {/* Send RFQ Form */}
              <Card id="contact" className="border border-border/80 shadow-xs rounded-lg overflow-hidden scroll-mt-20 bg-card transition-all duration-300 hover:shadow-xs">
                <CardHeader className="border-b border-border/60 py-3.5 px-5">
                  <CardTitle className="text-sm font-bold font-heading text-foreground">{tSup("requestQuote")}</CardTitle>
                </CardHeader>
                <CardContent className="p-5">
                  <form onSubmit={handleSendRFQ}>
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

                      <Button
                        type="submit"
                        disabled={isSubmitting}
                        className="w-full bg-primary text-primary-foreground hover:bg-primary/95 font-bold cursor-pointer rounded-lg text-xs h-9 hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 hover:shadow-md hover:shadow-primary/20"
                      >
                        {isSubmitting ? "Sending..." : tSup("btnSendQuote")}
                      </Button>
                    </FieldGroup>
                  </form>
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      </div>
    </PublicLayout>
  );
}
