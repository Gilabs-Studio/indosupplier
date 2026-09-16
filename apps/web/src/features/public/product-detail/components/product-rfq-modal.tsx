"use client";
/* eslint-disable @next/next/no-img-element */

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { NumericInput } from "@/components/ui/numeric-input";
import { Textarea } from "@/components/ui/textarea";
import { Field, FieldLabel, FieldGroup } from "@/components/ui/field";
import { rfqService } from "@/features/buyer/rfq/services/rfq.service";
import { resolveImageUrl, formatPrice } from "@/lib/utils";
import { CheckCircle2, Loader2, Send, Package, ArrowRight } from "lucide-react";
import { toast } from "sonner";
import type { PublicProductDto, PublicSupplierDto } from "@/features/public/search/types";

interface ProductRfqModalProps {
  isOpen: boolean;
  onClose: () => void;
  product: PublicProductDto;
  supplier: PublicSupplierDto;
  initialQuantity: number;
  selectedVariantText: string;
}

export function ProductRfqModal({
  isOpen,
  onClose,
  product,
  supplier,
  initialQuantity,
  selectedVariantText,
}: ProductRfqModalProps) {
  const t = useTranslations("public.productDetail");
  const [quantity, setQuantity] = useState(String(initialQuantity || 1));
  const [unit, setUnit] = useState(product.minOrder ? product.minOrder.replace(/^[0-9\s]+/, "") || "Unit" : "Unit");
  const [targetPort, setTargetPort] = useState("");
  const [budget, setBudget] = useState(String(product.price ? product.price * (initialQuantity || 1) : ""));
  const [deliveryTimeline, setDeliveryTimeline] = useState("14 Hari Kerja");
  const [notes, setNotes] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [createdRfqId, setCreatedRfqId] = useState<string | null>(null);
  const [imageError, setImageError] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!targetPort.trim()) {
      toast.error(t("rfqPortRequired") || "Lokasi pengiriman wajib diisi");
      return;
    }

    try {
      setIsSubmitting(true);
      const budgetNum = parseFloat(budget.replace(/[^0-9.]/g, "")) || 0;
      const res = await rfqService.createRfq({
        product_name: product.name,
        category: product.categoryName || "General",
        quantity: String(quantity),
        unit: unit.trim() || "Unit",
        target_port: targetPort.trim(),
        description: `[Varian: ${selectedVariantText}] ${notes.trim()}`,
        target_supplier_id: supplier.id,
        product_id: product.id,
        image_url: product.photos?.[0],
        budget: budgetNum > 0 ? budgetNum : undefined,
        delivery_timeline: deliveryTimeline.trim() || undefined,
      });

      setCreatedRfqId(res.id);
      toast.success(t("rfqSuccessTitle") || "RFQ berhasil dikirim ke supplier!");
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Gagal mengirim RFQ";
      toast.error(msg);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleReset = () => {
    setCreatedRfqId(null);
    onClose();
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && handleReset()}>
      <DialogContent size="md" className="rounded-xl border-border bg-card p-6 shadow-xl max-h-[90vh] overflow-y-auto">
        {createdRfqId ? (
          <div className="py-6 text-center space-y-4">
            <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-success/15 text-success">
              <CheckCircle2 className="h-8 w-8" />
            </div>
            <div className="space-y-1">
              <h3 className="text-lg font-bold text-foreground font-heading">
                {t("rfqSuccessTitle") || "RFQ Berhasil Dikirim!"}
              </h3>
              <p className="text-xs text-muted-foreground max-w-md mx-auto leading-relaxed">
                {t("rfqSuccessDesc") || "Permintaan penawaran Anda telah diteruskan langsung ke supplier. Anda akan mendapatkan notifikasi saat supplier memberikan penawaran harga."}
              </p>
            </div>
            <div className="pt-4 flex flex-col sm:flex-row items-center justify-center gap-2.5">
              <Button asChild className="w-full sm:w-auto cursor-pointer font-bold text-xs bg-primary text-primary-foreground hover:bg-primary/95 rounded-lg h-9">
                <Link href="/rfq" onClick={handleReset}>
                  <span>{t("rfqViewList") || "Lihat Daftar RFQ Saya"}</span>
                  <ArrowRight className="ml-1.5 h-3.5 w-3.5" />
                </Link>
              </Button>
              <Button variant="outline" onClick={handleReset} className="w-full sm:w-auto cursor-pointer font-semibold text-xs border-border rounded-lg h-9">
                Tutup
              </Button>
            </div>
          </div>
        ) : (
          <>
            <DialogHeader className="space-y-1 text-left">
              <DialogTitle className="text-base font-bold font-heading text-foreground">
                {t("rfqModalTitle") || "Kirim Permintaan Penawaran (RFQ)"}
              </DialogTitle>
              <DialogDescription className="text-xs text-muted-foreground">
                {t("rfqModalSubtitle") || "Ajukan penawaran harga dan spesifikasi kebutuhan langsung ke supplier ini."}
              </DialogDescription>
            </DialogHeader>

            {/* Product & Supplier Summary Card */}
            <div className="mt-3 flex items-center gap-3 rounded-lg border border-border bg-muted/20 p-3">
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
                <h4 className="truncate text-xs font-bold text-foreground">{product.name}</h4>
                <p className="truncate text-[11px] text-muted-foreground mt-0.5">
                  Supplier: <span className="font-semibold text-foreground">{supplier.companyName}</span> • Varian: <span className="text-primary font-semibold">{selectedVariantText}</span>
                </p>
                <p className="text-xs font-extrabold text-foreground mt-0.5">
                  {formatPrice(product.price, product.currency)}
                </p>
              </div>
            </div>

            <form onSubmit={handleSubmit} className="mt-4 space-y-4">
              <FieldGroup className="grid grid-cols-1 sm:grid-cols-2 gap-3.5">
                <Field>
                  <FieldLabel className="text-xs font-bold">{t("rfqQtyLabel") || "Jumlah Kebutuhan"}</FieldLabel>
                  <Input
                    type="number"
                    min="1"
                    value={quantity}
                    onChange={(e) => {
                      setQuantity(e.target.value);
                      const q = parseFloat(e.target.value) || 0;
                      if (product.price) setBudget(String(product.price * q));
                    }}
                    required
                    className="h-9 text-xs"
                  />
                </Field>
                <Field>
                  <FieldLabel className="text-xs font-bold">{t("rfqUnitLabel") || "Satuan"}</FieldLabel>
                  <Input
                    type="text"
                    value={unit}
                    onChange={(e) => setUnit(e.target.value)}
                    placeholder="Contoh: Ton, Kg, Pcs, Box"
                    required
                    className="h-9 text-xs"
                  />
                </Field>
              </FieldGroup>

              <Field>
                <FieldLabel className="text-xs font-bold">{t("rfqPortLabel") || "Lokasi Pengiriman / Pelabuhan Tujuan"}</FieldLabel>
                <Input
                  type="text"
                  value={targetPort}
                  onChange={(e) => setTargetPort(e.target.value)}
                  placeholder={t("rfqPortPlaceholder") || "Contoh: Pelabuhan Tanjung Priok atau Gudang Kawasan Industri MM2100"}
                  required
                  className="h-9 text-xs"
                />
              </Field>

              <FieldGroup className="grid grid-cols-1 sm:grid-cols-2 gap-3.5">
                <Field>
                  <FieldLabel className="text-xs font-bold">{t("rfqBudgetLabel") || "Target Budget Total (Rp)"}</FieldLabel>
                  <NumericInput
                    value={budget ? Number(budget) : undefined}
                    onChange={(val) => setBudget(val !== undefined ? String(val) : "")}
                    placeholder="Contoh: 12.500.000"
                    className="h-9 text-xs"
                  />
                </Field>
                <Field>
                  <FieldLabel className="text-xs font-bold">{t("rfqDeliveryLabel") || "Target Waktu Pengiriman"}</FieldLabel>
                  <Input
                    type="text"
                    value={deliveryTimeline}
                    onChange={(e) => setDeliveryTimeline(e.target.value)}
                    placeholder={t("rfqDeliveryPlaceholder") || "Contoh: 14 Hari Kerja"}
                    className="h-9 text-xs"
                  />
                </Field>
              </FieldGroup>

              <Field>
                <FieldLabel className="text-xs font-bold">{t("rfqNotesLabel") || "Catatan & Spesifikasi Khusus"}</FieldLabel>
                <Textarea
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                  placeholder={t("rfqNotesPlaceholder") || "Tuliskan spesifikasi detail, sertifikasi mutu SNI/ISO yang dibutuhkan, atau persyaratan kemasan..."}
                  rows={3}
                  className="resize-none text-xs"
                />
              </Field>

              <div className="flex items-center justify-end gap-2 pt-2">
                <Button
                  type="button"
                  variant="outline"
                  onClick={onClose}
                  disabled={isSubmitting}
                  className="cursor-pointer text-xs font-semibold border-border rounded-lg h-9 px-4"
                >
                  Batal
                </Button>
                <Button
                  type="submit"
                  disabled={isSubmitting}
                  className="cursor-pointer text-xs font-bold bg-primary text-primary-foreground hover:bg-primary/95 rounded-lg h-9 px-5 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0"
                >
                  {isSubmitting ? (
                    <div className="flex items-center gap-1.5">
                      <Loader2 className="h-3.5 w-3.5 animate-spin" />
                      <span>{t("rfqSubmitting") || "Mengirim RFQ..."}</span>
                    </div>
                  ) : (
                    <div className="flex items-center gap-1.5">
                      <Send className="h-3.5 w-3.5" />
                      <span>{t("rfqSubmitBtn") || "Kirim RFQ Sekarang"}</span>
                    </div>
                  )}
                </Button>
              </div>
            </form>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}
