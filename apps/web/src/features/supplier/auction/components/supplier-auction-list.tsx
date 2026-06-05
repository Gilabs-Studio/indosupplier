"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Gavel, Clock, ArrowUpRight } from "lucide-react";

export function SupplierAuctionList() {
  const t = useTranslations("supplier.auction");
  const [activeTab, setActiveTab] = useState("active");

  const [auctions] = useState([
    { id: "AUC-101", name: "Premium Search Result Slot #1 (Mineral Category)", currentBid: "Rp 2.500.000", timeLeft: "04h 32m", bids: 8, tab: "active" },
    { id: "AUC-102", name: "IndoSupplier Homepage Banner Slot A", currentBid: "Rp 5.200.000", timeLeft: "12h 15m", bids: 14, tab: "active" },
    { id: "AUC-103", name: "Top Supplier Spotlight Slot (Agriculture Category)", currentBid: "Rp 1.800.000", timeLeft: "1d 04h", bids: 4, tab: "active" },
    { id: "AUC-104", name: "Premium Sidebar Display Slot #2", currentBid: "Rp 1.000.000", timeLeft: "Starts in 2d", bids: 0, tab: "upcoming" },
    { id: "AUC-098", name: "Recommended Supplier Spotlight Slot", currentBid: "Rp 3.100.000", timeLeft: "Ended (Won)", bids: 11, tab: "won" },
  ]);

  const tabs = [
    { id: "active", label: t("activeAuctions") },
    { id: "upcoming", label: t("upcomingAuctions") },
    { id: "won", label: t("wonAuctions") },
  ];

  const filtered = auctions.filter((a) => a.tab === activeTab);

  return (
    <div className="space-y-6 text-left">
      {/* Page Header */}
      <div className="flex flex-col md:flex-row md:items-end md:justify-between gap-4 border-b border-border/80 pb-6">
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
            {t("listTitle")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("listSubtitle")}
          </p>
        </div>
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

      {/* Table */}
      <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow className="border-b border-border bg-muted/10">
                  <TableHead className="font-bold text-foreground py-3 px-6">{t("tableSlot")}</TableHead>
                  <TableHead className="font-bold text-foreground">{t("tableCurrentBid")}</TableHead>
                  <TableHead className="font-bold text-foreground">{t("tableTimeLeft")}</TableHead>
                  <TableHead className="font-bold text-foreground">{t("tableBidsCount")}</TableHead>
                  <TableHead className="font-bold text-foreground text-right px-6">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filtered.length > 0 ? (
                  filtered.map((a) => (
                    <TableRow key={a.id} className="hover:bg-muted/10 border-b border-border transition-colors">
                      <TableCell className="py-4 px-6 font-semibold text-foreground">
                        <div className="flex items-center gap-2">
                          <Gavel className="h-4.5 w-4.5 text-primary shrink-0" />
                          <span>{a.name}</span>
                        </div>
                        <p className="text-[10px] text-muted-foreground mt-0.5">{a.id}</p>
                      </TableCell>
                      <TableCell className="py-4 font-extrabold text-foreground">{a.currentBid}</TableCell>
                      <TableCell className="py-4 text-xs text-muted-foreground font-semibold">
                        <span className="flex items-center gap-1">
                          <Clock className="h-3.5 w-3.5 text-muted-foreground shrink-0" />
                          {a.timeLeft}
                        </span>
                      </TableCell>
                      <TableCell className="py-4 text-xs font-semibold">{a.bids} Bids</TableCell>
                      <TableCell className="py-4 px-6 text-right">
                        <Button asChild size="sm" variant={activeTab === "active" ? "default" : "outline"} className="text-xs font-semibold h-8 cursor-pointer border-border">
                          <Link href={`/supplier/auction/${a.id}`}>
                            {activeTab === "active" ? t("actionPlace") : t("actionDetail")}
                            <ArrowUpRight className="ml-1 h-3.5 w-3.5" />
                          </Link>
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell colSpan={5} className="py-8 text-center text-sm text-muted-foreground">
                      No auctions found in this category.
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
