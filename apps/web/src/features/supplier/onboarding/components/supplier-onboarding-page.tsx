"use client";

import React, { useState } from "react";
import { useRouter } from "@/i18n/routing";
import { useTranslations } from "next-intl";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Field, FieldLabel, FieldGroup } from "@/components/ui/field";
import { toast } from "sonner";
import { ArrowLeft, ArrowRight, Check, Sparkles } from "lucide-react";

export function SupplierOnboardingPage() {
  const router = useRouter();
  const t = useTranslations("supplier.onboarding");
  const [step, setStep] = useState(1);
  const [isSaving, setIsSaving] = useState(false);

  // Form aggregate state
  const [form, setForm] = useState({
    companyName: "",
    businessType: "",
    established: "",
    taxId: "",
    nib: "",
    firstProductName: "",
    firstProductPrice: "",
  });

  const handleNext = () => {
    if (step === 1 && (!form.companyName || !form.businessType)) {
      toast.error("Please fill out general information.");
      return;
    }
    if (step === 2 && (!form.taxId || !form.nib)) {
      toast.error("Please fill out NPWP and NIB details.");
      return;
    }
    setStep(step + 1);
  };

  const handleBack = () => {
    setStep(step - 1);
  };

  const handleComplete = (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.firstProductName || !form.firstProductPrice) {
      toast.error("Please specify your first product catalog listing.");
      return;
    }

    setIsSaving(true);
    setTimeout(() => {
      setIsSaving(false);
      toast.success(t("success"));
      router.push("/supplier/dashboard");
    }, 1200);
  };

  return (
    <div className="max-w-3xl mx-auto space-y-8 text-left py-6">
      {/* Header */}
      <div className="text-center space-y-2">
        <div className="inline-flex items-center gap-1.5 rounded-full border border-primary/20 bg-primary/10 text-primary px-3 py-1 text-xs font-semibold">
          <Sparkles className="h-3.5 w-3.5" /> Set Up Seller Account
        </div>
        <h1 className="text-2xl md:text-3xl font-extrabold tracking-tight text-foreground font-heading">
          {t("title")}
        </h1>
        <p className="text-sm text-muted-foreground max-w-lg mx-auto">
          {t("subtitle")}
        </p>
      </div>

      {/* Stepper Progress bar */}
      <div className="flex items-center justify-between relative max-w-md mx-auto before:absolute before:left-0 before:right-0 before:top-[18px] before:h-0.5 before:bg-border z-0">
        {[1, 2, 3].map((s) => (
          <div key={s} className="flex flex-col items-center gap-2 relative z-10">
            <div
              className={`h-9 w-9 rounded-full flex items-center justify-center font-bold text-sm border transition-all duration-300 ${
                s < step
                  ? "bg-success text-success-foreground border-success"
                  : s === step
                  ? "bg-primary text-primary-foreground border-primary ring-4 ring-primary/20"
                  : "bg-card text-muted-foreground border-border"
              }`}
            >
              {s < step ? <Check className="h-4 w-4" /> : s}
            </div>
            <span className={`text-[10px] font-bold uppercase ${s === step ? "text-primary" : "text-muted-foreground"}`}>
              {s === 1 ? t("step1") : s === 2 ? t("step2") : t("step3")}
            </span>
          </div>
        ))}
      </div>

      {/* Steps body */}
      <Card className="border border-border shadow-md rounded-2xl overflow-hidden bg-card">
        <CardContent className="p-6 md:p-8">
          {step === 1 && (
            <div className="space-y-4 animate-fade-in">
              <h3 className="text-base font-bold font-heading text-foreground">{t("step1")}</h3>
              <Field>
                <FieldLabel>Company Name</FieldLabel>
                <Input
                  required
                  placeholder="e.g. PT Nusantara Mineral"
                  value={form.companyName}
                  onChange={(e) => setForm({ ...form, companyName: e.target.value })}
                />
              </Field>
              <FieldGroup className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <Field>
                  <FieldLabel>Business Type</FieldLabel>
                  <Input
                    required
                    placeholder="e.g. Manufacturer / Trader"
                    value={form.businessType}
                    onChange={(e) => setForm({ ...form, businessType: e.target.value })}
                  />
                </Field>
                <Field>
                  <FieldLabel>Established Year</FieldLabel>
                  <Input
                    placeholder="e.g. 2018"
                    value={form.established}
                    onChange={(e) => setForm({ ...form, established: e.target.value })}
                  />
                </Field>
              </FieldGroup>
              <div className="flex justify-end pt-4">
                <Button onClick={handleNext} className="text-xs h-9 font-semibold cursor-pointer">
                  {t("btnNext")} <ArrowRight className="ml-1.5 h-4 w-4" />
                </Button>
              </div>
            </div>
          )}

          {step === 2 && (
            <div className="space-y-4 animate-fade-in">
              <h3 className="text-base font-bold font-heading text-foreground">{t("step2")}</h3>
              <FieldGroup className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <Field>
                  <FieldLabel>NPWP (Tax ID)</FieldLabel>
                  <Input
                    required
                    placeholder="01.234.567.8-999.000"
                    value={form.taxId}
                    onChange={(e) => setForm({ ...form, taxId: e.target.value })}
                  />
                </Field>
                <Field>
                  <FieldLabel>NIB (Registration No)</FieldLabel>
                  <Input
                    required
                    placeholder="9120001234567"
                    value={form.nib}
                    onChange={(e) => setForm({ ...form, nib: e.target.value })}
                  />
                </Field>
              </FieldGroup>
              <div className="flex justify-between pt-4">
                <Button variant="outline" onClick={handleBack} className="text-xs h-9 font-semibold cursor-pointer border-border">
                  <ArrowLeft className="mr-1.5 h-4 w-4" /> {t("btnBack")}
                </Button>
                <Button onClick={handleNext} className="text-xs h-9 font-semibold cursor-pointer">
                  {t("btnNext")} <ArrowRight className="ml-1.5 h-4 w-4" />
                </Button>
              </div>
            </div>
          )}

          {step === 3 && (
            <form onSubmit={handleComplete} className="space-y-4 animate-fade-in">
              <h3 className="text-base font-bold font-heading text-foreground">{t("step3")}</h3>
              <Field>
                <FieldLabel>Product Name</FieldLabel>
                <Input
                  required
                  placeholder="e.g. Fine Silica Powder 325 Mesh"
                  value={form.firstProductName}
                  onChange={(e) => setForm({ ...form, firstProductName: e.target.value })}
                />
              </Field>
              <Field>
                <FieldLabel>Price Terms (per Unit)</FieldLabel>
                <Input
                  required
                  placeholder="e.g. Rp 2.500.000 / Ton"
                  value={form.firstProductPrice}
                  onChange={(e) => setForm({ ...form, firstProductPrice: e.target.value })}
                />
              </Field>

              <div className="flex justify-between pt-4">
                <Button type="button" variant="outline" onClick={handleBack} className="text-xs h-9 font-semibold cursor-pointer border-border">
                  <ArrowLeft className="mr-1.5 h-4 w-4" /> {t("btnBack")}
                </Button>
                <Button
                  type="submit"
                  disabled={isSaving}
                  className="bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer font-semibold h-9 text-xs flex items-center justify-center gap-1.5 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg"
                >
                  <Check className="h-4 w-4" />
                  {isSaving ? "Completing..." : t("btnComplete")}
                </Button>
              </div>
            </form>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
