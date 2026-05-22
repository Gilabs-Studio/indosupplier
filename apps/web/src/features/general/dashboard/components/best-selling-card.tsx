"use client";

import { useState } from "react";
import { ChevronRight, Package } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useHasPermission } from "@/features/master-data/user-management/hooks/use-has-permission";
import { useProduct } from "@/features/master-data/product/hooks/use-products";
import { ProductDetailDialog } from "@/features/master-data/product/components/product/product-detail-dialog";
import { formatCurrency } from "@/lib/utils";
import type { TopProductRow } from "../types";

interface BestSellingCardProps {
  readonly data?: TopProductRow[];
  readonly isLoading?: boolean;
}

export function BestSellingCard({ data, isLoading }: BestSellingCardProps) {
  const t = useTranslations("dashboard");
  const rows = data ?? [];
  const [selectedProductId, setSelectedProductId] = useState<string | null>(null);
  const canViewProduct = useHasPermission("product.read");

  const { data: productData } = useProduct(selectedProductId ?? "", {
    enabled: !!selectedProductId,
  });

  return (
    <>
      <Card className="h-full">
      <CardHeader>
        <div className="flex items-start justify-between">
          <div>
            <CardTitle>{t("bestSelling.title")}</CardTitle>
            <CardDescription className="mt-1">
              {t("bestSelling.subtitle")}
            </CardDescription>
          </div>
          <Button variant="outline" size="icon">
            <ChevronRight className="h-4 w-4" aria-hidden="true" />
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {isLoading ? (
          <div className="space-y-3">
            {Array.from({ length: 5 }).map((_, i) => (
              <div
                key={i}
                className="h-14 animate-pulse rounded-md bg-muted"
              />
            ))}
          </div>
        ) : rows.length === 0 ? (
          <div className="flex h-36 items-center justify-center">
            <p className="text-sm text-muted-foreground">{t("noData")}</p>
          </div>
        ) : (
          rows.slice(0, 6).map((row, i) => {
            const rowClass =
              "hover:bg-muted flex items-center justify-between rounded-md border px-4 py-3" +
              (canViewProduct ? " cursor-pointer" : "");

            const handleClick = canViewProduct
              ? () => setSelectedProductId(row.id)
              : undefined;

            return (
              <div
                key={row.id || `bp-${i}`}
                className={rowClass}
                onClick={handleClick}
                role={canViewProduct ? "button" : undefined}
                tabIndex={canViewProduct ? 0 : undefined}
                onKeyDown={canViewProduct ? (e) => { if (e.key === "Enter") handleClick?.(); } : undefined}
              >
                <div className="flex items-center gap-4">
                  <div className="flex h-10 w-10 items-center justify-center rounded-md bg-muted">
                    <Package className="h-5 w-5 text-muted-foreground" />
                  </div>
                  <div>
                    <div className={"text-sm font-medium" + (canViewProduct ? " text-primary" : "")}>{row.name}</div>
                    <div className="text-xs text-muted-foreground">{row.sku}</div>
                  </div>
                </div>
                <div className="text-right">
                  <div className="text-sm font-semibold">{formatCurrency(row.revenue)}</div>
                  <div className="text-xs text-muted-foreground">
                    {row.quantity_sold} {t("bestSelling.itemsSold")}
                  </div>
                </div>
              </div>
            );
          })
        )}
      </CardContent>
    </Card>

    {canViewProduct && (
      <ProductDetailDialog
        open={!!selectedProductId}
        onOpenChange={(v) => { if (!v) setSelectedProductId(null); }}
        product={productData?.data ?? null}
      />
    )}
    </>
  );
}
