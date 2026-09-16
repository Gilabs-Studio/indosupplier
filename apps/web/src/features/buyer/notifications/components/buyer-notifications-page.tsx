"use client";

import React from "react";
import { useTranslations } from "next-intl";
import { BuyerLayout } from "../../components/buyer-layout";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { CenteredLoading } from "@/components/loading";
import { Link } from "@/i18n/routing";
import { Bell, Mail, RefreshCw, AlertCircle, MessageSquare, ArrowRight } from "lucide-react";
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
        <CenteredLoading />
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
                  const isRfq = notif.related_type === "rfq" || notif.type === "quote";
                  const isChat = notif.related_type === "chat" || notif.type === "message";
                  const targetHref = isChat && notif.related_id
                    ? `/chat?roomId=${notif.related_id}`
                    : notif.related_id
                    ? `/rfq/${notif.related_id}`
                    : "/rfq";

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
                      <div className="flex-1 space-y-2">
                        <div className="flex items-center justify-between gap-4">
                          <div className="flex items-center gap-2">
                            <h4 className="text-sm font-bold text-foreground leading-none">{notif.title}</h4>
                            {notif.unread && (
                              <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-semibold bg-destructive/10 text-destructive border border-destructive/20">
                                Baru
                              </span>
                            )}
                          </div>
                          <span className="text-[10px] text-muted-foreground shrink-0">{notif.date}</span>
                        </div>
                        <p className="text-xs text-muted-foreground leading-relaxed">{notif.desc}</p>

                        {/* Action Link & Conversation detail button */}
                        {(notif.related_id || isRfq || isChat) && (
                          <div className="pt-1.5 flex items-center gap-2">
                            <Link
                              href={targetHref}
                              className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold bg-primary/10 text-primary hover:bg-primary hover:text-primary-foreground transition-all duration-200 cursor-pointer"
                            >
                              <MessageSquare className="h-3.5 w-3.5" />
                              {isChat
                                ? "Buka Sesi Chat"
                                : isRfq
                                ? "Lihat Penawaran & Chat"
                                : "Buka Detail"}
                              <ArrowRight className="h-3 w-3" />
                            </Link>
                          </div>
                        )}
                      </div>
                      {notif.unread && (
                        <div className="h-2.5 w-2.5 rounded-full bg-destructive shrink-0 self-center" />
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
