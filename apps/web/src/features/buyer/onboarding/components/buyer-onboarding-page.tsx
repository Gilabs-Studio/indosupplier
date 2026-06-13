"use client";

import React from "react";
import { useTranslations } from "next-intl";
import { PublicNavbar } from "@/features/public/components/public-navbar";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Field, FieldLabel, FieldGroup, FieldError } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { onboardingSchema, type OnboardingFormData } from "../schemas/onboarding.schema";
import { useBuyerOnboarding } from "../hooks/useBuyerOnboarding";

export function BuyerOnboardingPage() {
  const t = useTranslations("buyer.onboarding");
  const { mutate: submitOnboarding, isPending } = useBuyerOnboarding();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<OnboardingFormData>({
    resolver: zodResolver(onboardingSchema),
    defaultValues: {
      company_name: "",
      industry: "manufacturing",
      phone: "",
    },
  });

  const onSubmit = (data: OnboardingFormData) => {
    submitOnboarding(data);
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
              <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
                <FieldGroup className="space-y-4">
                  <Field className="space-y-1">
                    <FieldLabel htmlFor="company_name">
                      Nama Perusahaan B2B <span className="text-destructive">*</span>
                    </FieldLabel>
                    <Input
                      id="company_name"
                      placeholder="PT Maju Bersama Corp"
                      {...register("company_name")}
                      className="cursor-pointer"
                    />
                    {errors.company_name && (
                      <FieldError>{errors.company_name.message}</FieldError>
                    )}
                  </Field>

                  <Field className="space-y-1">
                    <FieldLabel htmlFor="industry">Bidang Industri</FieldLabel>
                    <select
                      id="industry"
                      {...register("industry")}
                      className="w-full rounded-lg border border-border bg-card px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-hidden cursor-pointer"
                    >
                      <option value="manufacturing">Manufaktur & Material</option>
                      <option value="agriculture">Pertanian & Pangan</option>
                      <option value="textile">Tekstil & Konveksi</option>
                      <option value="furniture">Furnitur & Kayu</option>
                    </select>
                    {errors.industry && (
                      <FieldError>{errors.industry.message}</FieldError>
                    )}
                  </Field>

                  <Field className="space-y-1">
                    <FieldLabel htmlFor="phone">
                      Nomor Telepon Kantor / PIC <span className="text-destructive">*</span>
                    </FieldLabel>
                    <Input
                      id="phone"
                      placeholder="+62 812-3456-7890"
                      {...register("phone")}
                      className="cursor-pointer"
                    />
                    {errors.phone && (
                      <FieldError>{errors.phone.message}</FieldError>
                    )}
                  </Field>
                </FieldGroup>

                <Button
                  type="submit"
                  disabled={isPending}
                  className="w-full bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer py-2 transition-all hover:-translate-y-0.5 active:translate-y-0 shadow-lg hover:shadow-primary/20 font-bold"
                >
                  {isPending ? "Loading..." : t("btnSubmit")}
                </Button>
              </form>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
