"use client";

import React from "react";
import { useTranslations } from "next-intl";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Star,
  MessageCircleReply,
  CheckCircle2,
  Clock,
  Building2,
  Package,
  FileText,
  Edit2,
} from "lucide-react";
import { formatDate, getInitials } from "../utils/reviews.utils";
import type { SupplierReviewItem } from "../types/reviews.types";

interface ReviewCardProps {
  review: SupplierReviewItem;
  locale?: string;
  onOpenReply: (review: SupplierReviewItem) => void;
}

export function ReviewCard({ review, locale = "id", onOpenReply }: ReviewCardProps) {
  const t = useTranslations("supplier.reviews");
  const isReplied = Boolean(review.supplier_reply && review.supplier_reply.trim());

  return (
    <Card className="border border-border bg-card rounded-xl shadow-xs overflow-hidden transition-all duration-200 hover:border-border/80">
      <div className="p-5 space-y-4">
        {/* Top Header: Buyer Profile & Rating */}
        <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-3 border-b border-border/60 pb-3">
          <div className="flex items-center gap-3">
            {review.buyer_avatar_url ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img
                src={review.buyer_avatar_url}
                alt={review.buyer_user_name}
                className="h-10 w-10 rounded-full border border-border bg-muted object-cover shrink-0"
              />
            ) : (
              <div className="h-10 w-10 rounded-full border border-border bg-primary/10 text-primary font-bold text-xs flex items-center justify-center shrink-0">
                {getInitials(review.buyer_user_name)}
              </div>
            )}
            <div className="space-y-0.5">
              <div className="flex items-center gap-2 flex-wrap">
                <h4 className="text-sm font-bold text-foreground">
                  {review.buyer_company_name}
                </h4>
                <Badge
                  variant="secondary"
                  className="text-[10px] font-semibold flex items-center gap-1 bg-muted/60"
                >
                  <Building2 className="h-3 w-3" />
                  <span>{review.buyer_user_name}</span>
                </Badge>
              </div>
              <div className="flex items-center gap-2">
                <div className="flex items-center gap-0.5">
                  {[1, 2, 3, 4, 5].map((star) => (
                    <Star
                      key={star}
                      className={`h-3.5 w-3.5 ${
                        review.rating >= star
                          ? "fill-amber-400 text-amber-400"
                          : "text-muted-foreground/30"
                      }`}
                    />
                  ))}
                </div>
                <span className="text-[11px] text-muted-foreground font-medium">
                  {formatDate(review.created_at, locale)}
                </span>
              </div>
            </div>
          </div>

          {/* Status Badge */}
          <div className="flex items-center gap-2 self-start">
            {isReplied ? (
              <Badge
                variant="outline"
                className="text-[10px] font-semibold bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/30 flex items-center gap-1"
              >
                <CheckCircle2 className="h-3 w-3" />
                <span>{t("replied")}</span>
              </Badge>
            ) : (
              <Badge
                variant="outline"
                className="text-[10px] font-semibold bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/30 flex items-center gap-1"
              >
                <Clock className="h-3 w-3" />
                <span>{t("needReply")}</span>
              </Badge>
            )}
          </div>
        </div>

        {/* Product & PO Info Meta */}
        <div className="flex items-center gap-2 flex-wrap text-xs text-muted-foreground">
          {review.product_name && (
            <div className="flex items-center gap-1.5 bg-muted/30 px-2.5 py-1 rounded-md border border-border/60">
              <Package className="h-3.5 w-3.5 text-primary" />
              <span className="font-medium text-foreground">{review.product_name}</span>
            </div>
          )}
          {review.po_number && review.po_number !== "-" && (
            <div className="flex items-center gap-1.5 bg-muted/30 px-2.5 py-1 rounded-md border border-border/60">
              <FileText className="h-3.5 w-3.5 text-muted-foreground" />
              <span>
                {t("poNumber")}: <span className="font-mono font-medium">{review.po_number}</span>
              </span>
            </div>
          )}
        </div>

        {/* Review Comment Text */}
        <div className="bg-muted/15 border border-border/60 p-3.5 rounded-lg">
          <p className="text-xs sm:text-sm text-foreground leading-relaxed">
            {review.review_text}
          </p>
        </div>

        {/* Supplier Response Section */}
        {isReplied ? (
          <div className="bg-primary/5 border border-primary/15 rounded-lg p-3.5 space-y-2 ml-2 sm:ml-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-1.5 text-primary text-[11px] font-bold uppercase tracking-wider">
                <MessageCircleReply className="h-3.5 w-3.5" />
                <span>{t("yourReply")}</span>
                {review.supplier_replied_at && (
                  <span className="text-[10px] font-normal text-muted-foreground lowercase">
                    • {formatDate(review.supplier_replied_at, locale)}
                  </span>
                )}
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onOpenReply(review)}
                className="h-7 px-2 text-[11px] text-primary hover:text-primary hover:bg-primary/10 cursor-pointer flex items-center gap-1"
              >
                <Edit2 className="h-3 w-3" />
                <span>{t("btnEditReply")}</span>
              </Button>
            </div>
            <p className="text-xs text-foreground/90 leading-relaxed font-medium">
              {review.supplier_reply}
            </p>
          </div>
        ) : (
          <div className="flex items-center justify-between pt-1">
            <span className="text-xs text-muted-foreground flex items-center gap-1.5">
              <Clock className="h-3.5 w-3.5 text-amber-500" />
              <span>{t("unrepliedDesc")}</span>
            </span>
            <div className="flex items-center gap-2">
              <Button
                onClick={() => onOpenReply(review)}
                size="sm"
                className="text-xs h-8 bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer font-semibold shadow-xs hover:-translate-y-0.5 active:translate-y-0 transition-all flex items-center gap-1.5"
              >
                <MessageCircleReply className="h-3.5 w-3.5" />
                <span>{t("btnReply")}</span>
              </Button>
            </div>
          </div>
        )}
      </div>
    </Card>
  );
}
