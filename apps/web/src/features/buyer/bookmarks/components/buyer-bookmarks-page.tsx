"use client";

import { useLocale, useTranslations } from "next-intl";
import { Package, GitCompareArrows } from "lucide-react";

import { CenteredLoading } from "@/components/loading";
import { Button } from "@/components/ui/button";
import { DeleteDialog } from "@/components/ui/delete-dialog";
import { Link } from "@/i18n/routing";
import { ProductCard } from "@/components/ui/product-card";
import { getSmartFallbackImage } from "@/features/public/demo/components/marketplace-product-card";

import { BuyerLayout } from "../../components/buyer-layout";
import { useBuyerBookmarksPage } from "../hooks/use-buyer-bookmarks-page";

export function BuyerBookmarksPage() {
  const t = useTranslations("buyer.bookmarks");
  const locale = useLocale();
  const isEn = locale === "en";

  const {
    productBookmarks,
    compareProductsList,
    isLoading,
    deleteId,
    setDeleteId,
    handleToggleProductCompare,
    handleConfirmDelete,
  } = useBuyerBookmarksPage();

  if (isLoading) {
    return (
      <BuyerLayout>
        <CenteredLoading />
      </BuyerLayout>
    );
  }

  return (
    <BuyerLayout>
      <div className="space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="space-y-1 text-left">
            <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">{t("title")}</h1>
            <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
          </div>
          {compareProductsList.length > 0 && (
            <Button asChild className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/90 shadow-xs shrink-0 transition-all duration-200 hover:-translate-y-0.5 active:translate-y-0">
              <Link href="/compare?tab=products">
                <GitCompareArrows className="mr-1.5 h-4 w-4" />
                {t("compareCount", { count: compareProductsList.length })}
              </Link>
            </Button>
          )}
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
          <div className="grid grid-cols-2 gap-3 sm:gap-4 sm:grid-cols-3 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-4">
            {productBookmarks.map((bookmark) => {
              const productId = bookmark.supplierProductId || bookmark.id;
              const isCompared = compareProductsList.includes(productId);

              return (
                <ProductCard
                  key={bookmark.id}
                  id={bookmark.id}
                  name={bookmark.productName || ""}
                  price={bookmark.productPrice || 0}
                  currency="IDR"
                  unit={bookmark.productMinOrder || "unit"}
                  image={bookmark.productImage}
                  fallbackImage={getSmartFallbackImage(bookmark.productName || "", bookmark.category)}
                  href={`/demo/products/${productId}`}
                  isReadyStock={true}
                  isVerified={bookmark.isVerified ?? true}
                  rating={bookmark.rating || 4.9}
                  reviewCount={bookmark.reviewCount || 0}
                  ulasanLabel={isEn ? "reviews" : "terjual"}
                  supplierName={bookmark.companyName}
                  supplierLocation={bookmark.location || "Indonesia"}
                  supplierHref={bookmark.supplierSlug ? `/demo/suppliers/${bookmark.supplierSlug}` : undefined}
                  showBookmarkOverlayButton={true}
                  isBookmarked={true}
                  onBookmark={() => setDeleteId(bookmark.id)}
                  bookmarkAriaLabel={t("removeBookmarkAria")}
                  showCompareOverlayButton={true}
                  isCompared={isCompared}
                  onCompare={() => handleToggleProductCompare(productId)}
                  compareLabel={isCompared ? t("inCompareTooltip") : t("compareTooltip")}
                />
              );
            })}
          </div>
        )}

        <DeleteDialog
          open={deleteId !== null}
          onOpenChange={(open) => {
            if (!open) setDeleteId(null);
          }}
          onConfirm={handleConfirmDelete}
          itemName={t("itemTypeBookmark")}
        />
      </div>
    </BuyerLayout>
  );
}
