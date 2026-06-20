"use client";

import { useTranslations } from "next-intl";
import { MessageSquare, Package, Trash2 } from "lucide-react";

import { CenteredLoading } from "@/components/loading";
import { Button } from "@/components/ui/button";
import { DeleteDialog } from "@/components/ui/delete-dialog";
import { Link } from "@/i18n/routing";
import { ProductCard } from "@/components/ui/product-card";

import { BuyerLayout } from "../../components/buyer-layout";
import { useBuyerBookmarksPage } from "../hooks/use-buyer-bookmarks-page";

export function BuyerBookmarksPage() {
  const t = useTranslations("buyer.bookmarks");
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
              <ProductCard
                key={bookmark.id}
                id={bookmark.id}
                name={bookmark.productName || ""}
                price={bookmark.productPrice || 0}
                image={bookmark.productImage}
                href={`/demo/products/${bookmark.supplierProductId}`}
                supplierName={bookmark.companyName}
                supplierHref={`/demo/suppliers/${bookmark.supplierSlug}`}
                minOrder={bookmark.productMinOrder || undefined}
                moqLabel={t("moqDefault")}
                priceLabel={t("contactSupplier")}
                customOverlayButton={
                  <Button
                    type="button"
                    variant="secondary"
                    size="icon"
                    className="h-8 w-8 bg-card/90 text-foreground shadow-xs backdrop-blur-xs cursor-pointer hover:-translate-y-0.5 active:translate-y-0 active:scale-95 transition-all duration-200"
                    onClick={(event) => {
                      event.preventDefault();
                      event.stopPropagation();
                      setDeleteId(bookmark.id);
                    }}
                  >
                    <Trash2 className="h-4 w-4 text-destructive" />
                  </Button>
                }
                customFooter={
                  <div className="flex items-center justify-between gap-3">
                    <label className="flex cursor-pointer items-center gap-2 text-xs font-semibold text-foreground">
                      <input
                        type="checkbox"
                        checked={compareProductsList.includes(bookmark.supplierProductId || "")}
                        onChange={() => handleToggleProductCompare(bookmark.supplierProductId || "")}
                        onClick={(event) => event.stopPropagation()}
                        className="h-4 w-4 cursor-pointer rounded border-border text-primary focus:ring-primary"
                      />
                      <span>{t("compareCheckbox")}</span>
                    </label>
                    <Button asChild variant="outline" className="cursor-pointer text-xs font-semibold h-8 rounded-lg">
                      <Link href="/rfq/create" onClick={(event) => event.stopPropagation()}>
                        <MessageSquare className="mr-1.5 h-3.5 w-3.5" />
                        RFQ
                      </Link>
                    </Button>
                  </div>
                }
              />
            ))}
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
