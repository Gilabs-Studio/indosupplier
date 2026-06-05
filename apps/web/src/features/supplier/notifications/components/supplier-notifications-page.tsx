"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";
import { Bell, CreditCard, Inbox, Check } from "lucide-react";

export function SupplierNotificationsPage() {
  const t = useTranslations("supplier.notifications");
  const [activeTab, setActiveTab] = useState("all");

  const [notifications, setNotifications] = useState([
    { id: "N1", title: "New RFQ Inquiry Matching Category", text: "CV Borneo Abadi is looking for Bentonite Clay Powder (10 Ton). Submit your quote proposal now.", type: "rfq", date: "Just now", read: false },
    { id: "N2", title: "Billing Outstanding Charge", text: "Your advertising clicks balance is due. Click here to settle payment for invoice INV-2026-005.", type: "billing", date: "2 hours ago", read: false },
    { id: "N3", title: "Gold Tier Subscription Active", text: "PT Nusantara Supplier Utama membership has been successfully upgraded to Gold Enterprise tier.", type: "system", date: "1 day ago", read: true },
    { id: "N4", title: "Admin Support Chat Update", text: "Helpdesk agent has replied to ticket TCK-2026-908 regarding document evaluations.", type: "system", date: "2 days ago", read: true },
  ]);

  const handleMarkAllRead = () => {
    setNotifications(notifications.map(n => ({ ...n, read: true })));
    toast.success("All notifications marked as read!");
  };

  const handleMarkSingleRead = (id: string) => {
    setNotifications(notifications.map(n => n.id === id ? { ...n, read: true } : n));
  };

  const tabs = [
    { id: "all", label: "All Alerts" },
    { id: "rfq", label: "RFQ Sourcing" },
    { id: "billing", label: "Billing" },
    { id: "system", label: "System Notices" }
  ];

  const filtered = notifications.filter(n => activeTab === "all" || n.type === activeTab);

  const getIcon = (type: string) => {
    switch (type) {
      case "rfq":
        return <Inbox className="h-4 w-4 text-primary" />;
      case "billing":
        return <CreditCard className="h-4 w-4 text-warning" />;
      default:
        return <Bell className="h-4 w-4 text-muted-foreground" />;
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
        <Button onClick={handleMarkAllRead} variant="outline" className="cursor-pointer border-border font-semibold text-xs h-9">
          <Check className="mr-1.5 h-4 w-4" /> {t("btnMarkAll")}
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
        {filtered.length > 0 ? (
          filtered.map((n) => (
            <Card
              key={n.id}
              onClick={() => handleMarkSingleRead(n.id)}
              className={`border border-border shadow-xs rounded-xl overflow-hidden bg-card cursor-pointer transition-colors p-4 flex items-start gap-4 ${
                !n.read ? "border-l-4 border-l-primary" : ""
              }`}
            >
              {/* Icon */}
              <div className="h-9 w-9 rounded-lg bg-muted/40 border border-border flex items-center justify-center shrink-0">
                {getIcon(n.type)}
              </div>

              {/* Text contents */}
              <div className="space-y-1 flex-1">
                <h4 className="text-sm font-bold text-foreground flex items-center justify-between gap-2">
                  {n.title}
                  <span className="text-[10px] text-muted-foreground font-semibold shrink-0">{n.date}</span>
                </h4>
                <p className="text-xs text-muted-foreground leading-relaxed font-semibold">
                  {n.text}
                </p>
              </div>

              {/* Unread indicator */}
              {!n.read && (
                <div className="h-2 w-2 rounded-full bg-primary shrink-0 self-center" />
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
