"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Star, CheckCircle2, MessageSquare, ShoppingBag, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { BuyerLayout } from "../../components/buyer-layout";
import { useBuyerReviews } from "../hooks/useBuyerReviews";
import { useBuyerReviewsStore } from "../stores/useBuyerReviewsStore";

export function BuyerReviewsPage() {
  const t = useTranslations("buyerReviews");
  const { selectedItemId, setSelectedItemId } = useBuyerReviewsStore();
  const [rating, setRating] = useState<number>(5);
  const [reviewText, setReviewText] = useState<string>("");

  const {
    eligibleTransactions,
    isEligibleLoading,
    reviewHistory,
    isHistoryLoading,
    submitReview,
    isSubmitting,
  } = useBuyerReviews();

  const selectedItem = eligibleTransactions.find((item) => item.id === selectedItemId);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedItemId) return;

    submitReview(
      {
        purchaseOrderId: selectedItemId,
        rating,
        reviewText,
      },
      {
        onSuccess: () => {
          setSelectedItemId(null);
          setRating(5);
          setReviewText("");
        },
      }
    );
  };

  const getStatusColorClass = (status: string) => {
    switch (status) {
      case "approved":
        return "bg-success/15 text-success border-success/20";
      case "rejected":
        return "bg-destructive/15 text-destructive border-destructive/20";
      default:
        return "bg-warning/15 text-warning border-warning/20";
    }
  };

  const formatCurrency = (val: number) => {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
    }).format(val);
  };

  const renderContent = () => {
    if (isEligibleLoading || isHistoryLoading) {
      return (
        <div className="flex flex-col items-center justify-center py-20 gap-3 lg:col-span-3">
          <Loader2 className="h-8 w-8 text-primary animate-spin" />
          <p className="text-sm text-muted-foreground font-semibold">Memuat data ulasan...</p>
        </div>
      );
    }

    return (
      <>
        {/* Left Column: Eligible List & History */}
        <div className="lg:col-span-2 space-y-6">
          {/* Eligible List */}
          <div className="space-y-4">
            <h3 className="text-sm font-extrabold text-foreground tracking-tight flex items-center gap-2">
              <ShoppingBag className="h-4.5 w-4.5 text-primary" />
              {t("eligibleTitle")}
            </h3>

            {eligibleTransactions.length === 0 ? (
              <div className="bg-card rounded-xl border border-border p-10 text-center shadow-xs space-y-3">
                <CheckCircle2 className="h-8 w-8 text-success mx-auto" />
                <p className="text-sm text-muted-foreground font-semibold">{t("empty")}</p>
              </div>
            ) : (
              <div className="space-y-4">
                {eligibleTransactions.map((item) => (
                  <div
                    key={item.id}
                    className="bg-card rounded-xl border border-border p-5 shadow-xs hover:shadow-md transition-all duration-300 space-y-4"
                  >
                    {/* Card Header (Shopping bag icon, belanja label, date, status) */}
                    <div className="flex items-center justify-between border-b border-border pb-3">
                      <div className="flex items-center gap-2 text-xs text-muted-foreground font-semibold">
                        <ShoppingBag className="h-4 w-4 text-success" />
                        <span>Belanja</span>
                        <span>&bull;</span>
                        <span>{item.date}</span>
                      </div>
                      <span className="px-2.5 py-0.5 rounded-md text-[10px] font-extrabold bg-success/15 text-success border border-success/20 uppercase tracking-wider">
                        Selesai
                      </span>
                    </div>

                    {/* Supplier Name */}
                    <div className="font-extrabold text-sm text-foreground">
                      {item.supplierName}
                    </div>

                    {/* Card Body */}
                    <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                      {/* Product details */}
                      <div className="flex items-start gap-4">
                        <div className="bg-primary/10 h-14 w-14 rounded-lg flex items-center justify-center shrink-0">
                          <ShoppingBag className="h-6 w-6 text-primary" />
                        </div>
                        <div className="space-y-0.5 min-w-0">
                          <h4 className="font-extrabold text-sm text-foreground hover:text-primary transition-colors cursor-pointer truncate">
                            {item.productName}
                          </h4>
                          <p className="text-xs text-muted-foreground font-medium">
                            {item.quantityValue} {item.quantityUnit}
                          </p>
                          <p className="text-[10px] text-muted-foreground font-medium">
                            No. Transaksi: {item.poNumber}
                          </p>
                        </div>
                      </div>

                      {/* Total purchase amount */}
                      <div className="flex md:flex-col justify-between items-center md:items-end border-t md:border-t-0 border-border pt-3 md:pt-0 border-dashed md:border-l md:border-border md:pl-6 shrink-0 gap-1">
                        <span className="text-[10px] text-muted-foreground font-semibold uppercase tracking-wider">Total Belanja</span>
                        <span className="font-extrabold text-sm text-foreground">
                          {formatCurrency(item.totalAmount)}
                        </span>
                      </div>
                    </div>

                    {/* Footer Actions */}
                    <div className="flex items-center justify-end gap-3 pt-1">
                      <Button
                        size="sm"
                        onClick={() => setSelectedItemId(item.id)}
                        className={cn(
                          "h-8 text-xs font-bold shrink-0 cursor-pointer transition-all duration-300",
                          selectedItemId === item.id
                            ? "bg-secondary text-foreground hover:bg-secondary/80"
                            : "bg-primary text-primary-foreground hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/30"
                        )}
                      >
                        {t("pendingReview")}
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* History reviews */}
          <div className="space-y-4">
            <h3 className="text-sm font-extrabold text-foreground tracking-tight flex items-center gap-2">
              <Star className="h-4.5 w-4.5 text-warning fill-warning" />
              {t("historyTitle")}
            </h3>

            {reviewHistory.length === 0 ? (
              <div className="bg-card rounded-xl border border-border p-8 text-center">
                <p className="text-xs text-muted-foreground font-medium">{t("emptyHistory")}</p>
              </div>
            ) : (
              <div className="space-y-4">
                {reviewHistory.map((item) => (
                  <div
                    key={item.id}
                    className="bg-card rounded-xl border border-border p-5 shadow-xs hover:shadow-md transition-all duration-300 space-y-4"
                  >
                    {/* Card Header */}
                    <div className="flex items-center justify-between border-b border-border pb-3">
                      <div className="flex items-center gap-2 text-xs text-muted-foreground font-semibold">
                        <ShoppingBag className="h-4 w-4 text-success" />
                        <span>Belanja</span>
                        <span>&bull;</span>
                        <span>{item.date}</span>
                      </div>
                      <span className={cn(
                        "px-2.5 py-0.5 rounded-md text-[10px] font-extrabold border uppercase tracking-wider",
                        getStatusColorClass(item.status)
                      )}>
                        {item.status === "approved" ? t("reviewSubmitted") : item.status}
                      </span>
                    </div>

                    {/* Supplier Name */}
                    <div className="font-extrabold text-sm text-foreground">
                      {item.supplierName}
                    </div>

                    {/* Card Body */}
                    <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                      {/* Product details */}
                      <div className="flex items-start gap-4">
                        <div className="bg-primary/10 h-14 w-14 rounded-lg flex items-center justify-center shrink-0">
                          <ShoppingBag className="h-6 w-6 text-primary" />
                        </div>
                        <div className="space-y-0.5 min-w-0">
                          <h4 className="font-extrabold text-sm text-foreground truncate">
                            {item.productName}
                          </h4>
                          <p className="text-xs text-muted-foreground font-medium">
                            {item.quantityValue} {item.quantityUnit}
                          </p>
                          <p className="text-[10px] text-muted-foreground font-medium">
                            No. Transaksi: {item.poNumber}
                          </p>
                        </div>
                      </div>

                      {/* Total purchase amount */}
                      <div className="flex md:flex-col justify-between items-center md:items-end border-t md:border-t-0 border-border pt-3 md:pt-0 border-dashed md:border-l md:border-border md:pl-6 shrink-0 gap-1">
                        <span className="text-[10px] text-muted-foreground font-semibold uppercase tracking-wider">Total Belanja</span>
                        <span className="font-extrabold text-sm text-foreground">
                          {formatCurrency(item.totalAmount)}
                        </span>
                      </div>
                    </div>

                    {/* Stars & Review Content box */}
                    <div className="border-t border-border pt-4 space-y-2.5">
                      <div className="flex items-center gap-1">
                        {[1, 2, 3, 4, 5].map((s) => (
                          <Star
                            key={s}
                            className={cn(
                              "h-3.5 w-3.5",
                              s <= item.rating
                                ? "fill-warning text-warning"
                                : "text-muted-foreground/30"
                            )}
                          />
                        ))}
                      </div>

                      {item.reviewText && (
                        <div className="bg-muted/40 p-3 rounded-lg border border-border/50 text-xs italic text-muted-foreground leading-relaxed">
                          &quot;{item.reviewText}&quot;
                        </div>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Right Column: Submission Form */}
        <div className="space-y-6">
          {selectedItem ? (
            <div className="bg-card rounded-xl border border-border p-5 shadow-xs space-y-4 lg:sticky lg:top-24">
              <div className="border-b border-border pb-3">
                <h3 className="text-sm font-extrabold text-foreground tracking-tight">{t("formTitle")}</h3>
                <p className="text-xs text-muted-foreground mt-0.5 font-semibold">
                  {t("formSubtitle", { supplier: selectedItem.supplierName })}
                </p>
              </div>

              <form onSubmit={handleSubmit} className="space-y-4 font-medium text-sm">
                {/* Rating Pick */}
                <div className="space-y-1.5">
                  <label className="text-xs text-muted-foreground font-semibold">{t("rating")}</label>
                  <div className="flex items-center gap-1">
                    {[1, 2, 3, 4, 5].map((star) => (
                      <button
                        key={star}
                        type="button"
                        onClick={() => setRating(star)}
                        className="cursor-pointer transition-transform duration-200 hover:scale-110"
                      >
                        <Star
                          className={cn(
                            "h-7 w-7",
                            star <= rating
                              ? "text-warning fill-warning"
                              : "text-muted-foreground/40"
                          )}
                        />
                      </button>
                    ))}
                  </div>
                </div>

                {/* Review Text */}
                <div className="space-y-1.5">
                  <label className="text-xs text-muted-foreground font-semibold">{t("reviewText")}</label>
                  <Input
                    type="text"
                    value={reviewText}
                    onChange={(e) => setReviewText(e.target.value)}
                    required
                    placeholder={t("placeholderText")}
                    className="h-10 rounded-lg border-border focus-visible:ring-primary focus-visible:border-primary text-sm bg-card"
                  />
                </div>

                {/* Submit button */}
                <Button
                  type="submit"
                  disabled={isSubmitting}
                  className="w-full h-10 text-xs font-bold cursor-pointer bg-primary text-primary-foreground hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/30 transition-all duration-300"
                >
                  {isSubmitting ? (
                    <>
                      <Loader2 className="h-4 w-4 animate-spin mr-2" />
                      Memproses...
                    </>
                  ) : (
                    t("submitBtn")
                  )}
                </Button>
              </form>
            </div>
          ) : (
            <div className="bg-card rounded-xl border border-border p-6 shadow-xs text-center space-y-2 lg:sticky lg:top-24">
              <MessageSquare className="h-6 w-6 text-muted-foreground mx-auto" />
              <h4 className="font-extrabold text-sm text-foreground">{t("chooseProduct")}</h4>
              <p className="text-xs text-muted-foreground leading-normal">
                {t("chooseProductDesc")}
              </p>
            </div>
          )}
        </div>
      </>
    );
  };

  return (
    <BuyerLayout>
      <div className="space-y-6">
        {/* Title */}
        <div className="flex flex-col gap-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">{t("title")}</h1>
          <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
        </div>

        {/* Grid Content */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
          {renderContent()}
        </div>
      </div>
    </BuyerLayout>
  );
}
