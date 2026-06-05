"use client";

import React, { useState } from "react";
import { useRouter } from "@/i18n/routing";
import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Field, FieldLabel, FieldGroup } from "@/components/ui/field";
import { toast } from "sonner";
import { ArrowLeft, Save } from "lucide-react";

export function SupplierAdsCreate() {
  const router = useRouter();
  const t = useTranslations("supplier.ads");
  const [isSaving, setIsSaving] = useState(false);

  const [form, setForm] = useState({
    name: "",
    product: "Garnet Sand Mesh 80",
    budget: "",
    bid: "",
  });

  const products = [
    "Garnet Sand Mesh 80",
    "Bentonite Clay Powder",
    "Quartz Powder 325 Mesh",
  ];

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.name || !form.budget || !form.bid) {
      toast.error("Please fill out all fields.");
      return;
    }
    setIsSaving(true);
    setTimeout(() => {
      setIsSaving(false);
      toast.success(t("createSuccess"));
      router.push("/supplier/ads");
    }, 1000);
  };

  return (
    <div className="space-y-6 text-left">
      {/* Header */}
      <div className="flex items-center gap-4 border-b border-border/80 pb-6">
        <Button
          variant="outline"
          size="icon"
          onClick={() => router.push("/supplier/ads")}
          className="h-9 w-9 cursor-pointer border-border"
        >
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
            {t("createTitle")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("createSubtitle")}
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="max-w-2xl mx-auto">
        <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
          <CardHeader className="border-b border-border bg-muted/10">
            <CardTitle className="text-base font-bold font-heading">Ad Parameters</CardTitle>
            <CardDescription className="text-xs">Specify your campaign limits and target bids.</CardDescription>
          </CardHeader>
          <CardContent className="p-6 space-y-4">
            <Field>
              <FieldLabel>{t("formName")}</FieldLabel>
              <Input
                placeholder="e.g. Garnet Sand Search Promo"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                required
              />
            </Field>

            <Field>
              <FieldLabel>{t("formProduct")}</FieldLabel>
              <select
                value={form.product}
                onChange={(e) => setForm({ ...form, product: e.target.value })}
                className="w-full px-3 py-2 bg-card border border-border text-sm rounded-lg focus:outline-none focus:border-primary transition-all text-left"
              >
                {products.map((p) => (
                  <option key={p} value={p}>{p}</option>
                ))}
              </select>
            </Field>

            <FieldGroup className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <Field>
                <FieldLabel>{t("formBudget")}</FieldLabel>
                <Input
                  type="number"
                  placeholder="e.g. 150000"
                  value={form.budget}
                  onChange={(e) => setForm({ ...form, budget: e.target.value })}
                  required
                />
              </Field>
              <Field>
                <FieldLabel>{t("formBid")}</FieldLabel>
                <Input
                  type="number"
                  placeholder="e.g. 1500"
                  value={form.bid}
                  onChange={(e) => setForm({ ...form, bid: e.target.value })}
                  required
                />
              </Field>
            </FieldGroup>

            <div className="flex justify-end pt-4 gap-2">
              <Button type="button" variant="outline" onClick={() => router.push("/supplier/ads")} className="text-xs h-9 cursor-pointer border-border">
                Cancel
              </Button>
              <Button
                type="submit"
                disabled={isSaving}
                className="bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer font-semibold py-5 px-6 text-sm flex items-center justify-center gap-2 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/20"
              >
                {isSaving ? "Saving..." : (
                  <>
                    <Save className="h-4 w-4" /> {t("btnSubmit")}
                  </>
                )}
              </Button>
            </div>
          </CardContent>
        </Card>
      </form>
    </div>
  );
}
