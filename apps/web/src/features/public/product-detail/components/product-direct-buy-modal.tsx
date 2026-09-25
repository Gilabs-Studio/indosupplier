"use client";
/* eslint-disable @next/next/no-img-element */

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { useRouter } from "@/i18n/routing";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Field, FieldLabel } from "@/components/ui/field";
import { transactionService } from "@/features/buyer/transactions/services/transaction.service";
import { resolveImageUrl, formatPrice } from "@/lib/utils";
import { Loader2, ShoppingBag, Package } from "lucide-react";
import { toast } from "sonner";
import type { PublicProductDto, PublicSupplierDto } from "@/features/public/search/types";

interface ProductDirectBuyModalProps {
  isOpen: boolean;
  onClose: () => void;
  product: PublicProductDto;
  supplier: PublicSupplierDto;
  quantity: number;
  selectedVariantText: string;
}

export function ProductDirectBuyModal({
  isOpen,
  onClose,
  product,
  supplier,
  quantity,
  selectedVariantText,
}: ProductDirectBuyModalProps) {
  const t = useTranslations("public.productDetail");
  const router = useRouter();
  const [address, setAddress] = useState("");
  const [notes, setNotes] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [imageError, setImageError] = useState(false);

  const subtotal = product.price * quantity;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!address.trim()) {
      toast.error(t("directBuyAddressRequired"));
      return;
    }

    try {
      setIsSubmitting(true);
      const res = await transactionService.createTransaction({
        supplier_profile_id: supplier.id,
        product_name: product.name,
        quantity_value: quantity,
        quantity_unit: "Unit",
        price_per_unit: product.price,
        delivery_address: address.trim(),
        notes: `[${t("rfqVariantPrefix")}: ${selectedVariantText}] ${notes.trim()}`.trim(),
      });

      toast.success(t("directBuySuccess"));
      onClose();
      router.push(`/transactions/${res.id}`);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : t("directBuyError");
      toast.error(msg);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent size="md" className="rounded-lg border border-border bg-card p-6 shadow-xl max-h-[90vh] overflow-y-auto">
        <DialogHeader className="space-y-1 text-left">
          <DialogTitle className="text-base font-extrabold font-heading text-foreground">
            {t("directBuyTitle")}
          </DialogTitle>
          <DialogDescription className="text-xs text-muted-foreground">
            {t("directBuySubtitle")}
          </DialogDescription>
        </DialogHeader>

        {/* Product & Order Summary */}
        <div className="mt-3 rounded-lg border border-border bg-muted/20 p-3.5 space-y-3">
          <div className="flex items-center gap-3">
            <div className="h-12 w-12 shrink-0 overflow-hidden rounded-md border border-border bg-card flex items-center justify-center">
              {product.photos && product.photos.length > 0 && !imageError ? (
                <img
                  src={resolveImageUrl(product.photos[0])}
                  alt=""
                  onError={() => setImageError(true)}
                  className="h-full w-full object-cover"
                />
              ) : (
                <Package className="h-6 w-6 text-muted-foreground" />
              )}
            </div>
            <div className="min-w-0 flex-1">
              <h4 className="truncate text-xs font-bold text-foreground font-heading">{product.name}</h4>
              <p className="truncate text-[11px] text-muted-foreground mt-0.5">
                {t("rfqSummarySupplier")}{" "}
                <span className="font-semibold text-foreground">{supplier.companyName}</span>
                {" • "}
                {t("rfqSummaryVariant")}{" "}
                <span className="text-primary font-semibold">{selectedVariantText}</span>
              </p>
            </div>
          </div>

          <div className="border-t border-border/80 pt-2.5 space-y-1 text-xs">
            <div className="flex justify-between text-muted-foreground">
              <span>{t("unitPrice")}</span>
              <span className="font-semibold text-foreground">{formatPrice(product.price, product.currency)}</span>
            </div>
            <div className="flex justify-between text-muted-foreground">
              <span>{t("orderQuantity")}</span>
              <span className="font-semibold text-foreground">{quantity} Unit</span>
            </div>
            <div className="flex justify-between text-foreground font-bold border-t border-border/80 pt-1.5 text-sm">
              <span>{t("estimatedTotal")}</span>
              <span className="text-primary font-extrabold">{formatPrice(subtotal, product.currency)}</span>
            </div>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="mt-4 space-y-4">
          <Field>
            <FieldLabel className="text-xs font-bold text-foreground">{t("directBuyAddressLabel")}</FieldLabel>
            <Textarea
              value={address}
              onChange={(e) => setAddress(e.target.value)}
              placeholder={t("directBuyAddressPlaceholder")}
              rows={3}
              required
              className="resize-none text-xs"
            />
          </Field>

          <Field>
            <FieldLabel className="text-xs font-bold text-foreground">{t("directBuyNotesLabel")}</FieldLabel>
            <Input
              type="text"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder={t("directBuyNotesPlaceholder")}
              className="h-9 text-xs"
            />
          </Field>

          <div className="flex items-center justify-end gap-2 pt-2 border-t border-border">
            <Button
              type="button"
              variant="outline"
              onClick={onClose}
              disabled={isSubmitting}
              className="cursor-pointer text-xs font-semibold border-border rounded-lg h-9 px-4 hover:bg-muted"
            >
              {t("btnCancel")}
            </Button>
            <Button
              type="submit"
              disabled={isSubmitting}
              className="cursor-pointer text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/90 rounded-lg h-9 px-5 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 shadow-xs hover:shadow-primary/30"
            >
              {isSubmitting ? (
                <div className="flex items-center gap-1.5">
                  <Loader2 className="h-3.5 w-3.5 animate-spin" />
                  <span>{t("directBuySubmitting")}</span>
                </div>
              ) : (
                <div className="flex items-center gap-1.5">
                  <ShoppingBag className="h-3.5 w-3.5" />
                  <span>{t("directBuySubmitBtn")}</span>
                </div>
              )}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
