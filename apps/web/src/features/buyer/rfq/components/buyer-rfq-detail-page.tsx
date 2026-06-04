"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { BuyerLayout } from "../../components/buyer-layout";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Link } from "@/i18n/routing";
import {
  ArrowLeft,
  Calendar,
  MapPin,
  MessageSquare,
  ShieldCheck,
  Check,
  FileText,
} from "lucide-react";

interface BuyerRfqDetailPageProps {
  readonly id: string;
}

export function BuyerRfqDetailPage({ id }: BuyerRfqDetailPageProps) {
  const t = useTranslations("buyer.rfqDetail");
  const [activeTab, setActiveTab] = useState("quotes");

  const rfqDetails = {
    id: id || "RFQ-2026-003",
    product: "Bentonite Clay Powder",
    category: "Chemicals",
    quantity: "20 Ton",
    targetPort: "Tanjung Perak, Surabaya",
    date: "2026-05-28",
    status: "Offers Received",
    description: "Membutuhkan Bentonite Clay Powder kualitas industri untuk konstruksi sipil. Kemasan bag 25kg, total kebutuhan 20 ton dikirim ke pelabuhan Surabaya. Sertakan Certificate of Analysis (COA) terbaru.",
  };

  const bids = [
    {
      id: "bid-1",
      supplierName: "PT Rempah Nusantara",
      price: "Rp 12.500 / Kg",
      moq: "5 Ton",
      responseTime: "2 Jam",
      verified: true,
    },
    {
      id: "bid-2",
      supplierName: "CV Indo Mineral Utama",
      price: "Rp 11.800 / Kg",
      moq: "10 Ton",
      responseTime: "3 Jam",
      verified: true,
    },
    {
      id: "bid-3",
      supplierName: "PT Sumber Batuan Jaya",
      price: "Rp 13.000 / Kg",
      moq: "1 Ton",
      responseTime: "5 Jam",
      verified: false,
    },
  ];

  return (
    <BuyerLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="space-y-2">
          <Link href="/rfq" className="inline-flex items-center gap-1 text-xs font-semibold text-primary hover:underline cursor-pointer">
            <ArrowLeft className="h-3.5 w-3.5" /> Kembali ke Daftar RFQ
          </Link>
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
            <div>
              <div className="flex items-center gap-2 flex-wrap">
                <span className="text-xs font-bold text-muted-foreground">{rfqDetails.id}</span>
                <Badge variant="outline" className="bg-success/10 text-success border-success/20 rounded-full text-[10px]">
                  {rfqDetails.status}
                </Badge>
              </div>
              <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading mt-1">{rfqDetails.product}</h1>
            </div>
          </div>
        </div>

        {/* Tabs */}
        <div className="flex gap-1 border-b border-border">
          <button
            onClick={() => setActiveTab("quotes")}
            className={`px-4 py-2 text-sm font-semibold rounded-t-lg transition-all cursor-pointer ${
              activeTab === "quotes"
                ? "border-b-2 border-primary text-primary font-bold bg-muted/20"
                : "text-muted-foreground hover:text-foreground"
            }`}
          >
            {t("tabQuotes", { count: bids.length })}
          </button>
          <button
            onClick={() => setActiveTab("details")}
            className={`px-4 py-2 text-sm font-semibold rounded-t-lg transition-all cursor-pointer ${
              activeTab === "details"
                ? "border-b-2 border-primary text-primary font-bold bg-muted/20"
                : "text-muted-foreground hover:text-foreground"
            }`}
          >
            {t("tabDetails")}
          </button>
        </div>

        {activeTab === "quotes" ? (
          <div className="space-y-4">
            {bids.map((bid) => (
              <Card key={bid.id} className="border border-border rounded-xl bg-card shadow-xs overflow-hidden hover:shadow-md transition-shadow">
                <CardContent className="p-5 flex flex-col md:flex-row md:items-center md:justify-between gap-4">
                  <div className="space-y-2">
                    <div className="flex items-center gap-1.5">
                      <h4 className="text-sm font-bold text-foreground">{bid.supplierName}</h4>
                      {bid.verified && (
                        <Badge className="bg-success text-white border-0 text-[9px] px-1.5 py-0.5 rounded-full flex items-center gap-0.5">
                          <ShieldCheck className="h-2.5 w-2.5" /> Verified
                        </Badge>
                      )}
                    </div>
                    
                    <div className="grid grid-cols-2 sm:grid-cols-3 gap-x-6 gap-y-1 text-xs text-muted-foreground">
                      <div>
                        <span>{t("offerPrice")}: </span>
                        <strong className="text-foreground text-sm font-bold block mt-0.5">{bid.price}</strong>
                      </div>
                      <div>
                        <span>{t("moq")}: </span>
                        <strong className="text-foreground text-sm font-bold block mt-0.5">{bid.moq}</strong>
                      </div>
                      <div>
                        <span>Respon Rate: </span>
                        <strong className="text-foreground text-sm font-bold block mt-0.5">{bid.responseTime}</strong>
                      </div>
                    </div>
                  </div>

                  <div className="flex gap-2 shrink-0 md:self-center">
                    <Button asChild variant="outline" size="sm" className="text-xs font-semibold cursor-pointer border-border hover:border-muted-foreground transition-all hover:-translate-y-0.5 active:translate-y-0">
                      <Link href="/search">
                        <MessageSquare className="mr-1.5 h-3.5 w-3.5 text-muted-foreground" />
                        {t("btnChat")}
                      </Link>
                    </Button>
                    <Button
                      onClick={() => alert("Penawaran dari " + bid.supplierName + " berhasil disetujui! Tim sales kami akan menghubungi Anda.")}
                      size="sm"
                      className="text-xs font-semibold cursor-pointer bg-success hover:bg-success/90 text-white transition-all hover:-translate-y-0.5 active:translate-y-0"
                    >
                      <Check className="mr-1.5 h-3.5 w-3.5" />
                      {t("btnAccept")}
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        ) : (
          <Card className="border border-border rounded-xl bg-card shadow-xs overflow-hidden">
            <CardContent className="p-6 space-y-6">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-6 pb-6 border-b border-border">
                <div className="space-y-1">
                  <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">Kategori</span>
                  <p className="text-sm font-semibold text-foreground">{rfqDetails.category}</p>
                </div>
                <div className="space-y-1">
                  <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">Volume Kebutuhan</span>
                  <p className="text-sm font-semibold text-foreground">{rfqDetails.quantity}</p>
                </div>
                <div className="space-y-1">
                  <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">Tujuan Pengiriman</span>
                  <p className="text-sm font-semibold text-foreground flex items-center gap-1">
                    <MapPin className="h-4 w-4 text-muted-foreground" />
                    {rfqDetails.targetPort}
                  </p>
                </div>
                <div className="space-y-1">
                  <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">Tanggal Dibuat</span>
                  <p className="text-sm font-semibold text-foreground flex items-center gap-1">
                    <Calendar className="h-4 w-4 text-muted-foreground" />
                    {rfqDetails.date}
                  </p>
                </div>
              </div>

              <div className="space-y-2">
                <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">Deskripsi Kebutuhan</span>
                <p className="text-sm text-foreground leading-relaxed whitespace-pre-wrap">{rfqDetails.description}</p>
              </div>

              <div className="space-y-2">
                <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">Lampiran Dokumen</span>
                <div className="flex items-center gap-2 p-3 border border-border rounded-lg bg-muted/20 w-fit">
                  <FileText className="h-5 w-5 text-primary" />
                  <div className="text-xs">
                    <p className="font-semibold text-foreground">COA_Bentonite_Clay_2026.pdf</p>
                    <p className="text-[10px] text-muted-foreground">1.4 MB</p>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </BuyerLayout>
  );
}
