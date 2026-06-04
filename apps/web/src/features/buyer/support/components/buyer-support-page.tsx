"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { BuyerLayout } from "../../components/buyer-layout";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Link } from "@/i18n/routing";
import { Headset, Plus, ChevronRight, Calendar } from "lucide-react";

export function BuyerSupportPage() {
  const t = useTranslations("buyer.support");

  const [tickets] = useState([
    { id: "TK-2026-001", subject: "Pertanyaan Pengajuan Limit Kredit Sourcing", date: "2026-06-02", status: "Open" },
    { id: "TK-2026-002", subject: "Verifikasi Berkas SIUP Lambat", date: "2026-05-20", status: "Closed" },
  ]);

  return (
    <BuyerLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="space-y-1">
            <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">{t("title")}</h1>
            <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
          </div>
          <Button
            onClick={() => alert("Fitur tiket baru akan segera hadir!")}
            className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/95 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/20"
          >
            <Plus className="mr-2 h-4 w-4" />
            {t("btnCreate")}
          </Button>
        </div>

        {/* Tickets Card */}
        <Card className="border border-border rounded-xl bg-card shadow-xs overflow-hidden">
          <CardContent className="p-0">
            {tickets.length === 0 ? (
              <div className="text-center py-16">
                <Headset className="mx-auto h-12 w-12 text-muted-foreground opacity-45" />
                <h3 className="mt-4 text-sm font-semibold text-foreground">Tidak ada tiket bantuan</h3>
                <p className="mt-2 text-xs text-muted-foreground max-w-xs mx-auto">
                  Jika Anda mengalami masalah, silakan buat tiket baru di atas.
                </p>
              </div>
            ) : (
              <div className="divide-y divide-border">
                {tickets.map((ticket) => (
                  <div key={ticket.id} className="p-5 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 hover:bg-secondary/10 transition-colors">
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <span className="text-xs font-bold text-muted-foreground">{ticket.id}</span>
                        <h4 className="text-sm font-bold text-foreground hover:text-primary transition-colors cursor-pointer">
                          <Link href={`/support/${ticket.id}`}>{ticket.subject}</Link>
                        </h4>
                      </div>
                      <div className="flex items-center gap-1 text-[11px] text-muted-foreground">
                        <Calendar className="h-3 w-3" />
                        <span>Dibuat pada {ticket.date}</span>
                      </div>
                    </div>
                    <div className="flex items-center justify-between sm:justify-end gap-3 shrink-0">
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
                      <Button asChild variant="ghost" size="sm" className="text-primary hover:bg-primary/5 cursor-pointer font-semibold gap-1 transition-all">
                        <Link href={`/support/${ticket.id}`}>
                          Detail <ChevronRight className="h-4 w-4" />
                        </Link>
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </BuyerLayout>
  );
}
