"use client";
/* eslint-disable @next/next/no-img-element */

import React, { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "@/i18n/routing";
import { PublicLayout } from "@/features/public/components/public-layout";
import { searchService } from "@/features/public/search/services/search-service";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { useBuyerBookmarks } from "@/features/buyer/bookmarks/hooks/useBuyerBookmarks";
import { useBuyerCompare } from "@/features/buyer/compare/hooks/useBuyerCompare";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import { toast } from "sonner";
import {
  ChevronLeft,
  GitCompareArrows,
  Heart,
  MapPin,
  Package,
  ShieldCheck,
  Star,
  Store,
} from "lucide-react";

interface PublicProductDetailPageProps {
  locale: string;
  id: string;
  detailBasePath?: "" | "/demo";
}

function formatPrice(price: number, currency = "IDR") {
  if (!price) return "Hubungi Supplier";
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency,
    minimumFractionDigits: 0,
  }).format(price);
}

export function PublicProductDetailPage({ locale, id, detailBasePath = "" }: PublicProductDetailPageProps) {
  const [activePhoto, setActivePhoto] = useState(0);
  const { isAuthenticated } = useAuthStore();
  const { bookmarks, addBookmark, deleteBookmark } = useBuyerBookmarks();
  const { products: comparedProducts, addProduct, removeProduct } = useBuyerCompare();

  const { data, isLoading } = useQuery({
    queryKey: ["public-product-detail", id],
    queryFn: () => searchService.getProductById(id),
  });

  const product = data?.product;
  const supplier = data?.supplier;
  const isBookmarked = product ? bookmarks.some((item) => item.type === "product" && item.supplierProductId === product.id) : false;
  const isCompared = product ? comparedProducts.some((item) => item.id === product.id) : false;

  const requireAuth = (message: string) => {
    if (!isAuthenticated) {
      toast.error(message);
      return false;
    }
    return true;
  };

  const toggleBookmark = () => {
    if (!product) return;
    if (!requireAuth("Silakan masuk terlebih dahulu untuk menyimpan produk.")) return;
    const bookmark = bookmarks.find((item) => item.type === "product" && item.supplierProductId === product.id);
    if (bookmark) {
      deleteBookmark(bookmark.id);
    } else {
      addBookmark({ supplierProfileId: product.supplierId, supplierProductId: product.id });
    }
  };

  const toggleCompare = () => {
    if (!product) return;
    if (!requireAuth("Silakan masuk terlebih dahulu untuk membandingkan produk.")) return;
    if (isCompared) {
      removeProduct(product.id);
    } else {
      addProduct(product.id);
    }
  };

  if (isLoading) {
    return (
      <PublicLayout locale={locale}>
        <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
          <div className="h-[520px] animate-pulse rounded-lg border border-border bg-muted" />
        </div>
      </PublicLayout>
    );
  }

  if (!product || !supplier) {
    return (
      <PublicLayout locale={locale}>
        <div className="mx-auto max-w-7xl px-4 py-20 text-center sm:px-6 lg:px-8">
          <Package className="mx-auto h-12 w-12 text-muted-foreground" />
          <h1 className="mt-4 text-xl font-extrabold text-foreground">Produk tidak ditemukan</h1>
          <Button asChild className="mt-6 cursor-pointer">
            <Link href={`${detailBasePath}/search`}>Kembali ke pencarian</Link>
          </Button>
        </div>
      </PublicLayout>
    );
  }

  const photos = product.photos?.length ? product.photos : [];
  const currentPhoto = photos[activePhoto];

  return (
    <PublicLayout locale={locale}>
      <main className="min-h-screen bg-background">
        <div className="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
          <Link href={`${detailBasePath}/search?query=${encodeURIComponent(product.name)}`} className="mb-5 inline-flex cursor-pointer items-center gap-1 text-sm font-semibold text-muted-foreground hover:text-foreground">
            <ChevronLeft className="h-4 w-4" />
            Kembali ke katalog
          </Link>

          <section className="grid grid-cols-1 gap-8 lg:grid-cols-[420px_1fr_280px]">
            <div className="space-y-3">
              <div className="aspect-square overflow-hidden rounded-lg border border-border bg-muted">
                {currentPhoto ? (
                  <img src={currentPhoto} alt={product.name} className="h-full w-full object-cover" />
                ) : (
                  <div className="flex h-full items-center justify-center">
                    <Package className="h-12 w-12 text-muted-foreground" />
                  </div>
                )}
              </div>
              {photos.length > 1 && (
                <div className="grid grid-cols-5 gap-2">
                  {photos.slice(0, 5).map((photo, index) => (
                    <button
                      key={photo}
                      type="button"
                      onClick={() => setActivePhoto(index)}
                      className={`aspect-square overflow-hidden rounded border cursor-pointer ${activePhoto === index ? "border-primary" : "border-border"}`}
                    >
                      <img src={photo} alt={`${product.name} ${index + 1}`} className="h-full w-full object-cover" />
                    </button>
                  ))}
                </div>
              )}
            </div>

            <div className="space-y-6">
              <div className="border-b border-border pb-5">
                <div className="mb-3 flex flex-wrap gap-2">
                  {product.categoryName && <Badge variant="secondary">{product.categoryName}</Badge>}
                  {supplier.isVerified && (
                    <Badge className="border-0 bg-success text-success-foreground">
                      <ShieldCheck className="mr-1 h-3.5 w-3.5" />
                      Supplier Terverifikasi
                    </Badge>
                  )}
                </div>
                <h1 className="text-2xl font-extrabold leading-tight text-foreground">{product.name}</h1>
                <div className="mt-3 flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
                  <Star className="h-4 w-4 fill-warning text-warning" />
                  <span className="font-semibold text-foreground">{supplier.rating?.toFixed(1) || "0.0"}</span>
                  <span>•</span>
                  <span>{supplier.reviewCount || 0} ulasan supplier</span>
                  <span>•</span>
                  <span>MOQ {product.minOrder || "Nego"}</span>
                </div>
                <p className="mt-5 text-3xl font-extrabold text-foreground">{formatPrice(product.price, product.currency)}</p>
              </div>

              <Card className="border border-border shadow-xs">
                <CardHeader>
                  <CardTitle className="text-base font-extrabold">Detail Produk</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4 text-sm text-muted-foreground">
                  <p className="leading-relaxed">{product.description || "Supplier belum menambahkan deskripsi detail untuk produk ini."}</p>
                  <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                    <InfoRow label="Minimum Order" value={product.minOrder || "Nego"} />
                    <InfoRow label="Kapasitas" value={product.capacityText || "Hubungi supplier"} />
                    <InfoRow label="Kategori" value={product.categoryName || "-"} />
                    <InfoRow label="Lokasi Supplier" value={product.supplierLocation || supplier.location || "-"} />
                  </div>
                </CardContent>
              </Card>

              <Card className="border border-border shadow-xs">
                <CardHeader>
                  <CardTitle className="text-base font-extrabold">Ulasan Pembeli</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  {(data.reviews || []).length === 0 ? (
                    <div className="rounded-lg border border-dashed border-border py-10 text-center text-sm text-muted-foreground">
                      Belum ada ulasan untuk supplier produk ini.
                    </div>
                  ) : (
                    data.reviews.map((review) => (
                      <div key={review.id} className="rounded-lg border border-border p-4">
                        <div className="flex items-center justify-between gap-3">
                          <p className="font-semibold text-foreground">{review.buyerName}</p>
                          <span className="flex items-center gap-1 text-sm font-semibold text-foreground">
                            <Star className="h-4 w-4 fill-warning text-warning" />
                            {review.rating}
                          </span>
                        </div>
                        <p className="mt-2 text-sm leading-relaxed text-muted-foreground">{review.reviewText}</p>
                        {review.supplierReply && (
                          <div className="mt-3 rounded bg-muted p-3 text-xs text-muted-foreground">
                            <span className="font-semibold text-foreground">Balasan supplier:</span> {review.supplierReply}
                          </div>
                        )}
                      </div>
                    ))
                  )}
                </CardContent>
              </Card>
            </div>

            <aside className="space-y-4">
              <Card className="border border-border shadow-xs">
                <CardContent className="space-y-4 p-5">
                  <p className="text-sm font-extrabold text-foreground">Atur jumlah dan catatan</p>
                  <div className="rounded-lg border border-border p-3 text-sm text-muted-foreground">
                    MOQ: <span className="font-semibold text-foreground">{product.minOrder || "Nego"}</span>
                  </div>
                  <Button className="w-full cursor-pointer font-extrabold">
                    Kirim RFQ
                  </Button>
                  <Button variant="outline" className="w-full cursor-pointer font-extrabold">
                    Hubungi Supplier
                  </Button>
                  <div className="grid grid-cols-2 gap-2">
                    <Button variant="outline" onClick={toggleBookmark} className="cursor-pointer text-xs">
                      <Heart className={`mr-1 h-4 w-4 ${isBookmarked ? "fill-rose-600 text-rose-600" : ""}`} />
                      Wishlist
                    </Button>
                    <Button variant="outline" onClick={toggleCompare} className="cursor-pointer text-xs">
                      <GitCompareArrows className="mr-1 h-4 w-4" />
                      Compare
                    </Button>
                  </div>
                </CardContent>
              </Card>

              <Card className="border border-border shadow-xs">
                <CardContent className="space-y-4 p-5">
                  <Link href={`${detailBasePath}/suppliers/${supplier.slug}`} className="flex cursor-pointer items-center gap-3">
                    <div className="flex h-11 w-11 items-center justify-center rounded-lg bg-primary/10 text-sm font-extrabold text-primary">
                      {supplier.companyName.slice(0, 2).toUpperCase()}
                    </div>
                    <div className="min-w-0">
                      <p className="truncate text-sm font-extrabold text-foreground">{supplier.companyName}</p>
                      <p className="flex items-center gap-1 text-xs text-muted-foreground">
                        <MapPin className="h-3.5 w-3.5" />
                        {supplier.location || "Indonesia"}
                      </p>
                    </div>
                  </Link>
                  <div className="grid grid-cols-2 gap-2 text-center text-xs">
                    <div className="rounded border border-border p-2">
                      <p className="font-extrabold text-foreground">{Math.round(supplier.responseRate || 0)}%</p>
                      <p className="text-muted-foreground">Respons</p>
                    </div>
                    <div className="rounded border border-border p-2">
                      <p className="font-extrabold text-foreground">{supplier.responseTime || "-"}</p>
                      <p className="text-muted-foreground">Waktu</p>
                    </div>
                  </div>
                  <Button asChild variant="outline" className="w-full cursor-pointer">
                    <Link href={`${detailBasePath}/suppliers/${supplier.slug}`}>
                      <Store className="mr-2 h-4 w-4" />
                      Detail Supplier
                    </Link>
                  </Button>
                </CardContent>
              </Card>
            </aside>
          </section>

          {data.relatedProducts.length > 0 && (
            <section className="mt-10 space-y-4">
              <h2 className="text-lg font-extrabold text-foreground">Produk lain dari supplier ini</h2>
              <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
                {data.relatedProducts.map((item) => (
                  <Link key={item.id} href={`${detailBasePath}/products/${item.id}`} className="cursor-pointer rounded-lg border border-border bg-card p-3 transition-all hover:-translate-y-0.5 hover:shadow-md">
                    <div className="aspect-square overflow-hidden rounded bg-muted">
                      {item.photos?.[0] ? (
                        <img src={item.photos[0]} alt={item.name} className="h-full w-full object-cover" />
                      ) : (
                        <div className="flex h-full items-center justify-center">
                          <Package className="h-8 w-8 text-muted-foreground" />
                        </div>
                      )}
                    </div>
                    <p className="mt-3 line-clamp-2 text-sm font-semibold text-foreground">{item.name}</p>
                    <p className="mt-1 text-sm font-extrabold text-primary">{formatPrice(item.price, item.currency)}</p>
                  </Link>
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
    <div className="rounded-lg border border-border p-3">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="mt-1 font-semibold text-foreground">{value}</p>
    </div>
  );
}
