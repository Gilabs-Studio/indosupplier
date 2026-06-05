"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { motion } from "framer-motion";
import { Eye, EyeOff, ShieldCheck } from "lucide-react";
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

  return (
    <PublicLayout locale={locale}>
      <div className="min-h-[calc(100vh-4rem)] bg-muted/40 px-4 py-10 sm:py-14">
        <div className="mx-auto grid w-full max-w-5xl gap-8 lg:grid-cols-[1fr_440px] lg:items-start">
          <section className="space-y-6 pt-4 lg:pt-10">
            <div className="inline-flex items-center gap-2 rounded-full border border-border bg-background px-3 py-1 text-xs font-semibold text-muted-foreground">
              <ShieldCheck className="h-3.5 w-3.5 text-primary" />
              {tReg("badge")}
            </div>
            <div className="max-w-xl space-y-3">
              <h1 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
                {tReg("title")}
              </h1>
              <p className="text-sm leading-6 text-muted-foreground sm:text-base">
                {tReg("subtitle")}
              </p>
            </div>
            <div className="grid max-w-xl gap-3 sm:grid-cols-3">
              {["benefitSearch", "benefitRfq", "benefitSupplier"].map((key) => (
                <div key={key} className="rounded-lg border border-border bg-background p-4">
                  <p className="text-sm font-semibold text-foreground">{tReg(key)}</p>
                </div>
              ))}
            </div>
          </section>

          <motion.div
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.25 }}
            className="rounded-xl border border-border bg-background p-6 shadow-sm"
          >
            <div className="mb-6 space-y-1">
              <h2 className="text-xl font-bold text-foreground">{tReg("formTitle")}</h2>
              <p className="text-sm text-muted-foreground">{tReg("formSubtitle")}</p>
            </div>

            <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
              <FieldGroup className="space-y-4">
                <Field invalid={!!errors.name}>
                  <FieldLabel htmlFor="name">{tReg("name")}</FieldLabel>
                  <Input
                    id="name"
                    placeholder={tReg("namePlaceholder")}
                    {...register("name")}
                    disabled={isSubmitting}
                    className="h-11"
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
                    className="h-11"
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
                      className="h-11"
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
                      className="h-11"
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
                      className="h-11 pr-10"
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword((value) => !value)}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
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
                      className="h-11 pr-10"
                    />
                    <button
                      type="button"
                      onClick={() => setShowConfirmPassword((value) => !value)}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
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

                <Button type="submit" disabled={isSubmitting} className="h-11 w-full font-semibold">
                  {isSubmitting ? tReg("submitting") : tReg("submit")}
                </Button>
              </FieldGroup>
            </form>

            <p className="mt-5 text-center text-sm text-muted-foreground">
              {tReg("hasAccount")}{" "}
              <Link href="/login" className="font-semibold text-primary hover:underline">
                {tReg("loginLink")}
              </Link>
            </p>
          </motion.div>
        </div>
      </div>
    </PublicLayout>
  );
}
