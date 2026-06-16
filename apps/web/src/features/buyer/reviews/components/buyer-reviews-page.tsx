"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Star, CheckCircle2, MessageSquare, ShoppingBag } from "lucide-react";
import { toast } from "sonner";
import { cn } from "@/lib/utils";

interface ReviewItem {
  id: string;
  poNumber: string;
  supplierName: string;
  productName: string;
  date: string;
  reviewed: boolean;
}

export function BuyerReviewsPage() {
  const t = useTranslations("buyer.reviews");
  const [items, setItems] = useState<ReviewItem[]>([
    {
      id: "tx-2",
      poNumber: "PO-20260612-C02D3E",
      supplierName: "CV Tekstil Nusantara",
      productName: "Raw Indigo Denim Fabric 12oz",
      date: "2026-06-12",
      reviewed: false,
    },
    {
      id: "tx-10",
      poNumber: "PO-20260530-X99Z1A",
      supplierName: "PT Baja Sentosa",
      productName: "Reinforced Steel Bar (Rebar) D10",
      date: "2026-05-30",
      reviewed: true,
    },
  ]);

  const [selectedItemId, setSelectedItemId] = useState<string | null>(null);
  const [rating, setRating] = useState<number>(5);
  const [reviewText, setReviewText] = useState<string>("");

  const pendingItems = items.filter((item) => !item.reviewed);
  const completedItems = items.filter((item) => item.reviewed);

  const selectedItem = items.find((item) => item.id === selectedItemId);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedItemId) return;

    // Simulate submission
    setItems((prev) =>
      prev.map((item) =>
        item.id === selectedItemId ? { ...item, reviewed: true } : item
      )
    );

    toast.success(t("success"));
    setSelectedItemId(null);
    setRating(5);
    setReviewText("");
  };

  return (
    <div className="space-y-6">
      {/* Title */}
      <div className="flex flex-col gap-1">
        <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">{t("title")}</h1>
        <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
      </div>

      {/* Grid Content */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
        {/* Left Column: Eligible List */}
        <div className="lg:col-span-2 space-y-4">
          <div className="bg-card rounded-xl border border-border p-5 shadow-xs space-y-4">
            <h3 className="text-sm font-extrabold text-foreground tracking-tight border-b border-border pb-3 flex items-center gap-2">
              <ShoppingBag className="h-4.5 w-4.5 text-primary" />
              Menunggu Ulasan
            </h3>

            {pendingItems.length === 0 ? (
              <div className="text-center py-10 space-y-2">
                <CheckCircle2 className="h-8 w-8 text-success mx-auto" />
                <p className="text-sm text-muted-foreground font-semibold">{t("empty")}</p>
              </div>
            ) : (
              <div className="divide-y divide-border">
                {pendingItems.map((item) => (
                  <div key={item.id} className="py-4 flex items-start justify-between gap-4 first:pt-0 last:pb-0">
                    <div className="space-y-1">
                      <h4 className="font-extrabold text-sm text-foreground">{item.productName}</h4>
                      <p className="text-xs text-muted-foreground font-semibold">
                        Supplier: <span className="text-foreground">{item.supplierName}</span>
                      </p>
                      <p className="text-[10px] text-muted-foreground font-medium">
                        No. Transaksi: {item.poNumber} &bull; Selesai pada {item.date}
                      </p>
                    </div>
                    <Button
                      size="sm"
                      onClick={() => setSelectedItemId(item.id)}
                      className="h-8 text-xs font-bold shrink-0 cursor-pointer bg-primary text-primary-foreground hover:-translate-y-0.5 active:translate-y-0"
                    >
                      Beri Ulasan
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* History reviews */}
          <div className="bg-card rounded-xl border border-border p-5 shadow-xs space-y-4">
            <h3 className="text-sm font-extrabold text-foreground tracking-tight border-b border-border pb-3 flex items-center gap-2">
              <Star className="h-4.5 w-4.5 text-warning fill-warning" />
              Riwayat Ulasan Anda
            </h3>
            {completedItems.length === 0 ? (
              <p className="text-xs text-muted-foreground font-medium py-4">Belum ada riwayat ulasan.</p>
            ) : (
              <div className="divide-y divide-border">
                {completedItems.map((item) => (
                  <div key={item.id} className="py-4 flex items-start justify-between gap-4 first:pt-0 last:pb-0">
                    <div className="space-y-1">
                      <h4 className="font-extrabold text-sm text-foreground">{item.productName}</h4>
                      <p className="text-xs text-muted-foreground font-semibold">
                        Supplier: <span className="text-foreground">{item.supplierName}</span>
                      </p>
                      <div className="flex items-center gap-1">
                        {[1, 2, 3, 4, 5].map((s) => (
                          <Star key={s} className="h-3.5 w-3.5 fill-warning text-warning" />
                        ))}
                      </div>
                    </div>
                    <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-md text-[10px] font-semibold bg-success/10 text-success border border-success/20 uppercase tracking-wider">
                      Ulasan Dikirim
                    </span>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Right Column: Submission Form */}
        <div className="space-y-6">
          {selectedItem ? (
            <div className="bg-card rounded-xl border border-border p-5 shadow-xs space-y-4">
              <div className="border-b border-border pb-3">
                <h3 className="text-sm font-extrabold text-foreground tracking-tight">Form Ulasan</h3>
                <p className="text-xs text-muted-foreground mt-0.5 font-semibold">
                  Mendukung {selectedItem.supplierName}
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
                    placeholder="Contoh: Kualitas produk baja sangat kokoh, pengiriman cepat dan aman."
                    className="h-10 rounded-lg border-border focus-visible:ring-primary focus-visible:border-primary text-sm"
                  />
                </div>

                {/* Submit button */}
                <Button type="submit" className="w-full h-10 text-xs font-bold cursor-pointer bg-primary text-primary-foreground hover:-translate-y-0.5 active:translate-y-0">
                  {t("submitBtn")}
                </Button>
              </form>
            </div>
          ) : (
            <div className="bg-card rounded-xl border border-border p-6 shadow-xs text-center space-y-2">
              <MessageSquare className="h-6 w-6 text-muted-foreground mx-auto" />
              <h4 className="font-extrabold text-sm text-foreground">Pilih Produk</h4>
              <p className="text-xs text-muted-foreground">
                Silakan pilih salah satu produk di sebelah kiri untuk menulis ulasan.
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
