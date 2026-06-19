"use client";
/* eslint-disable @next/next/no-img-element */

import { useState } from "react";
import { useTranslations } from "next-intl";
import { MessageSquare, Package, Trash2 } from "lucide-react";

import { CenteredLoading } from "@/components/loading";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { DeleteDialog } from "@/components/ui/delete-dialog";
import { Link, useRouter } from "@/i18n/routing";
import { useBuyerCompare } from "@/features/buyer/compare/hooks/useBuyerCompare";

import { BuyerLayout } from "../../components/buyer-layout";
import { useBuyerBookmarks } from "../hooks/useBuyerBookmarks";

export function BuyerBookmarksPage() {
  const t = useTranslations("buyer.bookmarks");
  const router = useRouter();
  const { bookmarks, isLoading: isBookmarksLoading, deleteBookmark } = useBuyerBookmarks();
  const { products: comparedProducts, addProduct, removeProduct, isLoading: isCompareLoading } = useBuyerCompare();
  const [deleteId, setDeleteId] = useState<string | null>(null);

  const productBookmarks = bookmarks.filter((item) => item.type === "product");
  const compareProductsList = comparedProducts.map((item) => item.id);

  const handleToggleProductCompare = (supplierProductId: string) => {
    if (compareProductsList.includes(supplierProductId)) {
      removeProduct(supplierProductId);
      return;
    }
    if (compareProductsList.length >= 5) {
      return;
    }
    addProduct(supplierProductId);
  };

  const formatPrice = (price: number) =>
    new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
    }).format(price);

  if (isBookmarksLoading || isCompareLoading) {
    return (
      <BuyerLayout>
        <CenteredLoading />
      </BuyerLayout>
    );
  }

  return (
    <BuyerLayout>
      <div className="space-y-6">
        <div className="space-y-1 text-left">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">{t("title")}</h1>
          <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
        </div>

        {productBookmarks.length === 0 ? (
          <div className="rounded-lg border border-border bg-card p-12 text-center shadow-xs">
            <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-muted text-muted-foreground">
              <Package className="h-5 w-5" />
            </div>
            <h2 className="mt-4 text-base font-semibold text-foreground">{t("emptyTitle")}</h2>
            <p className="mx-auto mt-2 max-w-md text-sm text-muted-foreground">{t("emptyDesc")}</p>
            <Button asChild className="mt-5 cursor-pointer">
              <Link href="/search">{t("btnSearchProducts")}</Link>
            </Button>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
            {productBookmarks.map((bookmark) => (
              <Card
                key={bookmark.id}
                className="overflow-hidden rounded-lg border border-border bg-card shadow-xs transition-all duration-200 hover:-translate-y-0.5 hover:shadow-md cursor-pointer"
                role="link"
                tabIndex={0}
                onClick={() => router.push(`/demo/products/${bookmark.supplierProductId}`)}
                onKeyDown={(event) => {
                  if (event.key === "Enter" || event.key === " ") {
                    event.preventDefault();
                    router.push(`/demo/products/${bookmark.supplierProductId}`);
                  }
                }}
              >
                <div className="relative aspect-4/3 border-b border-border bg-muted">
                  {bookmark.productImage ? (
                    <img src={bookmark.productImage} alt={bookmark.productName} className="h-full w-full object-cover" />
                  ) : (
                    <div className="flex h-full items-center justify-center text-muted-foreground">
                      <Package className="h-8 w-8" />
                    </div>
                  )}
                  <Button
                    type="button"
                    variant="secondary"
                    size="icon"
                    className="absolute right-3 top-3 h-8 w-8 cursor-pointer"
                    onClick={(event) => {
                      event.stopPropagation();
                      setDeleteId(bookmark.id);
                    }}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>

                <CardContent className="space-y-4 p-5">
                  <div className="space-y-1">
                    <Link
                      href={`/demo/products/${bookmark.supplierProductId}`}
                      className="block"
                      onClick={(event) => event.stopPropagation()}
                    >
                      <h2 className="line-clamp-2 text-sm font-semibold text-foreground">{bookmark.productName}</h2>
                    </Link>
                    <Link
                      href={`/demo/suppliers/${bookmark.supplierSlug}`}
                      className="text-xs text-muted-foreground hover:text-primary"
                      onClick={(event) => event.stopPropagation()}
                    >
                      {bookmark.companyName}
                    </Link>
                  </div>

                  <div className="space-y-1">
                    <p className="text-xs text-muted-foreground">Harga mulai</p>
                    <p className="text-base font-semibold text-foreground">
                      {bookmark.productPrice ? formatPrice(bookmark.productPrice) : "Hubungi Supplier"}
                    </p>
                    <p className="text-xs text-muted-foreground">MOQ {bookmark.productMinOrder || "1 Pcs"}</p>
                  </div>

                  <div className="flex items-center justify-between gap-3 border-t border-border pt-4">
                    <label className="flex items-center gap-2 text-xs font-medium text-foreground">
                      <input
                        type="checkbox"
                        checked={compareProductsList.includes(bookmark.supplierProductId || "")}
                        onChange={() => handleToggleProductCompare(bookmark.supplierProductId || "")}
                        onClick={(event) => event.stopPropagation()}
                        className="h-4 w-4 rounded border-border text-primary focus:ring-primary"
                      />
                      <span>{t("compareCheckbox")}</span>
                    </label>
                    <Button asChild variant="outline" className="cursor-pointer text-xs font-medium">
                      <Link href="/rfq/create" onClick={(event) => event.stopPropagation()}>
                        <MessageSquare className="mr-1.5 h-3.5 w-3.5" />
                        RFQ
                      </Link>
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}

        <DeleteDialog
          open={deleteId !== null}
          onOpenChange={(open) => {
            if (!open) setDeleteId(null);
          }}
          onConfirm={() => {
            if (deleteId) {
              deleteBookmark(deleteId);
              setDeleteId(null);
            }
          }}
          itemName="bookmark"
        />
      </div>
    </BuyerLayout>
  );
}
