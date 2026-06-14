"use client";

import React, { useEffect, useRef, useState } from "react";
import { useTranslations } from "next-intl";
import { useSearchParams } from "next/navigation";
import { Link } from "@/i18n/routing";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { motion } from "framer-motion";
import { Eye, EyeOff } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import { loginSchema, type LoginFormData } from "../schemas/login.schema";
import { useLogin } from "../hooks/use-login";
import { useLoginGuard } from "../hooks/use-login-guard";
import type { AuthError } from "../types/errors";
import { useRateLimitCountdown } from "@/lib/hooks/useRateLimitCountdown";
import { useRateLimitStore } from "@/lib/stores/useRateLimitStore";
import { ButtonLoading } from "@/components/loading";

interface LoginFormProps {
  redirectTo?: string;
  registerHref?: string;
}

export default function LoginForm({
  redirectTo = "/dashboard",
  registerHref = "/register",
}: LoginFormProps) {
  const t = useTranslations("auth.login");
  const searchParams = useSearchParams();
  const hasShownPaymentToast = useRef(false);

  /**
   * useLoginGuard handles authentication verification:
   * 1. Calls /auth/refresh-token to verify session
   * 2. If 200 OK → redirects to dashboard (user already logged in)
   * 3. If 401/403 → clears localStorage and shows login form
   * 4. While checking → shows loading spinner
   *
   * CRITICAL: Never trust localStorage.isAuthenticated directly.
   */
  const { isLoading: isVerifying, shouldShowLoginForm } = useLoginGuard({ redirectTo });
  const { handleLogin, isLoading, error, clearError } = useLogin({ redirectTo });
  const [showPassword, setShowPassword] = useState(false);

  // Rate limit countdown hook - shows toast notification with countdown
  useRateLimitCountdown();

  // Get countdown text for display in form - update every second
  const resetTime = useRateLimitStore((state) => state.resetTime);
  const getCountdownText = useRateLimitStore((state) => state.getCountdownText);

  // Use tick state to trigger re-render every second for countdown updates
  // This avoids calling Date.now() during render and avoids synchronous setState in effects
  // The tick value is not used, only setTick is used to trigger re-renders
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const [tick, setTick] = useState(0);

  useEffect(() => {
    if (!resetTime) {
      return;
    }

    // Update tick every second to trigger re-render and recalculate countdown
    // setTick from useState is stable and doesn't need to be in dependencies
    const interval = setInterval(() => {
      setTick((prev) => prev + 1);
    }, 1000);

    return () => clearInterval(interval);
  }, [resetTime]);

  // Calculate countdown text and rate limited status
  // getCountdownText() is safe to call here because it's called during render
  // and the tick state ensures it updates every second
  const countdownText = resetTime
    ? (() => {
      const text = getCountdownText();
      if (text === "a moment") {
        return null;
      }
      return text;
    })()
    : null;

  const isRateLimited = countdownText !== null;

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    setError,
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: "",
      password: "",
      rememberMe: false,
    },
  });

  useEffect(() => {
    if (error) {
      setError("root", {
        message: error,
      });
      clearError();
    }
  }, [error, setError, clearError]);

  useEffect(() => {
    const paymentStatus = searchParams.get("payment");
    if (paymentStatus === "success" && !hasShownPaymentToast.current) {
      hasShownPaymentToast.current = true;
      toast.success(t("paymentSuccessToast"));
    }
  }, [searchParams, t]);

  const onSubmit = async (data: LoginFormData) => {
    try {
      await handleLogin(data);
      // rememberMe value is available here if needed later
    } catch (err) {
      const errorValue = err as AuthError;
      const isServerError = (errorValue.response?.status ?? 0) >= 500;
      const errorMessage =
        isServerError
          ? "Login failed. Please try again."
          : errorValue.response?.data?.error?.message ||
            errorValue.message ||
            "Login failed. Please try again.";
      setError("root", {
        message: errorMessage,
      });
    }
  };

  const isFormLoading = isLoading || isSubmitting;

  // The login form is rendered immediately to reduce perceived latency.
  // The session verification runs in the background and will redirect
  // automatically if the user is already authenticated.
  if (!shouldShowLoginForm) {
    return null;
  }

  return (
    <div className="bg-muted/10 relative overflow-hidden flex flex-col justify-center min-h-screen pt-14 w-full">
      <div className="mx-auto grid w-full max-w-7xl gap-8 lg:grid-cols-[1fr_440px] lg:items-center px-4 sm:px-6 lg:px-8 py-8 sm:py-12 relative z-10">
        {/* Left marketing column */}
        <section className="space-y-8 pt-4 lg:pt-10">
          <div className="inline-flex items-center gap-1.5 border border-border px-3 py-1 text-xs font-semibold text-muted-foreground uppercase tracking-widest">
            {t("badge") || "Platform Sourcing B2B"}
          </div>
          <div className="max-w-xl space-y-4">
            <h1 className="text-4xl font-normal font-serif leading-tight tracking-tight text-foreground">
              {t("marketingTitle") || "Sourcing Produk & Supplier Tanpa Hambatan"}
            </h1>
            <p className="text-[15px] leading-relaxed text-muted-foreground font-normal">
              {t("marketingSubtitle") || "Masuk untuk mengelola permintaan penawaran (RFQ), berdiskusi dengan supplier terverifikasi, dan memantau pesanan ekspor Anda."}
            </p>
          </div>
          <div className="space-y-6 max-w-lg">
            {[
              { title: t("benefit1Title") || "Akses Cepat ke RFQ", desc: t("benefit1Desc") || "Buat dan negosiasikan permintaan penawaran harga secara langsung." },
              { title: t("benefit2Title") || "Keamanan Dokumen Ekspor", desc: t("benefit2Desc") || "Unggah dan validasi perizinan kepabeanan dengan aman." },
              { title: t("benefit3Title") || "Notifikasi Real-time", desc: t("benefit3Desc") || "Dapatkan info terbaru tentang penawaran baru dari supplier." }
            ].map((b, i) => (
              <div key={i} className="flex gap-4 items-start pt-6 border-t border-border/60">
                <span className="font-serif text-[20px] font-normal text-[#E27D18] leading-none">
                  0{i + 1}
                </span>
                <div>
                  <h4 className="text-sm font-semibold text-foreground leading-none">{b.title}</h4>
                  <p className="text-xs text-muted-foreground font-normal leading-relaxed mt-2.5">{b.desc}</p>
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
              <h2 className="text-xl font-bold text-foreground">{t("title")}</h2>
              <p className="text-sm text-muted-foreground">{t("description")}</p>
            </div>

            {isVerifying && (
              <div className="rounded-md border border-border/60 bg-muted/50 px-3 py-2 text-sm text-muted-foreground">
                {t("verifyingSession") || "Verifying session..."}
              </div>
            )}

            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              <FieldGroup className="gap-4">
                <Field>
                  <FieldLabel htmlFor="email">{t("emailLabel") || "Email"}</FieldLabel>
                  <Input
                    id="email"
                    type="email"
                    placeholder={t("emailPlaceholder") || "Enter your email"}
                    {...register("email")}
                    disabled={isFormLoading}
                    aria-invalid={!!errors.email}
                    className="h-10"
                  />
                  {errors.email && (
                    <FieldError>{errors.email.message}</FieldError>
                  )}
                </Field>

                <Field>
                  <div className="flex items-center justify-between">
                    <FieldLabel htmlFor="password">
                      {t("passwordLabel")}
                    </FieldLabel>
                  </div>
                  <div className="relative">
                    <Input
                      id="password"
                      type={showPassword ? "text" : "password"}
                      placeholder={t("passwordPlaceholder")}
                      {...register("password")}
                      disabled={isFormLoading}
                      aria-invalid={!!errors.password}
                      className="h-10 pr-10"
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      disabled={isFormLoading}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
                      aria-label={
                        showPassword ? "Hide password" : "Show password"
                      }
                    >
                      {showPassword ? (
                        <EyeOff className="h-4 w-4" />
                      ) : (
                        <Eye className="h-4 w-4" />
                      )}
                    </button>
                  </div>
                  {errors.password && (
                    <FieldError>{errors.password.message}</FieldError>
                  )}
                </Field>

                <div className="flex items-center justify-between">
                  <Field>
                    <label className="flex cursor-pointer items-center gap-2">
                      <Checkbox
                        {...register("rememberMe")}
                        disabled={isFormLoading}
                      />
                      <span className="text-xs text-muted-foreground">
                        {t("rememberMe")}
                      </span>
                    </label>
                  </Field>
                </div>

                {errors.root && (
                  <Field>
                    <FieldError>{errors.root.message}</FieldError>
                  </Field>
                )}

                {/* Rate limit countdown display */}
                {countdownText && resetTime && (
                  <Field>
                    <div className="rounded-md bg-muted/50 px-3 py-2 text-sm text-muted-foreground">
                      <p className="font-medium">
                        {t("rateLimitMessage", { countdown: countdownText }) ||
                          `Too many login attempts. Please try again in ${countdownText}.`}
                      </p>
                    </div>
                  </Field>
                )}

                <Field className="pt-1">
                  <Button
                    type="submit"
                    disabled={isFormLoading || isRateLimited}
                    className="h-10 w-full text-[14px] font-semibold tracking-wider bg-neutral-900 hover:bg-neutral-800 text-white dark:bg-white dark:hover:bg-neutral-200 dark:text-neutral-900 border border-neutral-800/10 rounded-none hover:-translate-y-0.5 active:translate-y-0 active:scale-95 transition-all cursor-pointer"
                  >
                    <ButtonLoading loading={isFormLoading} loadingText={t("submitting")}>
                      {t("submit")}
                    </ButtonLoading>
                  </Button>
                </Field>
              </FieldGroup>
            </form>
            <p className="text-center text-sm text-muted-foreground">
              {t("noAccount")}{" "}
              <Link href={registerHref} className="font-semibold text-primary hover:underline cursor-pointer">
                {t("registerLink")}
              </Link>
            </p>
          </div>
        </motion.div>
      </div>
    </div>
  );
}
