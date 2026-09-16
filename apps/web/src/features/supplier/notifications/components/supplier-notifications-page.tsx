"use client";

import React from "react";
import { useTranslations } from "next-intl";
import { useRouter } from "@/i18n/routing";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Bell, CreditCard, Inbox, Check, Loader2, MessageSquare, ArrowRight } from "lucide-react";
import { useSupplierNotifications } from "../hooks/useSupplierNotifications";
import type { SupplierNotification } from "../types/notifications.types";

export function SupplierNotificationsPage() {
  const t = useTranslations("supplier.notifications");
  const router = useRouter();
  const {
    notifications,
    isLoading,
    activeTab,
    setActiveTab,
    markAllRead,
    isMarkingAllRead,
  } = useSupplierNotifications();

  const tabs = [
    { id: "all", label: "Semua Notifikasi" },
    { id: "rfq", label: "RFQ & Penawaran" },
    { id: "system", label: "Sistem & Info" }
  ];

  const getIcon = (type: string) => {
    switch (type) {
      case "quote":
        return <Inbox className="h-4 w-4 text-primary" />;
      case "alert":
        return <CreditCard className="h-4 w-4 text-warning" />;
      default:
        return <Bell className="h-4 w-4 text-muted-foreground" />;
    }
  };

  const handleItemClick = (n: SupplierNotification) => {
    if (n.related_type === "chat" && n.related_id) {
      router.push(`/chat?roomId=${n.related_id}`);
    } else if (n.related_id) {
      router.push(`/supplier/rfq/${n.related_id}`);
    } else if (n.type === "quote") {
      router.push("/supplier/rfq");
    }
  };

  return (
    <div className="space-y-6 text-left">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-border/80 pb-6">
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
            {t("title")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("subtitle")}
          </p>
        </div>
        <Button
          onClick={() => markAllRead()}
          disabled={isMarkingAllRead}
          variant="outline"
          className="cursor-pointer border-border font-semibold text-xs h-9"
        >
          {isMarkingAllRead ? (
            <Loader2 className="mr-1.5 h-3.5 w-3.5 animate-spin text-primary" />
          ) : (
            <Check className="mr-1.5 h-4 w-4" />
          )}
          {t("btnMarkAll")}
        </Button>
      </div>

      {/* Tabs */}
      <div className="flex flex-wrap gap-1 border-b border-border pb-1">
        {tabs.map((tab) => (
          <Button
            key={tab.id}
            variant="ghost"
            onClick={() => setActiveTab(tab.id)}
            className={`text-xs h-9 font-semibold relative cursor-pointer px-4 rounded-lg transition-all ${
              activeTab === tab.id
                ? "text-primary bg-primary/10"
                : "text-muted-foreground hover:text-foreground hover:bg-muted/40"
            }`}
          >
            {tab.label}
          </Button>
        ))}
      </div>

      {/* List */}
      <div className="space-y-3">
        {isLoading ? (
          <div className="flex justify-center py-12">
            <Loader2 className="h-6 w-6 animate-spin text-primary" />
          </div>
        ) : notifications.length > 0 ? (
          notifications.map((n) => (
            <Card
              key={n.id}
              onClick={() => handleItemClick(n)}
              className={`border border-border shadow-xs rounded-xl overflow-hidden bg-card cursor-pointer transition-colors p-4 flex items-start gap-4 hover:border-primary/40 ${
                n.unread ? "border-l-4 border-l-primary" : ""
              }`}
            >
              {/* Icon */}
              <div className="h-9 w-9 rounded-lg bg-muted/40 border border-border flex items-center justify-center shrink-0">
                {getIcon(n.type)}
              </div>

              {/* Text contents */}
              <div className="space-y-2 flex-1">
                <div className="flex items-center justify-between gap-2">
                  <div className="flex items-center gap-2">
                    <h4 className="text-sm font-bold text-foreground">
                      {n.title}
                    </h4>
                    {n.unread && (
                      <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-semibold bg-destructive/10 text-destructive border border-destructive/20">
                        Baru
                      </span>
                    )}
                  </div>
                  <span className="text-[10px] text-muted-foreground font-semibold shrink-0">{n.date}</span>
                </div>
                <p className="text-xs text-muted-foreground leading-relaxed">
                  {n.desc}
                </p>

                {/* Quick Action Button */}
                {(n.related_id || n.type === "quote") && (
                  <div className="pt-1 flex items-center gap-2">
                    <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-lg text-xs font-semibold bg-primary/10 text-primary hover:bg-primary hover:text-primary-foreground transition-all duration-200 cursor-pointer">
                      {n.related_type === "chat" ? (
                        <>
                          <MessageSquare className="h-3.5 w-3.5" />
                          Buka Sesi Chat
                        </>
                      ) : (
                        <>
                          <Inbox className="h-3.5 w-3.5" />
                          Buka Detail & Ajukan Penawaran
                        </>
                      )}
                      <ArrowRight className="h-3 w-3" />
                    </span>
                  </div>
                )}
              </div>

              {/* Unread indicator */}
              {n.unread && (
                <div className="h-2.5 w-2.5 rounded-full bg-destructive shrink-0 self-center" />
              )}
            </Card>
          ))
        ) : (
          <div className="py-12 text-center text-sm text-muted-foreground border border-dashed border-border rounded-xl">
            {t("noNotifications")}
          </div>
        )}
      </div>
    </div>
  );
}
