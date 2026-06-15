"use client";

import React, { useState } from "react";
import { BuyerLayout } from "../../components/buyer-layout";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Link } from "@/i18n/routing";
import { CenteredLoading } from "@/components/loading";
import { ArrowLeft, Send } from "lucide-react";
import { useBuyerSupportTicketDetail } from "../hooks/useBuyerSupport";

interface BuyerSupportDetailPageProps {
  readonly id: string;
}

export function BuyerSupportDetailPage({ id }: BuyerSupportDetailPageProps) {
  const [replyText, setReplyText] = useState("");
  const { ticket, isLoading, replyTicket, isReplying } = useBuyerSupportTicketDetail(id);

  const handleSendReply = (e: React.FormEvent) => {
    e.preventDefault();
    if (!replyText.trim()) return;

    replyTicket(replyText.trim());
    setReplyText("");
  };

  if (isLoading) {
    return (
      <BuyerLayout>
        <CenteredLoading />
      </BuyerLayout>
    );
  }

  if (!ticket) {
    return (
      <BuyerLayout>
        <div className="text-center py-20 bg-card rounded-xl border border-border">
          <p className="text-destructive font-semibold">Tiket tidak ditemukan.</p>
        </div>
      </BuyerLayout>
    );
  }

  return (
    <BuyerLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="space-y-2">
          <Link href="/support" className="inline-flex items-center gap-1 text-xs font-semibold text-primary hover:underline cursor-pointer">
            <ArrowLeft className="h-3.5 w-3.5" /> Kembali ke Pusat Bantuan
          </Link>
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
            <div>
              <div className="flex items-center gap-2 flex-wrap">
                <span className="text-xs font-bold text-muted-foreground">{ticket.id}</span>
                <Badge
                  variant="outline"
                  className={
                    ticket.status === "Open"
                      ? "bg-primary/10 text-primary border-primary/20 rounded-full text-[10px]"
                      : "bg-muted text-muted-foreground border-border rounded-full text-[10px]"
                  }
                >
                  {ticket.status}
                </Badge>
              </div>
              <h1 className="text-xl font-bold tracking-tight text-foreground font-heading mt-1">{ticket.subject}</h1>
            </div>
          </div>
        </div>

        {/* Messages Card */}
        <Card className="border border-border rounded-xl bg-card shadow-xs overflow-hidden">
          <CardContent className="p-0 flex flex-col h-[480px]">
            {/* Messages Area */}
            <div className="flex-1 overflow-y-auto p-5 space-y-4 bg-muted/5">
              {ticket.messages.length === 0 ? (
                <p className="text-sm text-muted-foreground text-center py-10">Belum ada percakapan dalam tiket ini.</p>
              ) : (
                ticket.messages.map((msg, idx) => (
                  <div
                    key={idx}
                    className={`flex flex-col max-w-[80%] space-y-1 ${
                      msg.sender === "buyer" ? "ml-auto items-end" : "mr-auto items-start"
                    }`}
                  >
                    <span className="text-[10px] text-muted-foreground font-semibold">{msg.name}</span>
                    <div
                      className={`p-3 text-sm rounded-lg leading-relaxed ${
                        msg.sender === "buyer"
                          ? "bg-primary text-primary-foreground rounded-br-none"
                          : "bg-secondary text-foreground rounded-bl-none border border-border"
                      }`}
                    >
                      {msg.text}
                    </div>
                    <span className="text-[9px] text-muted-foreground">{msg.time}</span>
                  </div>
                ))
              )}
            </div>

            {/* Input Box */}
            <form onSubmit={handleSendReply} className="p-4 border-t border-border flex gap-3 bg-card">
              <input
                type="text"
                value={replyText}
                onChange={(e) => setReplyText(e.target.value)}
                placeholder="Ketik balasan pesan Anda..."
                disabled={isReplying || ticket.status === "Closed"}
                className="flex-1 border border-border rounded-lg px-3 py-2 text-sm outline-hidden focus:border-primary transition-colors cursor-pointer disabled:bg-muted disabled:text-muted-foreground disabled:cursor-not-allowed"
              />
              <Button type="submit" disabled={isReplying || ticket.status === "Closed"} className="bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer px-4">
                <Send className="h-4 w-4" />
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </BuyerLayout>
  );
}
