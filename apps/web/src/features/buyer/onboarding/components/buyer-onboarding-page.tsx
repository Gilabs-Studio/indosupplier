"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { PublicNavbar } from "@/features/public/components/public-navbar";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Field, FieldLabel, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { useRouter } from "@/i18n/routing";

export function BuyerOnboardingPage() {
  const t = useTranslations("buyer.onboarding");
  const router = useRouter();

  const [companyName, setCompanyName] = useState("");
  const [industry, setIndustry] = useState("manufacturing");
  const [phone, setPhone] = useState("");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!companyName || !phone) {
      alert("Harap isi semua kolom wajib!");
      return;
    }
    // Simulate finish onboarding and redirect to dashboard
    alert("Profil B2B Anda berhasil diperbarui! Selamat berbelanja.");
    router.push("/dashboard");
  };

  return (
    <div className="min-h-screen bg-background text-foreground flex flex-col font-sans antialiased">
      <PublicNavbar locale="id" />
      <div className="flex-1 flex items-center justify-center p-4 py-16">
        <div className="max-w-md w-full space-y-6">
          <div className="text-center space-y-2">
            <h1 className="text-3xl font-extrabold tracking-tight text-foreground font-heading">{t("title")}</h1>
            <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
          </div>

          <Card className="border border-border rounded-xl bg-card shadow-lg">
            <CardContent className="p-6">
              <form onSubmit={handleSubmit} className="space-y-5">
                <FieldGroup className="space-y-4">
                  <Field className="space-y-1">
                    <FieldLabel htmlFor="companyName">Nama Perusahaan B2B <span className="text-destructive">*</span></FieldLabel>
                    <Input
                      id="companyName"
                      value={companyName}
                      onChange={(e) => setCompanyName(e.target.value)}
                      placeholder="PT Maju Bersama Corp"
                      required
                      className="cursor-pointer"
                    />
                  </Field>

                  <Field className="space-y-1">
                    <FieldLabel htmlFor="industry">Bidang Industri</FieldLabel>
                    <select
                      id="industry"
                      value={industry}
                      onChange={(e) => setIndustry(e.target.value)}
                      className="w-full rounded-lg border border-border bg-card px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-hidden cursor-pointer"
                    >
                      <option value="manufacturing">Manufaktur & Material</option>
                      <option value="agriculture">Pertanian & Pangan</option>
                      <option value="textile">Tekstil & Konveksi</option>
                      <option value="furniture">Furnitur & Kayu</option>
                    </select>
                  </Field>

                  <Field className="space-y-1">
                    <FieldLabel htmlFor="phone">Nomor Telepon Kantor / PIC <span className="text-destructive">*</span></FieldLabel>
                    <Input
                      id="phone"
                      value={phone}
                      onChange={(e) => setPhone(e.target.value)}
                      placeholder="+62 812-3456-7890"
                      required
                      className="cursor-pointer"
                    />
                  </Field>
                </FieldGroup>

                <Button type="submit" className="w-full bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer py-2 transition-all hover:-translate-y-0.5 active:translate-y-0 shadow-lg hover:shadow-primary/20 font-bold">
                  {t("btnSubmit")}
                </Button>
              </form>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
