"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import {
  Check,
  CheckCircle2,
  Send,
  Star,
} from "lucide-react";
import type { RFQThreadSupplier, RFQMessageItem } from "../types/rfq.types";
import {
  useBuyerRfqThreadMessages,
  useSendBuyerMessage,
  useAcceptBidInThread,
} from "../hooks/useBuyerRfqs";

interface BuyerRfqThreadProps {
  readonly rfqId: string;
  readonly supplier: RFQThreadSupplier;
  readonly isRfqCompleted: boolean;
}

export function BuyerRfqThread({
  rfqId,
  supplier,
  isRfqCompleted,
}: BuyerRfqThreadProps) {
  const t = useTranslations("buyerRfq.rfqDetail");
  const [replyText, setReplyText] = useState("");

  const { data: messages = [], isLoading } = useBuyerRfqThreadMessages(
    rfqId,
    supplier.supplierProfileId
  );
  const { mutate: sendMessage, isPending: isSending } = useSendBuyerMessage(
    rfqId,
    supplier.supplierProfileId
  );
  const { mutate: acceptBid, isPending: isAccepting } = useAcceptBidInThread(
    rfqId,
    supplier.supplierProfileId
  );

  const handleSend = (e: React.FormEvent) => {
    e.preventDefault();
    if (!replyText.trim() || isSending) return;
    sendMessage(
      { body: replyText.trim() },
      {
        onSuccess: () => {
          setReplyText("");
        },
      }
    );
  };

  const isSupplierAccepted =
    supplier.status === "accepted" ||
    messages.some((m) => m.messageType === "bid_accepted");

  return (
    <div className="space-y-8">
      {/* Messages Thread Timeline */}
      <div className="space-y-6">
        {isLoading ? (
          <div className="text-center py-8 text-xs text-muted-foreground">
            Memuat percakapan...
          </div>
        ) : messages.length === 0 ? (
          <div className="text-center py-8 text-xs text-muted-foreground">
            {t("emptyThreads")}
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
                            ? "Supplier"
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

                      {!isRfqCompleted && !isSupplierAccepted ? (
                        <Button
                          size="sm"
                          onClick={() => acceptBid()}
                          disabled={isAccepting}
                          className="cursor-pointer font-semibold shadow-xs hover:-translate-y-0.5 active:translate-y-0 transition-all self-start sm:self-auto"
                        >
                          <Check className="w-4 h-4 mr-1.5" />
                          Terima Penawaran
                        </Button>
                      ) : isSupplierAccepted ? (
                        <Badge
                          variant="outline"
                          className="bg-success/15 text-success border-success/30 px-3 py-1 text-xs font-bold flex items-center gap-1.5 self-start sm:self-auto"
                        >
                          <CheckCircle2 className="w-3.5 h-3.5" /> Penawaran Diterima
                        </Badge>
                      ) : null}
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

                {/* Bid Accepted Notification */}
                {isAcceptedBanner && (
                  <div className="rounded-lg border border-success/30 bg-success/10 p-4 text-xs font-medium text-success flex items-center gap-2">
                    <CheckCircle2 className="w-4 h-4 shrink-0" />
                    <span>Penawaran ini telah disetujui oleh Pembeli.</span>
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>

      {/* Reply Form */}
      <div id="buyer-thread-reply-box" className="rounded-lg border border-border bg-card p-6 space-y-4">
        <h3 className="text-sm font-bold text-foreground font-heading uppercase tracking-wide">
          Kirim Balasan
        </h3>

        <form onSubmit={handleSend} className="space-y-4">
          <Textarea
            value={replyText}
            onChange={(e) => setReplyText(e.target.value)}
            placeholder="Tulis balasan negosiasi harga atau pertanyaan spesifikasi teknis..."
            rows={4}
            className="resize-none text-sm border-input focus:border-primary rounded-lg bg-background p-4"
          />
          <div className="flex justify-end">
            <Button
              type="submit"
              disabled={!replyText.trim() || isSending}
              className="cursor-pointer font-medium hover:-translate-y-0.5 active:translate-y-0 transition-all"
            >
              <Send className="w-4 h-4 mr-2" />
              {isSending ? "Mengirim..." : "Kirim Balasan"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
