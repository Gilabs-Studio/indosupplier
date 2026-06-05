"use client";

import React, { useState, useEffect } from "react";
import { useRouter } from "@/i18n/routing";
import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Field, FieldLabel } from "@/components/ui/field";
import { toast } from "sonner";
import { ArrowLeft, Gavel, Clock, History } from "lucide-react";

interface SupplierAuctionDetailProps {
  id: string;
}

export function SupplierAuctionDetail({ id }: SupplierAuctionDetailProps) {
  const router = useRouter();
  const t = useTranslations("supplier.auction");

  const [auction, setAuction] = useState({
    id: "AUC-101",
    name: "Premium Search Result Slot #1 (Mineral Category)",
    startingPrice: "Rp 1.000.000",
    minIncrement: "Rp 100.000",
    currentHighBid: 2500000,
    timeLeft: "04h 32m",
    bids: [
      { id: "B1", bidder: "PT Barito Minerals", amount: "Rp 2.400.000", time: "20 mins ago" },
      { id: "B2", bidder: "PT Java Mining Corp", amount: "Rp 2.200.000", time: "1 hour ago" },
      { id: "B3", bidder: "PT Cipta Industrial", amount: "Rp 2.000.000", time: "2 hours ago" },
    ],
  });

  const [bidValue, setBidValue] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => {
      if (id === "AUC-102") {
        setAuction({
          id: "AUC-102",
          name: "IndoSupplier Homepage Banner Slot A",
          startingPrice: "Rp 2.000.000",
          minIncrement: "Rp 200.000",
          currentHighBid: 5200000,
          timeLeft: "12h 15m",
          bids: [
            { id: "B1", bidder: "PT Agro Nusantara", amount: "Rp 5.000.000", time: "5 mins ago" },
            { id: "B2", bidder: "PT Borneo Resources", amount: "Rp 4.600.000", time: "45 mins ago" },
          ],
        });
      }
    }, 0);
    return () => clearTimeout(timer);
  }, [id]);

  const handlePlaceBid = (e: React.FormEvent) => {
    e.preventDefault();
    const bidAmount = parseInt(bidValue, 10);
    const minRequiredBid = auction.currentHighBid + 100000;

    if (isNaN(bidAmount) || bidAmount < minRequiredBid) {
      toast.error(`Your bid must be at least Rp ${(minRequiredBid).toLocaleString("id-ID")}`);
      return;
    }

    setIsSubmitting(true);
    setTimeout(() => {
      setIsSubmitting(false);
      setAuction({
        ...auction,
        currentHighBid: bidAmount,
        bids: [
          { id: `B-${Date.now()}`, bidder: "You (PT Nusantara Supplier Utama)", amount: `Rp ${bidAmount.toLocaleString("id-ID")}`, time: "Just now" },
          ...auction.bids
        ]
      });
      setBidValue("");
      toast.success(t("bidSuccess"));
    }, 800);
  };

  return (
    <div className="space-y-6 text-left">
      {/* Header */}
      <div className="flex items-center gap-4 border-b border-border/80 pb-6">
        <Button
          variant="outline"
          size="icon"
          onClick={() => router.push("/supplier/auction")}
          className="h-9 w-9 cursor-pointer border-border"
        >
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
            {t("consoleTitle")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("consoleSubtitle")}
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
        {/* Main Console details */}
        <div className="lg:col-span-2 space-y-6">
          <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
            <CardHeader className="border-b border-border bg-muted/10">
              <CardTitle className="text-base font-bold font-heading">{auction.name}</CardTitle>
              <CardDescription className="text-xs">ID: {auction.id}</CardDescription>
            </CardHeader>
            <CardContent className="p-6 space-y-6">
              {/* Stats row */}
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <div className="p-4 border border-border bg-muted/20 rounded-xl flex flex-col">
                  <span className="text-[10px] font-bold text-muted-foreground uppercase">{t("startingPrice")}</span>
                  <span className="text-sm font-bold text-foreground mt-1">
                    {auction.startingPrice}
                  </span>
                </div>
                <div className="p-4 border border-border bg-primary/5 rounded-xl flex flex-col">
                  <span className="text-[10px] font-bold text-primary uppercase">{t("currentHighBid")}</span>
                  <span className="text-lg font-extrabold text-foreground mt-1 flex items-center gap-1.5">
                    <Gavel className="h-4.5 w-4.5 text-primary shrink-0 animate-bounce" />
                    Rp {auction.currentHighBid.toLocaleString("id-ID")}
                  </span>
                </div>
                <div className="p-4 border border-border bg-muted/20 rounded-xl flex flex-col">
                  <span className="text-[10px] font-bold text-muted-foreground uppercase">{t("tableTimeLeft")}</span>
                  <span className="text-base font-extrabold text-destructive mt-1 flex items-center gap-1.5">
                    <Clock className="h-4 w-4 shrink-0" />
                    {auction.timeLeft}
                  </span>
                </div>
              </div>

              {/* Increments */}
              <div className="text-xs font-semibold text-muted-foreground">
                Minimum increment: <span className="text-foreground">{auction.minIncrement}</span>
              </div>

              {/* Bidding history logs */}
              <div className="space-y-3">
                <h4 className="text-xs font-bold text-muted-foreground uppercase flex items-center gap-1.5">
                  <History className="h-4 w-4" />
                  {t("bidHistory")}
                </h4>
                <div className="border border-border rounded-xl divide-y divide-border overflow-hidden bg-card">
                  {auction.bids.map((b) => (
                    <div key={b.id} className="p-3.5 flex items-center justify-between text-xs font-medium hover:bg-muted/10 transition-colors">
                      <div className="space-y-0.5">
                        <p className="text-foreground font-bold">{b.bidder}</p>
                        <p className="text-[10px] text-muted-foreground">{b.time}</p>
                      </div>
                      <span className="text-foreground font-extrabold">{b.amount}</span>
                    </div>
                  ))}
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Place Bid Sidebar */}
        <div className="space-y-6">
          <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
            <CardHeader className="border-b border-border bg-muted/20 py-4 px-6">
              <CardTitle className="text-sm font-bold text-foreground">Place a New Bid</CardTitle>
            </CardHeader>
            <CardContent className="p-6">
              <form onSubmit={handlePlaceBid} className="space-y-4">
                <Field className="space-y-1">
                  <FieldLabel>{t("yourBid")}</FieldLabel>
                  <Input
                    type="number"
                    placeholder={`Min. Rp ${(auction.currentHighBid + 100000).toLocaleString("id-ID")}`}
                    value={bidValue}
                    onChange={(e) => setBidValue(e.target.value)}
                    required
                  />
                </Field>

                <Button
                  type="submit"
                  disabled={isSubmitting}
                  className="w-full bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer font-semibold py-5 text-sm flex items-center justify-center gap-2 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/20"
                >
                  <Gavel className="h-4.5 w-4.5" />
                  {t("btnBid")}
                </Button>
              </form>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
