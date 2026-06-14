"use client";

import React from "react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";
import { useBuyerTransactionDetail } from "../hooks/useBuyerTransactions";
import { Button } from "@/components/ui/button";
import { Loader2, ArrowLeft, Calendar, FileText, CheckCircle2, Circle, Truck, Wallet, ClipboardList, MessageSquare, Star, ShoppingBag } from "lucide-react";
import { cn, formatCurrency } from "@/lib/utils";

interface BuyerTransactionDetailPageProps {
  readonly id: string;
}

export function BuyerTransactionDetailPage({ id }: BuyerTransactionDetailPageProps) {
  const t = useTranslations("buyer.transactions");
  const { data: tx, isLoading } = useBuyerTransactionDetail(id);

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center py-40 gap-3">
        <Loader2 className="h-8 w-8 text-primary animate-spin" />
        <p className="text-sm text-muted-foreground">Memuat detail transaksi...</p>
      </div>
    );
  }

  if (!tx) {
    return (
      <div className="text-center py-20 space-y-4">
        <h2 className="text-xl font-bold text-foreground">Transaksi Tidak Ditemukan</h2>
        <Link href="/transactions">
          <Button variant="outline" className="cursor-pointer">
            <ArrowLeft className="mr-2 h-4 w-4" /> Kembali ke Daftar
          </Button>
        </Link>
      </div>
    );
  }

  // Stepper steps
  const steps = [
    { label: "Pending", desc: "Menunggu pembayaran" },
    { label: "Processing", desc: "Diproses supplier" },
    { label: "Shipped", desc: "Dalam pengiriman" },
    { label: "Completed", desc: "Pesanan selesai" },
  ];

  const getStepIndex = (status: string) => {
    switch (status) {
      case "cancelled":
        return -1;
      case "completed":
        return 3;
      case "shipped":
        return 2;
      case "processing":
        return 1;
      default:
        return 0; // pending
    }
  };

  const currentStepIdx = getStepIndex(tx.status);

  return (
    <div className="space-y-6">
      {/* Back button */}
      <div>
        <Link href="/transactions">
          <Button variant="ghost" size="sm" className="h-9 px-3 text-muted-foreground hover:text-foreground cursor-pointer">
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali ke Daftar Transaksi
          </Button>
        </Link>
      </div>

      {/* Main Info Card */}
      <div className="bg-card rounded-xl border border-border p-6 shadow-xs space-y-6">
        <div className="flex flex-wrap items-center justify-between gap-4 border-b border-border pb-4">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <span className="text-xs text-muted-foreground">PO Number:</span>
              <span className="text-base font-extrabold text-foreground flex items-center gap-1.5">
                <FileText className="h-4.5 w-4.5 text-primary" />
                {tx.po_number}
              </span>
            </div>
            <p className="text-xs text-muted-foreground flex items-center gap-1.5 font-medium">
              <Calendar className="h-3.5 w-3.5" />
              Dibuat pada:{" "}
              <span className="text-foreground">
                {new Date(tx.created_at).toLocaleDateString("id-ID", {
                  year: "numeric",
                  month: "long",
                  day: "numeric",
                  hour: "2-digit",
                  minute: "2-digit",
                })}
              </span>
            </p>
          </div>

          <div className="flex items-center gap-2">
            <span
              className={cn(
                "px-3 py-1 rounded-md text-xs font-bold uppercase tracking-wider border",
                tx.status === "completed"
                  ? "bg-success/10 text-success border-success/20"
                  : tx.status === "cancelled"
                  ? "bg-destructive/10 text-destructive border-destructive/20"
                  : "bg-warning/10 text-warning border-warning/20"
              )}
            >
              {tx.status}
            </span>
            <span
              className={cn(
                "px-3 py-1 rounded-md text-xs font-bold uppercase tracking-wider border",
                tx.payment_status === "paid"
                  ? "bg-success/10 text-success border-success/20"
                  : "bg-warning/10 text-warning border-warning/20"
              )}
            >
              {tx.payment_status}
            </span>
          </div>
        </div>

        {/* Stepper Status (Only show if not cancelled) */}
        {tx.status !== "cancelled" ? (
          <div className="space-y-4">
            <h3 className="text-sm font-extrabold text-foreground tracking-tight">{t("statusHistory")}</h3>
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4 relative">
              {steps.map((step, idx) => {
                const isCompleted = idx <= currentStepIdx;
                const isCurrent = idx === currentStepIdx;

                return (
                  <div
                    key={step.label}
                    className={cn(
                      "flex md:flex-col items-center md:items-start gap-3 p-3 rounded-lg border",
                      isCurrent
                        ? "bg-primary/5 border-primary/30"
                        : isCompleted
                        ? "bg-muted/30 border-border"
                        : "bg-transparent border-dashed border-border"
                    )}
                  >
                    <div className="shrink-0">
                      {isCompleted ? (
                        <CheckCircle2 className="h-5 w-5 text-primary" />
                      ) : (
                        <Circle className="h-5 w-5 text-muted-foreground" />
                      )}
                    </div>
                    <div className="min-w-0">
                      <p
                        className={cn(
                          "text-xs font-extrabold",
                          isCompleted ? "text-foreground" : "text-muted-foreground"
                        )}
                      >
                        {step.label}
                      </p>
                      <p className="text-[10px] text-muted-foreground truncate font-medium">{step.desc}</p>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        ) : (
          <div className="bg-destructive/5 border border-destructive/20 rounded-lg p-4 flex items-center gap-3">
            <CheckCircle2 className="h-5 w-5 text-destructive shrink-0" />
            <div className="space-y-0.5">
              <p className="text-xs font-extrabold text-destructive">Transaksi Dibatalkan</p>
              <p className="text-[10px] text-destructive/80 font-medium">Transaksi ini telah dibatalkan oleh pembeli.</p>
            </div>
          </div>
        )}
      </div>

      {/* Grid Content */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
        {/* Left column (Order Details, Shipping) */}
        <div className="lg:col-span-2 space-y-6">
          {/* Product Items */}
          <div className="bg-card rounded-xl border border-border p-6 shadow-xs space-y-4">
            <h3 className="text-sm font-extrabold text-foreground tracking-tight flex items-center gap-2 border-b border-border pb-3">
              <ClipboardList className="h-4.5 w-4.5 text-primary" />
              Detail Produk
            </h3>

            <div className="flex items-start gap-4 py-2">
              <div className="bg-primary/10 h-12 w-12 rounded-lg flex items-center justify-center shrink-0">
                <ShoppingBag className="h-6 w-6 text-primary" />
              </div>
              <div className="min-w-0 flex-1 space-y-1">
                <h4 className="font-extrabold text-sm text-foreground">{tx.product_name}</h4>
                <p className="text-xs text-muted-foreground font-semibold">
                  {t("supplier")}: <span className="text-foreground">{tx.supplier_name}</span>
                </p>
                <p className="text-xs text-muted-foreground font-medium">
                  {tx.quantity_value} {tx.quantity_unit} x {formatCurrency(tx.price_per_unit)}
                </p>
              </div>
              <div className="text-right shrink-0">
                <p className="text-sm font-extrabold text-foreground">{formatCurrency(tx.total_amount)}</p>
              </div>
            </div>
          </div>

          {/* Logistics & Delivery Address */}
          <div className="bg-card rounded-xl border border-border p-6 shadow-xs space-y-4">
            <h3 className="text-sm font-extrabold text-foreground tracking-tight flex items-center gap-2 border-b border-border pb-3">
              <Truck className="h-4.5 w-4.5 text-primary" />
              Informasi Pengiriman
            </h3>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 text-sm font-medium">
              <div className="space-y-1.5">
                <p className="text-xs text-muted-foreground font-semibold">{t("deliveryAddress")}</p>
                <p className="text-foreground leading-relaxed font-semibold">{tx.delivery_address}</p>
              </div>

              <div className="space-y-1.5">
                <p className="text-xs text-muted-foreground font-semibold">{t("notes")}</p>
                <p className="text-foreground italic leading-relaxed font-semibold">
                  {tx.notes ? `"${tx.notes}"` : "-"}
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* Right column (Summary, Supplier Profile card) */}
        <div className="space-y-6">
          {/* Payment Summary */}
          <div className="bg-card rounded-xl border border-border p-6 shadow-xs space-y-4">
            <h3 className="text-sm font-extrabold text-foreground tracking-tight flex items-center gap-2 border-b border-border pb-3">
              <Wallet className="h-4.5 w-4.5 text-primary" />
              Ringkasan Belanja
            </h3>

            <div className="space-y-3 text-sm font-medium">
              <div className="flex items-center justify-between text-muted-foreground">
                <span>Subtotal Produk</span>
                <span>{formatCurrency(tx.total_amount)}</span>
              </div>
              <div className="flex items-center justify-between text-muted-foreground">
                <span>Pajak & Biaya B2B (0%)</span>
                <span>Rp 0</span>
              </div>
              <div className="flex items-center justify-between text-muted-foreground">
                <span>Metode Pembayaran</span>
                <span className="text-foreground font-semibold">Term of Payment</span>
              </div>
              <div className="border-t border-border pt-3 flex items-center justify-between text-foreground">
                <span className="font-extrabold">Total Pembayaran</span>
                <span className="text-base font-extrabold text-primary">{formatCurrency(tx.total_amount)}</span>
              </div>
            </div>

            {/* Stepper CTAs */}
            <div className="pt-2 flex flex-col gap-2">
              <Link href="/chat">
                <Button className="w-full h-10 text-xs font-bold gap-2 cursor-pointer bg-primary text-primary-foreground hover:-translate-y-0.5 active:translate-y-0">
                  <MessageSquare className="h-4 w-4" />
                  Hubungi Supplier
                </Button>
              </Link>
              {tx.status === "completed" && (
                <Link href="/reviews">
                  <Button variant="outline" className="w-full h-10 text-xs font-bold gap-2 cursor-pointer hover:-translate-y-0.5 active:translate-y-0">
                    <Star className="h-4 w-4 text-warning fill-warning" />
                    Beri Ulasan Supplier
                  </Button>
                </Link>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
