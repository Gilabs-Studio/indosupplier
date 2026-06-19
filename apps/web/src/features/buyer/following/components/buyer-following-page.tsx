"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { BadgeCheck, MapPin, Package, Store } from "lucide-react";

import { CenteredLoading } from "@/components/loading";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { DeleteDialog } from "@/components/ui/delete-dialog";
import { Link, useRouter } from "@/i18n/routing";

import { BuyerLayout } from "../../components/buyer-layout";
import { useBuyerFollowing } from "../hooks/useBuyerFollowing";

export function BuyerFollowingPage() {
  const t = useTranslations("buyer.following");
  const router = useRouter();
  const { following, isLoading, unfollowSupplier } = useBuyerFollowing();

  const [deleteTarget, setDeleteTarget] = useState<{ id: string; name: string } | null>(null);

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
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="space-y-1 text-left">
            <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">{t("title")}</h1>
            <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
          </div>
          <Button asChild variant="outline" className="w-full cursor-pointer sm:w-auto font-semibold transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-md hover:bg-secondary">
            <Link href="/search">{t("btnSearchSuppliers")}</Link>
          </Button>
        </div>

        {following.length === 0 ? (
          <div className="rounded-lg border border-border bg-card p-12 text-center shadow-xs">
            <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-muted text-muted-foreground">
              <Store className="h-5 w-5" />
            </div>
            <h2 className="mt-4 text-base font-semibold text-foreground">{t("emptyTitle")}</h2>
            <p className="mx-auto mt-2 max-w-md text-sm text-muted-foreground">{t("emptyDesc")}</p>
            <Button asChild className="mt-5 cursor-pointer">
              <Link href="/search">{t("btnSearchSuppliers")}</Link>
            </Button>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
            {following.map((supplier) => {
              return (
                <Card
                  key={supplier.id}
                  className="group overflow-hidden rounded-lg border border-border bg-card shadow-xs transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg hover:shadow-primary/5 cursor-pointer flex flex-col justify-between"
                  role="link"
                  tabIndex={0}
                  onClick={() => router.push(`/demo/suppliers/${supplier.supplierSlug}`)}
                  onKeyDown={(event) => {
                    if (event.key === "Enter" || event.key === " ") {
                      event.preventDefault();
                      router.push(`/demo/suppliers/${supplier.supplierSlug}`);
                    }
                  }}
                >
                  <CardContent className="p-5 flex flex-col h-full justify-between">
                    <div>
                      <div className="flex items-center justify-between gap-4">
                        <div className="flex min-w-0 items-center gap-3">
                          <div className="relative h-11 w-11 shrink-0 overflow-hidden rounded-lg border border-border bg-muted flex items-center justify-center">
                            {supplier.logo ? (
                              // eslint-disable-next-line @next/next/no-img-element
                              <img
                                src={supplier.logo}
                                alt={supplier.companyName}
                                className="h-full w-full object-cover"
                              />
                            ) : (
                              <div className="flex h-full w-full items-center justify-center bg-primary/10 text-sm font-bold text-primary">
                                {supplier.companyName.slice(0, 2).toUpperCase()}
                              </div>
                            )}
                          </div>
                          <div className="min-w-0 space-y-0.5 text-left">
                            <div className="flex flex-wrap items-center gap-1.5">
                              <h2 className="truncate text-sm font-semibold text-foreground group-hover:text-primary transition-colors duration-300">
                                {supplier.companyName}
                              </h2>
                              {supplier.isVerified && (
                                <Badge variant="outline" className="h-4.5 rounded-md border-success/30 bg-success/5 px-1 text-[9px] font-bold text-success flex items-center gap-0.5 shrink-0">
                                  <BadgeCheck className="h-3 w-3 shrink-0" />
                                  {t("badgeVerified")}
                                </Badge>
                              )}
                            </div>
                            <p className="flex items-center gap-1 text-xs text-muted-foreground">
                              <MapPin className="h-3.5 w-3.5 shrink-0" />
                              <span className="truncate">{supplier.location || "Indonesia"}</span>
                            </p>
                          </div>
                        </div>

                        <Button
                          type="button"
                          variant="outline"
                          className="cursor-pointer text-xs font-semibold px-3 py-1.5 h-8 border-border text-foreground hover:bg-destructive/10 hover:text-destructive hover:border-destructive/30 transition-all duration-300 rounded-lg shrink-0"
                          onClick={(event) => {
                            event.stopPropagation();
                            setDeleteTarget({ id: supplier.supplierProfileId, name: supplier.companyName });
                          }}
                        >
                          {t("btnFollowing")}
                        </Button>
                      </div>

                      <div className="border-t border-border/80 my-4" />

                      {supplier.keyProducts && supplier.keyProducts.length > 0 ? (
                        <div className="grid grid-cols-3 gap-3">
                          {supplier.keyProducts.slice(0, 3).map((product, idx) => (
                            <Link
                              key={`${product.id || idx}-${idx}`}
                              href={`/demo/products/${product.id}`}
                              className="group/product relative aspect-square overflow-hidden rounded-lg border border-border/80 bg-muted cursor-pointer"
                              onClick={(event) => event.stopPropagation()}
                            >
                              {product.image ? (
                                // eslint-disable-next-line @next/next/no-img-element
                                <img
                                  src={product.image}
                                  alt={product.name}
                                  className="h-full w-full object-cover transition-transform duration-500 group-hover/product:scale-105"
                                />
                              ) : (
                                <div className="flex h-full w-full items-center justify-center text-muted-foreground">
                                  <Package className="h-5 w-5" />
                                </div>
                              )}
                              <div className="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/60 to-transparent p-1.5 opacity-0 group-hover/product:opacity-100 transition-opacity duration-300">
                                <p className="truncate text-[9px] font-medium text-white">{product.name}</p>
                              </div>
                            </Link>
                          ))}
                        </div>
                      ) : (
                        <div className="h-20 flex items-center justify-center rounded-lg border border-dashed border-border bg-muted/10">
                          <p className="text-xs text-muted-foreground">Tidak ada produk</p>
                        </div>
                      )}
                    </div>
                  </CardContent>
                </Card>
              );
            })}
          </div>
        )}
      </div>

      <DeleteDialog
        open={deleteTarget !== null}
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null);
        }}
        onConfirm={() => {
          if (deleteTarget) {
            unfollowSupplier(deleteTarget.id);
            setDeleteTarget(null);
          }
        }}
        title={t("btnUnfollow")}
        description={deleteTarget ? `Supplier ${deleteTarget.name} akan dihapus dari daftar mengikuti Anda.` : ""}
        isLoading={false}
        itemName="supplier"
      />
    </BuyerLayout>
  );
}
