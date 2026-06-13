"use client";

import React, { useState, useEffect } from "react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Field, FieldLabel, FieldGroup } from "@/components/ui/field";
import { Check, X, Edit2, AlertCircle, Loader2 } from "lucide-react";
import { useVerificationData, useUpdateVerificationData, useSubmitVerification } from "../hooks/useVerification";

export function SupplierVerificationPage() {
  const t = useTranslations("supplier.verification");
  const { data: verificationData, isLoading } = useVerificationData();
  const updateMutation = useUpdateVerificationData();
  const submitMutation = useSubmitVerification();

  const [activeStep, setActiveStep] = useState<number>(1);

  // Form states matching steps
  const [businessForm, setBusinessForm] = useState({
    legalName: "",
    nibNumber: "",
    establishedDate: "",
    industry: "",
    phone: "",
    description: "",
    address: "",
  });

  const [stakeholderForm, setStakeholderForm] = useState({
    directorName: "",
    directorNik: "",
  });

  const [bankForm, setBankForm] = useState({
    bankName: "",
    accountNumber: "",
    accountName: "",
  });

  // Load fetched data into form states
  useEffect(() => {
    if (verificationData) {
      const timer = setTimeout(() => {
        setBusinessForm({
          legalName: verificationData.businessInfo.legalName || "",
          nibNumber: verificationData.businessInfo.nibNumber || "",
          establishedDate: verificationData.businessInfo.establishedDate || "",
          industry: verificationData.businessInfo.industry || "",
          phone: verificationData.businessInfo.phone || "",
          description: verificationData.businessInfo.description || "",
          address: verificationData.businessInfo.address || "",
        });
        setStakeholderForm({
          directorName: verificationData.stakeholderInfo.directorName || "",
          directorNik: verificationData.stakeholderInfo.directorNik || "",
        });
        setBankForm({
          bankName: verificationData.bankAccountInfo.bankName || "",
          accountNumber: verificationData.bankAccountInfo.accountNumber || "",
          accountName: verificationData.bankAccountInfo.accountName || "",
        });
      }, 0);
      return () => clearTimeout(timer);
    }
  }, [verificationData]);

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center p-12 min-h-[400px]">
        <Loader2 className="h-8 w-8 animate-spin text-primary mb-4" />
        <span className="text-sm font-semibold text-muted-foreground">{t("loading")}</span>
      </div>
    );
  }

  const steps = [
    { id: 1, label: t("step1Label") },
    { id: 2, label: t("step2Label") },
    { id: 3, label: t("step3Label") },
    { id: 4, label: t("step4Label") },
  ];

  const handleSaveStep = (step: number, data: Record<string, string>) => {
    const payload: Record<string, unknown> = {};
    if (step === 1) payload.businessInfo = data;
    if (step === 2) payload.stakeholderInfo = data;
    if (step === 3) payload.bankAccountInfo = data;

    updateMutation.mutate(payload, {
      onSuccess: () => {
        setActiveStep(step + 1);
      },
    });
  };

  const handleFinalSubmit = () => {
    submitMutation.mutate();
  };

  // Helper validation indicators
  const isBusinessValid =
    !!businessForm.legalName &&
    !!businessForm.nibNumber &&
    !!businessForm.establishedDate &&
    !!businessForm.industry &&
    !!businessForm.phone &&
    !!businessForm.address;

  const isStakeholderValid = !!stakeholderForm.directorName && !!stakeholderForm.directorNik;

  const isBankValid =
    !!bankForm.bankName && !!bankForm.accountNumber && !!bankForm.accountName;

  const isAllValid = isBusinessValid && isStakeholderValid && isBankValid;

  return (
    <div className="min-h-screen bg-background text-foreground transition-all duration-300 font-sans flex flex-col">
      {/* Header bar */}
      <header className="sticky top-0 z-30 h-16 bg-card border-b border-border flex items-center justify-between px-6">
        <h1 className="text-lg font-bold font-heading">{t("stepperTitle")}</h1>
        <Link
          href="/supplier/dashboard"
          className="text-muted-foreground hover:text-foreground p-2 rounded-lg transition-colors cursor-pointer"
        >
          <X className="h-5 w-5" />
        </Link>
      </header>

      {/* Main stepper split */}
      <div className="flex-1 max-w-6xl w-full mx-auto grid grid-cols-1 md:grid-cols-4 gap-8 p-6 md:p-8 text-left items-start">
        {/* Left Side: Stepper Navigation */}
        <div className="md:col-span-1 space-y-4">
          {steps.map((s) => {
            const isActive = activeStep === s.id;
            const isCompleted =
              s.id < activeStep ||
              (s.id === 1 && isBusinessValid) ||
              (s.id === 2 && isStakeholderValid) ||
              (s.id === 3 && isBankValid);

            return (
              <button
                key={s.id}
                onClick={() => setActiveStep(s.id)}
                className={`w-full flex items-center gap-3.5 p-3 rounded-lg text-left transition-all duration-200 cursor-pointer ${
                  isActive
                    ? "bg-primary/10 text-primary font-bold shadow-xs border-l-4 border-primary"
                    : "text-muted-foreground hover:text-foreground hover:bg-muted/50"
                }`}
              >
                <div
                  className={`h-7 w-7 rounded-full flex items-center justify-center text-xs font-bold shrink-0 border transition-colors ${
                    isCompleted && s.id !== 4
                      ? "bg-success/20 text-success border-success/30"
                      : isActive
                      ? "bg-primary text-primary-foreground border-primary"
                      : "bg-muted text-muted-foreground border-border"
                  }`}
                >
                  {isCompleted && s.id !== 4 ? <Check className="h-4 w-4" /> : s.id}
                </div>
                <span className="text-sm font-semibold select-none">{s.label}</span>
              </button>
            );
          })}
        </div>

        {/* Right Side: Step Content Area */}
        <div className="md:col-span-3 bg-card border border-border/80 rounded-xl p-6 md:p-8 shadow-xs min-h-[450px] flex flex-col justify-between whitespace-pre-wrap">
          {/* Step 1 Form */}
          {activeStep === 1 && (
            <div className="space-y-6 animate-fade-in">
              <div>
                <h3 className="text-lg font-bold font-heading text-foreground">{t("step1Label")}</h3>
                <p className="text-xs text-muted-foreground mt-1 font-medium">
                  {t("step1Desc")}
                </p>
              </div>

              <FieldGroup className="space-y-4">
                <Field className="space-y-1">
                  <FieldLabel>{t("fieldLegalName")}</FieldLabel>
                  <Input
                    value={businessForm.legalName}
                    onChange={(e) => setBusinessForm({ ...businessForm, legalName: e.target.value })}
                    placeholder={t("placeholderLegalName")}
                  />
                </Field>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <Field className="space-y-1">
                    <FieldLabel>{t("fieldNib")}</FieldLabel>
                    <Input
                      value={businessForm.nibNumber}
                      onChange={(e) => setBusinessForm({ ...businessForm, nibNumber: e.target.value })}
                      placeholder={t("placeholderNib")}
                    />
                  </Field>
                  <Field className="space-y-1">
                    <FieldLabel>{t("fieldEstDate")}</FieldLabel>
                    <Input
                      type="date"
                      value={businessForm.establishedDate}
                      onChange={(e) =>
                        setBusinessForm({ ...businessForm, establishedDate: e.target.value })
                      }
                    />
                  </Field>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <Field className="space-y-1">
                    <FieldLabel>{t("fieldIndustry")}</FieldLabel>
                    <Input
                      value={businessForm.industry}
                      onChange={(e) => setBusinessForm({ ...businessForm, industry: e.target.value })}
                      placeholder={t("placeholderIndustry")}
                    />
                  </Field>
                  <Field className="space-y-1">
                    <FieldLabel>{t("fieldPhone")}</FieldLabel>
                    <Input
                      value={businessForm.phone}
                      onChange={(e) => setBusinessForm({ ...businessForm, phone: e.target.value })}
                      placeholder={t("placeholderPhone")}
                    />
                  </Field>
                </div>

                <Field className="space-y-1">
                  <FieldLabel>{t("fieldAddress")}</FieldLabel>
                  <Input
                    value={businessForm.address}
                    onChange={(e) => setBusinessForm({ ...businessForm, address: e.target.value })}
                    placeholder={t("placeholderAddress")}
                  />
                </Field>

                <Field className="space-y-1">
                  <FieldLabel>{t("fieldDesc")}</FieldLabel>
                  <Textarea
                    value={businessForm.description}
                    onChange={(e) =>
                      setBusinessForm({ ...businessForm, description: e.target.value })
                    }
                    placeholder={t("placeholderDesc")}
                    rows={3}
                  />
                </Field>
              </FieldGroup>

              <div className="flex justify-end pt-4 border-t border-border/80">
                <Button
                  onClick={() => handleSaveStep(1, businessForm)}
                  className="bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer font-bold px-6 py-4 rounded-lg transition-all hover:-translate-y-0.5"
                >
                  {t("btnSaveContinue")}
                </Button>
              </div>
            </div>
          )}

          {/* Step 2 Form */}
          {activeStep === 2 && (
            <div className="space-y-6 animate-fade-in">
              <div>
                <h3 className="text-lg font-bold font-heading text-foreground">{t("step2Label")}</h3>
                <p className="text-xs text-muted-foreground mt-1 font-medium">
                  {t("step2Desc")}
                </p>
              </div>

              <FieldGroup className="space-y-4">
                <Field className="space-y-1">
                  <FieldLabel>{t("fieldDirectorName")}</FieldLabel>
                  <Input
                    value={stakeholderForm.directorName}
                    onChange={(e) =>
                      setStakeholderForm({ ...stakeholderForm, directorName: e.target.value })
                    }
                    placeholder={t("placeholderDirectorName")}
                  />
                </Field>

                <Field className="space-y-1">
                  <FieldLabel>{t("fieldDirectorNik")}</FieldLabel>
                  <Input
                    value={stakeholderForm.directorNik}
                    onChange={(e) =>
                      setStakeholderForm({ ...stakeholderForm, directorNik: e.target.value })
                    }
                    placeholder={t("placeholderDirectorNik")}
                  />
                </Field>
              </FieldGroup>

              <div className="flex justify-between pt-4 border-t border-border/80">
                <Button
                  variant="outline"
                  onClick={() => setActiveStep(1)}
                  className="cursor-pointer font-semibold border-border hover:bg-muted/50"
                >
                  {t("btnBack")}
                </Button>
                <Button
                  onClick={() => handleSaveStep(2, stakeholderForm)}
                  className="bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer font-bold px-6 py-4 rounded-lg transition-all hover:-translate-y-0.5"
                >
                  {t("btnSaveContinue")}
                </Button>
              </div>
            </div>
          )}

          {/* Step 3 Form */}
          {activeStep === 3 && (
            <div className="space-y-6 animate-fade-in">
              <div>
                <h3 className="text-lg font-bold font-heading text-foreground">{t("step3Label")}</h3>
                <p className="text-xs text-muted-foreground mt-1 font-medium">
                  {t("step3Desc")}
                </p>
              </div>

              <FieldGroup className="space-y-4">
                <Field className="space-y-1">
                  <FieldLabel>{t("fieldBankName")}</FieldLabel>
                  <Input
                    value={bankForm.bankName}
                    onChange={(e) => setBankForm({ ...bankForm, bankName: e.target.value })}
                    placeholder={t("placeholderBankName")}
                  />
                </Field>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <Field className="space-y-1">
                    <FieldLabel>{t("fieldAccountNumber")}</FieldLabel>
                    <Input
                      value={bankForm.accountNumber}
                      onChange={(e) =>
                        setBankForm({ ...bankForm, accountNumber: e.target.value })
                      }
                      placeholder={t("placeholderAccountNumber")}
                    />
                  </Field>
                  <Field className="space-y-1">
                    <FieldLabel>{t("fieldAccountName")}</FieldLabel>
                    <Input
                      value={bankForm.accountName}
                      onChange={(e) => setBankForm({ ...bankForm, accountName: e.target.value })}
                      placeholder={t("placeholderAccountName")}
                    />
                  </Field>
                </div>
              </FieldGroup>

              <div className="flex justify-between pt-4 border-t border-border/80">
                <Button
                  variant="outline"
                  onClick={() => setActiveStep(2)}
                  className="cursor-pointer font-semibold border-border hover:bg-muted/50"
                >
                  {t("btnBack")}
                </Button>
                <Button
                  onClick={() => handleSaveStep(3, bankForm)}
                  className="bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer font-bold px-6 py-4 rounded-lg transition-all hover:-translate-y-0.5"
                >
                  {t("btnSaveContinue")}
                </Button>
              </div>
            </div>
          )}

          {/* Step 4 Review Page */}
          {activeStep === 4 && (
            <div className="space-y-6 animate-fade-in text-left">
              <div>
                <h3 className="text-lg font-bold font-heading text-foreground">{t("reviewTitle")}</h3>
                <p className="text-xs text-muted-foreground mt-1 font-medium">
                  {t("step4Desc")}
                </p>
              </div>

              <div className="space-y-5">
                {/* Detail Bisnis */}
                <div
                  className={`border rounded-lg p-5 bg-card space-y-4 transition-colors ${
                    isBusinessValid ? "border-border" : "border-destructive/50 bg-destructive/5"
                  }`}
                >
                  <div className="flex items-center justify-between border-b border-border/50 pb-2">
                    <h4 className="text-sm font-bold text-foreground">{t("reviewBusinessHeader")}</h4>
                    <button
                      onClick={() => setActiveStep(1)}
                      className="text-xs text-primary font-bold flex items-center gap-1 hover:underline cursor-pointer"
                    >
                      <Edit2 className="h-3.5 w-3.5" /> {t("reviewEditLink")}
                    </button>
                  </div>

                  <div className="space-y-3 text-xs">
                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1 py-1">
                      <span className="text-muted-foreground font-semibold">{t("fieldLegalName")}</span>
                      <span className={businessForm.legalName ? "font-bold text-foreground" : "text-destructive font-bold flex items-center gap-1"}>
                        {businessForm.legalName || (
                          <>
                            <AlertCircle className="h-3.5 w-3.5" /> {t("reviewStatusError")}
                          </>
                        )}
                      </span>
                    </div>

                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1 py-1">
                      <span className="text-muted-foreground font-semibold">{t("fieldNib")}</span>
                      <span className={businessForm.nibNumber ? "font-bold text-foreground" : "text-destructive font-bold flex items-center gap-1"}>
                        {businessForm.nibNumber || (
                          <>
                            <AlertCircle className="h-3.5 w-3.5" /> {t("reviewStatusError")}
                          </>
                        )}
                      </span>
                    </div>

                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1 py-1">
                      <span className="text-muted-foreground font-semibold">{t("fieldEstDate")}</span>
                      <span className={businessForm.establishedDate ? "font-bold text-foreground" : "text-destructive font-bold flex items-center gap-1"}>
                        {businessForm.establishedDate || (
                          <>
                            <AlertCircle className="h-3.5 w-3.5" /> {t("reviewStatusError")}
                          </>
                        )}
                      </span>
                    </div>

                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1 py-1">
                      <span className="text-muted-foreground font-semibold">{t("fieldIndustry")}</span>
                      <span className={businessForm.industry ? "font-bold text-foreground" : "text-destructive font-bold flex items-center gap-1"}>
                        {businessForm.industry || (
                          <>
                            <AlertCircle className="h-3.5 w-3.5" /> {t("reviewStatusError")}
                          </>
                        )}
                      </span>
                    </div>

                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1 py-1">
                      <span className="text-muted-foreground font-semibold">{t("fieldPhone")}</span>
                      <span className={businessForm.phone ? "font-bold text-foreground" : "text-destructive font-bold flex items-center gap-1"}>
                        {businessForm.phone || (
                          <>
                            <AlertCircle className="h-3.5 w-3.5" /> {t("reviewStatusError")}
                          </>
                        )}
                      </span>
                    </div>

                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1 py-1">
                      <span className="text-muted-foreground font-semibold">{t("fieldAddress")}</span>
                      <span className={businessForm.address ? "font-bold text-foreground" : "text-destructive font-bold flex items-center gap-1"}>
                        {businessForm.address || (
                          <>
                            <AlertCircle className="h-3.5 w-3.5" /> {t("reviewStatusError")}
                          </>
                        )}
                      </span>
                    </div>
                  </div>
                </div>

                {/* Dokumen Bisnis */}
                <div className="border border-border rounded-lg p-5 bg-card space-y-4">
                  <div className="flex items-center justify-between border-b border-border/50 pb-2">
                    <h4 className="text-sm font-bold text-foreground">{t("reviewDocsHeader")}</h4>
                    <button
                      onClick={() => setActiveStep(1)}
                      className="text-xs text-primary font-bold flex items-center gap-1 hover:underline cursor-pointer"
                    >
                      <Edit2 className="h-3.5 w-3.5" /> {t("reviewEditLink")}
                    </button>
                  </div>

                  <div className="space-y-3 text-xs">
                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1 py-1">
                      <span className="text-muted-foreground font-semibold">{t("reviewDocNib")}</span>
                      <span className={businessForm.nibNumber ? "text-success font-bold flex items-center gap-1" : "text-destructive font-bold flex items-center gap-1"}>
                        {businessForm.nibNumber ? (
                          <>
                            <Check className="h-3.5 w-3.5" /> {t("reviewStatusSuccess")}
                          </>
                        ) : (
                          <>
                            <AlertCircle className="h-3.5 w-3.5" /> {t("reviewStatusError")}
                          </>
                        )}
                      </span>
                    </div>

                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1 py-1">
                      <span className="text-muted-foreground font-semibold">{t("reviewDocNpwp")}</span>
                      <span className={businessForm.nibNumber ? "text-success font-bold flex items-center gap-1" : "text-destructive font-bold flex items-center gap-1"}>
                        {businessForm.nibNumber ? (
                          <>
                            <Check className="h-3.5 w-3.5" /> {t("reviewStatusSuccess")}
                          </>
                        ) : (
                          <>
                            <AlertCircle className="h-3.5 w-3.5" /> {t("reviewStatusError")}
                          </>
                        )}
                      </span>
                    </div>
                  </div>
                </div>

                {/* Aktivitas Bisnis */}
                <div className="border border-border rounded-lg p-5 bg-card space-y-4">
                  <div className="flex items-center justify-between border-b border-border/50 pb-2">
                    <h4 className="text-sm font-bold text-foreground">{t("reviewActivityHeader")}</h4>
                  </div>

                  <div className="space-y-3 text-xs">
                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1 py-1">
                      <span className="text-muted-foreground font-semibold">{t("reviewActivityQuestion")}</span>
                      <span className="font-bold text-foreground">{t("reviewActivityAnswer")}</span>
                    </div>
                  </div>
                </div>
              </div>

              {/* Bottom Buttons */}
              <div className="flex justify-between pt-4 border-t border-border/80">
                <Button
                  variant="outline"
                  onClick={() => setActiveStep(3)}
                  className="cursor-pointer font-semibold border-border hover:bg-muted/50"
                >
                  {t("btnBack")}
                </Button>
                <Button
                  onClick={handleFinalSubmit}
                  disabled={!isAllValid || submitMutation.isPending}
                  className="bg-primary text-primary-foreground hover:bg-primary/95 disabled:bg-muted disabled:text-muted-foreground cursor-pointer font-bold px-6 py-4 rounded-lg transition-all hover:-translate-y-0.5 active:translate-y-0 flex items-center gap-2"
                >
                  {submitMutation.isPending ? (
                    <>
                      <Loader2 className="h-4 w-4 animate-spin" /> {t("btnSubmitting")}
                    </>
                  ) : (
                    t("btnSubmit")
                  )}
                </Button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
