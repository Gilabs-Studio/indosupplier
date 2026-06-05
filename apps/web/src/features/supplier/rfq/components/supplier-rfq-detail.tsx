"use client";

import React, { useState, useEffect } from "react";
import { useRouter } from "@/i18n/routing";
import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Field, FieldLabel, FieldGroup } from "@/components/ui/field";
import { toast } from "sonner";
import { ArrowLeft, Send, ShieldCheck, Calendar, MapPin, DollarSign, Scale } from "lucide-react";

interface SupplierRfqDetailProps {
  id: string;
}

export function SupplierRfqDetail({ id }: SupplierRfqDetailProps) {
  const router = useRouter();
  const t = useTranslations("supplier.rfq");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [rfq, setRfq] = useState({
    id: "RFQ-2026-102",
    product: "Bentonite Clay Powder",
    category: "Industrial Minerals",
    quantity: "20 Ton",
    port: "Tanjung Perak, Surabaya",
    date: "2026-06-03",
    budget: "Rp 4.000.000 / Ton",
    description: "Looking for high expansion grade sodium bentonite clay powder for drilling mud applications. Moisture content must be less than 12%. Mesh size 200 min 98% passing.",
    shippingTerm: "FOB (Free On Board)",
    targetDelivery: "2026-07-15",
    buyer: {
      name: "PT Nusantara Drilling Service",
      established: "2015",
      location: "Surabaya, Jawa Timur",
      rating: "4.7 / 5.0",
    }
  });

  const [form, setForm] = useState({
    price: "",
    moq: "",
    deliveryTime: "",
    notes: "",
  });

  useEffect(() => {
    const timer = setTimeout(() => {
      if (id === "RFQ-2026-101") {
        setRfq({
          id: "RFQ-2026-101",
          product: "Garnet Sand Mesh 80",
          category: "Industrial Minerals",
          quantity: "50 Ton",
          port: "Tanjung Priok, Jakarta",
          date: "2026-05-30",
          budget: "Rp 3.300.000 / Ton",
          description: "Need garnet sand mesh 80 for waterjet cutting machine. Low dust, washed grade, chloride content under 15ppm. Packing in 1.5-ton big bags.",
          shippingTerm: "CIF (Cost, Insurance & Freight)",
          targetDelivery: "2026-06-30",
          buyer: {
            name: "PT Metal Fabrication Indonesia",
            established: "2010",
            location: "Bekasi, Jawa Barat",
            rating: "4.9 / 5.0",
          }
        });
      }
    }, 0);
    return () => clearTimeout(timer);
  }, [id]);

  const handleSubmitProposal = (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setTimeout(() => {
      setIsSubmitting(false);
      toast.success(t("submitSuccess"));
      router.push("/supplier/rfq");
    }, 1000);
  };

  return (
    <div className="space-y-6 text-left">
      {/* Header */}
      <div className="flex items-center gap-4 border-b border-border/80 pb-6">
        <Button
          variant="outline"
          size="icon"
          onClick={() => router.push("/supplier/rfq")}
          className="h-9 w-9 cursor-pointer border-border"
        >
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
            {t("detailTitle")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("detailSubtitle")}
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
        {/* Specifications & details */}
        <div className="lg:col-span-2 space-y-6">
          <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
            <CardHeader className="border-b border-border bg-muted/10">
              <CardTitle className="text-base font-bold font-heading">{rfq.product}</CardTitle>
              <CardDescription className="text-xs">ID: {rfq.id} • Posted on {rfq.date}</CardDescription>
            </CardHeader>
            <CardContent className="p-6 space-y-6">
              {/* Core metrics */}
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <div className="p-3 border border-border bg-muted/20 rounded-xl flex flex-col">
                  <span className="text-[10px] font-bold text-muted-foreground uppercase">{t("reqQty")}</span>
                  <span className="text-base font-extrabold text-foreground mt-1 flex items-center gap-1.5">
                    <Scale className="h-4 w-4 text-primary" />
                    {rfq.quantity}
                  </span>
                </div>
                <div className="p-3 border border-border bg-muted/20 rounded-xl flex flex-col">
                  <span className="text-[10px] font-bold text-muted-foreground uppercase">{t("targetBudget")}</span>
                  <span className="text-base font-extrabold text-foreground mt-1 flex items-center gap-1.5">
                    <DollarSign className="h-4 w-4 text-success" />
                    {rfq.budget}
                  </span>
                </div>
                <div className="p-3 border border-border bg-muted/20 rounded-xl flex flex-col">
                  <span className="text-[10px] font-bold text-muted-foreground uppercase">{t("shippingTerm")}</span>
                  <span className="text-base font-extrabold text-foreground mt-1">
                    {rfq.shippingTerm}
                  </span>
                </div>
              </div>

              {/* Description */}
              <div className="space-y-2">
                <h4 className="text-xs font-bold text-muted-foreground uppercase">{t("specs")}</h4>
                <p className="text-sm text-foreground leading-relaxed bg-muted/10 p-4 rounded-xl border border-border">
                  {rfq.description}
                </p>
              </div>

              {/* Delivery and Destination */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs font-semibold">
                <div className="flex items-center gap-2 border border-border/80 p-3 rounded-lg bg-card">
                  <MapPin className="h-4 w-4 text-muted-foreground shrink-0" />
                  <div>
                    <p className="text-[10px] font-bold text-muted-foreground uppercase">Destination Port</p>
                    <p className="text-foreground text-sm font-bold mt-0.5">{rfq.port}</p>
                  </div>
                </div>
                <div className="flex items-center gap-2 border border-border/80 p-3 rounded-lg bg-card">
                  <Calendar className="h-4 w-4 text-muted-foreground shrink-0" />
                  <div>
                    <p className="text-[10px] font-bold text-muted-foreground uppercase">{t("targetDate")}</p>
                    <p className="text-foreground text-sm font-bold mt-0.5">{rfq.targetDelivery}</p>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Proposal form */}
          <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
            <CardHeader>
              <CardTitle className="text-base font-bold font-heading">{t("formTitle")}</CardTitle>
              <CardDescription className="text-xs">Provide details for your bidding proposal.</CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmitProposal} className="space-y-4">
                <FieldGroup className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                  <Field>
                    <FieldLabel>{t("formPrice")}</FieldLabel>
                    <Input
                      placeholder="e.g. Rp 3.200.000 / Ton"
                      value={form.price}
                      onChange={(e) => setForm({ ...form, price: e.target.value })}
                      required
                    />
                  </Field>
                  <Field>
                    <FieldLabel>{t("formMinQty")}</FieldLabel>
                    <Input
                      placeholder="e.g. 15 Ton"
                      value={form.moq}
                      onChange={(e) => setForm({ ...form, moq: e.target.value })}
                      required
                    />
                  </Field>
                  <Field>
                    <FieldLabel>{t("formDeliveryTime")}</FieldLabel>
                    <Input
                      placeholder="e.g. 14 Days"
                      value={form.deliveryTime}
                      onChange={(e) => setForm({ ...form, deliveryTime: e.target.value })}
                      required
                    />
                  </Field>
                </FieldGroup>

                <Field>
                  <FieldLabel>{t("formNotes")}</FieldLabel>
                  <Textarea
                    placeholder="Provide shipping availability, packaging details, cargo assurances..."
                    rows={4}
                    value={form.notes}
                    onChange={(e) => setForm({ ...form, notes: e.target.value })}
                    required
                    className="resize-none"
                  />
                </Field>

                <div className="flex justify-end pt-2">
                  <Button
                    type="submit"
                    disabled={isSubmitting}
                    className="bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer font-semibold py-5 px-6 text-sm flex items-center justify-center gap-2 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/20"
                  >
                    {isSubmitting ? "Sending..." : (
                      <>
                        <Send className="h-4 w-4" /> {t("btnSubmit")}
                      </>
                    )}
                  </Button>
                </div>
              </form>
            </CardContent>
          </Card>
        </div>

        {/* Buyer info sidebar */}
        <div className="space-y-6">
          <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
            <CardHeader>
              <CardTitle className="text-sm font-bold font-heading">{t("buyerInfo")}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center gap-3">
                <div className="h-10 w-10 bg-primary/10 text-primary border border-border rounded-lg font-heading font-bold text-lg flex items-center justify-center">
                  {rfq.buyer.name.substring(3, 5).toUpperCase()}
                </div>
                <div>
                  <h4 className="text-sm font-bold text-foreground truncate max-w-[150px]">{rfq.buyer.name}</h4>
                  <span className="text-[10px] text-muted-foreground font-semibold">Buyer Member</span>
                </div>
              </div>

              <div className="h-px bg-border" />

              <div className="space-y-2.5 text-xs">
                <div className="flex items-center justify-between font-semibold">
                  <span className="text-muted-foreground">Location</span>
                  <span className="text-foreground">{rfq.buyer.location}</span>
                </div>
                <div className="flex items-center justify-between font-semibold">
                  <span className="text-muted-foreground">Established</span>
                  <span className="text-foreground">{rfq.buyer.established}</span>
                </div>
                <div className="flex items-center justify-between font-semibold">
                  <span className="text-muted-foreground">Reputation</span>
                  <span className="text-foreground flex items-center gap-1">
                    <ShieldCheck className="h-3.5 w-3.5 text-success" />
                    {rfq.buyer.rating}
                  </span>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
