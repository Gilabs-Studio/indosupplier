"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { BuyerLayout } from "../../components/buyer-layout";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Link } from "@/i18n/routing";
import {
  ShieldCheck,
  Star,
  MapPin,
  MessageSquare,
  X,
  Plus,
} from "lucide-react";

export function BuyerComparePage() {
  const t = useTranslations("buyer.compare");

  const [suppliers, setSuppliers] = useState([
    {
      id: "1",
      companyName: "PT Rempah Nusantara",
      location: "Surabaya, Jawa Timur",
      businessType: "Manufacturer",
      establishedYear: 2012,
      rating: 4.8,
      reviewCount: 128,
      verified: true,
      moq: "500 Kg",
      responseTime: "2 Jam",
      capacity: "20 Ton / Bulan",
      certifications: ["BPOM", "Halal", "HACCP"],
    },
    {
      id: "2",
      companyName: "CV Nusantara Garment",
      location: "Bandung, Jawa Barat",
      businessType: "Manufacturer",
      establishedYear: 2015,
      rating: 4.6,
      reviewCount: 94,
      verified: true,
      moq: "100 Pcs",
      responseTime: "4 Jam",
      capacity: "10,000 Pcs / Bulan",
      certifications: ["SNI", "OEKO-TEX"],
    },
  ]);

  const handleRemove = (id: string) => {
    setSuppliers(suppliers.filter((s) => s.id !== id));
  };

  return (
    <BuyerLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">{t("title")}</h1>
          <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
        </div>

        {suppliers.length === 0 ? (
          <div className="text-center py-20 bg-card rounded-xl border border-border">
            <h3 className="text-base font-semibold text-foreground">{t("emptyTitle")}</h3>
            <p className="mt-2 text-sm text-muted-foreground max-w-xs mx-auto">
              {t("emptyDesc")}
            </p>
            <div className="mt-6 flex justify-center gap-3">
              <Button asChild variant="outline" className="cursor-pointer transition-all hover:-translate-y-0.5 active:translate-y-0 shadow-xs">
                <Link href="/bookmarks">{t("btnBack")}</Link>
              </Button>
              <Button asChild className="cursor-pointer transition-all hover:-translate-y-0.5 active:translate-y-0 shadow-lg hover:shadow-primary/20">
                <Link href="/search">Cari Supplier</Link>
              </Button>
            </div>
          </div>
        ) : (
          <div className="overflow-x-auto border border-border rounded-xl bg-card shadow-xs">
            <table className="w-full border-collapse text-left text-sm text-foreground">
              <thead>
                <tr className="border-b border-border bg-muted/20">
                  <th className="p-4 font-bold text-muted-foreground w-1/4">{t("criteria")}</th>
                  {suppliers.map((supplier) => (
                    <th key={supplier.id} className="p-4 font-bold relative w-1/3 min-w-[240px]">
                      <div className="flex items-start justify-between gap-4">
                        <div className="space-y-1">
                          <h3 className="font-bold text-foreground text-sm tracking-tight">{supplier.companyName}</h3>
                          <div className="flex items-center gap-1.5 text-xs font-normal text-muted-foreground">
                            <MapPin className="h-3.5 w-3.5 shrink-0" />
                            <span>{supplier.location.split(",")[0]}</span>
                          </div>
                        </div>
                        <Button
                          onClick={() => handleRemove(supplier.id)}
                          variant="ghost"
                          size="icon"
                          className="h-7 w-7 text-muted-foreground hover:text-destructive hover:bg-secondary rounded-full cursor-pointer absolute right-2 top-2 transition-colors"
                        >
                          <X className="h-3.5 w-3.5" />
                        </Button>
                      </div>
                    </th>
                  ))}
                  {suppliers.length < 3 && (
                    <th className="p-4 text-center border-l border-dashed border-border w-1/4 min-w-[200px] bg-muted/5">
                      <Button asChild variant="outline" className="border-dashed hover:border-solid cursor-pointer transition-all hover:-translate-y-0.5 active:translate-y-0">
                        <Link href="/bookmarks">
                          <Plus className="mr-1 h-4 w-4" /> Tambah Supplier
                        </Link>
                      </Button>
                    </th>
                  )}
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {/* Rating */}
                <tr>
                  <td className="p-4 font-semibold text-muted-foreground">{t("rating")}</td>
                  {suppliers.map((s) => (
                    <td key={s.id} className="p-4">
                      <div className="flex items-center gap-1 text-sm font-semibold text-foreground">
                        <Star className="h-4 w-4 fill-amber-400 stroke-amber-400" />
                        <span>{s.rating}</span>
                        <span className="text-muted-foreground font-normal text-xs">({s.reviewCount} ulasan)</span>
                      </div>
                    </td>
                  ))}
                  {suppliers.length < 3 && <td className="p-4 border-l border-dashed border-border bg-muted/5" />}
                </tr>

                {/* Verification */}
                <tr>
                  <td className="p-4 font-semibold text-muted-foreground">{t("verification")}</td>
                  {suppliers.map((s) => (
                    <td key={s.id} className="p-4">
                      {s.verified ? (
                        <Badge className="bg-success text-white border-0 text-[10px] px-2 py-0.5 rounded-full flex items-center gap-1 w-fit">
                          <ShieldCheck className="h-3 w-3" /> Terverifikasi
                        </Badge>
                      ) : (
                        <Badge variant="secondary" className="text-[10px] px-2 py-0.5 rounded-full w-fit">Belum Terverifikasi</Badge>
                      )}
                    </td>
                  ))}
                  {suppliers.length < 3 && <td className="p-4 border-l border-dashed border-border bg-muted/5" />}
                </tr>

                {/* MOQ */}
                <tr>
                  <td className="p-4 font-semibold text-muted-foreground">{t("moq")}</td>
                  {suppliers.map((s) => (
                    <td key={s.id} className="p-4 font-medium text-foreground">{s.moq}</td>
                  ))}
                  {suppliers.length < 3 && <td className="p-4 border-l border-dashed border-border bg-muted/5" />}
                </tr>

                {/* Production Capacity */}
                <tr>
                  <td className="p-4 font-semibold text-muted-foreground">{t("capacity")}</td>
                  {suppliers.map((s) => (
                    <td key={s.id} className="p-4 text-foreground">{s.capacity}</td>
                  ))}
                  {suppliers.length < 3 && <td className="p-4 border-l border-dashed border-border bg-muted/5" />}
                </tr>

                {/* Response Time */}
                <tr>
                  <td className="p-4 font-semibold text-muted-foreground">{t("response")}</td>
                  {suppliers.map((s) => (
                    <td key={s.id} className="p-4 text-foreground">{s.responseTime}</td>
                  ))}
                  {suppliers.length < 3 && <td className="p-4 border-l border-dashed border-border bg-muted/5" />}
                </tr>

                {/* Certifications */}
                <tr>
                  <td className="p-4 font-semibold text-muted-foreground">{t("certifications")}</td>
                  {suppliers.map((s) => (
                    <td key={s.id} className="p-4">
                      <div className="flex flex-wrap gap-1">
                        {s.certifications.map((c, i) => (
                          <Badge key={i} variant="outline" className="text-[10px] border-border text-muted-foreground rounded-lg">
                            {c}
                          </Badge>
                        ))}
                      </div>
                    </td>
                  ))}
                  {suppliers.length < 3 && <td className="p-4 border-l border-dashed border-border bg-muted/5" />}
                </tr>

                {/* Actions */}
                <tr className="bg-muted/5">
                  <td className="p-4 font-semibold text-muted-foreground">{t("action")}</td>
                  {suppliers.map((s) => (
                    <td key={s.id} className="p-4">
                      <div className="flex gap-2">
                        <Button asChild size="sm" className="bg-primary text-primary-foreground hover:bg-primary/95 text-xs font-semibold cursor-pointer transition-all hover:-translate-y-0.5 active:translate-y-0">
                          <Link href="/rfq/create">
                            <MessageSquare className="mr-1 h-3.5 w-3.5" /> {t("sendRfq")}
                          </Link>
                        </Button>
                        <Button asChild size="sm" variant="outline" className="text-xs font-semibold border-border hover:border-muted-foreground cursor-pointer transition-all hover:-translate-y-0.5 active:translate-y-0">
                          <Link href="/search">Detail</Link>
                        </Button>
                      </div>
                    </td>
                  ))}
                  {suppliers.length < 3 && <td className="p-4 border-l border-dashed border-border bg-muted/5" />}
                </tr>
              </tbody>
            </table>
          </div>
        )}
      </div>
    </BuyerLayout>
  );
}
