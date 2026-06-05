"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { toast } from "sonner";
import { Megaphone, Plus, Search, Eye, TrendingUp, MousePointer, CreditCard } from "lucide-react";

export function SupplierAdsList() {
  const t = useTranslations("supplier.ads");
  const [search, setSearch] = useState("");

  const [campaigns, setCampaigns] = useState([
    { id: "CAM-01", name: "Garnet Sand Search Ads", product: "Garnet Sand Mesh 80", budget: "Rp 150.000 / Day", bid: "Rp 1.500", status: "active", clicks: 342, impressions: 8540, spend: "Rp 513.000" },
    { id: "CAM-02", name: "Bentonite Banner Promo", product: "Bentonite Clay Powder", budget: "Rp 200.000 / Day", bid: "Rp 2.000", status: "paused", clicks: 120, impressions: 5600, spend: "Rp 240.000" },
    { id: "CAM-03", name: "Quartz Minerals Boost", product: "Quartz Powder 325 Mesh", budget: "Rp 100.000 / Day", bid: "Rp 1.200", status: "ended", clicks: 450, impressions: 12400, spend: "Rp 540.000" },
  ]);

  const handleToggleStatus = (id: string, currentStatus: string) => {
    const nextStatus = currentStatus === "active" ? "paused" : "active";
    setCampaigns(campaigns.map(c => c.id === id ? { ...c, status: nextStatus } : c));
    toast.success(`Campaign status updated to ${nextStatus}!`);
  };

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

  const filtered = campaigns.filter(c =>
    c.name.toLowerCase().includes(search.toLowerCase()) ||
    c.product.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="space-y-6 text-left">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-border/80 pb-6">
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
            {t("listTitle")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("listSubtitle")}
          </p>
        </div>
        <Button asChild className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/95 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/20 font-semibold">
          <Link href="/supplier/ads/create">
            <Plus className="mr-2 h-4 w-4" /> {t("btnCreate")}
          </Link>
        </Button>
      </div>

      {/* Metrics Section */}
      <div className="grid gap-4 grid-cols-2 md:grid-cols-4">
        <Card className="border-border shadow-xs rounded-xl bg-card">
          <CardContent className="p-4 flex items-center justify-between">
            <div>
              <p className="text-[10px] font-bold text-muted-foreground uppercase">{t("statImpressions")}</p>
              <h3 className="text-xl font-bold tracking-tight text-foreground mt-1">26,540</h3>
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
              <h3 className="text-xl font-bold tracking-tight text-foreground mt-1">912</h3>
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
              <h3 className="text-xl font-bold tracking-tight text-foreground mt-1">3.44%</h3>
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
              <h3 className="text-xl font-bold tracking-tight text-foreground mt-1">Rp 1.293.000</h3>
            </div>
            <div className="h-10 w-10 bg-purple/10 text-purple border border-border rounded-lg flex items-center justify-center">
              <CreditCard className="h-4.5 w-4.5" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Filter and Table */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="relative max-w-xs w-full">
          <Search className="absolute left-3 top-2.5 h-4.5 w-4.5 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search campaigns..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-9 pr-4 py-1.5 bg-card border border-border text-sm rounded-lg outline-hidden focus:border-primary transition-all text-left"
          />
        </div>
      </div>

      <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow className="border-b border-border bg-muted/10">
                  <TableHead className="font-bold text-foreground py-3 px-6">{t("tableCampaign")}</TableHead>
                  <TableHead className="font-bold text-foreground">{t("tableProduct")}</TableHead>
                  <TableHead className="font-bold text-foreground">{t("tableBudget")}</TableHead>
                  <TableHead className="font-bold text-foreground">{t("tableBid")}</TableHead>
                  <TableHead className="font-bold text-foreground">Status</TableHead>
                  <TableHead className="font-bold text-foreground text-right px-6">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filtered.map((c) => (
                  <TableRow key={c.id} className="hover:bg-muted/10 border-b border-border transition-colors">
                    <TableCell className="py-4 px-6">
                      <Link href={`/supplier/ads/${c.id}`} className="font-bold text-foreground hover:text-primary transition-colors flex items-center gap-2">
                        <Megaphone className="h-4.5 w-4.5 text-muted-foreground shrink-0" />
                        {c.name}
                      </Link>
                      <p className="text-[10px] text-muted-foreground mt-0.5">{c.id}</p>
                    </TableCell>
                    <TableCell className="py-4 font-semibold text-xs text-muted-foreground">{c.product}</TableCell>
                    <TableCell className="py-4 font-semibold text-xs">{c.budget}</TableCell>
                    <TableCell className="py-4 font-semibold text-xs">{c.bid}</TableCell>
                    <TableCell className="py-4">{getStatusBadge(c.status)}</TableCell>
                    <TableCell className="py-4 px-6 text-right space-x-1.5">
                      {c.status !== "ended" && (
                        <Button onClick={() => handleToggleStatus(c.id, c.status)} variant="outline" size="sm" className="text-xs h-8 cursor-pointer border-border font-semibold">
                          {c.status === "active" ? "Pause" : "Activate"}
                        </Button>
                      )}
                      <Button asChild variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground hover:text-primary cursor-pointer hover:bg-primary/5 border border-border">
                        <Link href={`/supplier/ads/${c.id}`}>
                          <Eye className="h-3.5 w-3.5" />
                        </Link>
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
