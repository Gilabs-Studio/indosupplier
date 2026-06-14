"use client";
/* eslint-disable @next/next/no-img-element */

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { BuyerLayout } from "../../components/buyer-layout";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Link } from "@/i18n/routing";
import { DeleteDialog } from "@/components/ui/delete-dialog";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import {
  MapPin,
  Star,
  Trash2,
  ExternalLink,
  MessageSquare,
  Columns3,
  Heart,
  Store,
  Package,
} from "lucide-react";
import { useBuyerBookmarks } from "../hooks/useBuyerBookmarks";

export function BuyerBookmarksPage() {
  const t = useTranslations("buyer.bookmarks");
  const { bookmarks, isLoading, deleteBookmark } = useBuyerBookmarks();

  const [compareList, setCompareList] = useState<string[]>([]);
  const [deleteId, setDeleteId] = useState<string | null>(null);

  const handleToggleCompare = (id: string) => {
    if (compareList.includes(id)) {
      setCompareList(compareList.filter((item) => item !== id));
    } else {
      if (compareList.length >= 3) {
        alert("Maksimal bandingkan 3 supplier sekaligus!");
        return;
      }
      setCompareList([...compareList, id]);
    }
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
      <BuyerLayout>
        <div className="flex items-center justify-center py-20">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
        </div>
      </BuyerLayout>
    );
  }

  const supplierBookmarks = bookmarks.filter((b) => b.type === "supplier");
  const productBookmarks = bookmarks.filter((b) => b.type === "product");

  return (
    <BuyerLayout>
      <div className="space-y-8">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-6 pb-6 border-b border-border">
          <div className="space-y-1">
            <h1 className="text-3xl font-bold tracking-tight text-foreground font-heading">
              {t("title")}
            </h1>
            <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
          </div>
          {compareList.length > 0 && (
            <Button
              asChild
              className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/95 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/30 rounded-lg px-4 py-2"
            >
              <Link href="/compare">
                <Columns3 className="mr-2 h-4 w-4" />
                {t("compareCount", { count: compareList.length })}
              </Link>
            </Button>
          )}
        </div>

        {/* Tabs for Suppliers vs Products */}
        <Tabs defaultValue="suppliers" className="w-full">
          <TabsList className="mb-6 w-full sm:w-auto border-b border-border pb-0 bg-transparent gap-6">
            <TabsTrigger
              value="suppliers"
              className="cursor-pointer pb-3 data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none px-2 text-sm font-medium"
            >
              <Store className="h-4 w-4 mr-2" />
              Supplier ({supplierBookmarks.length})
            </TabsTrigger>
            <TabsTrigger
              value="products"
              className="cursor-pointer pb-3 data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none px-2 text-sm font-medium"
            >
              <Package className="h-4 w-4 mr-2" />
              Produk ({productBookmarks.length})
            </TabsTrigger>
          </TabsList>

          {/* Suppliers Tab */}
          <TabsContent value="suppliers" className="outline-none">
            {supplierBookmarks.length === 0 ? (
              <div className="text-center py-20 bg-card rounded-lg border border-border">
                <Heart className="mx-auto h-12 w-12 text-muted-foreground opacity-40" />
                <h3 className="mt-4 text-base font-semibold text-foreground">
                  Belum ada supplier disimpan
                </h3>
                <p className="mt-2 text-sm text-muted-foreground max-w-xs mx-auto">
                  Cari supplier dan simpan untuk membandingkan mereka di sini.
                </p>
                <Button
                  asChild
                  className="mt-6 cursor-pointer hover:-translate-y-0.5 active:translate-y-0 transition-transform bg-primary text-primary-foreground rounded-lg"
                >
                  <Link href="/search">Cari Supplier</Link>
                </Button>
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {supplierBookmarks.map((supplier) => (
                  <Card
                    key={supplier.id}
                    className="overflow-hidden border border-border shadow-xs hover:shadow-md transition-all duration-300 rounded-lg bg-card"
                  >
                    <CardContent className="p-6 space-y-5">
                      <div className="flex items-start justify-between gap-4">
                        <Link
                          href={`/demo/suppliers/${supplier.supplierSlug}`}
                          className="cursor-pointer group flex items-start gap-4"
                        >
                          <div className="h-12 w-12 rounded bg-muted flex items-center justify-center text-foreground font-bold text-base border border-border group-hover:border-primary transition-colors">
                            {supplier.companyName.substring(0, 2).toUpperCase()}
                          </div>
                          <div className="space-y-1">
                            <h3 className="font-semibold text-foreground text-base group-hover:text-primary transition-colors leading-snug">
                              {supplier.companyName}
                            </h3>
                            <div className="flex items-center gap-1 text-muted-foreground text-xs font-normal">
                              <MapPin className="h-3.5 w-3.5 shrink-0" />
                              <span>{supplier.location}</span>
                            </div>
                          </div>
                        </Link>

                        <Button
                          onClick={() => setDeleteId(supplier.id)}
                          variant="ghost"
                          size="icon"
                          className="h-8 w-8 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-lg cursor-pointer transition-colors"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>

                      {/* Metadata row */}
                      <div className="flex flex-wrap items-center gap-x-4 gap-y-2 text-xs text-muted-foreground border-y border-border py-3">
                        <span>{supplier.businessType}</span>
                        <span className="w-1.5 h-1.5 rounded-full bg-border" />
                        <span>Est. {supplier.establishedYear}</span>
                        <span className="w-1.5 h-1.5 rounded-full bg-border" />
                        <div className="flex items-center gap-1 text-foreground font-medium">
                          <Star className="h-3.5 w-3.5 fill-warning text-warning" />
                          <span>
                            {supplier.rating} ({supplier.reviewCount})
                          </span>
                        </div>
                        {supplier.isVerified && (
                          <>
                            <span className="w-1.5 h-1.5 rounded-full bg-border" />
                            <Badge variant="outline" className="text-[10px] text-success border-success bg-success/5 font-semibold py-0 px-2 rounded-lg">
                              Terverifikasi
                            </Badge>
                          </>
                        )}
                      </div>

                      {/* Key Products */}
                      {supplier.keyProducts && supplier.keyProducts.length > 0 && (
                        <div className="space-y-2">
                          <h4 className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground">
                            {t("cardProducts")}
                          </h4>
                          <div className="flex flex-wrap gap-1.5">
                            {supplier.keyProducts.map((product, idx) => (
                              <Badge
                                key={idx}
                                variant="secondary"
                                className="text-[10px] text-foreground bg-muted border-0 px-2 py-0.5 rounded-lg"
                              >
                                {product}
                              </Badge>
                            ))}
                          </div>
                        </div>
                      )}

                      {/* Footer Actions */}
                      <div className="pt-2 flex items-center justify-between gap-4 border-t border-border">
                        <label className="flex items-center gap-2.5 text-xs font-medium text-foreground cursor-pointer select-none">
                          <input
                            type="checkbox"
                            checked={compareList.includes(supplier.id)}
                            onChange={() => handleToggleCompare(supplier.id)}
                            className="h-4 w-4 rounded border-border text-primary focus:ring-primary cursor-pointer transition-all"
                          />
                          <span>{t("compareCheckbox")}</span>
                        </label>

                        <div className="flex gap-2">
                          <Button
                            asChild
                            variant="outline"
                            size="sm"
                            className="text-xs font-semibold border-border hover:border-muted-foreground cursor-pointer transition-all hover:-translate-y-0.5 active:translate-y-0 rounded-lg"
                          >
                            <Link href={`/demo/suppliers/${supplier.supplierSlug}`}>
                              <ExternalLink className="mr-1.5 h-3.5 w-3.5" />
                              {t("btnProfile")}
                            </Link>
                          </Button>
                          <Button
                            asChild
                            size="sm"
                            className="bg-primary text-primary-foreground hover:bg-primary/95 text-xs font-semibold cursor-pointer transition-all hover:-translate-y-0.5 active:translate-y-0 shadow-lg hover:shadow-primary/20 rounded-lg"
                          >
                            <Link href="/rfq/create">
                              <MessageSquare className="mr-1.5 h-3.5 w-3.5" />
                              {t("btnRfq")}
                            </Link>
                          </Button>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}
          </TabsContent>

          {/* Products Tab */}
          <TabsContent value="products" className="outline-none">
            {productBookmarks.length === 0 ? (
              <div className="text-center py-20 bg-card rounded-lg border border-border">
                <Heart className="mx-auto h-12 w-12 text-muted-foreground opacity-40" />
                <h3 className="mt-4 text-base font-semibold text-foreground">
                  Belum ada produk disimpan
                </h3>
                <p className="mt-2 text-sm text-muted-foreground max-w-xs mx-auto">
                  Jelajahi produk dari supplier terpilih dan simpan di sini.
                </p>
                <Button
                  asChild
                  className="mt-6 cursor-pointer hover:-translate-y-0.5 active:translate-y-0 transition-transform bg-primary text-primary-foreground rounded-lg"
                >
                  <Link href="/search">Cari Produk</Link>
                </Button>
              </div>
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-6">
                {productBookmarks.map((bookmark) => (
                  <Card
                    key={bookmark.id}
                    className="overflow-hidden border border-border shadow-xs hover:shadow-md transition-all duration-300 rounded-lg bg-card flex flex-col justify-between"
                  >
                    <div>
                      {/* Product Image / Placeholder */}
                      <div className="relative aspect-video bg-muted border-b border-border flex items-center justify-center overflow-hidden">
                        {bookmark.productImage ? (
                          <img
                            src={bookmark.productImage}
                            alt={bookmark.productName}
                            className="object-cover w-full h-full hover:scale-105 transition-transform duration-500"
                          />
                        ) : (
                          <Package className="h-10 w-10 text-muted-foreground opacity-30" />
                        )}
                        <Button
                          onClick={() => setDeleteId(bookmark.id)}
                          variant="ghost"
                          size="icon"
                          className="absolute right-2.5 top-2.5 h-8 w-8 text-foreground/75 bg-card/85 backdrop-blur-xs hover:text-destructive hover:bg-destructive/10 rounded-lg cursor-pointer transition-colors shadow-xs"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>

                      <CardContent className="p-5 space-y-4">
                        <div className="space-y-1">
                          <Link
                            href={`/demo/suppliers/${bookmark.supplierSlug}#product-${bookmark.supplierProductId}`}
                            className="cursor-pointer group block"
                          >
                            <h3 className="font-semibold text-foreground text-sm group-hover:text-primary transition-colors leading-tight line-clamp-2">
                              {bookmark.productName}
                            </h3>
                          </Link>
                          <p className="text-xs text-muted-foreground">
                            Dari:{" "}
                            <Link
                              href={`/demo/suppliers/${bookmark.supplierSlug}`}
                              className="text-foreground font-medium hover:underline hover:text-primary cursor-pointer"
                            >
                              {bookmark.companyName}
                            </Link>
                          </p>
                        </div>

                        <div className="space-y-1 border-t border-border pt-3">
                          <p className="text-xs text-muted-foreground">Harga Mulai</p>
                          <p className="text-base font-bold text-primary">
                            {bookmark.productPrice ? formatPrice(bookmark.productPrice) : "Hubungi Kami"}
                          </p>
                        </div>
                      </CardContent>
                    </div>

                    <div className="px-5 pb-5 pt-0">
                      <div className="flex items-center justify-between gap-3 text-xs text-muted-foreground border-t border-border pt-3">
                        <span>Min. Order: {bookmark.productMinOrder || "1 Pcs"}</span>
                        <Button
                          asChild
                          size="sm"
                          className="bg-primary text-primary-foreground hover:bg-primary/95 text-xs font-semibold cursor-pointer transition-all hover:-translate-y-0.5 active:translate-y-0 shadow-lg hover:shadow-primary/20 rounded-lg"
                        >
                          <Link href="/rfq/create">
                            <MessageSquare className="mr-1.5 h-3.5 w-3.5" />
                            Minta Penawaran
                          </Link>
                        </Button>
                      </div>
                    </div>
                  </Card>
                ))}
              </div>
            )}
          </TabsContent>
        </Tabs>

        <DeleteDialog
          open={!!deleteId}
          onOpenChange={(open) => !open && setDeleteId(null)}
          onConfirm={async () => {
            if (deleteId) {
              await deleteBookmark(deleteId);
              setDeleteId(null);
            }
          }}
          itemName="bookmark"
        />
      </div>
    </BuyerLayout>
  );
}
