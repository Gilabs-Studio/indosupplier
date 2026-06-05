"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { motion } from "framer-motion";
import { Eye, EyeOff, ShieldCheck, CheckCircle } from "lucide-react";
import { useQueryClient } from "@tanstack/react-query";
import { PublicLayout } from "@/features/public/components/public-layout";
import { Field, FieldLabel, FieldError, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Link, useRouter } from "@/i18n/routing";
import { authService } from "@/features/auth/services/auth-service";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import type { AuthError } from "@/features/auth/types/errors";

interface PublicRegistrationPageProps {
  locale: string;
}

const getRegisterSchema = (t: (key: string) => string) =>
  z
    .object({
      name: z.string().min(2, t("nameError")),
      email: z.string().email(t("emailError")),
      password: z.string().min(6, t("passwordError")),
      confirmPassword: z.string().min(6, t("passwordError")),
      companyName: z.string().optional(),
      industry: z.string().optional(),
    })
    .refine((data) => data.password === data.confirmPassword, {
      message: t("confirmPasswordError"),
      path: ["confirmPassword"],
    });

type RegisterFormData = z.infer<ReturnType<typeof getRegisterSchema>>;

function getSafeRegisterError(error: unknown): string {
  const authError = error as AuthError;
  return (
    authError.response?.data?.error?.message ||
    authError.message ||
    "Registration failed. Please try again."
  );
}

export function PublicRegistrationPage({ locale }: PublicRegistrationPageProps) {
  const tReg = useTranslations("public.register");
  const router = useRouter();
  const queryClient = useQueryClient();
  const { setUser, setSessionVerified } = useAuthStore();
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    setError,
  } = useForm<RegisterFormData>({
    resolver: zodResolver(getRegisterSchema(tReg)),
    defaultValues: {
      name: "",
      email: "",
      password: "",
      confirmPassword: "",
      companyName: "",
      industry: "",
    },
  });

  const onSubmit = async (data: RegisterFormData) => {
    try {
      const csrfToken = await authService.prefetchCSRFToken();
      const response = await authService.register(
        {
          name: data.name,
          email: data.email,
          password: data.password,
          company_name: data.companyName?.trim() || undefined,
          industry: data.industry?.trim() || undefined,
        },
        csrfToken,
      );

      if (response.success && response.data?.user) {
        queryClient.clear();
        setUser(response.data.user);
        setSessionVerified(true);
        router.replace("/dashboard");
      }
    } catch (error) {
      setError("root", { message: getSafeRegisterError(error) });
    }
  };

  const isId = locale === "id";

  return (
    <PublicLayout locale={locale} showFooter={false} overlapNavbar={true}>
      <div className="bg-muted/30 relative overflow-hidden flex flex-col justify-center min-h-screen pt-14 w-full">
        {/* Ambient background glows */}
        <div className="absolute top-0 left-1/4 -translate-x-1/2 w-96 h-96 bg-primary/5 rounded-full blur-3xl pointer-events-none" />
        <div className="absolute bottom-12 right-1/4 translate-x-1/2 w-96 h-96 bg-cyan/5 rounded-full blur-3xl pointer-events-none" />

        <div className="mx-auto grid w-full max-w-7xl gap-8 lg:grid-cols-[1fr_440px] lg:items-center px-4 sm:px-6 lg:px-8 py-8 sm:py-12 relative z-10">
          {/* Left marketing column */}
          <section className="space-y-6 pt-4 lg:pt-10">
            <div className="inline-flex items-center gap-1.5 rounded-full border border-primary/20 bg-primary/5 px-3.5 py-1 text-xs font-semibold text-primary">
              <ShieldCheck className="h-3.5 w-3.5" />
              {tReg("badge")}
            </div>
            <div className="max-w-xl space-y-3">
              <h1 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl font-heading leading-tight">
                {tReg("title")}
              </h1>
              <p className="text-sm leading-6 text-muted-foreground font-light">
                {tReg("subtitle")}
              </p>
            </div>
            <div className="space-y-6 pt-4">
              {[
                { title: tReg("benefitSearch"), desc: isId ? "Temukan produsen dengan NIB & sertifikasi ekspor lengkap." : "Find manufacturers with complete NIB & export certificates." },
                { title: tReg("benefitRfq"), desc: isId ? "Kirim spesifikasi produk Anda ke ratusan supplier sekaligus." : "Send your product specs to hundreds of suppliers at once." },
                { title: tReg("benefitSupplier"), desc: isId ? "Ubah akun pembeli Anda menjadi toko penjualan kapan saja." : "Convert your buyer account to a seller store at any time." }
              ].map((b, i) => (
                <div key={i} className="flex gap-4 items-start group">
                  <div className="h-10 w-10 rounded-lg bg-card border border-border flex items-center justify-center text-primary shrink-0 group-hover:scale-110 transition-transform">
                    <CheckCircle className="h-5 w-5" />
                  </div>
                  <div>
                    <h4 className="text-sm font-bold text-foreground leading-snug">{b.title}</h4>
                    <p className="text-xs text-muted-foreground font-light leading-relaxed mt-0.5">{b.desc}</p>
                  </div>
                </div>
              ))}
            </div>
          </section>

          {/* Right form column */}
          <motion.div
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.25 }}
            className="w-full"
          >
            <div className="bg-card text-card-foreground border border-border rounded-lg p-6 shadow-xs space-y-4">
              <div className="space-y-1">
                <h2 className="text-xl font-bold text-foreground">{tReg("formTitle")}</h2>
              </div>

              <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
                <FieldGroup className="gap-4">
                  <Field invalid={!!errors.name}>
                    <FieldLabel htmlFor="name">{tReg("name")}</FieldLabel>
                    <Input
                      id="name"
                      placeholder={tReg("namePlaceholder")}
                      {...register("name")}
                      disabled={isSubmitting}
                      className="h-10"
                    />
                    {errors.name && <FieldError>{errors.name.message}</FieldError>}
                  </Field>

                  <Field invalid={!!errors.email}>
                    <FieldLabel htmlFor="email">{tReg("email")}</FieldLabel>
                    <Input
                      id="email"
                      type="email"
                      placeholder={tReg("emailPlaceholder")}
                      {...register("email")}
                      disabled={isSubmitting}
                      className="h-10"
                    />
                    {errors.email && <FieldError>{errors.email.message}</FieldError>}
                  </Field>

                  <div className="grid gap-4 sm:grid-cols-2">
                    <Field invalid={!!errors.companyName}>
                      <FieldLabel htmlFor="companyName">{tReg("companyName")}</FieldLabel>
                      <Input
                        id="companyName"
                        placeholder={tReg("companyNamePlaceholder")}
                        {...register("companyName")}
                        disabled={isSubmitting}
                        className="h-10"
                      />
                      {errors.companyName && <FieldError>{errors.companyName.message}</FieldError>}
                    </Field>

                    <Field invalid={!!errors.industry}>
                      <FieldLabel htmlFor="industry">{tReg("industry")}</FieldLabel>
                      <Input
                        id="industry"
                        placeholder={tReg("industryPlaceholder")}
                        {...register("industry")}
                        disabled={isSubmitting}
                        className="h-10"
                      />
                      {errors.industry && <FieldError>{errors.industry.message}</FieldError>}
                    </Field>
                  </div>

                  <Field invalid={!!errors.password}>
                    <FieldLabel htmlFor="password">{tReg("password")}</FieldLabel>
                    <div className="relative">
                      <Input
                        id="password"
                        type={showPassword ? "text" : "password"}
                        placeholder={tReg("passwordPlaceholder")}
                        {...register("password")}
                        disabled={isSubmitting}
                        className="h-10 pr-10"
                      />
                      <button
                        type="button"
                        onClick={() => setShowPassword((value) => !value)}
                        className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground cursor-pointer"
                        aria-label={showPassword ? tReg("hidePassword") : tReg("showPassword")}
                      >
                        {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                      </button>
                    </div>
                    {errors.password && <FieldError>{errors.password.message}</FieldError>}
                  </Field>

                  <Field invalid={!!errors.confirmPassword}>
                    <FieldLabel htmlFor="confirmPassword">{tReg("confirmPassword")}</FieldLabel>
                    <div className="relative">
                      <Input
                        id="confirmPassword"
                        type={showConfirmPassword ? "text" : "password"}
                        placeholder={tReg("confirmPasswordPlaceholder")}
                        {...register("confirmPassword")}
                        disabled={isSubmitting}
                        className="h-10 pr-10"
                      />
                      <button
                        type="button"
                        onClick={() => setShowConfirmPassword((value) => !value)}
                        className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground cursor-pointer"
                        aria-label={showConfirmPassword ? tReg("hidePassword") : tReg("showPassword")}
                      >
                        {showConfirmPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                      </button>
                    </div>
                    {errors.confirmPassword && <FieldError>{errors.confirmPassword.message}</FieldError>}
                  </Field>

                  {errors.root && (
                    <Field>
                      <FieldError>{errors.root.message}</FieldError>
                    </Field>
                  )}

                  <Button
                    type="submit"
                    disabled={isSubmitting}
                    className="h-10 w-full font-semibold bg-primary hover:bg-primary/90 text-primary-foreground hover:-translate-y-0.5 active:translate-y-0 active:scale-95 transition-all shadow-xs cursor-pointer"
                  >
                    {isSubmitting ? tReg("submitting") : tReg("submit")}
                  </Button>
                </FieldGroup>
              </form>

              <p className="mt-4 text-center text-sm text-muted-foreground">
                {tReg("hasAccount")}{" "}
                <Link href="/login" className="font-semibold text-primary hover:underline cursor-pointer">
                  {tReg("loginLink")}
                </Link>
              </p>
            </div>
          </motion.div>
        </div>
      </div>
    </PublicLayout>
  );
}
