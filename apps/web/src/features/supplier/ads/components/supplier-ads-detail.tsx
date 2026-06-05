"use client";

import React, { useState, useEffect } from "react";
import { useRouter } from "@/i18n/routing";
import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ArrowLeft, Eye, MousePointer, TrendingUp, CreditCard } from "lucide-react";

interface SupplierAdsDetailProps {
  id: string;
}

export function SupplierAdsDetail({ id }: SupplierAdsDetailProps) {
  const router = useRouter();
  const t = useTranslations("supplier.ads");

  const [campaign, setCampaign] = useState({
    id: "CAM-01",
    name: "Garnet Sand Search Ads",
    product: "Garnet Sand Mesh 80",
    budget: "Rp 150.000 / Day",
    bid: "Rp 1.500",
    status: "active",
    clicks: 342,
    impressions: 8540,
    spend: "Rp 513.000",
    ctr: "4.00%",
  });

  useEffect(() => {
    const timer = setTimeout(() => {
      if (id === "CAM-02") {
        setCampaign({
          id: "CAM-02",
          name: "Bentonite Banner Promo",
          product: "Bentonite Clay Powder",
          budget: "Rp 200.000 / Day",
          bid: "Rp 2.000",
          status: "paused",
          clicks: 120,
          impressions: 5600,
          spend: "Rp 240.000",
          ctr: "2.14%",
        });
      } else if (id === "CAM-03") {
        setCampaign({
          id: "CAM-03",
          name: "Quartz Minerals Boost",
          product: "Quartz Powder 325 Mesh",
          budget: "Rp 100.000 / Day",
          bid: "Rp 1.200",
          status: "ended",
          clicks: 450,
          impressions: 12400,
          spend: "Rp 540.000",
          ctr: "3.63%",
        });
      }
    }, 0);
    return () => clearTimeout(timer);
  }, [id]);

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "active":
        return <Badge className="bg-success/15 text-success border border-success/30 font-bold">{t("statusActive")}</Badge>;
      case "paused":
        return <Badge className="bg-warning/15 text-warning border border-warning/30 font-bold">{t("statusPaused")}</Badge>;
      default:
        return <Badge variant="secondary" className="font-bold">{t("statusEnded")}</Badge>;
    }
  };

  return (
    <div className="space-y-6 text-left">
      {/* Header */}
      <div className="flex items-center gap-4 border-b border-border/80 pb-6">
        <Button
          variant="outline"
          size="icon"
          onClick={() => router.push("/supplier/ads")}
          className="h-9 w-9 cursor-pointer border-border"
        >
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div className="space-y-1 flex-1">
          <div className="flex items-center gap-2 flex-wrap">
            <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
              {campaign.name}
            </h1>
            {getStatusBadge(campaign.status)}
          </div>
          <p className="text-sm text-muted-foreground">
            {t("detailSubtitle")}
          </p>
        </div>
      </div>

      {/* Metrics Section */}
      <div className="grid gap-4 grid-cols-2 md:grid-cols-4">
        <Card className="border-border shadow-xs rounded-xl bg-card">
          <CardContent className="p-4 flex items-center justify-between">
            <div>
              <p className="text-[10px] font-bold text-muted-foreground uppercase">{t("statImpressions")}</p>
              <h3 className="text-xl font-bold tracking-tight text-foreground mt-1">{campaign.impressions}</h3>
            </div>
            <div className="h-10 w-10 bg-primary/10 text-primary border border-border rounded-lg flex items-center justify-center">
              <Eye className="h-4.5 w-4.5" />
            </div>
          </CardContent>
        </Card>
        <Card className="border-border shadow-xs rounded-xl bg-card">
          <CardContent className="p-4 flex items-center justify-between">
            <div>
              <p className="text-[10px] font-bold text-muted-foreground uppercase">{t("statClicks")}</p>
              <h3 className="text-xl font-bold tracking-tight text-foreground mt-1">{campaign.clicks}</h3>
            </div>
            <div className="h-10 w-10 bg-success/10 text-success border border-border rounded-lg flex items-center justify-center">
              <MousePointer className="h-4.5 w-4.5" />
            </div>
          </CardContent>
        </Card>
        <Card className="border-border shadow-xs rounded-xl bg-card">
          <CardContent className="p-4 flex items-center justify-between">
            <div>
              <p className="text-[10px] font-bold text-muted-foreground uppercase">{t("statCtr")}</p>
              <h3 className="text-xl font-bold tracking-tight text-foreground mt-1">{campaign.ctr}</h3>
            </div>
            <div className="h-10 w-10 bg-warning/10 text-warning border border-border rounded-lg flex items-center justify-center">
              <TrendingUp className="h-4.5 w-4.5" />
            </div>
          </CardContent>
        </Card>
        <Card className="border-border shadow-xs rounded-xl bg-card">
          <CardContent className="p-4 flex items-center justify-between">
            <div>
              <p className="text-[10px] font-bold text-muted-foreground uppercase">{t("statSpend")}</p>
              <h3 className="text-xl font-bold tracking-tight text-foreground mt-1">{campaign.spend}</h3>
            </div>
            <div className="h-10 w-10 bg-purple/10 text-purple border border-border rounded-lg flex items-center justify-center">
              <CreditCard className="h-4.5 w-4.5" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Chart Block Mockup */}
      <Card className="border border-border rounded-xl shadow-xs overflow-hidden bg-card">
        <CardHeader className="border-b border-border bg-muted/20 py-4 px-6">
          <CardTitle className="text-sm font-bold text-foreground">Click Performance History</CardTitle>
          <CardDescription className="text-xs">Past 7 days clicks analytics.</CardDescription>
        </CardHeader>
        <CardContent className="p-6">
          <div className="h-[180px] flex items-end justify-between gap-4 pt-4">
            {[10, 24, 15, 30, 45, 38, 55].map((val, idx) => (
              <div key={idx} className="flex-1 flex flex-col items-center gap-2">
                <div
                  className="w-full bg-primary rounded-t-sm hover:opacity-85 transition-opacity"
                  style={{ height: `${(val / 60) * 140}px` }}
                />
                <span className="text-[10px] text-muted-foreground font-semibold">
                  {["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"][idx]}
                </span>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
