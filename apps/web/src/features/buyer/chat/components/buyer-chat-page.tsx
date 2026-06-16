"use client";

import React, { useState, useEffect, useRef, useMemo } from "react";
import { useTranslations } from "next-intl";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Send, Check, MessageSquare, ShieldCheck, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";

interface ChatPartner {
  id: string;
  name: string;
  avatar: string;
  verified: boolean;
  lastMessage: string;
  lastMessageTime: string;
}

interface Message {
  id: string;
  sender: "buyer" | "supplier";
  text: string;
  time: string;
}

export function BuyerChatPage() {
  const t = useTranslations("buyer.chat");
  const [selectedPartnerId, setSelectedPartnerId] = useState<string>("partner-1");
  const [inputText, setInputText] = useState<string>("");
  const [isTyping, setIsTyping] = useState<boolean>(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const partners: ChatPartner[] = [
    {
      id: "partner-1",
      name: "PT Baja Sentosa",
      avatar: "https://api.dicebear.com/7.x/initials/svg?seed=PT%20Baja%20Sentosa",
      verified: true,
      lastMessage: "Harga penawaran kami Rp 12.500 / Kg sudah include PPN.",
      lastMessageTime: "10:30",
    },
    {
      id: "partner-2",
      name: "CV Tekstil Nusantara",
      avatar: "https://api.dicebear.com/7.x/initials/svg?seed=CV%20Tekstil%20Nusantara",
      verified: true,
      lastMessage: "Gulungan denim akan dilapis plastik tebal tebal antiair.",
      lastMessageTime: "Kemarin",
    },
    {
      id: "partner-3",
      name: "PT Agro Indo Sejahtera",
      avatar: "https://api.dicebear.com/7.x/initials/svg?seed=PT%20Agro%20Indo",
      verified: false,
      lastMessage: "Sertifikat asal biji kopi gayo siap dilampirkan.",
      lastMessageTime: "11 Jun",
    },
  ];

  const [messageThreads, setMessageThreads] = useState<Record<string, Message[]>>({
    "partner-1": [
      { id: "m1", sender: "buyer", text: "Halo, saya tertarik dengan produk Reinforced Steel Bar D10 Anda.", time: "10:00" },
      { id: "m2", sender: "supplier", text: "Halo! Terima kasih telah menghubungi PT Baja Sentosa. Berapa banyak kebutuhan Anda?", time: "10:05" },
      { id: "m3", sender: "buyer", text: "Kebutuhan kami sekitar 20 Ton untuk proyek di Cibitung.", time: "10:15" },
      { id: "m4", sender: "supplier", text: "Baik, harga penawaran kami Rp 12.500 / Kg sudah include PPN. Untuk MOQ 20 Ton, kami bisa langsung kirim minggu depan.", time: "10:30" },
    ],
    "partner-2": [
      { id: "m5", sender: "buyer", text: "Apakah ada biaya tambahan untuk kemasan plastik tebal antiair?", time: "Kemarin" },
      { id: "m6", sender: "supplier", text: "Tidak ada biaya tambahan. Gulungan denim akan dilapis plastik tebal tebal antiair.", time: "Kemarin" },
    ],
    "partner-3": [
      { id: "m7", sender: "buyer", text: "Bisa kirimkan contoh sertifikat asal kopi?", time: "11 Jun" },
      { id: "m8", sender: "supplier", text: "Bisa, sertifikat asal biji kopi gayo siap dilampirkan saat pengiriman.", time: "11 Jun" },
    ],
  });

  const selectedPartner = partners.find((p) => p.id === selectedPartnerId);
  const currentMessages = useMemo(() => messageThreads[selectedPartnerId] || [], [messageThreads, selectedPartnerId]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    scrollToBottom();
  }, [currentMessages, isTyping]);

  const handleSend = () => {
    if (!inputText.trim()) return;

    const newMsg: Message = {
      id: "m-" + Math.random().toString(36).substring(2, 9),
      sender: "buyer",
      text: inputText,
      time: new Date().toLocaleTimeString("id-ID", { hour: "2-digit", minute: "2-digit" }),
    };

    setMessageThreads((prev) => ({
      ...prev,
      [selectedPartnerId]: [...(prev[selectedPartnerId] || []), newMsg],
    }));

    setInputText("");
    setIsTyping(true);

    // Simulate reply after 1.5 seconds
    setTimeout(() => {
      setIsTyping(false);
      const replyMsg: Message = {
        id: "m-" + Math.random().toString(36).substring(2, 9),
        sender: "supplier",
        text: `Terima kasih atas pesannya. Tim B2B kami dari ${selectedPartner?.name} akan meninjau pesan ini dan segera menghubungi Anda kembali dengan penawaran khusus.`,
        time: new Date().toLocaleTimeString("id-ID", { hour: "2-digit", minute: "2-digit" }),
      };

      setMessageThreads((prev) => ({
        ...prev,
        [selectedPartnerId]: [...(prev[selectedPartnerId] || []), replyMsg],
      }));
    }, 1500);
  };

  return (
    <div className="space-y-6">
      {/* Page Title */}
      <div className="flex flex-col gap-1">
        <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">{t("title")}</h1>
        <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
      </div>

      {/* Chat Container Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 bg-card border border-border rounded-xl overflow-hidden min-h-[580px] shadow-xs">
        {/* Left Side: Partner List */}
        <div className="border-r border-border flex flex-col h-full bg-muted/10">
          <div className="p-4 border-b border-border bg-card">
            <p className="text-xs font-extrabold text-foreground tracking-wider uppercase">NEGOSIASI AKTIF</p>
          </div>
          <div className="flex-1 overflow-y-auto divide-y divide-border">
            {partners.map((partner) => {
              const isSelected = partner.id === selectedPartnerId;
              return (
                <div
                  key={partner.id}
                  onClick={() => setSelectedPartnerId(partner.id)}
                  className={cn(
                    "p-4 flex items-start gap-3 transition-colors duration-200 cursor-pointer",
                    isSelected ? "bg-primary/10" : "bg-card hover:bg-muted/30"
                  )}
                >
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img
                    src={partner.avatar}
                    alt={partner.name}
                    className="h-10 w-10 rounded-lg shrink-0 border border-border"
                  />
                  <div className="min-w-0 flex-1 space-y-1">
                    <div className="flex items-center justify-between gap-2">
                      <span className="font-extrabold text-xs text-foreground flex items-center gap-1">
                        {partner.name}
                        {partner.verified && <ShieldCheck className="h-3.5 w-3.5 text-primary fill-primary/10 shrink-0" />}
                      </span>
                      <span className="text-[10px] text-muted-foreground font-semibold shrink-0">{partner.lastMessageTime}</span>
                    </div>
                    <p className="text-[11px] text-muted-foreground truncate font-semibold leading-relaxed">
                      {partner.lastMessage}
                    </p>
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Right Side: Message Room */}
        <div className="md:col-span-2 flex flex-col h-full bg-card">
          {selectedPartner ? (
            <>
              {/* Partner Header */}
              <div className="p-4 border-b border-border flex items-center gap-3 shrink-0">
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img
                    src={selectedPartner.avatar}
                    alt={selectedPartner.name}
                    className="h-10 w-10 rounded-lg shrink-0 border border-border"
                  />
                <div>
                  <h3 className="text-xs font-extrabold text-foreground flex items-center gap-1">
                    {selectedPartner.name}
                    {selectedPartner.verified && (
                      <ShieldCheck className="h-4 w-4 text-primary fill-primary/10 shrink-0" />
                    )}
                  </h3>
                  <span className="inline-flex items-center gap-1 mt-0.5 text-[10px] text-muted-foreground font-semibold">
                    <span className="h-1.5 w-1.5 rounded-full bg-success" />
                    Online
                  </span>
                </div>
              </div>

              {/* Chat Thread */}
              <div className="flex-1 overflow-y-auto p-4 space-y-4 max-h-[400px]">
                {currentMessages.map((msg) => {
                  const isMe = msg.sender === "buyer";
                  return (
                    <div
                      key={msg.id}
                      className={cn("flex items-end gap-2", isMe ? "justify-end" : "justify-start")}
                    >
                      <div className="max-w-[75%] space-y-1">
                        <div
                          className={cn(
                            "p-3 text-sm font-semibold rounded-lg leading-relaxed shadow-2xs border",
                            isMe
                              ? "bg-primary text-primary-foreground border-primary/20"
                              : "bg-muted text-foreground border-border"
                          )}
                        >
                          {msg.text}
                        </div>
                        <div className="flex items-center justify-end gap-1 text-[10px] text-muted-foreground font-semibold">
                          <span>{msg.time}</span>
                          {isMe && <Check className="h-3 w-3 text-primary" />}
                        </div>
                      </div>
                    </div>
                  );
                })}

                {/* Typings simulation indicator */}
                {isTyping && (
                  <div className="flex items-center gap-2 text-xs text-muted-foreground font-semibold bg-muted/40 p-2 rounded-lg max-w-[200px] border border-border">
                    <Loader2 className="h-3.5 w-3.5 animate-spin text-primary" />
                    <span>{selectedPartner.name} sedang mengetik...</span>
                  </div>
                )}
                <div ref={messagesEndRef} />
              </div>

              {/* Chat Input Field */}
              <div className="p-4 border-t border-border bg-card shrink-0">
                <form
                  onSubmit={(e) => {
                    e.preventDefault();
                    handleSend();
                  }}
                  className="flex items-center gap-2"
                >
                  <Input
                    type="text"
                    value={inputText}
                    onChange={(e) => setInputText(e.target.value)}
                    placeholder={t("placeholder")}
                    className="flex-1 h-10 rounded-lg border-border focus-visible:ring-primary focus-visible:border-primary text-sm font-medium"
                  />
                  <Button type="submit" size="icon" className="h-10 w-10 shrink-0 cursor-pointer bg-primary text-primary-foreground hover:-translate-y-0.5 active:translate-y-0">
                    <Send className="h-4.5 w-4.5" />
                  </Button>
                </form>
              </div>
            </>
          ) : (
            <div className="flex-1 flex flex-col items-center justify-center text-center p-8 gap-3">
              <MessageSquare className="h-8 w-8 text-muted-foreground" />
              <p className="text-sm text-muted-foreground">{t("empty")}</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
