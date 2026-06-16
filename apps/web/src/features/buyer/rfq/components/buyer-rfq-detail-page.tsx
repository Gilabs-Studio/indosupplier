"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { BuyerLayout } from "../../components/buyer-layout";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Link } from "@/i18n/routing";
import { CenteredLoading } from "@/components/loading";
import { cn } from "@/lib/utils";
import {
  ArrowLeft,
  Calendar,
  MapPin,
  MessageSquare,
  ShieldCheck,
  Check,
  FileText,
} from "lucide-react";
import { useBuyerRfqDetail } from "../hooks/useBuyerRfqs";

interface BuyerRfqDetailPageProps {
  readonly id: string;
}

export function BuyerRfqDetailPage({ id }: BuyerRfqDetailPageProps) {
  const t = useTranslations("buyerRfq.rfqDetail");
  const [activeTab, setActiveTab] = useState("quotes");

  const {
    rfq,
    isLoadingRfq,
    bids,
    isLoadingBids,
    acceptBid,
    isAccepting,
  } = useBuyerRfqDetail(id);

  if (isLoadingRfq || isLoadingBids) {
    return (
      <BuyerLayout>
        <CenteredLoading />
      </BuyerLayout>
    );
  }

  if (!rfq) {
    return (
      <BuyerLayout>
        <div className="text-center py-20 bg-card rounded-xl border border-border">
          <p className="text-destructive font-semibold">{t("notFound")}</p>
        </div>
      </BuyerLayout>
    );
  }

  return (
    <BuyerLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="space-y-2">
          <Link href="/rfq" className="inline-flex items-center gap-1 text-xs font-semibold text-primary hover:underline cursor-pointer">
            <ArrowLeft className="h-3.5 w-3.5" /> {t("backLink")}
          </Link>
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
            <div>
              <div className="flex items-center gap-2 flex-wrap">
                <span className="text-xs font-bold text-muted-foreground">{rfq.id}</span>
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
              </div>
              <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading mt-1">{rfq.product}</h1>
            </div>
          </div>
        </div>

        {/* Tabs */}
        <div className="flex gap-1.5 border-b border-border pb-1">
          <button
            onClick={() => setActiveTab("quotes")}
            className={cn(
              "px-4 py-2 text-sm font-medium transition-all duration-200 whitespace-nowrap cursor-pointer hover:text-primary hover:-translate-y-0.5 active:translate-y-0",
              activeTab === "quotes"
                ? "text-primary font-semibold border-b-2 border-primary -mb-[5px]"
                : "text-muted-foreground"
            )}
          >
            {t("tabQuotes", { count: bids.length })}
          </button>
          <button
            onClick={() => setActiveTab("details")}
            className={cn(
              "px-4 py-2 text-sm font-medium transition-all duration-200 whitespace-nowrap cursor-pointer hover:text-primary hover:-translate-y-0.5 active:translate-y-0",
              activeTab === "details"
                ? "text-primary font-semibold border-b-2 border-primary -mb-[5px]"
                : "text-muted-foreground"
            )}
          >
            {t("tabDetails")}
          </button>
        </div>

        {activeTab === "quotes" ? (
          <div className="space-y-4">
            {bids.length === 0 ? (
              <p className="text-sm text-muted-foreground text-center py-10 bg-card rounded-xl border border-border">
                {t("emptyBids")}
              </p>
            ) : (
              bids.map((bid) => (
                <Card key={bid.id} className="border border-border rounded-xl bg-card shadow-xs overflow-hidden hover:shadow-md transition-shadow">
                  <CardContent className="p-5 flex flex-col md:flex-row md:items-center md:justify-between gap-4">
                    <div className="space-y-2">
                      <div className="flex items-center gap-1.5">
                        <h4 className="text-sm font-bold text-foreground">{bid.supplierName}</h4>
                        {bid.verified && (
                          <Badge className="bg-success text-white border-0 text-[9px] px-1.5 py-0.5 rounded-full flex items-center gap-0.5">
                            <ShieldCheck className="h-2.5 w-2.5" /> {t("verified")}
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
                          <span>{t("responseRate")}: </span>
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
                        onClick={() => acceptBid(bid.id)}
                        disabled={isAccepting}
                        size="sm"
                        className="text-xs font-semibold cursor-pointer bg-success hover:bg-success/90 text-white transition-all hover:-translate-y-0.5 active:translate-y-0 shadow-lg hover:shadow-success/20"
                      >
                        <Check className="mr-1.5 h-3.5 w-3.5" />
                        {t("btnAccept")}
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              ))
            )}
          </div>
        ) : (
          <Card className="border border-border rounded-xl bg-card shadow-xs overflow-hidden">
            <CardContent className="p-6 space-y-6">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-6 pb-6 border-b border-border">
                <div className="space-y-1">
                  <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">{t("labelCategory")}</span>
                  <p className="text-sm font-semibold text-foreground">{rfq.category}</p>
                </div>
                <div className="space-y-1">
                  <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">{t("labelVolume")}</span>
                  <p className="text-sm font-semibold text-foreground">{rfq.quantity}</p>
                </div>
                <div className="space-y-1">
                  <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">{t("labelDestination")}</span>
                  <p className="text-sm font-semibold text-foreground flex items-center gap-1">
                    <MapPin className="h-4 w-4 text-muted-foreground" />
                    {rfq.targetPort}
                  </p>
                </div>
                <div className="space-y-1">
                  <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">{t("labelCreatedAt")}</span>
                  <p className="text-sm font-semibold text-foreground flex items-center gap-1">
                    <Calendar className="h-4 w-4 text-muted-foreground" />
                    {rfq.date}
                  </p>
                </div>
              </div>

              <div className="space-y-2">
                <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">{t("labelDescription")}</span>
                <p className="text-sm text-foreground leading-relaxed whitespace-pre-wrap">{rfq.description || "-"}</p>
              </div>

              {rfq.attachmentUrl && (
                <div className="space-y-2">
                  <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">{t("labelAttachment")}</span>
                  <a href={rfq.attachmentUrl} target="_blank" rel="noopener noreferrer" className="flex items-center gap-2 p-3 border border-border rounded-lg bg-muted/20 w-fit hover:bg-muted/30 transition-colors">
                    <FileText className="h-5 w-5 text-primary" />
                    <div className="text-xs">
                      <p className="font-semibold text-foreground">{rfq.attachmentName || "Spesifikasi_Teknis.pdf"}</p>
                      <p className="text-[10px] text-muted-foreground">{rfq.attachmentSize || "1.4 MB"}</p>
                    </div>
                  </a>
                </div>
              )}
            </CardContent>
          </Card>
        )}
      </div>
    </BuyerLayout>
  );
}
