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
import { Field, FieldLabel } from "@/components/ui/field";
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
}

export function PublicSupplierProfilePage({ locale, slug }: PublicSupplierProfilePageProps) {
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
            <Link href="/search">Kembali ke Pencarian</Link>
          </Button>
        </div>
      </PublicLayout>
    );
  }

  return (
    <PublicLayout locale={locale}>
      <div className="bg-muted py-8">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          {/* Back button */}
          <Link
            href="/search"
            className="inline-flex items-center gap-1.5 text-sm font-semibold text-muted-foreground hover:text-foreground transition-colors mb-6 cursor-pointer"
          >
            <ChevronLeft className="h-4 w-4" />
            {tSup("backButton")}
          </Link>

          {/* Profile Header */}
          <div className="bg-card border border-border shadow-xs rounded-lg overflow-hidden p-6 md:p-8 flex flex-col md:flex-row items-start md:items-center justify-between gap-6">
            <div className="flex items-center gap-5">
              <div className="h-16 w-16 rounded-lg bg-primary text-primary-foreground font-heading font-bold text-2xl flex items-center justify-center">
                {supplier.companyName.substring(0, 2).toUpperCase()}
              </div>
              <div className="space-y-1.5">
                <div className="flex flex-wrap items-center gap-2">
                  <h1 className="text-2xl font-bold text-foreground font-heading tracking-tight leading-none">
                    {supplier.companyName}
                  </h1>
                  {supplier.isVerified ? (
                    <Badge variant="outline" className="border-success text-success bg-success/5 font-semibold rounded-lg px-2.5">
                      Terverifikasi
                    </Badge>
                  ) : (
                    <Badge variant="outline" className="border-border text-muted-foreground font-semibold rounded-lg px-2.5">
                      {tSup("notVerified")}
                    </Badge>
                  )}
                </div>
                <div className="flex items-center gap-3 text-xs text-muted-foreground font-medium">
                  <div className="flex items-center gap-1">
                    <MapPin className="h-4 w-4" />
                    <span>{supplier.location}</span>
                  </div>
                  <span className="w-1.5 h-1.5 rounded-full bg-border" />
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
              className={`cursor-pointer rounded-lg border-border font-semibold transition-all hover:-translate-y-0.5 active:translate-y-0 ${
                isSupplierBookmarked
                  ? "bg-rose-50 text-rose-600 border-rose-200 hover:bg-rose-100"
                  : "hover:bg-muted"
              }`}
            >
              <Heart className={`mr-2 h-4 w-4 ${isSupplierBookmarked ? "fill-rose-600 text-rose-600" : ""}`} />
              {isSupplierBookmarked ? "Tersimpan" : "Simpan Supplier"}
            </Button>
          </div>

          {/* Content Layout */}
          <div className="mt-8 grid grid-cols-1 lg:grid-cols-3 gap-8">
            {/* Left/Main Column: Overview, Catalog, Certs */}
            <div className="lg:col-span-2 space-y-8">
              {/* Company Overview */}
              <Card className="border border-border shadow-xs rounded-lg overflow-hidden bg-card">
                <CardHeader>
                  <CardTitle className="text-base font-bold font-heading">{tSup("overview")}</CardTitle>
                </CardHeader>
                <CardContent className="text-sm text-muted-foreground leading-relaxed space-y-4">
                  <p>{supplier.description || tSup("emptyOverview")}</p>
                </CardContent>
              </Card>

              {/* Products Catalog */}
              <Card className="border border-border shadow-xs rounded-lg overflow-hidden bg-card">
                <CardHeader>
                  <CardTitle className="text-base font-bold font-heading">{tSup("products")}</CardTitle>
                </CardHeader>
                <CardContent>
                  {!supplier.products || supplier.products.length === 0 ? (
                    <div className="text-center py-10 bg-muted border border-dashed border-border rounded-lg">
                      <Package className="mx-auto h-10 w-10 text-muted-foreground" />
                      <h3 className="mt-4 text-sm font-semibold text-foreground">{tSup("emptyProductsTitle")}</h3>
                      <p className="mt-1 text-xs text-muted-foreground">{tSup("emptyProductsDesc")}</p>
                    </div>
                  ) : (
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                      {supplier.products.map((product) => {
                        const isProductBookmarked = !!getProductBookmarkId(product.id);
                        return (
                          <div
                            key={product.id}
                            id={`product-${product.id}`}
                            className="border border-border rounded-lg overflow-hidden bg-card flex flex-col justify-between hover:shadow-md transition-shadow relative scroll-mt-20"
                          >
                            <div>
                              {/* Product Image */}
                              <div className="aspect-video bg-muted border-b border-border flex items-center justify-center relative overflow-hidden">
                                {product.photos && product.photos.length > 0 ? (
                                  <img
                                    src={product.photos[0]}
                                    alt={product.name}
                                    className="object-cover w-full h-full"
                                  />
                                ) : (
                                  <Package className="h-8 w-8 text-muted-foreground opacity-30" />
                                )}

                                {/* Favorite Product Toggle */}
                                <Button
                                  onClick={() => handleToggleProductBookmark(product.id)}
                                  variant="ghost"
                                  size="icon"
                                  className="absolute right-2.5 top-2.5 h-8 w-8 text-foreground/75 bg-card/80 backdrop-blur-xs hover:text-rose-600 hover:bg-rose-50 rounded-lg cursor-pointer transition-colors shadow-xs"
                                >
                                  <Heart
                                    className={`h-4 w-4 ${
                                      isProductBookmarked ? "fill-rose-600 text-rose-600" : ""
                                    }`}
                                  />
                                </Button>
                              </div>

                              <div className="p-4 space-y-2">
                                <h3 className="font-semibold text-foreground text-sm leading-tight line-clamp-2">
                                  {product.name}
                                </h3>
                                <p className="text-xs text-muted-foreground line-clamp-2">
                                  {product.description || "Tidak ada deskripsi tambahan."}
                                </p>
                              </div>
                            </div>

                            <div className="p-4 border-t border-border bg-muted/5 flex items-center justify-between gap-4">
                              <div className="space-y-0.5">
                                <p className="text-[10px] text-muted-foreground uppercase tracking-wider">Mulai Dari</p>
                                <p className="text-sm font-bold text-primary">
                                  {product.price ? formatPrice(product.price) : "Hubungi Kami"}
                                </p>
                              </div>
                              <Button
                                asChild
                                size="sm"
                                className="text-xs bg-primary text-primary-foreground hover:bg-primary/95 font-semibold cursor-pointer rounded-lg transition-transform hover:-translate-y-0.5"
                              >
                                <Link href="#contact">Minta Harga</Link>
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
              <Card className="border border-border shadow-xs rounded-lg overflow-hidden bg-card">
                <CardHeader>
                  <CardTitle className="text-base font-bold font-heading">{tSup("certifications")}</CardTitle>
                </CardHeader>
                <CardContent>
                  {!supplier.certificationList || supplier.certificationList.length === 0 ? (
                    <div className="text-center py-10 bg-muted border border-dashed border-border rounded-lg">
                      <Award className="mx-auto h-10 w-10 text-muted-foreground" />
                      <h3 className="mt-4 text-sm font-semibold text-foreground">{tSup("emptyCertsTitle")}</h3>
                      <p className="mt-1 text-xs text-muted-foreground">{tSup("emptyCertsDesc")}</p>
                    </div>
                  ) : (
                    <div className="space-y-4">
                      {supplier.certificationList.map((cert) => (
                        <div
                          key={cert.id}
                          className="flex items-center justify-between p-4 border border-border rounded-lg bg-muted/5"
                        >
                          <div className="flex items-center gap-3">
                            <div className="h-10 w-10 bg-primary/10 text-primary flex items-center justify-center rounded-lg border border-primary/20 shrink-0">
                              <Award className="h-5 w-5" />
                            </div>
                            <div>
                              <p className="text-sm font-semibold text-foreground">{cert.name}</p>
                              <p className="text-xs text-muted-foreground">
                                Diterbitkan oleh: {cert.institution} {cert.year ? `(${cert.year})` : ""}
                              </p>
                            </div>
                          </div>
                          <Badge variant="outline" className="border-success text-success bg-success/5 font-semibold rounded-lg">
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
              <Card className="border border-border shadow-xs rounded-lg overflow-hidden bg-card">
                <CardHeader>
                  <CardTitle className="text-base font-bold font-heading">{tSup("businessInfo")}</CardTitle>
                </CardHeader>
                <CardContent className="p-0">
                  <div className="divide-y divide-border text-sm">
                    <div className="flex justify-between px-6 py-4">
                      <span className="text-muted-foreground font-medium flex items-center gap-2">
                        <Building className="h-4 w-4 text-muted-foreground" />
                        {tSup("businessType")}
                      </span>
                      <span className="font-semibold text-foreground">
                        {supplier.businessType || tSup("na")}
                      </span>
                    </div>
                    <div className="flex justify-between px-6 py-4">
                      <span className="text-muted-foreground font-medium flex items-center gap-2">
                        <Calendar className="h-4 w-4 text-muted-foreground" />
                        {tSup("established")}
                      </span>
                      <span className="font-semibold text-foreground">
                        {supplier.establishedYear || tSup("na")}
                      </span>
                    </div>
                    <div className="flex justify-between px-6 py-4">
                      <span className="text-muted-foreground font-medium flex items-center gap-2">
                        <Users className="h-4 w-4 text-muted-foreground" />
                        {tSup("employees")}
                      </span>
                      <span className="font-semibold text-foreground">
                        {supplier.employeeCount || tSup("na")} Orang
                      </span>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* Contact Information */}
              <Card className="border border-border shadow-xs rounded-lg overflow-hidden bg-card">
                <CardHeader>
                  <CardTitle className="text-base font-bold font-heading">{tSup("contactInfo")}</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4 text-sm text-muted-foreground">
                  <div className="flex items-center gap-3">
                    <Mail className="h-4 w-4 text-muted-foreground" />
                    <span>{supplier.email || tSup("na")}</span>
                  </div>
                  <div className="flex items-center gap-3">
                    <Phone className="h-4 w-4 text-muted-foreground" />
                    <span>{supplier.phone || tSup("na")}</span>
                  </div>
                  <div className="flex items-center gap-3">
                    <Globe className="h-4 w-4 text-muted-foreground" />
                    {supplier.website ? (
                      <a
                        href={supplier.website}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-primary hover:underline"
                      >
                        {supplier.website}
                      </a>
                    ) : (
                      <span>{tSup("na")}</span>
                    )}
                  </div>
                </CardContent>
              </Card>

              {/* Send RFQ Form */}
              <Card id="contact" className="border border-border shadow-xs rounded-lg overflow-hidden scroll-mt-20 bg-card">
                <CardHeader>
                  <CardTitle className="text-base font-bold font-heading">{tSup("requestQuote")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <form onSubmit={handleSendRFQ} className="space-y-4">
                    <Field>
                      <FieldLabel>{tSup("quoteSubject")}</FieldLabel>
                      <Input
                        type="text"
                        placeholder="e.g. Bulk Procurement for Teak Wood chairs"
                        value={subject}
                        onChange={(e) => setSubject(e.target.value)}
                        required
                        className="bg-card border-border focus-visible:border-muted-foreground"
                      />
                    </Field>

                    <Field>
                      <FieldLabel>{tSup("quoteQuantity")}</FieldLabel>
                      <Input
                        type="number"
                        min="1"
                        value={quantity}
                        onChange={(e) => setQuantity(e.target.value)}
                        required
                        className="bg-card border-border focus-visible:border-muted-foreground"
                      />
                    </Field>

                    <Field>
                      <FieldLabel>{tSup("quoteMessage")}</FieldLabel>
                      <Textarea
                        rows={4}
                        placeholder={tSup("quoteMessage")}
                        value={message}
                        onChange={(e) => setMessage(e.target.value)}
                        required
                        className="bg-card border-border focus-visible:border-muted-foreground resize-none"
                      />
                    </Field>

                    <Button
                      type="submit"
                      disabled={isSubmitting}
                      className="w-full bg-primary text-primary-foreground hover:bg-primary/95 font-semibold cursor-pointer rounded-lg"
                    >
                      {isSubmitting ? "Sending..." : tSup("btnSendQuote")}
                    </Button>
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
