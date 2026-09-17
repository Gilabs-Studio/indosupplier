"use client";

import React, { useState, useEffect } from "react";
import { useTranslations } from "next-intl";
import { BuyerLayout } from "../../components/buyer-layout";
import { Card, CardContent } from "@/components/ui/card";
import { Link } from "@/i18n/routing";
import { CenteredLoading } from "@/components/loading";
import { cn } from "@/lib/utils";
import {
  ArrowLeft,
  Clock,
  Check,
} from "lucide-react";
import { useBuyerRfqDetail, useBuyerRfqThreads } from "../hooks/useBuyerRfqs";
import { BuyerRfqThread } from "./buyer-rfq-thread";
import { useRfqViewedStore } from "../stores/use-rfq-viewed-store";

interface BuyerRfqDetailPageProps {
  readonly id: string;
}

export function BuyerRfqDetailPage({ id }: BuyerRfqDetailPageProps) {
  const t = useTranslations("buyerRfq.rfqDetail");
  const [selectedSupplierId, setSelectedSupplierId] = useState<string | null>(null);
  const { markRfqAsViewed } = useRfqViewedStore();

  const { rfq, isLoadingRfq } = useBuyerRfqDetail(id);
  const { data: threads = [], isLoading: isLoadingThreads } = useBuyerRfqThreads(id);

  // Automatically mark this RFQ chat thread as viewed/read
  useEffect(() => {
    if (id) {
      markRfqAsViewed(id);
    }
  }, [id, markRfqAsViewed]);

  // Auto-select first supplier if available and not yet set
  useEffect(() => {
    if (threads.length > 0 && !selectedSupplierId) {
      // Prioritize supplier with accepted status or latest offer
      const accepted = threads.find((th) => th.status === "accepted");
      setSelectedSupplierId(accepted ? accepted.supplierProfileId : threads[0].supplierProfileId);
    }
  }, [threads, selectedSupplierId]);

  if (isLoadingRfq || isLoadingThreads) {
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

  const selectedSupplier = threads.find((th) => th.supplierProfileId === selectedSupplierId) || threads[0];
  const isCompleted = rfq.status === "Completed" || rfq.status === "accepted";

  return (
    <BuyerLayout>
      <div className="space-y-8">
        {/* Navigation & Header */}
        <div className="space-y-2">
          <Link
            href="/rfq"
            className="inline-flex items-center gap-1.5 text-xs font-semibold text-primary hover:underline cursor-pointer"
          >
            <ArrowLeft className="h-4 w-4" /> {t("backLink")}
          </Link>
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-foreground font-heading">
              {rfq.product}
            </h1>
          </div>
        </div>

        {/* Unified Bidding & Negotiation Thread Section */}
        <div className="space-y-6">
          {threads.length === 0 ? (
            <Card className="border border-border rounded-lg bg-card shadow-xs">
              <CardContent className="py-16 text-center space-y-3">
                <Clock className="w-8 h-8 text-muted-foreground/60 mx-auto" />
                <p className="text-sm font-semibold text-foreground">{t("emptyBids")}</p>
                <p className="text-xs text-muted-foreground max-w-md mx-auto">
                  {t("emptyThreads")}
                </p>
              </CardContent>
            </Card>
          ) : (
            <div className="space-y-6">
              {/* Supplier Selector Tabs (only shown when multiple suppliers exist) */}
              {threads.length > 1 && (
                <div className="flex items-center gap-2 overflow-x-auto pb-2 scrollbar-none">
                  {threads.map((th) => {
                    const isSelected = th.supplierProfileId === selectedSupplier?.supplierProfileId;
                    const isAccepted = th.status === "accepted";

                    return (
                      <button
                        key={th.supplierProfileId}
                        type="button"
                        onClick={() => setSelectedSupplierId(th.supplierProfileId)}
                        className={cn(
                          "flex items-center gap-2 px-3.5 py-2 rounded-lg text-xs font-semibold whitespace-nowrap transition-colors cursor-pointer border shrink-0",
                          isSelected
                            ? "bg-primary text-primary-foreground border-primary shadow-xs"
                            : "bg-card text-foreground border-border hover:border-border/80 hover:bg-muted/40"
                        )}
                      >
                        <span className="truncate max-w-[160px]">
                          {th.supplierName}
                        </span>
                        {th.latestOffer && (
                          <span
                            className={cn(
                              "text-[11px] font-bold px-2 py-0.5 rounded-md",
                              isSelected
                                ? "bg-primary-foreground/20 text-primary-foreground"
                                : "bg-primary/10 text-primary"
                            )}
                          >
                            {th.latestOffer}
                          </span>
                        )}
                        {isAccepted && (
                          <span
                            className={cn(
                              "inline-flex items-center justify-center h-4.5 w-4.5 rounded-full shadow-xs shrink-0",
                              isSelected ? "bg-white text-emerald-600" : "bg-emerald-600 text-white"
                            )}
                            title="Penawaran Diterima"
                          >
                            <Check className="h-2.5 w-2.5 stroke-[3]" />
                          </span>
                        )}
                      </button>
                    );
                  })}
                </div>
              )}

              {/* Active Supplier Thread */}
              {selectedSupplier && (
                <BuyerRfqThread
                  rfqId={id}
                  supplier={selectedSupplier}
                  isRfqCompleted={isCompleted}
                />
              )}
            </div>
          )}
        </div>
      </div>
    </BuyerLayout>
  );
}
