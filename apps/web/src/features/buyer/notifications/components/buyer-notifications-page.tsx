"use client";

import React from "react";
import { useTranslations } from "next-intl";
import { BuyerLayout } from "../../components/buyer-layout";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Bell, Mail, RefreshCw, AlertCircle } from "lucide-react";
import { useBuyerNotifications } from "../hooks/useBuyerNotifications";

export function BuyerNotificationsPage() {
  const t = useTranslations("buyer.notifications");
  const { notifications, isLoading, markAllRead } = useBuyerNotifications();

  const handleMarkAllRead = () => {
    markAllRead();
  };

  const getIcon = (type: string) => {
    switch (type) {
      case "quote":
        return RefreshCw;
      case "message":
        return Mail;
      case "alert":
        return AlertCircle;
      default:
        return Bell;
    }
  };

  const getColorClass = (type: string) => {
    switch (type) {
      case "quote":
        return "bg-primary/10 text-primary";
      case "message":
        return "bg-success/10 text-success";
      case "alert":
        return "bg-destructive/10 text-destructive";
      default:
        return "bg-cyan/10 text-cyan";
    }
  };

  if (isLoading) {
    return (
      <BuyerLayout>
        <div className="flex items-center justify-center py-20">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
        </div>
      </BuyerLayout>
    );
  }

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
                  const IconComponent = getIcon(notif.type);
                  return (
                    <div
                      key={notif.id}
                      className={`p-5 flex gap-4 transition-colors hover:bg-secondary/10 ${
                        notif.unread ? "bg-muted/10 font-medium" : "bg-card"
                      }`}
                    >
                      <div className={`p-2.5 rounded-lg h-fit ${getColorClass(notif.type)} shrink-0`}>
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
