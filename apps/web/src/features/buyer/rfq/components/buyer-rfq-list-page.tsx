"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { BuyerLayout } from "../../components/buyer-layout";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Link } from "@/i18n/routing";
import {
  Plus,
  Search,
  ChevronRight,
  FileText,
  Calendar,
  MessageSquare,
} from "lucide-react";

export function BuyerRfqListPage() {
  const t = useTranslations("buyer.rfqList");
  const [activeTab, setActiveTab] = useState("all");
  
  const [rfqs] = useState([
    {
      id: "RFQ-2026-004",
      product: "Garnet Sand Mesh 80",
      category: "Industrial Minerals",
      quantity: "50 Ton",
      targetPort: "Tanjung Priok, Jakarta",
      date: "2026-06-01",
      status: "Waiting for Quotes",
      replies: 3,
    },
    {
      id: "RFQ-2026-003",
      product: "Bentonite Clay Powder",
      category: "Chemicals",
      quantity: "20 Ton",
      targetPort: "Tanjung Perak, Surabaya",
      date: "2026-05-28",
      status: "Offers Received",
      replies: 8,
    },
    {
      id: "RFQ-2026-002",
      product: "Quartz Powder 325 Mesh",
      category: "Industrial Minerals",
      quantity: "100 Ton",
      targetPort: "Tanjung Priok, Jakarta",
      date: "2026-05-15",
      status: "Completed",
      replies: 5,
    },
    {
      id: "RFQ-2026-001",
      product: "Organic Coconut Sugar Organic Grade",
      category: "Agriculture & Food",
      quantity: "5 Ton",
      targetPort: "Port of Rotterdam (CIF)",
      date: "2026-05-01",
      status: "Completed",
      replies: 12,
    },
  ]);

  const tabs = [
    { id: "all", name: t("tabAll") },
    { id: "waiting", name: t("tabWaiting") },
    { id: "received", name: t("tabReceived") },
    { id: "completed", name: t("tabCompleted") },
  ];

  const getFilteredRfqs = () => {
    if (activeTab === "waiting") return rfqs.filter((r) => r.status === "Waiting for Quotes");
    if (activeTab === "received") return rfqs.filter((r) => r.status === "Offers Received");
    if (activeTab === "completed") return rfqs.filter((r) => r.status === "Completed");
    return rfqs;
  };

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
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-border pb-1">
          <div className="flex flex-wrap gap-1">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`px-4 py-2 text-sm font-semibold rounded-t-lg transition-all cursor-pointer ${
                  activeTab === tab.id
                    ? "border-b-2 border-primary text-primary font-bold bg-muted/20"
                    : "text-muted-foreground hover:text-foreground"
                }`}
              >
                {tab.name}
              </button>
            ))}
          </div>

          <div className="relative max-w-xs w-full">
            <Search className="absolute left-3 top-2.5 h-4.5 w-4.5 text-muted-foreground" />
            <input
              type="text"
              placeholder={t("searchPlaceholder")}
              className="w-full pl-9 pr-4 py-1.5 bg-card border border-border text-sm rounded-lg outline-hidden focus:border-primary transition-all cursor-pointer"
            />
          </div>
        </div>

        {/* RFQ List Table */}
        <Card className="border border-border rounded-xl shadow-xs overflow-hidden bg-card">
          <CardContent className="p-0">
            {getFilteredRfqs().length === 0 ? (
              <div className="text-center py-16">
                <FileText className="mx-auto h-12 w-12 text-muted-foreground opacity-40" />
                <h3 className="mt-4 text-sm font-semibold text-foreground">{t("emptyRfqs")}</h3>
                <p className="mt-2 text-xs text-muted-foreground max-w-xs mx-auto">
                  Belum ada RFQ yang sesuai dengan filter tab ini.
                </p>
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left text-sm text-foreground">
                  <thead className="bg-muted/30 border-b border-border text-muted-foreground text-xs font-bold uppercase tracking-wider">
                    <tr>
                      <th className="p-4 px-6">{t("colId")}</th>
                      <th className="p-4">{t("colProduct")}</th>
                      <th className="p-4">{t("colQty")}</th>
                      <th className="p-4">{t("colDestination")}</th>
                      <th className="p-4">{t("colStatus")}</th>
                      <th className="p-4 text-right px-6">{t("colAction")}</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border">
                    {getFilteredRfqs().map((rfq) => (
                      <tr key={rfq.id} className="hover:bg-muted/10 transition-colors">
                        <td className="p-4 px-6 space-y-0.5">
                          <span className="text-xs font-bold text-muted-foreground">{rfq.id}</span>
                          <div className="flex items-center gap-1 text-[11px] text-muted-foreground">
                            <Calendar className="h-3 w-3" />
                            <span>{rfq.date}</span>
                          </div>
                        </td>
                        <td className="p-4 font-semibold text-foreground">
                          <Link href={`/rfq/${rfq.id}`} className="hover:text-primary transition-colors cursor-pointer">
                            {rfq.product}
                          </Link>
                          <p className="text-[11px] font-normal text-muted-foreground mt-0.5">{rfq.category}</p>
                        </td>
                        <td className="p-4 font-semibold text-foreground">{rfq.quantity}</td>
                        <td className="p-4 text-xs text-muted-foreground">{rfq.targetPort}</td>
                        <td className="p-4">
                          <Badge
                            variant="outline"
                            className={
                              rfq.status === "Waiting for Quotes"
                                ? "bg-cyan/10 text-cyan border-cyan/20 rounded-full text-[10px]"
                                : rfq.status === "Offers Received"
                                ? "bg-success/10 text-success border-success/25 rounded-full text-[10px]"
                                : "bg-muted text-muted-foreground border-border rounded-full text-[10px]"
                            }
                          >
                            {rfq.status}
                          </Badge>
                        </td>
                        <td className="p-4 px-6 text-right">
                          <Button asChild variant="ghost" size="sm" className="text-primary hover:bg-primary/5 cursor-pointer font-semibold gap-1 transition-all">
                            <Link href={`/rfq/${rfq.id}`}>
                              {rfq.status === "Offers Received" ? (
                                <span className="flex items-center gap-1">
                                  <MessageSquare className="h-3.5 w-3.5" /> Lihat ({rfq.replies})
                                </span>
                              ) : (
                                "Detail"
                              )}
                              <ChevronRight className="h-4 w-4" />
                            </Link>
                          </Button>
                        </td>
                      </tr>
                    ))}
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
