"use client";

import React from "react";
import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Link } from "@/i18n/routing";
import {
  TrendingUp,
  Inbox,
  ArrowUpRight,
  Package,
  MessageSquare,
  ShieldCheck,
  Star,
  Users
} from "lucide-react";

export function SupplierDashboardPage() {
  const t = useTranslations("supplier.dashboard");

  const stats = [
    { label: "Total Sales", value: "Rp 128.500.000", desc: "+12.4% from last month", icon: TrendingUp, tone: "success" },
    { label: "Active Products", value: "48 Items", desc: "45 active, 3 draft", icon: Package, tone: "info" },
    { label: "Matching RFQs", value: "14 Open", desc: "3 expiring soon", icon: Inbox, tone: "warning" },
  ];

  const recentRfqs = [
    { id: "RFQ-2026-102", product: "Bentonite Clay Powder", quantity: "20 Ton", date: "2026-06-03", replies: 2 },
    { id: "RFQ-2026-101", product: "Garnet Sand Mesh 80", quantity: "50 Ton", date: "2026-05-30", replies: 5 },
    { id: "RFQ-2026-099", product: "Industrial Quartz Sand", quantity: "15 Ton", date: "2026-05-27", replies: 1 },
  ];

  return (
    <div className="space-y-6">
      {/* Welcome Section */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-border/80 pb-6 text-left">
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
            {t("title")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("subtitle")}
          </p>
        </div>
        <div className="flex gap-3">
          <Button asChild variant="outline" className="cursor-pointer transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-md border-border">
            <Link href="/supplier/products/create">
              <PlusIcon className="mr-2 h-4 w-4" />
              Upload Product
            </Link>
          </Button>
          <Button asChild className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/95 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg">
            <Link href="/supplier/rfq">
              Find RFQs
            </Link>
          </Button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-6">
        {stats.map((stat, idx) => {
          const Icon = stat.icon;
          return (
            <Card key={idx} className="border-border rounded-xl shadow-xs overflow-hidden bg-card text-left">
              <CardContent className="p-6 flex items-center justify-between">
                <div className="space-y-1">
                  <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">{stat.label}</p>
                  <div className="flex items-baseline gap-2 mt-1">
                    <span className="text-2xl font-bold tracking-tight text-foreground">{stat.value}</span>
                  </div>
                  <p className="text-xs text-muted-foreground mt-1">{stat.desc}</p>
                </div>
                <div className={`h-12 w-12 rounded-lg flex items-center justify-center bg-primary/10 text-primary border border-border`}>
                  <Icon className="h-5 w-5" />
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      {/* Sales Performance Graphic Mock */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <Card className="lg:col-span-2 border border-border rounded-xl shadow-xs overflow-hidden bg-card text-left">
          <CardHeader className="border-b border-border bg-muted/20 py-4 px-6">
            <CardTitle className="text-sm font-bold text-foreground">Sales Performance Graph</CardTitle>
          </CardHeader>
          <CardContent className="p-6">
            <div className="h-[200px] flex items-end justify-between gap-2 pt-4">
              {[45, 60, 50, 75, 90, 85, 110, 95, 120, 130, 115, 140].map((val, idx) => (
                <div key={idx} className="flex-1 flex flex-col items-center gap-2">
                  <div
                    className="w-full bg-primary rounded-t-sm hover:opacity-85 transition-opacity"
                    style={{ height: `${(val / 150) * 160}px` }}
                  />
                  <span className="text-[9px] text-muted-foreground font-semibold">
                    {["J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"][idx]}
                  </span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Store Health / Profile Summary */}
        <Card className="border border-border rounded-xl shadow-xs overflow-hidden bg-card text-left">
          <CardHeader className="border-b border-border bg-muted/20 py-4 px-6">
            <CardTitle className="text-sm font-bold text-foreground">Seller Performance Level</CardTitle>
          </CardHeader>
          <CardContent className="p-6 space-y-4">
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 bg-success/10 text-success rounded-full flex items-center justify-center border border-success/20">
                <ShieldCheck className="h-5 w-5" />
              </div>
              <div>
                <h4 className="text-sm font-bold text-foreground">Gold Level Verified</h4>
                <p className="text-xs text-muted-foreground">High trust level on IndoSupplier</p>
              </div>
            </div>

            <div className="h-px bg-border" />

            <div className="space-y-3">
              <div className="flex items-center justify-between text-xs font-semibold">
                <span className="text-muted-foreground flex items-center gap-1">
                  <Star className="h-3.5 w-3.5 fill-amber-400 stroke-amber-400" /> Store Rating
                </span>
                <span className="text-foreground">4.8 / 5.0</span>
              </div>
              <div className="flex items-center justify-between text-xs font-semibold">
                <span className="text-muted-foreground flex items-center gap-1">
                  <MessageSquare className="h-3.5 w-3.5 text-muted-foreground" /> Chat Response Rate
                </span>
                <span className="text-foreground">98.5% (Very Fast)</span>
              </div>
              <div className="flex items-center justify-between text-xs font-semibold">
                <span className="text-muted-foreground flex items-center gap-1">
                  <Users className="h-3.5 w-3.5 text-muted-foreground" /> Catalog Visitors
                </span>
                <span className="text-foreground">1,420 / month</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Matching RFQ Inquiries */}
      <Card className="border border-border rounded-xl shadow-xs overflow-hidden bg-card text-left">
        <CardHeader className="flex flex-row items-center justify-between py-4 px-6 border-b border-border bg-muted/20">
          <CardTitle className="text-sm font-bold text-foreground">{t("recentRfqs")}</CardTitle>
          <Button asChild variant="ghost" size="sm" className="text-xs font-semibold text-primary cursor-pointer p-0 h-auto hover:bg-transparent transition-colors">
            <Link href="/supplier/rfq" className="flex items-center gap-1">
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
                      <Link href={`/supplier/rfq/${rfq.id}`}>{rfq.product}</Link>
                    </h3>
                  </div>
                  <p className="text-xs text-muted-foreground">Requested Quantity: <span className="font-semibold text-foreground">{rfq.quantity}</span> • Received Bids: {rfq.replies}</p>
                </div>
                <div className="flex items-center justify-between sm:justify-end gap-3 shrink-0">
                  <Badge className="bg-primary/10 text-primary border border-primary/20 rounded-full text-[10px] font-semibold">
                    Matching Category
                  </Badge>
                  <Button asChild variant="outline" size="sm" className="text-xs font-semibold cursor-pointer h-8 border-border">
                    <Link href={`/supplier/rfq/${rfq.id}`}>
                      Submit Quote
                    </Link>
                  </Button>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function PlusIcon(props: React.SVGProps<SVGSVGElement>) {
  return (
    <svg
      {...props}
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M5 12h14" />
      <path d="M12 5v14" />
    </svg>
  );
}
