"use client";

import React, { useState } from "react";
import { useRouter } from "@/i18n/routing";
import { useTranslations } from "next-intl";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { NumericInput } from "@/components/ui/numeric-input";
import { Textarea } from "@/components/ui/textarea";
import { Field, FieldLabel, FieldGroup } from "@/components/ui/field";
import {
  ArrowLeft,
  Send,
  CheckCircle2,
  Star,
  Tag,
  Clock,
} from "lucide-react";
import { CenteredLoading } from "@/components/loading";
import {
  useSupplierRfqDetail,
  useSupplierRfqThread,
  useSendSupplierRfqMessage,
} from "../hooks/useSupplierRfqs";
import type { RFQMessageItem } from "../types/rfq.types";

interface SupplierRfqDetailProps {
  readonly id: string;
}

export function SupplierRfqDetail({ id }: SupplierRfqDetailProps) {
  const router = useRouter();
  const t = useTranslations("supplier.rfq");

  const { data: rfq, isLoading: isLoadingRfq } = useSupplierRfqDetail(id);
  const { data: messages = [], isLoading: isLoadingThread } = useSupplierRfqThread(id);
  const { mutate: sendMessage, isPending: isSending } = useSendSupplierRfqMessage(id);

  const [includeOffer, setIncludeOffer] = useState(false);
  const [messageBody, setMessageBody] = useState("");
  const [priceValue, setPriceValue] = useState<number | undefined>(undefined);
  const [moq, setMoq] = useState("");
  const [deliveryTime, setDeliveryTime] = useState("");

  const unit = rfq?.quantity ? rfq.quantity.split(" ").slice(1).join(" ") : "";
  const unitSuffix = unit ? ` / ${unit}` : "";

  const handleSendMessage = (e: React.FormEvent) => {
    e.preventDefault();
    if (!messageBody.trim() || isSending) return;

    const payload: {
      body: string;
      price?: string;
      moq?: string;
      deliveryTime?: string;
    } = {
      body: messageBody.trim(),
    };

    if (includeOffer && priceValue) {
      payload.price = `Rp ${priceValue.toLocaleString("id-ID")}${unitSuffix}`;
      payload.moq = moq.trim() || undefined;
      payload.deliveryTime = deliveryTime.trim() || undefined;
    }

    sendMessage(payload, {
      onSuccess: () => {
        setMessageBody("");
        if (priceValue) {
          setIncludeOffer(false);
        }
      },
    });
  };

  if (isLoadingRfq || isLoadingThread) {
    return <CenteredLoading />;
  }

  if (!rfq) {
    return (
      <div className="text-center py-20 bg-card rounded-lg border border-border">
        <p className="text-destructive font-semibold">RFQ tidak ditemukan.</p>
      </div>
    );
  }

  const isAccepted =
    rfq.status === "accepted" ||
    messages.some((m) => m.messageType === "bid_accepted");

  return (
    <div className="space-y-8 text-left">
      {/* Top Header */}
      <div className="space-y-2">
        <button
          type="button"
          onClick={() => router.push("/supplier/rfq")}
          className="inline-flex items-center gap-1.5 text-xs font-semibold text-primary hover:underline cursor-pointer"
        >
          <ArrowLeft className="h-4 w-4" /> Kembali ke Daftar RFQ
        </button>
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Badge
                variant="outline"
                className="text-xs font-semibold px-2.5 py-0.5 rounded-lg"
              >
                {isAccepted ? (
                  <span className="text-success font-bold">✓ Diterima</span>
                ) : (
                  <span className="text-muted-foreground">
                    {rfq.status === "responded" ? "Direspon" : "Terbuka"}
                  </span>
                )}
              </Badge>
            </div>
            <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-foreground font-heading">
              {rfq.product}
            </h1>
          </div>
        </div>
      </div>

      {/* Messages Thread Timeline */}
      <div className="space-y-6">
        {messages.length === 0 ? (
          <div className="text-center py-16 bg-card rounded-lg border border-border space-y-3">
            <Clock className="w-8 h-8 text-muted-foreground/60 mx-auto" />
            <p className="text-sm font-semibold text-foreground">
              Belum ada penawaran atau percakapan yang dikirim.
            </p>
          </div>
        ) : (
          messages.map((m: RFQMessageItem) => {
            const isOffer = m.messageType === "offer" || !!m.priceFormatted;
            const isAcceptedBanner = m.messageType === "bid_accepted";

            return (
              <div
                key={m.id}
                className="rounded-lg border border-border bg-card p-6 space-y-4"
              >
                {/* Header: Avatar, Name, Role, Rating, Timestamp */}
                <div className="flex items-center justify-between gap-4 flex-wrap pb-4 border-b border-border/60">
                  <div className="flex items-center gap-3 min-w-0">
                    <Avatar className="h-10 w-10 rounded-lg border border-border bg-muted/40 shrink-0">
                      <AvatarImage
                        src={m.senderAvatar}
                        alt={m.senderName}
                        className="object-cover"
                      />
                      <AvatarFallback className="font-bold text-xs bg-primary/10 text-primary">
                        {m.senderName ? m.senderName.slice(0, 2).toUpperCase() : "US"}
                      </AvatarFallback>
                    </Avatar>

                    <div className="min-w-0 space-y-0.5">
                      <div className="flex items-center gap-2 flex-wrap">
                        <span className="text-sm font-bold text-foreground truncate">
                          {m.senderName}
                        </span>
                        <span className="text-xs text-muted-foreground">•</span>
                        <span className="text-xs font-medium text-muted-foreground">
                          {m.senderType === "supplier"
                            ? m.isMine
                              ? "Anda (Supplier)"
                              : "Supplier"
                            : m.senderType === "buyer"
                            ? "Pembeli"
                            : "Sistem"}
                        </span>
                        {m.senderRating && m.senderType === "supplier" && (
                          <>
                            <span className="text-xs text-muted-foreground">•</span>
                            <span className="flex items-center gap-1 text-xs font-semibold text-amber-500">
                              <Star className="w-3 h-3 fill-amber-400 text-amber-400" />
                              {m.senderRating.toFixed(1)}
                            </span>
                          </>
                        )}
                      </div>
                    </div>
                  </div>

                  <span className="text-xs text-muted-foreground shrink-0">
                    {m.createdAtFormatted || m.createdAt}
                  </span>
                </div>

                {/* Message Body */}
                {m.body && (
                  <div className="text-sm text-foreground/90 leading-relaxed whitespace-pre-wrap">
                    {m.body}
                  </div>
                )}

                {/* Offer Callout Box */}
                {isOffer && (
                  <div className="rounded-lg border border-primary/30 bg-primary/5 p-4 space-y-3">
                    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                      <div className="space-y-1">
                        <span className="text-xs font-medium text-muted-foreground block">
                          Penawaran Harga
                        </span>
                        <p className="text-lg font-bold text-primary tracking-tight">
                          {m.priceFormatted}
                        </p>
                      </div>

                      {isAccepted ? (
                        <Badge
                          variant="outline"
                          className="bg-success/15 text-success border-success/30 px-3 py-1 text-xs font-bold flex items-center gap-1.5 self-start sm:self-auto"
                        >
                          <CheckCircle2 className="w-3.5 h-3.5" /> Penawaran Diterima
                        </Badge>
                      ) : (
                        <Badge
                          variant="outline"
                          className="bg-muted text-muted-foreground border-border px-3 py-1 text-xs font-medium self-start sm:self-auto"
                        >
                          Menunggu Keputusan Pembeli
                        </Badge>
                      )}
                    </div>

                    {(m.moq || m.deliveryTime) && (
                      <div className="flex items-center gap-6 text-xs text-muted-foreground pt-3 border-t border-border/60">
                        {m.moq && (
                          <span>
                            MOQ: <strong className="text-foreground">{m.moq}</strong>
                          </span>
                        )}
                        {m.deliveryTime && (
                          <span>
                            Estimasi Pengiriman:{" "}
                            <strong className="text-foreground">{m.deliveryTime}</strong>
                          </span>
                        )}
                      </div>
                    )}
                  </div>
                )}

                {/* Accepted Notification Banner */}
                {isAcceptedBanner && (
                  <div className="rounded-lg border border-success/30 bg-success/10 p-4 text-xs font-medium text-success flex items-center gap-2">
                    <CheckCircle2 className="w-4 h-4 shrink-0" />
                    <span>Penawaran Anda telah disetujui oleh Pembeli.</span>
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>

      {/* Reply & Offer Form */}
      <div id="supplier-reply-section" className="rounded-lg border border-border bg-card p-6 space-y-4">
        <div className="flex items-center justify-between flex-wrap gap-2">
          <h3 className="text-sm font-bold text-foreground font-heading uppercase tracking-wide">
            Kirim Balasan
          </h3>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => setIncludeOffer(!includeOffer)}
            className="text-xs font-semibold cursor-pointer h-8"
          >
            <Tag className="w-3.5 h-3.5 mr-1.5 text-primary" />
            {includeOffer
              ? "Tutup Form Penawaran Harga"
              : "Sertakan / Perbarui Penawaran Harga"}
          </Button>
        </div>

        <form onSubmit={handleSendMessage} className="space-y-4">
          <Textarea
            value={messageBody}
            onChange={(e) => setMessageBody(e.target.value)}
            placeholder="Tulis pesan respon teknis atau penawaran ke pembeli..."
            rows={4}
            required
            className="resize-none text-sm border-input focus:border-primary rounded-lg bg-background p-4"
          />

          {includeOffer && (
            <div className="p-4 rounded-lg border border-border bg-muted/40 space-y-4 animate-in fade-in duration-200">
              <span className="text-xs font-bold text-foreground block">
                Detail Penawaran Harga
              </span>

              <FieldGroup className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <Field className="space-y-2">
                  <FieldLabel className="text-xs">
                    {t("formPrice")}
                  </FieldLabel>
                  <NumericInput
                    placeholder="Contoh: 12.000.000"
                    value={priceValue}
                    onChange={(val) => setPriceValue(val)}
                  />
                </Field>
                <Field className="space-y-2">
                  <FieldLabel className="text-xs">
                    {t("formMinQty")}
                  </FieldLabel>
                  <Input
                    placeholder="Contoh: 10 Ton"
                    value={moq}
                    onChange={(e) => setMoq(e.target.value)}
                    className="text-sm h-10"
                  />
                </Field>
                <Field className="space-y-2">
                  <FieldLabel className="text-xs">
                    {t("formDeliveryTime")}
                  </FieldLabel>
                  <Input
                    placeholder="Contoh: 14 Hari Kerja"
                    value={deliveryTime}
                    onChange={(e) => setDeliveryTime(e.target.value)}
                    className="text-sm h-10"
                  />
                </Field>
              </FieldGroup>
            </div>
          )}

          <div className="flex justify-end">
            <Button
              type="submit"
              disabled={!messageBody.trim() || isSending}
              className="cursor-pointer font-medium hover:-translate-y-0.5 active:translate-y-0 transition-all"
            >
              <Send className="w-4 h-4 mr-2" />
              {isSending
                ? "Mengirim..."
                : includeOffer && priceValue
                ? "Kirim Balasan & Penawaran"
                : "Kirim Balasan"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
