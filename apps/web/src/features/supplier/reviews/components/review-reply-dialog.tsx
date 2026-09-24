"use client";

import React, { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
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
import { Badge } from "@/components/ui/badge";
import { Star, MessageSquareReply, Sparkles, Loader2 } from "lucide-react";
import { replyReviewSchema, type ReplyReviewFormValues } from "../schemas/reviews.schema";
import type { SupplierReviewItem, ReplyReviewPayload } from "../types/reviews.types";

interface ReviewReplyDialogProps {
  isOpen: boolean;
  review: SupplierReviewItem | null;
  onClose: () => void;
  onSubmit: (payload: ReplyReviewPayload) => void;
  isSubmitting: boolean;
}

export function ReviewReplyDialog({
  isOpen,
  review,
  onClose,
  onSubmit,
  isSubmitting,
}: ReviewReplyDialogProps) {
  const t = useTranslations("supplier.reviews");

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    reset,
    formState: { errors },
  } = useForm<ReplyReviewFormValues>({
    resolver: zodResolver(replyReviewSchema),
    defaultValues: {
      reply: "",
    },
  });

  const replyValue = watch("reply") || "";

  useEffect(() => {
    if (review) {
      reset({
        reply: review.supplier_reply || "",
      });
    }
  }, [review, reset]);

  const onFormSubmit = (data: ReplyReviewFormValues) => {
    onSubmit({ reply: data.reply });
  };

  const applyTemplate = (templateKey: "template1" | "template2" | "template3") => {
    setValue("reply", t(templateKey), { shouldValidate: true });
  };

  if (!review) return null;

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-xl w-full p-0 overflow-hidden rounded-xl border border-border bg-card shadow-2xl">
        <DialogHeader className="p-6 pb-4 border-b border-border bg-muted/20">
          <div className="flex items-center gap-2 text-primary font-semibold text-xs tracking-wider uppercase mb-1">
            <MessageSquareReply className="h-4 w-4" />
            <span>{t("replyModalTitle")}</span>
          </div>
          <DialogTitle className="text-lg font-bold font-heading text-foreground">
            {review.supplier_reply ? t("btnEditReply") : t("btnReply")}
          </DialogTitle>
          <DialogDescription className="text-xs text-muted-foreground">
            {t("replyModalSubtitle")}
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit(onFormSubmit)} className="space-y-4 p-6 pt-4">
          {/* Buyer Review Context Card */}
          <div className="bg-muted/30 border border-border rounded-lg p-3.5 space-y-2">
            <div className="flex items-center justify-between text-xs">
              <div className="flex items-center gap-2">
                <span className="font-bold text-foreground">
                  {review.buyer_company_name}
                </span>
                <span className="text-muted-foreground text-[11px]">
                  ({review.buyer_user_name})
                </span>
              </div>
              <div className="flex items-center gap-0.5">
                {[1, 2, 3, 4, 5].map((star) => (
                  <Star
                    key={star}
                    className={`h-3 w-3 ${
                      review.rating >= star
                        ? "fill-amber-400 text-amber-400"
                        : "text-muted-foreground/30"
                    }`}
                  />
                ))}
              </div>
            </div>

            {review.product_name && (
              <Badge variant="outline" className="text-[10px] font-medium bg-background">
                {review.product_name}
                {review.po_number && review.po_number !== "-" && (
                  <span className="ml-1 text-muted-foreground">
                    • {review.po_number}
                  </span>
                )}
              </Badge>
            )}

            <p className="text-xs text-foreground italic border-l-2 border-primary/40 pl-2.5 py-0.5">
              &ldquo;{review.review_text}&rdquo;
            </p>
          </div>

          {/* Quick Response Templates */}
          <div className="space-y-1.5">
            <div className="flex items-center gap-1.5 text-[11px] font-semibold text-muted-foreground">
              <Sparkles className="h-3.5 w-3.5 text-primary" />
              <span>{t("quickTemplates")}:</span>
            </div>
            <div className="flex flex-col gap-1.5">
              {(["template1", "template2", "template3"] as const).map((key) => (
                <button
                  key={key}
                  type="button"
                  onClick={() => applyTemplate(key)}
                  className="text-left text-xs bg-muted/40 hover:bg-muted p-2 rounded-lg border border-border/80 transition-colors text-foreground line-clamp-1 cursor-pointer hover:border-primary/40"
                >
                  <span className="font-medium text-primary mr-1">✦</span>
                  {t(key)}
                </button>
              ))}
            </div>
          </div>

          {/* Reply Textarea */}
          <div className="space-y-1.5">
            <div className="flex justify-between items-center text-xs">
              <label
                htmlFor="supplier-reply-input"
                className="font-semibold text-foreground text-xs"
              >
                {t("yourReply")}
              </label>
              <span
                className={`text-[11px] font-mono ${
                  replyValue.length > 900
                    ? "text-destructive font-bold"
                    : "text-muted-foreground"
                }`}
              >
                {replyValue.length}/1000 {t("charCount")}
              </span>
            </div>
            <textarea
              id="supplier-reply-input"
              rows={4}
              placeholder={t("replyPlaceholder")}
              {...register("reply")}
              className={`w-full px-3 py-2 bg-background border text-sm rounded-lg focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary transition-all text-left text-foreground resize-none ${
                errors.reply ? "border-destructive focus:ring-destructive" : "border-border"
              }`}
            />
            {errors.reply && (
              <p className="text-xs text-destructive font-medium">
                {errors.reply.message}
              </p>
            )}
          </div>

          <DialogFooter className="pt-2 flex items-center justify-end gap-2 border-t border-border">
            <Button
              type="button"
              variant="outline"
              onClick={onClose}
              disabled={isSubmitting}
              className="text-xs h-9 cursor-pointer border-border"
            >
              {t("btnCancel")}
            </Button>
            <Button
              type="submit"
              disabled={isSubmitting || replyValue.trim().length < 3}
              className="text-xs h-9 bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer font-semibold flex items-center gap-1.5"
            >
              {isSubmitting ? (
                <>
                  <Loader2 className="h-3.5 w-3.5 animate-spin" />
                  <span>{t("submitting")}</span>
                </>
              ) : (
                <span>{t("btnSubmitReply")}</span>
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
