"use client";

import React from "react";
import { useTranslations } from "next-intl";
import { BuyerLayout } from "../../components/buyer-layout";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Link } from "@/i18n/routing";
import {
  FileText,
  Plus,
  Search,
  ExternalLink,
  ShieldCheck,
  Star,
  MapPin,
  ArrowUpRight,
} from "lucide-react";

export function BuyerDashboardPage() {
  const t = useTranslations("buyer.dashboard");

  const stats = [
    { label: t("statsActiveRfqs"), value: "6", desc: "+2 this week", color: "text-primary" },
    { label: t("statsSavedSuppliers"), value: "24", desc: "Shortlisted", color: "text-success" },
    { label: t("statsNotifications"), value: "9", desc: "3 urgent", color: "text-warning" },
  ];

  const recentRfqs = [
    { id: "RFQ-2026-004", product: "Garnet Sand Mesh 80", date: "2026-06-01", status: "Waiting for Quotes", replies: 3 },
    { id: "RFQ-2026-003", product: "Bentonite Clay Powder", date: "2026-05-28", status: "Offers Received", replies: 8 },
    { id: "RFQ-2026-002", product: "Quartz Powder 325 Mesh", date: "2026-05-15", status: "Completed", replies: 5 },
  ];

  const suggestedSuppliers = [
    { name: "PT Rempah Nusantara", category: "Agriculture", location: "Surabaya", rating: 4.8, verified: true },
    { name: "CV Nusantara Garment", category: "Textile & Apparel", location: "Bandung", rating: 4.6, verified: true },
    { name: "PT Logam Steel Jaya", category: "Manufacturing", location: "Jakarta", rating: 4.7, verified: false },
  ];

  return (
    <BuyerLayout>
      <div className="space-y-6">
        {/* Welcome Section */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="space-y-1">
            <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">{t("title")}</h1>
            <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
          </div>
          <div className="flex gap-3">
            <Button asChild variant="outline" className="cursor-pointer transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-md">
              <Link href="/search">
                <Search className="mr-2 h-4 w-4" />
                {t("searchSupplier")}
              </Link>
            </Button>
            <Button asChild className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/95 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/20">
              <Link href="/rfq/create">
                <Plus className="mr-2 h-4 w-4" />
                {t("createRfq")}
              </Link>
            </Button>
          </div>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-6">
          {stats.map((stat, i) => (
            <Card key={i} className="border-border rounded-xl shadow-xs overflow-hidden bg-card">
              <CardContent className="p-6">
                <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">{stat.label}</p>
                <div className="mt-2 flex items-baseline gap-2">
                  <span className={`text-3xl font-bold tracking-tight ${stat.color}`}>{stat.value}</span>
                  <span className="text-xs text-muted-foreground">{stat.desc}</span>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>

        {/* Recent RFQs */}
        <Card className="border border-border rounded-xl shadow-xs overflow-hidden bg-card">
          <CardHeader className="flex flex-row items-center justify-between py-4 px-6 border-b border-border bg-muted/20">
            <CardTitle className="text-sm font-bold text-foreground">{t("recentRfqs")}</CardTitle>
            <Button asChild variant="ghost" size="sm" className="text-xs font-semibold text-primary cursor-pointer p-0 h-auto hover:bg-transparent transition-colors hover:text-primary/80">
              <Link href="/rfq" className="flex items-center gap-1">
                {t("allRfqs")}
                <ArrowUpRight className="h-4 w-4" />
              </Link>
            </Button>
          </CardHeader>
          <CardContent className="p-0">
            <div className="divide-y divide-border">
              {recentRfqs.map((rfq) => (
                <div key={rfq.id} className="p-4 sm:px-6 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 hover:bg-secondary/20 transition-colors">
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-bold text-muted-foreground">{rfq.id}</span>
                      <h3 className="text-sm font-semibold text-foreground hover:text-primary transition-colors cursor-pointer">
                        <Link href={`/rfq/${rfq.id}`}>{rfq.product}</Link>
                      </h3>
                    </div>
                    <p className="text-xs text-muted-foreground">Dibuat pada {rfq.date} • {t("repliesCount", { count: rfq.replies })}</p>
                  </div>
                  <div className="flex items-center justify-between sm:justify-end gap-3 shrink-0">
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
                    <Button asChild variant="ghost" size="icon" className="h-8 w-8 rounded-full border border-border cursor-pointer hover:bg-secondary transition-colors">
                      <Link href={`/rfq/${rfq.id}`}>
                        <ExternalLink className="h-4 w-4 text-muted-foreground hover:text-foreground" />
                      </Link>
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Suggested / Verified Suppliers */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <Card className="border border-border rounded-xl shadow-xs overflow-hidden bg-card">
            <CardHeader className="py-4 px-6 border-b border-border bg-muted/20">
              <CardTitle className="text-sm font-bold text-foreground">{t("suggestedTitle")}</CardTitle>
            </CardHeader>
            <CardContent className="p-4 space-y-4">
              {suggestedSuppliers.map((supplier, idx) => (
                <div key={idx} className="flex items-start justify-between gap-3 p-3 rounded-lg border border-border hover:border-muted-foreground/30 transition-all bg-card">
                  <div className="space-y-1">
                    <div className="flex items-center gap-1.5">
                      <h4 className="text-sm font-semibold text-foreground">{supplier.name}</h4>
                      {supplier.verified && (
                        <Badge className="bg-success text-white border-0 text-[9px] px-1.5 py-0.5 rounded-full flex items-center gap-0.5">
                          <ShieldCheck className="h-2.5 w-2.5" /> Verified
                        </Badge>
                      )}
                    </div>
                    <div className="flex items-center gap-3 text-xs text-muted-foreground">
                      <span className="flex items-center gap-1"><MapPin className="h-3 w-3" /> {supplier.location}</span>
                      <span>{supplier.category}</span>
                    </div>
                  </div>
                  <div className="flex flex-col items-end gap-1.5">
                    <div className="flex items-center gap-1 text-xs font-semibold text-foreground">
                      <Star className="h-3.5 w-3.5 fill-amber-400 stroke-amber-400" />
                      <span>{supplier.rating}</span>
                    </div>
                    <Button asChild size="sm" variant="ghost" className="text-xs h-7 text-primary hover:text-primary-foreground hover:bg-primary cursor-pointer px-2 transition-colors">
                      <Link href="/search">Hubungi</Link>
                    </Button>
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>

          {/* Quick Guidance Card */}
          <Card className="border border-border rounded-xl shadow-xs overflow-hidden bg-primary/5 relative">
            <CardContent className="p-6 space-y-4">
              <div className="p-2 bg-primary/10 text-primary rounded-lg w-fit">
                <FileText className="h-6 w-6" />
              </div>
              <h3 className="text-base font-bold text-foreground">{t("guideTitle")}</h3>
              <ul className="space-y-3 text-xs text-muted-foreground">
                <li className="flex items-start gap-2">
                  <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-primary/15 text-primary font-bold">1</span>
                  <span><strong>{t("guideStep1Title")}:</strong> {t("guideStep1Desc")}</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-primary/15 text-primary font-bold">2</span>
                  <span><strong>{t("guideStep2Title")}:</strong> {t("guideStep2Desc")}</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-primary/15 text-primary font-bold">3</span>
                  <span><strong>{t("guideStep3Title")}:</strong> {t("guideStep3Desc")}</span>
                </li>
              </ul>
            </CardContent>
          </Card>
        </div>
      </div>
    </BuyerLayout>
  );
}
