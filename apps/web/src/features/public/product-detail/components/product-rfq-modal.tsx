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
import { Field, FieldLabel, FieldGroup, FieldError } from "@/components/ui/field";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import { rfqService } from "@/features/buyer/rfq/services/rfq.service";
import { resolveImageUrl, formatPrice } from "@/lib/utils";
import { CheckCircle2, Loader2, Send, Package, ArrowRight, ArrowLeft, Info } from "lucide-react";
import { toast } from "sonner";
import type { PublicProductDto, PublicSupplierDto } from "@/features/public/search/types";

interface ProductRfqModalProps {
  isOpen: boolean;
  onClose: () => void;
  product: PublicProductDto;
  supplier: PublicSupplierDto;
  initialQuantity?: number;
  selectedVariantText?: string;
  initialNotes?: string;
}

export function ProductRfqModal({
  isOpen,
  onClose,
  product,
  supplier,
  initialQuantity = 1,
  selectedVariantText,
  initialNotes = "",
}: ProductRfqModalProps) {
  const t = useTranslations("public.productDetail");
  const { user } = useAuthStore();

  const effectiveVariantText = selectedVariantText || t("rfqDefaultVariant");

  const [step, setStep] = useState<1 | 2>(1);

  // Step 1: Buyer Contact Information
  const [buyerName, setBuyerName] = useState(user?.name || "");
  const [buyerPhone, setBuyerPhone] = useState("");
  const [buyerEmail, setBuyerEmail] = useState(user?.email || "");
  const [step1Errors, setStep1Errors] = useState<{ name?: string; phone?: string; email?: string }>({});

  // Step 2: Product Specifications
  const [quantity, setQuantity] = useState(String(initialQuantity || 1));
  const [unit, setUnit] = useState(
    product.minOrder ? product.minOrder.replace(/^[0-9\s]+/, "") || t("rfqDefaultUnit") : t("rfqDefaultUnit")
  );
  const [budget, setBudget] = useState(String(product.price ? product.price * (initialQuantity || 1) : ""));
  const [deliveryTimeline, setDeliveryTimeline] = useState(t("rfqDefaultTimeline"));
  const [notes, setNotes] = useState(initialNotes || "");

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [createdRfqId, setCreatedRfqId] = useState<string | null>(null);
  const [imageError, setImageError] = useState(false);

  const handleNextStep = (e: React.FormEvent) => {
    e.preventDefault();
    const errors: { name?: string; phone?: string; email?: string } = {};

    if (!buyerName.trim()) {
      errors.name = t("fieldFullNameRequired");
    }
    if (!buyerPhone.trim()) {
      errors.phone = t("fieldPhoneRequired");
    }
    if (!buyerEmail.trim()) {
      errors.email = t("fieldEmailRequired");
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(buyerEmail.trim())) {
      errors.email = t("fieldEmailInvalid");
    }

    if (Object.keys(errors).length > 0) {
      setStep1Errors(errors);
      return;
    }

    setStep1Errors({});
    setStep(2);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      setIsSubmitting(true);
      const budgetNum = parseFloat(budget.replace(/[^0-9.]/g, "")) || 0;

      const descParts = [
        `[${t("rfqContactPrefix")}: ${buyerName.trim()} | ${t("rfqTelPrefix")}: ${buyerPhone.trim()} | ${t("rfqEmailPrefix")}: ${buyerEmail.trim()}]`,
        effectiveVariantText ? `[${t("rfqVariantPrefix")}: ${effectiveVariantText}]` : "",
        notes.trim() || t("rfqDefaultDesc"),
      ]
        .filter(Boolean)
        .join(" ");

      const res = await rfqService.createRfq({
        product_name: product.name,
        category: product.categoryName || t("catalogDefault"),
        quantity: String(quantity),
        unit: unit.trim() || t("rfqDefaultUnit"),
        description: descParts,
        target_supplier_id: supplier.id,
        product_id: product.id,
        image_url: product.photos?.[0],
        budget: budgetNum > 0 ? budgetNum : undefined,
        delivery_timeline: deliveryTimeline.trim() || undefined,
      });

      setCreatedRfqId(res.id);
      toast.success(t("rfqSuccessTitle"));
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : t("rfqErrorSubmit");
      toast.error(msg);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleReset = () => {
    setCreatedRfqId(null);
    setStep(1);
    setStep1Errors({});
    onClose();
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && handleReset()}>
      <DialogContent size="md" className="rounded-lg border border-border bg-card p-6 shadow-xl max-h-[90vh] overflow-y-auto">
        {createdRfqId ? (
          <div className="py-6 text-center space-y-4">
            <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-success/15 text-success">
              <CheckCircle2 className="h-8 w-8" />
            </div>
            <div className="space-y-1">
              <h3 className="text-lg font-extrabold text-foreground font-heading">
                {t("rfqSuccessTitle")}
              </h3>
              <p className="text-xs text-muted-foreground max-w-md mx-auto leading-relaxed">
                {t("rfqSuccessDesc")}
              </p>
            </div>
            <div className="pt-4 flex flex-col sm:flex-row items-center justify-center gap-2.5">
              <Button
                asChild
                className="w-full sm:w-auto cursor-pointer font-semibold text-xs bg-primary text-primary-foreground hover:bg-primary/90 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 rounded-lg h-9 px-5"
              >
                <Link href="/rfq" onClick={handleReset}>
                  <span>{t("rfqViewList")}</span>
                  <ArrowRight className="ml-1.5 h-3.5 w-3.5" />
                </Link>
              </Button>
              <Button
                variant="outline"
                onClick={handleReset}
                className="w-full sm:w-auto cursor-pointer font-semibold text-xs border border-border rounded-lg h-9 px-4 hover:bg-muted"
              >
                {t("btnClose")}
              </Button>
            </div>
          </div>
        ) : (
          <>
            <DialogHeader className="space-y-1 text-left">
              <DialogTitle className="text-lg font-extrabold font-heading text-foreground">
                {t("rfqModalTitle")}
              </DialogTitle>
              <DialogDescription className="text-xs text-muted-foreground">
                {t("sendToSupplier", { supplier: supplier.companyName })}
              </DialogDescription>
            </DialogHeader>

            {/* Informative Guidance Banner */}
            <div className="mt-3 flex items-start gap-2.5 rounded-lg border border-primary/20 bg-primary/10 p-3 text-xs text-primary">
              <Info className="h-4 w-4 shrink-0 mt-0.5 text-primary" />
              <p className="text-xs font-medium leading-relaxed text-foreground">
                {t("modalRfqBanner")}
              </p>
            </div>

            {/* Stepper Navigation */}
            <div className="flex items-center justify-between border-b border-border py-2.5">
              <span className="text-[11px] font-bold tracking-wider text-muted-foreground uppercase font-heading">
                {step === 1 ? t("step1Title") : t("step2Title")}
              </span>
              <span className="text-xs font-extrabold text-primary">
                {step} / 2
              </span>
            </div>

            {/* Step 1: Contact Details */}
            {step === 1 && (
              <form onSubmit={handleNextStep} className="mt-4 space-y-4">
                <FieldGroup className="space-y-3.5">
                  <Field>
                    <FieldLabel className="text-xs font-bold text-foreground">
                      {t("fieldFullName")} <span className="text-destructive">*</span>
                    </FieldLabel>
                    <Input
                      type="text"
                      value={buyerName}
                      onChange={(e) => {
                        setBuyerName(e.target.value);
                        if (step1Errors.name) setStep1Errors((prev) => ({ ...prev, name: undefined }));
                      }}
                      placeholder={t("fieldFullNamePlaceholder")}
                      required
                      className="h-9 text-xs"
                    />
                    {step1Errors.name && <FieldError>{step1Errors.name}</FieldError>}
                  </Field>

                  <Field>
                    <FieldLabel className="text-xs font-bold text-foreground">
                      {t("fieldPhone")} <span className="text-destructive">*</span>
                    </FieldLabel>
                    <Input
                      type="tel"
                      value={buyerPhone}
                      onChange={(e) => {
                        setBuyerPhone(e.target.value);
                        if (step1Errors.phone) setStep1Errors((prev) => ({ ...prev, phone: undefined }));
                      }}
                      placeholder={t("fieldPhonePlaceholder")}
                      required
                      className="h-9 text-xs"
                    />
                    {step1Errors.phone && <FieldError>{step1Errors.phone}</FieldError>}
                  </Field>

                  <Field>
                    <FieldLabel className="text-xs font-bold text-foreground">
                      {t("fieldEmail")} <span className="text-destructive">*</span>
                    </FieldLabel>
                    <Input
                      type="email"
                      value={buyerEmail}
                      onChange={(e) => {
                        setBuyerEmail(e.target.value);
                        if (step1Errors.email) setStep1Errors((prev) => ({ ...prev, email: undefined }));
                      }}
                      placeholder={t("fieldEmailPlaceholder")}
                      required
                      className="h-9 text-xs"
                    />
                    {step1Errors.email && <FieldError>{step1Errors.email}</FieldError>}
                  </Field>
                </FieldGroup>

                <div className="flex items-center justify-end gap-2 pt-3 border-t border-border">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={onClose}
                    className="cursor-pointer text-xs font-semibold border-border rounded-lg h-9 px-4 hover:bg-muted"
                  >
                    {t("btnCancel")}
                  </Button>
                  <Button
                    type="submit"
                    className="cursor-pointer text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/90 rounded-lg h-9 px-6 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 shadow-xs hover:shadow-primary/30"
                  >
                    <span>{t("btnNext")}</span>
                    <ArrowRight className="ml-1.5 h-3.5 w-3.5" />
                  </Button>
                </div>
              </form>
            )}

            {/* Step 2: Order Specifications */}
            {step === 2 && (
              <form onSubmit={handleSubmit} className="mt-4 space-y-4">
                {/* Product Summary Card */}
                <div className="flex items-center gap-3 rounded-lg border border-border bg-muted/30 p-3">
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
                      <span className="text-primary font-semibold">{effectiveVariantText}</span>
                    </p>
                    <p className="text-xs font-extrabold text-primary mt-0.5">
                      {formatPrice(product.price, product.currency) || t("negotiable")}
                    </p>
                  </div>
                </div>

                <FieldGroup className="grid grid-cols-1 sm:grid-cols-2 gap-3.5">
                  <Field>
                    <FieldLabel className="text-xs font-bold text-foreground">
                      {t("rfqQtyLabel")} <span className="text-destructive">*</span>
                    </FieldLabel>
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
                    <FieldLabel className="text-xs font-bold text-foreground">
                      {t("rfqUnitLabel")} <span className="text-destructive">*</span>
                    </FieldLabel>
                    <Input
                      type="text"
                      value={unit}
                      onChange={(e) => setUnit(e.target.value)}
                      placeholder={t("rfqUnitPlaceholder")}
                      required
                      className="h-9 text-xs"
                    />
                  </Field>
                </FieldGroup>

                <FieldGroup className="grid grid-cols-1 sm:grid-cols-2 gap-3.5">
                  <Field>
                    <FieldLabel className="text-xs font-bold text-foreground">
                      {t("rfqBudgetLabel")}
                    </FieldLabel>
                    <NumericInput
                      value={budget ? Number(budget) : undefined}
                      onChange={(val) => setBudget(val !== undefined ? String(val) : "")}
                      placeholder={t("rfqBudgetPlaceholder")}
                      className="h-9 text-xs"
                    />
                  </Field>
                  <Field>
                    <FieldLabel className="text-xs font-bold text-foreground">
                      {t("rfqDeliveryLabel")}
                    </FieldLabel>
                    <Input
                      type="text"
                      value={deliveryTimeline}
                      onChange={(e) => setDeliveryTimeline(e.target.value)}
                      placeholder={t("rfqDeliveryPlaceholder")}
                      className="h-9 text-xs"
                    />
                  </Field>
                </FieldGroup>

                <Field>
                  <FieldLabel className="text-xs font-bold text-foreground">
                    {t("rfqNotesLabel")}
                  </FieldLabel>
                  <Textarea
                    value={notes}
                    onChange={(e) => setNotes(e.target.value)}
                    placeholder={t("rfqNotesPlaceholder")}
                    rows={3}
                    className="resize-none text-xs"
                  />
                </Field>

                <div className="flex items-center justify-between gap-2 pt-3 border-t border-border">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => setStep(1)}
                    disabled={isSubmitting}
                    className="cursor-pointer text-xs font-semibold border-border rounded-lg h-9 px-4 hover:bg-muted"
                  >
                    <ArrowLeft className="mr-1.5 h-3.5 w-3.5" />
                    <span>{t("btnBack")}</span>
                  </Button>
                  <Button
                    type="submit"
                    disabled={isSubmitting}
                    className="cursor-pointer text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/90 rounded-lg h-9 px-6 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 shadow-xs hover:shadow-primary/30"
                  >
                    {isSubmitting ? (
                      <div className="flex items-center gap-1.5">
                        <Loader2 className="h-3.5 w-3.5 animate-spin" />
                        <span>{t("rfqSubmitting")}</span>
                      </div>
                    ) : (
                      <div className="flex items-center gap-1.5">
                        <Send className="h-3.5 w-3.5" />
                        <span>{t("rfqSubmitBtn")}</span>
                      </div>
                    )}
                  </Button>
                </div>
              </form>
            )}
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}
