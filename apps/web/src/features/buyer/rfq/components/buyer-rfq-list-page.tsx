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
import { cn } from "@/lib/utils";
import {
  Plus,
  Search,
  ChevronRight,
  FileText,
  Calendar,
  MessageSquare,
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

  const tabs = [
    { id: "all", name: t("tabAll") },
    { id: "waiting", name: t("tabWaiting") },
    { id: "received", name: t("tabReceived") },
    { id: "completed", name: t("tabCompleted") },
  ];

  if (isLoading) {
    return (
      <BuyerLayout>
        <CenteredLoading />
      </BuyerLayout>
    );
  }

  const rfqs = rfqData?.items || [];
  const filteredRfqs = rfqs.filter((rfq) =>
    rfq.product.toLowerCase().includes(searchQuery.toLowerCase()) ||
    rfq.id.toLowerCase().includes(searchQuery.toLowerCase()) ||
    rfq.category.toLowerCase().includes(searchQuery.toLowerCase())
  );

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
          <div className="flex flex-wrap gap-1.5 overflow-x-auto scrollbar-none">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => {
                  setActiveTab(tab.id);
                  setPage(1);
                }}
                className={cn(
                  "px-4 py-2 text-sm font-medium transition-all duration-200 whitespace-nowrap cursor-pointer hover:text-primary hover:-translate-y-0.5 active:translate-y-0",
                  activeTab === tab.id
                    ? "text-primary font-semibold border-b-2 border-primary -mb-[5px]"
                    : "text-muted-foreground"
                )}
              >
                {tab.name}
              </button>
            ))}
          </div>

          <div className="relative max-w-xs w-full">
            <Search className="absolute left-3 top-3 h-4.5 w-4.5 text-muted-foreground" />
            <Input
              type="text"
              placeholder={t("searchPlaceholder")}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9 pr-4 py-1.5 bg-card border border-border text-sm rounded-lg outline-hidden focus-visible:ring-primary focus-visible:border-primary transition-all cursor-pointer h-9"
            />
          </div>
        </div>

        {/* RFQ List Table */}
        <Card className="border border-border rounded-xl shadow-xs overflow-hidden bg-card">
          <CardContent className="p-0">
            {filteredRfqs.length === 0 ? (
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
                    {filteredRfqs.map((rfq) => (
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
                                  <MessageSquare className="h-3.5 w-3.5" /> {t("viewReplies", { count: rfq.replies })}
                                </span>
                              ) : (
                                t("viewDetail")
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
