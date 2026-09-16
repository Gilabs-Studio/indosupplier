"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { FileText, Search, ArrowUpRight, Calendar, MapPin, Package } from "lucide-react";
import { CenteredLoading } from "@/components/loading";
import { resolveImageUrl } from "@/lib/utils";
import { useSupplierRfqs } from "../hooks/useSupplierRfqs";

export function SupplierRfqList() {
  const t = useTranslations("supplier.rfq");
  const [search, setSearch] = useState("");
  const { data, isLoading } = useSupplierRfqs({ page: 1, per_page: 20 });
  const rfqs = data?.items || [];

  const filtered = rfqs.filter(r =>
    r.product.toLowerCase().includes(search.toLowerCase()) ||
    r.category.toLowerCase().includes(search.toLowerCase()) ||
    r.port.toLowerCase().includes(search.toLowerCase())
  );

  if (isLoading) {
    return <CenteredLoading />;
  }

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
                <TableRow className="border-b border-border text-muted-foreground text-[11px] font-medium uppercase tracking-wider">
                  <TableHead className="py-3 px-6">{t("tableProduct")}</TableHead>
                  <TableHead className="py-3 px-4">{t("tableQty")}</TableHead>
                  <TableHead className="py-3 px-4">{t("tablePort")}</TableHead>
                  <TableHead className="py-3 px-4">Status</TableHead>
                  <TableHead className="py-3 px-6 text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody className="divide-y divide-border/60">
                {filtered.map((r) => {
                  const defaultThumbnail = "/images/categories/cat-bahan-baku.webp";
                  const imageSrc = resolveImageUrl(r.imageUrl || defaultThumbnail);

                  return (
                    <TableRow key={r.id} className="hover:bg-muted/20 border-b border-border/60 transition-colors">
                      <TableCell className="py-3.5 px-6">
                        <div className="flex items-center gap-3">
                          <div className="h-11 w-11 shrink-0 rounded-lg overflow-hidden border border-border/80 bg-muted/20">
                            <img
                              src={imageSrc}
                              alt={r.product}
                              className="h-full w-full object-cover"
                              onError={(e) => {
                                e.currentTarget.src = defaultThumbnail;
                              }}
                            />
                          </div>
                          <div className="min-w-0 max-w-md space-y-0.5">
                            <Link
                              href={`/supplier/rfq/${r.id}`}
                              className="font-medium text-sm text-foreground hover:text-primary transition-colors cursor-pointer line-clamp-1"
                            >
                              {r.product}
                            </Link>
                            <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                              <span>{r.category}</span>
                              <span>•</span>
                              <span>{r.date}</span>
                            </div>
                          </div>
                        </div>
                      </TableCell>
                      <TableCell className="py-3.5 px-4 text-sm text-foreground whitespace-nowrap">
                        {r.quantity}
                      </TableCell>
                      <TableCell className="py-3.5 px-4 text-xs text-muted-foreground max-w-xs truncate">
                        {r.port}
                      </TableCell>
                      <TableCell className="py-3.5 px-4 whitespace-nowrap">
                        {r.status === "open" || r.status === "new" ? (
                          <div className="flex items-center gap-1.5 text-xs font-medium text-foreground">
                            <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 shrink-0" />
                            <span>Open</span>
                          </div>
                        ) : r.status === "accepted" ? (
                          <div className="flex items-center gap-1.5 text-xs font-medium text-foreground">
                            <span className="h-1.5 w-1.5 rounded-full bg-primary shrink-0" />
                            <span>Diterima</span>
                          </div>
                        ) : (
                          <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                            <span className="h-1.5 w-1.5 rounded-full bg-muted-foreground/30 shrink-0" />
                            <span className="capitalize">{r.status}</span>
                          </div>
                        )}
                      </TableCell>
                      <TableCell className="py-3.5 px-6 text-right whitespace-nowrap">
                        <Link
                          href={`/supplier/rfq/${r.id}`}
                          className="inline-flex items-center justify-end gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors cursor-pointer py-1 px-2 rounded-md hover:bg-muted/40"
                        >
                          <span>{r.status === "open" || r.status === "new" ? t("actionSubmit") : "Lihat Detail"}</span>
                          <ArrowUpRight className="ml-1 h-3.5 w-3.5" />
                        </Link>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
