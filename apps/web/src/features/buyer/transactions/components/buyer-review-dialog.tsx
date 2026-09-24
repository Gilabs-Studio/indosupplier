"use client";

import React, { useState } from "react";
import Image from "next/image";
import { useTranslations } from "next-intl";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Star, ShoppingBag, Loader2, Building2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { TransactionItem } from "../types/transaction.types";
import { useBuyerReviews } from "@/features/buyer/reviews/hooks/useBuyerReviews";

interface BuyerReviewDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  transaction: TransactionItem | null;
  onSuccess?: () => void;
}

interface ReviewDialogInnerProps {
  transaction: TransactionItem;
  onClose: () => void;
  onSuccess?: () => void;
}

function ReviewDialogInner({
  transaction,
  onClose,
  onSuccess,
}: ReviewDialogInnerProps) {
  const t = useTranslations("buyer.transactions");
  const [rating, setRating] = useState<number>(5);
  const [hoverRating, setHoverRating] = useState<number>(0);
  const [reviewText, setReviewText] = useState<string>("");
  const { submitReview, isSubmitting } = useBuyerReviews();

  const quickPresets = [
    "Kualitas material sangat baik & presisi sesuai pesanan.",
    "Pengiriman cepat dan packing bundle rapi tanpa cacat.",
    "Pelayanan supplier sangat responsif dan informatif.",
    "Barang sesuai spesifikasi teknis dan standar SNI.",
  ];

  const handleTogglePreset = (preset: string) => {
    if (reviewText.includes(preset)) {
      const updated = reviewText
        .replace(preset, "")
        .replace(/\s{2,}/g, " ")
        .trim();
      setReviewText(updated);
    } else {
      setReviewText((prev) => (prev.trim() ? `${prev.trim()} ${preset}` : preset));
    }
  };

  const getRatingLabel = (score: number) => {
    switch (score) {
      case 5:
        return "Sangat Puas";
      case 4:
        return "Puas";
      case 3:
        return "Cukup";
      case 2:
        return "Kurang Memuaskan";
      default:
        return "Kecewa";
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (reviewText.trim().length < 5 || isSubmitting) return;

    submitReview(
      {
        purchaseOrderId: transaction.id,
        rating,
        reviewText: reviewText.trim(),
      },
      {
        onSuccess: () => {
          onClose();
          onSuccess?.();
        },
      }
    );
  };

  const effectiveRating = hoverRating || rating;

  return (
    <DialogContent className="sm:max-w-lg p-0 gap-0 overflow-hidden rounded-xl border border-border bg-card">
      {/* Header */}
      <DialogHeader className="p-6 pb-4 border-b border-border bg-muted/20">
        <DialogTitle className="text-base font-bold text-foreground">
          {t("reviewModalTitle")}
        </DialogTitle>
        <DialogDescription className="text-xs text-muted-foreground mt-1">
          {t("reviewModalSubtitle")}
        </DialogDescription>
      </DialogHeader>

        <form onSubmit={handleSubmit} className="p-6 space-y-5">
          {/* Target Product / Transaction Info Card */}
          <div className="flex items-center gap-3.5 p-3 rounded-lg border border-border bg-muted/40">
            <div className="relative h-14 w-14 rounded-lg border border-border bg-muted overflow-hidden shrink-0">
              {transaction.product_image ? (
                <Image
                  src={transaction.product_image}
                  alt={transaction.product_name}
                  fill
                  sizes="56px"
                  className="object-cover"
                />
              ) : (
                <div className="h-full w-full flex items-center justify-center bg-primary/10 text-primary">
                  <ShoppingBag className="h-6 w-6" />
                </div>
              )}
            </div>

            <div className="min-w-0 flex-1 space-y-0.5">
              <h4 className="font-extrabold text-xs text-foreground truncate">
                {transaction.product_name}
              </h4>
              <p className="text-[11px] text-muted-foreground flex items-center gap-1 font-medium truncate">
                <Building2 className="h-3 w-3 shrink-0" />
                <span>{transaction.supplier_name}</span>
              </p>
              <p className="text-[10px] text-muted-foreground font-semibold">
                PO: <span className="font-mono text-foreground">{transaction.po_number}</span> &bull; {transaction.quantity_value} {transaction.quantity_unit}
              </p>
            </div>
          </div>

          {/* Interactive Star Rating */}
          <div className="space-y-2 text-center py-1">
            <label className="text-xs font-bold text-foreground block">
              {t("ratingLabel")}
            </label>
            <div className="flex items-center justify-center gap-1.5">
              {[1, 2, 3, 4, 5].map((star) => (
                <button
                  key={star}
                  type="button"
                  onClick={() => setRating(star)}
                  onMouseEnter={() => setHoverRating(star)}
                  onMouseLeave={() => setHoverRating(0)}
                  className="cursor-pointer p-1 transition-transform active:scale-95 focus:outline-hidden"
                  aria-label={`Beri bintang ${star}`}
                >
                  <Star
                    className={cn(
                      "h-8 w-8 transition-colors duration-150",
                      star <= effectiveRating
                        ? "text-amber-400 fill-amber-400 drop-shadow-xs"
                        : "text-muted-foreground/25 fill-transparent hover:text-amber-300"
                    )}
                  />
                </button>
              ))}
            </div>
            <p className="text-xs font-semibold text-amber-500">
              {getRatingLabel(effectiveRating)}
            </p>
          </div>

          {/* Review Textarea */}
          <div className="space-y-1.5">
            <div className="flex items-center justify-between text-xs">
              <label htmlFor="review-feedback" className="font-bold text-foreground">
                {t("feedbackLabel")}
              </label>
              <span
                className={cn(
                  "text-[10px] font-semibold",
                  reviewText.length < 5
                    ? "text-muted-foreground"
                    : "text-primary"
                )}
              >
                {reviewText.length}/1000
              </span>
            </div>

            <textarea
              id="review-feedback"
              rows={4}
              maxLength={1000}
              placeholder={t("feedbackPlaceholder")}
              value={reviewText}
              onChange={(e) => setReviewText(e.target.value)}
              className="w-full rounded-lg border border-border bg-card p-3 text-xs font-medium text-foreground placeholder:text-muted-foreground focus:border-primary focus:outline-hidden focus:ring-1 focus:ring-primary transition-all resize-none"
            />
          </div>

          {/* Quick Suggestions Chips */}
          <div className="space-y-1.5">
            <p className="text-xs text-muted-foreground font-medium">
              {t("quickPresets")}
            </p>
            <div className="flex flex-wrap gap-1.5">
              {quickPresets.map((preset) => {
                const isSelected = reviewText.includes(preset);
                return (
                  <button
                    key={preset}
                    type="button"
                    onClick={() => handleTogglePreset(preset)}
                    className={cn(
                      "px-2.5 py-1 text-[11px] rounded-md border transition-all duration-150 cursor-pointer text-left",
                      isSelected
                        ? "bg-primary/15 text-primary border-primary/40 font-semibold shadow-xs"
                        : "bg-muted/40 text-muted-foreground border-border hover:bg-muted/80 hover:text-foreground"
                    )}
                  >
                    {isSelected ? "✓" : "+"} {preset}
                  </button>
                );
              })}
            </div>
          </div>

          {/* Dialog Footer Actions */}
          <DialogFooter className="pt-2 flex items-center justify-end gap-2 border-t border-border">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={onClose}
              disabled={isSubmitting}
              className="h-9 px-4 text-xs font-semibold cursor-pointer"
            >
              {t("cancel")}
            </Button>
            <Button
              type="submit"
              size="sm"
              disabled={reviewText.trim().length < 5 || isSubmitting}
              className="h-9 px-5 text-xs font-semibold bg-primary text-primary-foreground hover:-translate-y-0.5 active:translate-y-0 cursor-pointer shadow-sm shadow-primary/20 disabled:cursor-not-allowed disabled:pointer-events-none disabled:opacity-50 disabled:transform-none"
            >
              {isSubmitting ? (
                <>
                  <Loader2 className="mr-1.5 h-3.5 w-3.5 animate-spin" />
                  {t("submitting")}
                </>
              ) : (
                t("submitReview")
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
  );
}

export function BuyerReviewDialog({
  open,
  onOpenChange,
  transaction,
  onSuccess,
}: BuyerReviewDialogProps) {
  if (!transaction) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      {open && (
        <ReviewDialogInner
          key={transaction.id}
          transaction={transaction}
          onClose={() => onOpenChange(false)}
          onSuccess={onSuccess}
        />
      )}
    </Dialog>
  );
}
