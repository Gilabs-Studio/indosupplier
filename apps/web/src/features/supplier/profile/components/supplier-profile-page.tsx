"use client";

import React, { useState, useEffect } from "react";
import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Field, FieldLabel, FieldGroup } from "@/components/ui/field";
import { Save, Loader2, Eye } from "lucide-react";
import { useSupplierProfile, useUpdateSupplierProfile } from "../hooks/useProfile";
import { SupplierProfilePreview } from "./supplier-profile-preview";

export function SupplierProfilePage() {
  const t = useTranslations("supplier.profile");
  const { data: profile, isLoading } = useSupplierProfile();
  const updateMutation = useUpdateSupplierProfile();
  const [isPreview, setIsPreview] = useState(false);

  const [form, setForm] = useState({
    companyName: "",
    businessType: "",
    established: "",
    employees: "",
    email: "",
    phone: "",
    website: "",
    taxId: "",
    nib: "",
    overview: "",
    location: "",
  });

  useEffect(() => {
    if (profile) {
      const timer = setTimeout(() => {
        setForm({
          companyName: profile.companyName || "",
          businessType: profile.businessType || "",
          established: profile.established || "",
          employees: profile.employees || "",
          email: profile.email || "",
          phone: profile.phone || "",
          website: profile.website || "",
          taxId: profile.taxId || "",
          nib: profile.nib || "",
          overview: profile.overview || "",
          location: profile.location || "",
        });
      }, 0);
      return () => clearTimeout(timer);
    }
  }, [profile]);

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault();
    updateMutation.mutate(form);
  };

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center p-12 min-h-[400px]">
        <Loader2 className="h-8 w-8 animate-spin text-primary mb-4" />
        <span className="text-sm font-semibold text-muted-foreground">{t("loading")}</span>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-border/80 pb-6 text-left">
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
            {isPreview ? t("pageTitlePreview") : t("pageTitleEdit")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {isPreview ? t("pageSubtitlePreview") : t("pageSubtitleEdit")}
          </p>
        </div>
        <Button
          onClick={() => setIsPreview(!isPreview)}
          variant="outline"
          className="cursor-pointer font-semibold flex items-center gap-2 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-md"
        >
          <Eye className="h-4 w-4" />
          {isPreview ? t("btnBackToEdit") : t("btnPreviewAsBuyer")}
        </Button>
      </div>

      {isPreview ? (
        <SupplierProfilePreview />
      ) : (
        <form onSubmit={handleSave}>
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
            {/* Main Info */}
            <div className="lg:col-span-2 space-y-6">
              <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card text-left">
                <CardHeader>
                  <CardTitle className="text-base font-bold font-heading">{t("sectionGeneralTitle")}</CardTitle>
                  <CardDescription className="text-xs">{t("sectionGeneralDesc")}</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <FieldGroup className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <Field>
                      <FieldLabel>{t("fieldCompanyName")}</FieldLabel>
                      <Input
                        value={form.companyName}
                        onChange={(e) => setForm({ ...form, companyName: e.target.value })}
                        required
                      />
                    </Field>
                    <Field>
                      <FieldLabel>{t("fieldBusinessType")}</FieldLabel>
                      <Input
                        value={form.businessType}
                        onChange={(e) => setForm({ ...form, businessType: e.target.value })}
                        required
                      />
                    </Field>
                  </FieldGroup>

                  <FieldGroup className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <Field>
                      <FieldLabel>{t("fieldEstablished")}</FieldLabel>
                      <Input
                        value={form.established}
                        onChange={(e) => setForm({ ...form, established: e.target.value })}
                        required
                      />
                    </Field>
                    <Field>
                      <FieldLabel>{t("fieldEmployees")}</FieldLabel>
                      <Input
                        value={form.employees}
                        onChange={(e) => setForm({ ...form, employees: e.target.value })}
                        required
                      />
                    </Field>
                  </FieldGroup>

                  <Field>
                    <FieldLabel>{t("fieldLocation")}</FieldLabel>
                    <Input
                      value={form.location}
                      onChange={(e) => setForm({ ...form, location: e.target.value })}
                      placeholder="e.g. Kawasan Industri Jababeka, Cikarang, Jawa Barat, Indonesia"
                      required
                    />
                  </Field>

                  <Field>
                    <FieldLabel>{t("fieldOverview")}</FieldLabel>
                    <Textarea
                      rows={4}
                      value={form.overview}
                      onChange={(e) => setForm({ ...form, overview: e.target.value })}
                      required
                      className="resize-none"
                    />
                  </Field>
                </CardContent>
              </Card>

              <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card text-left">
                <CardHeader>
                  <CardTitle className="text-base font-bold font-heading">{t("sectionLegalTitle")}</CardTitle>
                  <CardDescription className="text-xs">{t("sectionLegalDesc")}</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <FieldGroup className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <Field>
                      <FieldLabel>{t("fieldTaxId")}</FieldLabel>
                      <Input
                        value={form.taxId}
                        onChange={(e) => setForm({ ...form, taxId: e.target.value })}
                        required
                      />
                    </Field>
                    <Field>
                      <FieldLabel>{t("fieldNib")}</FieldLabel>
                      <Input
                        value={form.nib}
                        onChange={(e) => setForm({ ...form, nib: e.target.value })}
                        required
                      />
                    </Field>
                  </FieldGroup>
                </CardContent>
              </Card>
            </div>

            {/* Contact Details & Save Bar */}
            <div className="space-y-6">
              <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card text-left">
                <CardHeader>
                  <CardTitle className="text-base font-bold font-heading">{t("sectionContactTitle")}</CardTitle>
                  <CardDescription className="text-xs">{t("sectionContactDesc")}</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <Field>
                    <FieldLabel>{t("fieldEmail")}</FieldLabel>
                    <Input
                      type="email"
                      value={form.email}
                      onChange={(e) => setForm({ ...form, email: e.target.value })}
                      required
                    />
                  </Field>
                  <Field>
                    <FieldLabel>{t("fieldPhone")}</FieldLabel>
                    <Input
                      value={form.phone}
                      onChange={(e) => setForm({ ...form, phone: e.target.value })}
                      required
                    />
                  </Field>
                  <Field>
                    <FieldLabel>{t("fieldWebsite")}</FieldLabel>
                    <Input
                      type="url"
                      value={form.website}
                      onChange={(e) => setForm({ ...form, website: e.target.value })}
                      required
                    />
                  </Field>
                </CardContent>
              </Card>

              <Button
                type="submit"
                disabled={updateMutation.isPending}
                className="w-full bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer font-semibold py-6 text-sm flex items-center justify-center gap-2 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/30"
              >
                {updateMutation.isPending ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin" /> {t("btnSaving")}
                  </>
                ) : (
                  <>
                    <Save className="h-4 w-4" /> {t("btnSave")}
                  </>
                )}
              </Button>
            </div>
          </div>
        </form>
      )}
    </div>
  );
}
