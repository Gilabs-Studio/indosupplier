"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { BuyerLayout } from "../../components/buyer-layout";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Link } from "@/i18n/routing";
import { Field, FieldLabel, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { CenteredLoading } from "@/components/loading";
import { Headset, Plus, ChevronRight, Calendar, X } from "lucide-react";
import { useBuyerSupportTickets } from "../hooks/useBuyerSupport";

export function BuyerSupportPage() {
  const t = useTranslations("buyer.support");
  const { tickets, isLoading, createTicket } = useBuyerSupportTickets();

  const [showCreateForm, setShowCreateForm] = useState(false);
  const [subject, setSubject] = useState("");
  const [message, setMessage] = useState("");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!subject.trim() || !message.trim()) return;

    createTicket({
      subject: subject.trim(),
      message: message.trim(),
    });

    setSubject("");
    setMessage("");
    setShowCreateForm(false);
  };

  if (isLoading) {
    return (
      <BuyerLayout>
        <CenteredLoading />
      </BuyerLayout>
    );
  }

  return (
    <BuyerLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="space-y-1">
            <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">{t("title")}</h1>
            <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
          </div>
          {!showCreateForm && (
            <Button
              onClick={() => setShowCreateForm(true)}
              className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/95 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/25"
            >
              <Plus className="mr-2 h-4 w-4" />
              {t("btnCreate")}
            </Button>
          )}
        </div>

        {/* Inline Create Ticket Form */}
        {showCreateForm && (
          <Card className="border border-border rounded-xl bg-card shadow-md overflow-hidden">
            <CardContent className="p-6 space-y-4">
              <div className="flex items-center justify-between border-b border-border pb-3">
                <h3 className="text-sm font-bold text-foreground">Buat Tiket Dukungan Baru</h3>
                <Button
                  onClick={() => setShowCreateForm(false)}
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7 text-muted-foreground hover:text-foreground rounded-full cursor-pointer"
                >
                  <X className="h-4 w-4" />
                </Button>
              </div>

              <form onSubmit={handleSubmit} className="space-y-4">
                <FieldGroup className="space-y-4">
                  <Field className="space-y-1">
                    <FieldLabel htmlFor="subject">Subjek Masalah</FieldLabel>
                    <Input
                      id="subject"
                      value={subject}
                      onChange={(e) => setSubject(e.target.value)}
                      placeholder="Contoh: Pertanyaan Limit Kredit"
                      required
                      className="cursor-pointer"
                    />
                  </Field>

                  <Field className="space-y-1">
                    <FieldLabel htmlFor="message">Detail Laporan</FieldLabel>
                    <textarea
                      id="message"
                      rows={4}
                      value={message}
                      onChange={(e) => setMessage(e.target.value)}
                      placeholder="Tuliskan secara lengkap detail masalah atau pertanyaan Anda..."
                      required
                      className="w-full rounded-lg border border-border bg-card px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-hidden"
                    />
                  </Field>
                </FieldGroup>

                <div className="flex justify-end gap-2 border-t border-border pt-4">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => setShowCreateForm(false)}
                    className="cursor-pointer shadow-xs"
                  >
                    Batal
                  </Button>
                  <Button
                    type="submit"
                    className="bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer px-5 shadow-lg hover:shadow-primary/20"
                  >
                    Kirim Tiket
                  </Button>
                </div>
              </form>
            </CardContent>
          </Card>
        )}

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
