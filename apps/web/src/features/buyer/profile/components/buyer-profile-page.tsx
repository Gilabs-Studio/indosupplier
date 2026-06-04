"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { BuyerLayout } from "../../components/buyer-layout";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Field, FieldLabel, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";

export function BuyerProfilePage() {
  const t = useTranslations("buyer.profile");
  const { user } = useAuthStore();
  const [activeTab, setActiveTab] = useState("personal");

  const [formData, setFormData] = useState({
    name: user?.name || "Yohanes",
    email: user?.email || "yohanes@example.com",
    phone: "+62 812-3456-7890",
    companyName: "PT Global Sourcing Mandiri",
    industry: "Logistics & Supply Chain",
    address: "Sudirman Central Business District (SCBD), Jakarta Selatan",
    website: "https://www.globalsourcing.co.id",
  });

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault();
    alert(t("saveSuccess"));
  };

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
            <form onSubmit={handleSave} className="space-y-6">
              {activeTab === "personal" ? (
                <FieldGroup className="space-y-5">
                  <Field className="space-y-2">
                    <FieldLabel htmlFor="fullName">Nama Lengkap</FieldLabel>
                    <Input
                      id="fullName"
                      value={formData.name}
                      onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                      required
                      className="cursor-pointer"
                    />
                  </Field>
                  <Field className="space-y-2">
                    <FieldLabel htmlFor="email">Alamat Email</FieldLabel>
                    <Input
                      id="email"
                      type="email"
                      value={formData.email}
                      disabled
                      className="bg-muted text-muted-foreground cursor-not-allowed opacity-80"
                    />
                  </Field>
                  <Field className="space-y-2">
                    <FieldLabel htmlFor="phone">Nomor Telepon</FieldLabel>
                    <Input
                      id="phone"
                      value={formData.phone}
                      onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
                      required
                      className="cursor-pointer"
                    />
                  </Field>
                </FieldGroup>
              ) : (
                <FieldGroup className="space-y-5">
                  <Field className="space-y-2">
                    <FieldLabel htmlFor="companyName">Nama Perusahaan B2B</FieldLabel>
                    <Input
                      id="companyName"
                      value={formData.companyName}
                      onChange={(e) => setFormData({ ...formData, companyName: e.target.value })}
                      required
                      className="cursor-pointer"
                    />
                  </Field>
                  <Field className="space-y-2">
                    <FieldLabel htmlFor="industry">Bidang Industri</FieldLabel>
                    <Input
                      id="industry"
                      value={formData.industry}
                      onChange={(e) => setFormData({ ...formData, industry: e.target.value })}
                      required
                      className="cursor-pointer"
                    />
                  </Field>
                  <Field className="space-y-2">
                    <FieldLabel htmlFor="website">Website Resmi (Opsional)</FieldLabel>
                    <Input
                      id="website"
                      value={formData.website}
                      onChange={(e) => setFormData({ ...formData, website: e.target.value })}
                      className="cursor-pointer"
                    />
                  </Field>
                  <Field className="space-y-2">
                    <FieldLabel htmlFor="address">Alamat Kantor Pusat</FieldLabel>
                    <textarea
                      id="address"
                      rows={3}
                      value={formData.address}
                      onChange={(e) => setFormData({ ...formData, address: e.target.value })}
                      className="w-full rounded-lg border border-border bg-card px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-hidden"
                    />
                  </Field>
                </FieldGroup>
              )}

              {/* Action Button */}
              <div className="pt-4 border-t border-border flex justify-end">
                <Button type="submit" className="bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer px-6 transition-all hover:-translate-y-0.5 active:translate-y-0 shadow-lg hover:shadow-primary/20 font-semibold">
                  {t("saveBtn")}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </BuyerLayout>
  );
}
