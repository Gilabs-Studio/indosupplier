"use client";

import React from "react";
import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Field, FieldLabel, FieldError, FieldGroup } from "@/components/ui/field";
import { Save, Loader2, Eye } from "lucide-react";
import { useSupplierProfileForm } from "../hooks/useProfile";
import { SupplierProfilePreview } from "./supplier-profile-preview";
import { ImageUpload } from "@/components/ui/image-upload";

export function SupplierProfilePage() {
  const t = useTranslations("supplier.profile");
  const {
    form,
    isLoading,
    isUpdating,
    onSubmit,
    isPreview,
    setIsPreview,
  } = useSupplierProfileForm();

  const {
    register,
    watch,
    setValue,
    formState: { errors },
  } = form;

  const currentLogo = watch("logo");

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
          type="button"
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
        <form onSubmit={onSubmit} noValidate>
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
                    <Field invalid={!!errors.companyName}>
                      <FieldLabel htmlFor="companyName">{t("fieldCompanyName")}</FieldLabel>
                      <Input
                        id="companyName"
                        {...register("companyName")}
                        placeholder="e.g. PT Baja Sentosa"
                      />
                      {errors.companyName && <FieldError>{errors.companyName.message}</FieldError>}
                    </Field>
                    <Field invalid={!!errors.businessType}>
                      <FieldLabel htmlFor="businessType">{t("fieldBusinessType")}</FieldLabel>
                      <Input
                        id="businessType"
                        {...register("businessType")}
                        placeholder="e.g. PT, CV, atau Manufacturer"
                      />
                      {errors.businessType && <FieldError>{errors.businessType.message}</FieldError>}
                    </Field>
                  </FieldGroup>

                  <FieldGroup className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <Field invalid={!!errors.established}>
                      <FieldLabel htmlFor="established">{t("fieldEstablished")}</FieldLabel>
                      <Input
                        id="established"
                        type="number"
                        min="1800"
                        max={new Date().getFullYear() + 1}
                        inputMode="numeric"
                        {...register("established")}
                        placeholder="e.g. 2018"
                      />
                      {errors.established && <FieldError>{errors.established.message}</FieldError>}
                    </Field>
                    <Field invalid={!!errors.employees}>
                      <FieldLabel htmlFor="employees">{t("fieldEmployees")}</FieldLabel>
                      <Input
                        id="employees"
                        {...register("employees")}
                        placeholder="e.g. 50-100"
                      />
                      {errors.employees && <FieldError>{errors.employees.message}</FieldError>}
                    </Field>
                  </FieldGroup>

                  <Field invalid={!!errors.location}>
                    <FieldLabel htmlFor="location">{t("fieldLocation")}</FieldLabel>
                    <Input
                      id="location"
                      {...register("location")}
                      placeholder="e.g. Kawasan Industri Jababeka, Cikarang, Jawa Barat, Indonesia"
                    />
                    {errors.location && <FieldError>{errors.location.message}</FieldError>}
                  </Field>

                  <Field invalid={!!errors.overview}>
                    <FieldLabel htmlFor="overview">{t("fieldOverview")}</FieldLabel>
                    <Textarea
                      id="overview"
                      rows={4}
                      {...register("overview")}
                      placeholder="e.g. Profil dan spesialisasi perusahaan Anda..."
                      className="resize-none"
                    />
                    {errors.overview && <FieldError>{errors.overview.message}</FieldError>}
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
                    <Field invalid={!!errors.taxId}>
                      <FieldLabel htmlFor="taxId">{t("fieldTaxId")}</FieldLabel>
                      <Input
                        id="taxId"
                        type="text"
                        inputMode="numeric"
                        {...register("taxId")}
                        placeholder="01.234.567.8-901.000"
                      />
                      {errors.taxId && <FieldError>{errors.taxId.message}</FieldError>}
                    </Field>
                    <Field invalid={!!errors.nib}>
                      <FieldLabel htmlFor="nib">{t("fieldNib")}</FieldLabel>
                      <Input
                        id="nib"
                        type="text"
                        inputMode="numeric"
                        {...register("nib")}
                        placeholder="9120001234567"
                      />
                      {errors.nib && <FieldError>{errors.nib.message}</FieldError>}
                    </Field>
                  </FieldGroup>
                </CardContent>
              </Card>
            </div>

            {/* Contact Details & Save Bar */}
            <div className="space-y-6">
              <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card text-left">
                <CardHeader>
                  <CardTitle className="text-base font-bold font-heading">{t("sectionLogoTitle")}</CardTitle>
                  <CardDescription className="text-xs">{t("sectionLogoDesc")}</CardDescription>
                </CardHeader>
                <CardContent className="flex flex-col items-center gap-4">
                  <ImageUpload
                    value={currentLogo}
                    onChange={(url) => setValue("logo", url, { shouldDirty: true })}
                    uploadFolder="logos"
                  />
                  {errors.logo && <FieldError>{errors.logo.message}</FieldError>}
                </CardContent>
              </Card>

              <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card text-left">
                <CardHeader>
                  <CardTitle className="text-base font-bold font-heading">{t("sectionContactTitle")}</CardTitle>
                  <CardDescription className="text-xs">{t("sectionContactDesc")}</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <Field invalid={!!errors.email}>
                    <FieldLabel htmlFor="email">{t("fieldEmail")}</FieldLabel>
                    <Input
                      id="email"
                      type="email"
                      {...register("email")}
                      placeholder="contact@company.com"
                    />
                    {errors.email && <FieldError>{errors.email.message}</FieldError>}
                  </Field>
                  <Field invalid={!!errors.phone}>
                    <FieldLabel htmlFor="phone">{t("fieldPhone")}</FieldLabel>
                    <Input
                      id="phone"
                      type="tel"
                      inputMode="tel"
                      {...register("phone")}
                      placeholder="+6281234567890"
                    />
                    {errors.phone && <FieldError>{errors.phone.message}</FieldError>}
                  </Field>
                  <Field invalid={!!errors.website}>
                    <FieldLabel htmlFor="website">{t("fieldWebsite")}</FieldLabel>
                    <Input
                      id="website"
                      type="url"
                      {...register("website")}
                      placeholder="https://company.com"
                    />
                    {errors.website && <FieldError>{errors.website.message}</FieldError>}
                  </Field>
                </CardContent>
              </Card>

              <Button
                type="submit"
                disabled={isUpdating}
                className="w-full bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer font-semibold py-6 text-sm flex items-center justify-center gap-2 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/30"
              >
                {isUpdating ? (
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
