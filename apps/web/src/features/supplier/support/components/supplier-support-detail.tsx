"use client";

import React, { useState, useEffect, useRef } from "react";
import { useRouter } from "@/i18n/routing";
import { useTranslations } from "next-intl";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { ArrowLeft, Send } from "lucide-react";

interface SupplierSupportDetailProps {
  id: string;
}

export function SupplierSupportDetail({ id }: SupplierSupportDetailProps) {
  const router = useRouter();
  const t = useTranslations("supplier.support");
  const scrollRef = useRef<HTMLDivElement>(null);

  const [messages, setMessages] = useState([
    { sender: "system", text: "Ticket opened automatically. Staf helpdesk IndoSupplier akan segera menghubungi Anda.", time: "2026-06-04 10:00" },
    { sender: "user", text: "Halo, saya sudah mengunggah semua dokumen NPWP dan NIB saya kemarin siang. Mengapa verifikasi saya masih pending?", time: "2026-06-04 10:05" },
    { sender: "admin", text: "Halo PT Nusantara Supplier Utama. Berkas NIB Anda sedang berada di antrean evaluasi manual. Proses ini membutuhkan waktu maksimal 3 hari kerja. Mohon ditunggu terlebih dahulu ya.", time: "2026-06-04 11:30" },
  ]);

  const [text, setText] = useState("");

  const handleSend = (e: React.FormEvent) => {
    e.preventDefault();
    if (!text.trim()) return;

    const userMsg = {
      sender: "user",
      text: text,
      time: "Just now",
    };

    setMessages((prev) => [...prev, userMsg]);
    setText("");

    // Simulate Admin Response after 1.5 seconds
    setTimeout(() => {
      const adminMsg = {
        sender: "admin",
        text: "Terima kasih atas tanggapan Anda. Kami telah menandai tiket ini agar diprioritaskan oleh tim kepatuhan kami. Kami akan mengabari Anda begitu verifikasi selesai.",
        time: "Just now",
      };
      setMessages((prev) => [...prev, adminMsg]);
    }, 1500);
  };

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [messages]);

  return (
    <div className="space-y-6 text-left">
      {/* Header */}
      <div className="flex items-center gap-4 border-b border-border/80 pb-6">
        <Button
          variant="outline"
          size="icon"
          onClick={() => router.push("/supplier/support")}
          className="h-9 w-9 cursor-pointer border-border"
        >
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
            {t("chatTitle")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("chatSubtitle")} (ID: {id})
          </p>
        </div>
      </div>

      {/* Chat workspace */}
      <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card flex flex-col h-[500px]">
        {/* Messages Body */}
        <div className="flex-1 p-6 overflow-y-auto space-y-4">
          {messages.map((m, idx) => {
            return (
              <div
                key={idx}
                className={`flex gap-3 max-w-[80%] ${
                  m.sender === "user" ? "ml-auto flex-row-reverse" : "mr-auto"
                }`}
              >
                {/* Avatar */}
                <div
                  className={`h-8 w-8 rounded-full flex items-center justify-center shrink-0 text-xs font-bold ${
                    m.sender === "user"
                      ? "bg-primary text-primary-foreground"
                      : m.sender === "system"
                      ? "bg-muted text-muted-foreground"
                      : "bg-success/15 text-success border border-success/30"
                  }`}
                >
                  {m.sender === "user" ? "U" : m.sender === "system" ? "S" : "A"}
                </div>

                {/* Bubble */}
                <div className="space-y-1">
                  <div
                    className={`p-3 rounded-2xl text-sm font-semibold leading-relaxed ${
                      m.sender === "user"
                        ? "bg-primary text-primary-foreground rounded-tr-none text-right"
                        : m.sender === "system"
                        ? "bg-muted/40 text-muted-foreground rounded-tl-none border border-border"
                        : "bg-secondary text-foreground rounded-tl-none border border-border"
                    }`}
                  >
                    {m.text}
                  </div>
                  <p className={`text-[9px] text-muted-foreground font-bold ${m.sender === "user" ? "text-right" : ""}`}>
                    {m.time}
                  </p>
                </div>
              </div>
            );
          })}
          <div ref={scrollRef} />
        </div>

        {/* Input Bar */}
        <div className="p-4 border-t border-border bg-muted/10">
          <form onSubmit={handleSend} className="flex gap-2">
            <input
              type="text"
              placeholder={t("placeholder")}
              value={text}
              onChange={(e) => setText(e.target.value)}
              className="flex-1 px-4 py-2 bg-card border border-border text-sm rounded-lg outline-hidden focus:border-primary transition-all text-left"
            />
            <Button type="submit" size="icon" className="h-9 w-9 cursor-pointer shrink-0">
              <Send className="h-4 w-4" />
            </Button>
          </form>
        </div>
      </Card>
    </div>
  );
}
