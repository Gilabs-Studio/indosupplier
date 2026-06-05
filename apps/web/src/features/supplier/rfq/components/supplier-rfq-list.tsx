"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { FileText, Search, ArrowUpRight, Calendar, MapPin } from "lucide-react";

export function SupplierRfqList() {
  const t = useTranslations("supplier.rfq");
  const [search, setSearch] = useState("");

  const [rfqs] = useState([
    { id: "RFQ-2026-102", product: "Bentonite Clay Powder", category: "Industrial Minerals", quantity: "20 Ton", port: "Tanjung Perak, Surabaya", date: "2026-06-03", budget: "Rp 4.000.000 / Ton", status: "open" },
    { id: "RFQ-2026-101", product: "Garnet Sand Mesh 80", category: "Industrial Minerals", quantity: "50 Ton", port: "Tanjung Priok, Jakarta", date: "2026-05-30", budget: "Rp 3.300.000 / Ton", status: "open" },
    { id: "RFQ-2026-099", product: "Industrial Quartz Sand", category: "Industrial Minerals", quantity: "15 Ton", port: "Tanjung Priok, Jakarta", date: "2026-05-27", budget: "Rp 2.700.000 / Ton", status: "open" },
    { id: "RFQ-2026-095", product: "Organic Coconut Sugar", category: "Agriculture", quantity: "5 Ton", port: "Port of Rotterdam (CIF)", date: "2026-05-18", budget: "Rp 22.000.000 / Ton", status: "closed" },
  ]);

  const filtered = rfqs.filter(r =>
    r.product.toLowerCase().includes(search.toLowerCase()) ||
    r.id.toLowerCase().includes(search.toLowerCase())
  );

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

      {/* Filter Section */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="relative max-w-xs w-full">
          <Search className="absolute left-3 top-2.5 h-4.5 w-4.5 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search RFQs..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-9 pr-4 py-1.5 bg-card border border-border text-sm rounded-lg outline-hidden focus:border-primary transition-all text-left"
          />
        </div>
      </div>

      {/* Table */}
      <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow className="border-b border-border bg-muted/10">
                  <TableHead className="font-bold text-foreground py-3 px-6">{t("tableId")}</TableHead>
                  <TableHead className="font-bold text-foreground">{t("tableProduct")}</TableHead>
                  <TableHead className="font-bold text-foreground">{t("tableQty")}</TableHead>
                  <TableHead className="font-bold text-foreground">{t("tablePort")}</TableHead>
                  <TableHead className="font-bold text-foreground">{t("tableDate")}</TableHead>
                  <TableHead className="font-bold text-foreground">Status</TableHead>
                  <TableHead className="font-bold text-foreground text-right px-6">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filtered.map((r) => (
                  <TableRow key={r.id} className="hover:bg-muted/10 border-b border-border transition-colors">
                    <TableCell className="py-4 px-6 font-bold text-muted-foreground">{r.id}</TableCell>
                    <TableCell className="py-4">
                      <div className="flex items-center gap-2 font-semibold text-foreground">
                        <FileText className="h-4.5 w-4.5 text-primary shrink-0" />
                        <span>{r.product}</span>
                      </div>
                      <p className="text-[10px] text-muted-foreground mt-0.5">{r.category}</p>
                    </TableCell>
                    <TableCell className="py-4">
                      <Badge variant="outline" className="bg-primary/5 text-primary border-primary/20">{r.quantity}</Badge>
                    </TableCell>
                    <TableCell className="py-4 text-xs text-muted-foreground font-semibold">
                      <span className="flex items-center gap-1">
                        <MapPin className="h-3.5 w-3.5" />
                        {r.port}
                      </span>
                    </TableCell>
                    <TableCell className="py-4 text-xs font-semibold text-muted-foreground">
                      <span className="flex items-center gap-1">
                        <Calendar className="h-3.5 w-3.5" />
                        {r.date}
                      </span>
                    </TableCell>
                    <TableCell className="py-4">
                      {r.status === "open" ? (
                        <Badge className="bg-success/15 text-success border border-success/30 font-bold">Open</Badge>
                      ) : (
                        <Badge variant="secondary" className="font-bold">Closed</Badge>
                      )}
                    </TableCell>
                    <TableCell className="py-4 px-6 text-right space-x-2">
                      <Button asChild size="sm" variant={r.status === "open" ? "default" : "outline"} disabled={r.status !== "open"} className="text-xs font-semibold h-8 cursor-pointer border-border">
                        <Link href={`/supplier/rfq/${r.id}`}>
                          {t("actionSubmit")}
                          <ArrowUpRight className="ml-1 h-3.5 w-3.5" />
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
