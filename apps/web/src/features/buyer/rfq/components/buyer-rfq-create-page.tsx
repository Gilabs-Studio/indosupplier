"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { BuyerLayout } from "../../components/buyer-layout";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Link, useRouter } from "@/i18n/routing";
import { Field, FieldLabel, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  FileUp,
  ArrowLeft,
  Info,
} from "lucide-react";

export function BuyerRfqCreatePage() {
  const t = useTranslations("buyer.rfqCreate");
  const router = useRouter();
  
  const [productName, setProductName] = useState("");
  const [category, setCategory] = useState("manufacturing");
  const [quantity, setQuantity] = useState("");
  const [unit, setUnit] = useState("Ton");
  const [targetPort, setTargetPort] = useState("");
  const [description, setDescription] = useState("");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!productName || !quantity || !targetPort) {
      alert("Harap isi semua kolom wajib!");
      return;
    }

    alert("RFQ berhasil dibuat dan disebarkan ke supplier terverifikasi!");
    router.push("/rfq");
  };

  return (
    <BuyerLayout>
      <div className="space-y-6 max-w-3xl">
        {/* Header */}
        <div className="space-y-2">
          <Link href="/rfq" className="inline-flex items-center gap-1 text-xs font-semibold text-primary hover:underline cursor-pointer">
            <ArrowLeft className="h-3.5 w-3.5" /> {t("backLink")}
          </Link>
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">{t("title")}</h1>
          <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
        </div>

        {/* Info Banner */}
        <div className="bg-primary/5 border border-primary/20 rounded-xl p-4 flex gap-3 text-sm text-foreground">
          <Info className="h-5 w-5 text-primary shrink-0 mt-0.5" />
          <p>
            <strong>{t("tipsTitle")}:</strong> {t("tipsDesc")}
          </p>
        </div>

        {/* Form Card */}
        <Card className="border border-border rounded-xl bg-card shadow-xs overflow-hidden">
          <CardContent className="p-6">
            <form onSubmit={handleSubmit} className="space-y-6">
              <FieldGroup className="space-y-5">
                {/* Product Name */}
                <Field className="space-y-2">
                  <FieldLabel htmlFor="productName">
                    {t("labelProduct")} <span className="text-destructive">*</span>
                  </FieldLabel>
                  <Input
                    id="productName"
                    value={productName}
                    onChange={(e) => setProductName(e.target.value)}
                    placeholder={t("placeholderProduct")}
                    required
                    className="cursor-pointer"
                  />
                </Field>

                {/* Category & Unit */}
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <Field className="space-y-2">
                    <FieldLabel htmlFor="category">{t("labelCategory")}</FieldLabel>
                    <select
                      id="category"
                      value={category}
                      onChange={(e) => setCategory(e.target.value)}
                      className="w-full rounded-lg border border-border bg-card px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-hidden cursor-pointer"
                    >
                      <option value="manufacturing">Manufaktur & Material</option>
                      <option value="agriculture">Pertanian & Pangan</option>
                      <option value="textile">Tekstil & Konveksi</option>
                      <option value="furniture">Furnitur & Kayu</option>
                    </select>
                  </Field>

                  <div className="grid grid-cols-2 gap-2">
                    <Field className="space-y-2">
                      <FieldLabel htmlFor="quantity">
                        {t("labelVolume")} <span className="text-destructive">*</span>
                      </FieldLabel>
                      <Input
                        id="quantity"
                        type="text"
                        value={quantity}
                        onChange={(e) => setQuantity(e.target.value)}
                        placeholder="Qty"
                        required
                        className="cursor-pointer"
                      />
                    </Field>
                    <Field className="space-y-2">
                      <FieldLabel htmlFor="unit">{t("labelUnit")}</FieldLabel>
                      <select
                        id="unit"
                        value={unit}
                        onChange={(e) => setUnit(e.target.value)}
                        className="w-full rounded-lg border border-border bg-card px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-hidden cursor-pointer"
                      >
                        <option value="Ton">Ton</option>
                        <option value="Kg">Kg</option>
                        <option value="Pcs">Pcs</option>
                        <option value="Container">20ft Container</option>
                      </select>
                    </Field>
                  </div>
                </div>

                {/* Shipping Destination */}
                <Field className="space-y-2">
                  <FieldLabel htmlFor="targetPort">
                    {t("labelDestination")} <span className="text-destructive">*</span>
                  </FieldLabel>
                  <Input
                    id="targetPort"
                    value={targetPort}
                    onChange={(e) => setTargetPort(e.target.value)}
                    placeholder={t("placeholderDestination")}
                    required
                    className="cursor-pointer"
                  />
                </Field>

                {/* Description Requirements */}
                <Field className="space-y-2">
                  <FieldLabel htmlFor="description">{t("labelDescription")}</FieldLabel>
                  <textarea
                    id="description"
                    rows={4}
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    placeholder={t("placeholderDescription")}
                    className="w-full rounded-lg border border-border bg-card px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:border-primary focus:outline-hidden"
                  />
                </Field>

                {/* File Attachment */}
                <Field className="space-y-2">
                  <FieldLabel>{t("labelAttachment")}</FieldLabel>
                  <div className="border border-dashed border-border hover:border-primary/50 transition-colors rounded-lg p-6 text-center cursor-pointer space-y-2">
                    <FileUp className="mx-auto h-8 w-8 text-muted-foreground opacity-60" />
                    <p className="text-xs font-semibold text-foreground">{t("uploadPlaceholder")}</p>
                    <p className="text-[10px] text-muted-foreground">{t("uploadLimit")}</p>
                  </div>
                </Field>
              </FieldGroup>

              {/* Actions */}
              <div className="pt-4 flex items-center justify-end gap-3 border-t border-border">
                <Button asChild variant="outline" className="cursor-pointer transition-all hover:-translate-y-0.5 active:translate-y-0 shadow-xs">
                  <Link href="/rfq">{t("btnCancel")}</Link>
                </Button>
                <Button type="submit" className="bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer px-6 transition-all hover:-translate-y-0.5 active:translate-y-0 shadow-lg hover:shadow-primary/20">
                  {t("btnSubmit")}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </BuyerLayout>
  );
}
