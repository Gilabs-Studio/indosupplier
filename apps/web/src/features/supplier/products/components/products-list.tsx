"use client";

import React, { useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Link } from "@/i18n/routing";
import { Edit2, Plus, Search, Trash2, Layers, Star, ShoppingBag } from "lucide-react";
import { useTranslations } from "next-intl";
import { useSupplierProducts, useDeleteProduct, useCategories } from "../hooks/useProducts";
import { Skeleton } from "@/components/ui/skeleton";
import { DeleteDialog } from "@/components/ui/delete-dialog";

export function SupplierProductsList() {
  const t = useTranslations("supplier.products");
  const [search, setSearch] = useState("");
  const [categoryId, setCategoryId] = useState("");
  const [deleteId, setDeleteId] = useState<string | null>(null);
  const page = 1;

  const { data, isLoading } = useSupplierProducts({
    search,
    category_id: categoryId || undefined,
    page,
    per_page: 12,
  });

  const { data: categories } = useCategories();
  const { mutate: deleteProduct } = useDeleteProduct();

  const handleDelete = (id: string) => {
    setDeleteId(id);
  };

  return (
    <div className="space-y-6 text-left">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-border/80 pb-6">
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
            {t("title")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("subtitle")}
          </p>
        </div>
        <Button asChild className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/95 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/20">
          <Link href="/supplier/products/create">
            <Plus className="mr-2 h-4 w-4" /> {t("addProduct")}
          </Link>
        </Button>
      </div>

      {/* Filter and Search Controls */}
      <div className="flex flex-col sm:flex-row gap-4 items-center justify-between">
        <div className="relative w-full sm:max-w-xs">
          <Search className="absolute left-3 top-2.5 h-4.5 w-4.5 text-muted-foreground" />
          <input
            type="text"
            placeholder={t("searchPlaceholder")}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-9 pr-4 py-1.5 bg-card border border-border text-sm rounded-lg outline-hidden focus:border-primary transition-all"
          />
        </div>

        <div className="w-full sm:max-w-xs">
          <select
            value={categoryId}
            onChange={(e) => setCategoryId(e.target.value)}
            className="w-full px-3 py-1.5 bg-card border border-border text-sm rounded-lg outline-hidden focus:border-primary transition-all cursor-pointer text-muted-foreground"
          >
            <option value="">{t("allCategories")}</option>
            {categories?.map((cat) => (
              <option key={cat.id} value={cat.id}>
                {cat.name}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Products Loading/Empty/List */}
      {isLoading ? (
        <div className="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-6">
          {[...Array(6)].map((_, idx) => (
            <Card key={idx} className="border border-border rounded-xl bg-card overflow-hidden h-[380px] flex flex-col justify-between">
              <Skeleton className="h-[180px] w-full" />
              <div className="p-5 space-y-3 flex-1">
                <Skeleton className="h-4 w-1/3" />
                <Skeleton className="h-6 w-3/4" />
                <Skeleton className="h-4 w-full" />
              </div>
              <div className="p-5 border-t border-border flex justify-between">
                <Skeleton className="h-9 w-20" />
                <Skeleton className="h-9 w-24" />
              </div>
            </Card>
          ))}
        </div>
      ) : !data || data.items.length === 0 ? (
        <div className="text-center py-20 bg-card rounded-xl border border-border">
          <ShoppingBag className="mx-auto h-12 w-12 text-muted-foreground opacity-40" />
          <h3 className="mt-4 text-base font-semibold text-foreground">{t("emptyTitle")}</h3>
          <p className="mt-2 text-sm text-muted-foreground max-w-xs mx-auto">
            {t("emptyDesc")}
          </p>
          <Button asChild className="mt-6 cursor-pointer hover:-translate-y-0.5 active:translate-y-0 transition-transform">
            <Link href="/supplier/products/create">{t("addProduct")}</Link>
          </Button>
        </div>
      ) : (
        <div className="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-6">
          {data.items.map((product) => {
            const hasPhotos = product.photos && product.photos.length > 0;
            const coverPhoto = hasPhotos ? product.photos[0].file_url : "";

            return (
              <Card
                key={product.id}
                className="overflow-hidden border border-border shadow-xs hover:shadow-md transition-all duration-300 rounded-xl bg-card flex flex-col justify-between h-full group"
              >
                <Link
                  href={`/supplier/products/${product.id}`}
                  className="flex-1 flex flex-col cursor-pointer"
                >
                  {/* Product Cover Photo */}
                  <div className="h-[180px] bg-muted/20 relative overflow-hidden shrink-0 border-b border-border w-full">
                    {coverPhoto ? (
                      // eslint-disable-next-line @next/next/no-img-element
                      <img
                        src={coverPhoto}
                        alt={product.name}
                        className="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
                      />
                    ) : (
                      <div className="w-full h-full flex flex-col items-center justify-center text-muted-foreground gap-2">
                        <ShoppingBag className="h-8 w-8 opacity-40" />
                        <span className="text-[10px] font-semibold uppercase tracking-wider">No Image</span>
                      </div>
                    )}

                    {product.is_featured && (
                      <Badge className="absolute top-3 left-3 bg-amber-500 hover:bg-amber-600 text-white border-0 flex items-center gap-1 text-[10px] px-2 py-0.5 rounded-full">
                        <Star className="h-3 w-3 fill-white" /> Featured
                      </Badge>
                    )}
                  </div>

                  {/* Content Body */}
                  <CardContent className="p-5 flex-1 flex flex-col justify-between">
                    <div className="space-y-2.5">
                      {/* Category Label */}
                      <div className="flex items-center gap-1 text-[11px] text-muted-foreground font-semibold uppercase tracking-wider">
                        <Layers className="h-3 w-3" />
                        <span>{product.category?.name || "Uncategorized"}</span>
                      </div>

                      {/* Title */}
                      <h3 className="font-bold text-foreground text-base leading-tight tracking-tight line-clamp-2">
                        {product.name}
                      </h3>

                      {/* Description excerpt */}
                      <p className="text-xs text-muted-foreground line-clamp-2">
                        {product.description}
                      </p>
                    </div>

                    {/* Metadata and Pricing */}
                    <div className="space-y-2 pt-3 border-t border-border mt-3">
                      <div className="flex justify-between items-center text-xs">
                        <span className="text-muted-foreground">MOQ</span>
                        <Badge variant="outline" className="bg-primary/5 text-primary border-primary/20 text-[10px] font-bold rounded-full px-2 py-0.5">
                          {product.moq || "N/A"}
                        </Badge>
                      </div>

                      <div className="flex justify-between items-center text-xs">
                        <span className="text-muted-foreground">Capacity</span>
                        <span className="font-semibold text-foreground truncate max-w-[150px]">{product.capacity_text || "N/A"}</span>
                      </div>

                      <div className="flex justify-between items-baseline pt-1">
                        <span className="text-[11px] text-muted-foreground uppercase font-bold tracking-wider">Price</span>
                        <span className="font-extrabold text-foreground text-sm">
                          {product.starting_price > 0 ? (
                            <>
                              {product.currency} {product.starting_price.toLocaleString("id-ID")}
                            </>
                          ) : (
                            "Inquire"
                          )}
                        </span>
                      </div>
                    </div>
                  </CardContent>
                </Link>

                {/* Actions Footer */}
                <div className="p-5 pt-0 border-t border-border/80 flex items-center justify-end gap-2 shrink-0">
                  <Button
                    asChild
                    variant="outline"
                    size="sm"
                    className="h-8 text-xs font-semibold border-border hover:border-muted-foreground cursor-pointer transition-all hover:-translate-y-0.5 active:translate-y-0"
                  >
                    <Link href={`/supplier/products/${product.id}/edit`}>
                      <Edit2 className="mr-1.5 h-3.5 w-3.5" /> Edit
                    </Link>
                  </Button>
                  <Button
                    onClick={() => handleDelete(product.id)}
                    variant="ghost"
                    size="sm"
                    className="h-8 text-xs font-semibold text-muted-foreground hover:text-destructive hover:bg-destructive/10 cursor-pointer border border-border"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </Button>
                </div>
              </Card>
            );
          })}
        </div>
      )}

      <DeleteDialog
        open={!!deleteId}
        onOpenChange={(open) => !open && setDeleteId(null)}
        onConfirm={() => {
          if (deleteId) {
            deleteProduct(deleteId);
          }
        }}
        itemName="product"
      />
    </div>
  );
}
