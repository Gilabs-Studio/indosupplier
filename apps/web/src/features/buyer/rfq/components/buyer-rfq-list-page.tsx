"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { BuyerLayout } from "../../components/buyer-layout";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Link } from "@/i18n/routing";
import { CenteredLoading } from "@/components/loading";
import { Input } from "@/components/ui/input";
import { cn, resolveImageUrl } from "@/lib/utils";
import {
  Plus,
  Search,
  ChevronRight,
  FileText,
  Calendar,
  MessageSquare,
  Package,
  MapPin,
} from "lucide-react";
import { useBuyerRfqs } from "../hooks/useBuyerRfqs";

export function BuyerRfqListPage() {
  const t = useTranslations("buyerRfq.rfqList");
  const [activeTab, setActiveTab] = useState("all");
  const [page, setPage] = useState(1);
  const [searchQuery, setSearchQuery] = useState("");

  const { data: rfqData, isLoading } = useBuyerRfqs({
    page,
    per_page: 20,
    status: activeTab,
  });

  // Dedicated query to keep the received offers count accurate and persistent across all tabs
  const { data: receivedCountData } = useBuyerRfqs({
    status: "received",
    per_page: 1,
  });

  const rfqs = rfqData?.items || [];

  // Count active offers for notification badge from server total, falling back to current list
  const offersCount = receivedCountData?.total ?? rfqs.filter((r) => r.status === "Offers Received" || r.replies > 0).length;

  const tabs = [
    { id: "all", name: t("tabAll"), count: 0 },
    { id: "waiting", name: t("tabWaiting"), count: 0 },
    { id: "received", name: t("tabReceived"), count: offersCount },
    { id: "completed", name: t("tabCompleted"), count: 0 },
  ];

  if (isLoading) {
    return (
      <BuyerLayout>
        <CenteredLoading />
      </BuyerLayout>
    );
  }

  const filteredRfqs = rfqs.filter((rfq) =>
    rfq.product.toLowerCase().includes(searchQuery.toLowerCase()) ||
    rfq.category.toLowerCase().includes(searchQuery.toLowerCase()) ||
    rfq.targetPort.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // Sort priority: RFQs with offers received come first, then by date
  const sortedRfqs = [...filteredRfqs].sort((a, b) => {
    const aHasOffer = a.status === "Offers Received" || a.replies > 0;
    const bHasOffer = b.status === "Offers Received" || b.replies > 0;
    if (aHasOffer && !bHasOffer) return -1;
    if (!aHasOffer && bHasOffer) return 1;
    return new Date(b.date).getTime() - new Date(a.date).getTime();
  });

  return (
    <BuyerLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="space-y-1">
            <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">{t("title")}</h1>
            <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
          </div>
          <Button asChild className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/95 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/20">
            <Link href="/rfq/create">
              <Plus className="mr-2 h-4 w-4" />
              {t("btnCreate")}
            </Link>
          </Button>
        </div>

        {/* Tabs & Search */}
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-border">
          <div className="flex items-center gap-1 overflow-x-auto scrollbar-none">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => {
                  setActiveTab(tab.id);
                  setPage(1);
                }}
                className={cn(
                  "relative pb-3 px-3 text-sm font-medium transition-colors whitespace-nowrap cursor-pointer flex items-center gap-2",
                  activeTab === tab.id
                    ? "text-primary font-semibold after:absolute after:bottom-0 after:left-0 after:right-0 after:h-0.5 after:bg-primary"
                    : "text-muted-foreground hover:text-foreground"
                )}
              >
                <span>{tab.name}</span>
                {tab.count > 0 && (
                  <span className="inline-flex items-center justify-center h-4.5 min-w-4.5 px-1 rounded-full text-[11px] font-semibold bg-destructive/15 text-destructive leading-none shrink-0">
                    {tab.count}
                  </span>
                )}
              </button>
            ))}
          </div>

          <div className="relative max-w-xs w-full pb-2 md:pb-0">
            <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              type="text"
              placeholder={t("searchPlaceholder")}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9 pr-4 py-1 bg-card border border-border text-sm rounded-lg outline-hidden focus-visible:ring-primary focus-visible:border-primary transition-all h-8.5"
            />
          </div>
        </div>

        {/* RFQ List Table */}
        <Card className="border border-border rounded-xl shadow-xs overflow-hidden bg-card">
          <CardContent className="p-0">
            {sortedRfqs.length === 0 ? (
              <div className="text-center py-16">
                <FileText className="mx-auto h-12 w-12 text-muted-foreground opacity-40" />
                <h3 className="mt-4 text-sm font-semibold text-foreground">{t("emptyRfqs")}</h3>
                <p className="mt-2 text-xs text-muted-foreground max-w-xs mx-auto">
                  {t("emptyDesc")}
                </p>
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left text-sm text-foreground">
                  <thead className="border-b border-border text-muted-foreground text-[11px] font-medium uppercase tracking-wider">
                    <tr>
                      <th className="py-3 px-6">{t("colProduct")}</th>
                      <th className="py-3 px-4">{t("colQty")}</th>
                      <th className="py-3 px-4">{t("colDestination")}</th>
                      <th className="py-3 px-4">{t("colStatus")}</th>
                      <th className="py-3 px-6 text-right">{t("colAction")}</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border/60">
                    {sortedRfqs.map((rfq) => {
                      const hasOffer = rfq.status === "Offers Received" || rfq.replies > 0;
                      const defaultThumbnail = "/images/categories/cat-bahan-baku.webp";
                      const imageSrc = resolveImageUrl(rfq.imageUrl || defaultThumbnail);

                      return (
                        <tr
                          key={rfq.id}
                          className="hover:bg-muted/20 border-b border-border/60 transition-colors"
                        >
                          <td className="py-3.5 px-6">
                            <div className="flex items-center gap-3">
                              <div className="h-11 w-11 shrink-0 rounded-lg overflow-hidden border border-border/80 bg-muted/20">
                                <img
                                  src={imageSrc}
                                  alt={rfq.product}
                                  className="h-full w-full object-cover"
                                  onError={(e) => {
                                    e.currentTarget.src = defaultThumbnail;
                                  }}
                                />
                              </div>
                              <div className="min-w-0 max-w-md space-y-0.5">
                                <Link
                                  href={`/rfq/${rfq.id}`}
                                  className="font-medium text-sm text-foreground hover:text-primary transition-colors cursor-pointer line-clamp-1"
                                >
                                  {rfq.product}
                                </Link>
                                <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                                  <span>{rfq.category}</span>
                                  <span>•</span>
                                  <span>{rfq.date}</span>
                                </div>
                              </div>
                            </div>
                          </td>
                          <td className="py-3.5 px-4 text-sm text-foreground whitespace-nowrap">
                            {rfq.quantity}
                          </td>
                          <td className="py-3.5 px-4 text-xs text-muted-foreground max-w-xs truncate">
                            {rfq.targetPort}
                          </td>
                          <td className="py-3.5 px-4 whitespace-nowrap">
                            {hasOffer ? (
                              <div className="flex items-center gap-1.5 text-xs font-medium text-foreground">
                                <span className="h-1.5 w-1.5 rounded-full bg-primary shrink-0" />
                                <span>Ada Penawaran</span>
                              </div>
                            ) : rfq.status === "Waiting for Quotes" ? (
                              <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                                <span className="h-1.5 w-1.5 rounded-full bg-muted-foreground/30 shrink-0" />
                                <span>Menunggu</span>
                              </div>
                            ) : (
                              <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                                <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 shrink-0" />
                                <span>{rfq.status}</span>
                              </div>
                            )}
                          </td>
                          <td className="py-3.5 px-6 text-right whitespace-nowrap">
                            {hasOffer ? (
                              <Link
                                href={`/rfq/${rfq.id}`}
                                className="inline-flex items-center justify-end gap-1.5 text-foreground hover:text-primary transition-colors cursor-pointer py-1 px-2 rounded-md hover:bg-muted/40 group"
                                title={`${rfq.replies || 1} Penawaran Masuk`}
                              >
                                <MessageSquare className="h-4 w-4 text-primary shrink-0 group-hover:scale-110 transition-transform" />
                                <span className="text-xs font-semibold">{rfq.replies || 1}</span>
                                <ChevronRight className="h-3.5 w-3.5 text-muted-foreground/60 group-hover:text-primary group-hover:translate-x-0.5 transition-all" />
                              </Link>
                            ) : (
                              <Link
                                href={`/rfq/${rfq.id}`}
                                className="inline-flex items-center justify-end gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors cursor-pointer py-1 px-2 rounded-md hover:bg-muted/40"
                              >
                                <span>{t("viewDetail")}</span>
                                <ChevronRight className="h-3.5 w-3.5 text-muted-foreground/60" />
                              </Link>
                            )}
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>

                </table>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </BuyerLayout>
  );
}
