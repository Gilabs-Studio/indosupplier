"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { BuyerLayout } from "../../components/buyer-layout";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Bell, Mail, RefreshCw, AlertCircle } from "lucide-react";

export function BuyerNotificationsPage() {
  const t = useTranslations("buyer.notifications");

  const [notifications, setNotifications] = useState([
    {
      id: "1",
      title: "Penawaran Masuk Baru",
      desc: "PT Rempah Nusantara mengirimkan penawaran harga sebesar Rp 12.500 / Kg untuk RFQ Bentonite Clay Powder.",
      date: "2 Jam yang lalu",
      icon: RefreshCw,
      color: "bg-primary/10 text-primary",
      unread: true,
    },
    {
      id: "2",
      title: "Pesan Baru dari Supplier",
      desc: "CV Nusantara Garment membalas chat Anda tentang ukuran sampel uniform.",
      date: "5 Jam yang lalu",
      icon: Mail,
      color: "bg-success/10 text-success",
      unread: true,
    },
    {
      id: "3",
      title: "RFQ Berhasil Disiarkan",
      desc: "RFQ-2026-004 (Garnet Sand) Anda telah disetujui oleh admin dan disiarkan ke 12 supplier terdaftar.",
      date: "1 Hari yang lalu",
      icon: Bell,
      color: "bg-cyan/10 text-cyan",
      unread: false,
    },
    {
      id: "4",
      title: "Peringatan Dokumen Kedaluwarsa",
      desc: "Masa berlaku dokumen SIUP Anda akan berakhir dalam 30 hari. Segera perbarui untuk menghindari pemblokiran akun.",
      date: "3 Hari yang lalu",
      icon: AlertCircle,
      color: "bg-warning/10 text-warning",
      unread: false,
    },
  ]);

  const handleMarkAllRead = () => {
    setNotifications(notifications.map((n) => ({ ...n, unread: false })));
  };

  return (
    <BuyerLayout>
      <div className="space-y-6 max-w-3xl">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="space-y-1">
            <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">{t("title")}</h1>
            <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
          </div>
          {notifications.some((n) => n.unread) && (
            <Button
              onClick={handleMarkAllRead}
              variant="outline"
              size="sm"
              className="text-xs font-semibold cursor-pointer border-border hover:border-muted-foreground transition-all"
            >
              Tandai Semua Dibaca
            </Button>
          )}
        </div>

        {/* Notifications Card */}
        <Card className="border border-border rounded-xl bg-card shadow-xs overflow-hidden">
          <CardContent className="p-0">
            {notifications.length === 0 ? (
              <div className="text-center py-16">
                <Bell className="mx-auto h-12 w-12 text-muted-foreground opacity-40" />
                <h3 className="mt-4 text-sm font-semibold text-foreground">Tidak ada notifikasi</h3>
                <p className="mt-2 text-xs text-muted-foreground max-w-xs mx-auto">
                  Semua aktivitas Anda sudah dibaca.
                </p>
              </div>
            ) : (
              <div className="divide-y divide-border">
                {notifications.map((notif) => {
                  const IconComponent = notif.icon;
                  return (
                    <div
                      key={notif.id}
                      className={`p-5 flex gap-4 transition-colors hover:bg-secondary/10 ${
                        notif.unread ? "bg-muted/10 font-medium" : "bg-card"
                      }`}
                    >
                      <div className={`p-2.5 rounded-lg h-fit ${notif.color} shrink-0`}>
                        <IconComponent className="h-5 w-5" />
                      </div>
                      <div className="flex-1 space-y-1">
                        <div className="flex items-center justify-between gap-4">
                          <h4 className="text-sm font-bold text-foreground leading-none">{notif.title}</h4>
                          <span className="text-[10px] text-muted-foreground shrink-0">{notif.date}</span>
                        </div>
                        <p className="text-xs text-muted-foreground leading-relaxed">{notif.desc}</p>
                      </div>
                      {notif.unread && (
                        <div className="h-2.5 w-2.5 rounded-full bg-primary shrink-0 self-center" />
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </BuyerLayout>
  );
}
