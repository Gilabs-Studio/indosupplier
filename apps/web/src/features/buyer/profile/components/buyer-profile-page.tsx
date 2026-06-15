"use client";

import React, { useState, useEffect } from "react";
import { useTranslations } from "next-intl";
import { BuyerLayout } from "../../components/buyer-layout";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Field, FieldLabel, FieldGroup, FieldError } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { CenteredLoading } from "@/components/loading";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  personalProfileSchema,
  companyProfileSchema,
  type PersonalProfileFormData,
  type CompanyProfileFormData,
} from "../schemas/profile.schema";
import { useBuyerProfile } from "../hooks/useBuyerProfile";

export function BuyerProfilePage() {
  const t = useTranslations("buyer.profile");
  const [activeTab, setActiveTab] = useState("personal");

  const {
    profile,
    isLoading,
    updatePersonal,
    isUpdatingPersonal,
    updateCompany,
    isUpdatingCompany,
  } = useBuyerProfile();

  // Personal Form
  const {
    register: registerPersonal,
    handleSubmit: handleSubmitPersonal,
    reset: resetPersonal,
    formState: { errors: errorsPersonal },
  } = useForm<PersonalProfileFormData>({
    resolver: zodResolver(personalProfileSchema),
  });

  // Company Form
  const {
    register: registerCompany,
    handleSubmit: handleSubmitCompany,
    reset: resetCompany,
    formState: { errors: errorsCompany },
  } = useForm<CompanyProfileFormData>({
    resolver: zodResolver(companyProfileSchema),
  });

  // Load defaults when profile data is available
  useEffect(() => {
    if (profile) {
      resetPersonal({
        full_name: profile.full_name || "",
        phone: profile.phone || "",
      });
      resetCompany({
        company_name: profile.company_name || "",
        industry: profile.industry || "manufacturing",
        website: profile.website || "",
        address: profile.address || "",
      });
    }
  }, [profile, resetPersonal, resetCompany]);

  const onSavePersonal = (data: PersonalProfileFormData) => {
    updatePersonal(data);
  };

  const onSaveCompany = (data: CompanyProfileFormData) => {
    updateCompany(data);
  };

  if (isLoading) {
    return (
      <BuyerLayout>
        <CenteredLoading />
      </BuyerLayout>
    );
  }

  return (
    <BuyerLayout>
      <div className="space-y-6 max-w-3xl">
        {/* Header */}
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">{t("title")}</h1>
          <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
        </div>

        {/* Tabs */}
        <div className="flex gap-1 border-b border-border">
          <button
            onClick={() => setActiveTab("personal")}
            className={`px-4 py-2 text-sm font-semibold rounded-t-lg transition-all cursor-pointer ${
              activeTab === "personal"
                ? "border-b-2 border-primary text-primary font-bold bg-muted/20"
                : "text-muted-foreground hover:text-foreground"
            }`}
          >
            {t("tabPersonal")}
          </button>
          <button
            onClick={() => setActiveTab("company")}
            className={`px-4 py-2 text-sm font-semibold rounded-t-lg transition-all cursor-pointer ${
              activeTab === "company"
                ? "border-b-2 border-primary text-primary font-bold bg-muted/20"
                : "text-muted-foreground hover:text-foreground"
            }`}
          >
            {t("tabCompany")}
          </button>
        </div>

        {/* Form Card */}
        <Card className="border border-border rounded-xl bg-card shadow-xs overflow-hidden">
          <CardContent className="p-6">
            {activeTab === "personal" ? (
              <form onSubmit={handleSubmitPersonal(onSavePersonal)} className="space-y-6">
                <FieldGroup className="space-y-5">
                  <Field className="space-y-2">
                    <FieldLabel htmlFor="full_name">Nama Lengkap</FieldLabel>
                    <Input
                      id="full_name"
                      {...registerPersonal("full_name")}
                      className="cursor-pointer"
                    />
                    {errorsPersonal.full_name && (
                      <FieldError>{errorsPersonal.full_name.message}</FieldError>
                    )}
                  </Field>
                  <Field className="space-y-2">
                    <FieldLabel htmlFor="email">Alamat Email</FieldLabel>
                    <Input
                      id="email"
                      type="email"
                      value={profile?.user_id ? "yohanes@example.com" : ""}
                      disabled
                      className="bg-muted text-muted-foreground cursor-not-allowed opacity-80"
                    />
                  </Field>
                  <Field className="space-y-2">
                    <FieldLabel htmlFor="phone">Nomor Telepon</FieldLabel>
                    <Input
                      id="phone"
                      {...registerPersonal("phone")}
                      className="cursor-pointer"
                    />
                    {errorsPersonal.phone && (
                      <FieldError>{errorsPersonal.phone.message}</FieldError>
                    )}
                  </Field>
                </FieldGroup>

                {/* Action Button */}
                <div className="pt-4 border-t border-border flex justify-end">
                  <Button
                    type="submit"
                    disabled={isUpdatingPersonal}
                    className="bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer px-6 transition-all hover:-translate-y-0.5 active:translate-y-0 shadow-lg hover:shadow-primary/20 font-semibold"
                  >
                    {isUpdatingPersonal ? "Loading..." : t("saveBtn")}
                  </Button>
                </div>
              </form>
            ) : (
              <form onSubmit={handleSubmitCompany(onSaveCompany)} className="space-y-6">
                <FieldGroup className="space-y-5">
                  <Field className="space-y-2">
                    <FieldLabel htmlFor="company_name">Nama Perusahaan B2B</FieldLabel>
                    <Input
                      id="company_name"
                      {...registerCompany("company_name")}
                      className="cursor-pointer"
                    />
                    {errorsCompany.company_name && (
                      <FieldError>{errorsCompany.company_name.message}</FieldError>
                    )}
                  </Field>
                  <Field className="space-y-2">
                    <FieldLabel htmlFor="industry">Bidang Industri</FieldLabel>
                    <Input
                      id="industry"
                      {...registerCompany("industry")}
                      className="cursor-pointer"
                    />
                    {errorsCompany.industry && (
                      <FieldError>{errorsCompany.industry.message}</FieldError>
                    )}
                  </Field>
                  <Field className="space-y-2">
                    <FieldLabel htmlFor="website">Website Resmi (Opsional)</FieldLabel>
                    <Input
                      id="website"
                      {...registerCompany("website")}
                      className="cursor-pointer"
                    />
                    {errorsCompany.website && (
                      <FieldError>{errorsCompany.website.message}</FieldError>
                    )}
                  </Field>
                  <Field className="space-y-2">
                    <FieldLabel htmlFor="address">Alamat Kantor Pusat</FieldLabel>
                    <textarea
                      id="address"
                      rows={3}
                      {...registerCompany("address")}
                      className="w-full rounded-lg border border-border bg-card px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-hidden"
                    />
                    {errorsCompany.address && (
                      <FieldError>{errorsCompany.address.message}</FieldError>
                    )}
                  </Field>
                </FieldGroup>

                {/* Action Button */}
                <div className="pt-4 border-t border-border flex justify-end">
                  <Button
                    type="submit"
                    disabled={isUpdatingCompany}
                    className="bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer px-6 transition-all hover:-translate-y-0.5 active:translate-y-0 shadow-lg hover:shadow-primary/20 font-semibold"
                  >
                    {isUpdatingCompany ? "Loading..." : t("saveBtn")}
                  </Button>
                </div>
              </form>
            )}
          </CardContent>
        </Card>
      </div>
    </BuyerLayout>
  );
}
