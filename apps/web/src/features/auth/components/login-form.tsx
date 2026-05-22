"use client";

import React, { useEffect, useRef, useState } from "react";
import { useTranslations } from "next-intl";
import { useSearchParams } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { motion } from "framer-motion";
import { Eye, EyeOff } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import { Link } from "@/i18n/routing";
import { AuthLayout } from "./auth-layout";
import { loginSchema, type LoginFormData } from "../schemas/login.schema";
import { useLogin } from "../hooks/use-login";
import { useLoginGuard } from "../hooks/use-login-guard";
import type { AuthError } from "../types/errors";
import { useRateLimitCountdown } from "@/lib/hooks/useRateLimitCountdown";
import { useRateLimitStore } from "@/lib/stores/useRateLimitStore";
import { ButtonLoading } from "@/components/loading";

export default function LoginForm() {
  const t = useTranslations("auth.login");
  const searchParams = useSearchParams();
  const hasShownPaymentToast = useRef(false);
  const isPaymentSuccessView = searchParams.get("payment") === "success";

  /**
   * useLoginGuard handles authentication verification:
   * 1. Calls /auth/refresh-token to verify session
   * 2. If 200 OK → redirects to dashboard (user already logged in)
   * 3. If 401/403 → clears localStorage and shows login form
   * 4. While checking → shows loading spinner
   *
   * CRITICAL: Never trust localStorage.isAuthenticated directly.
   */
  const { isLoading: isVerifying, shouldShowLoginForm } = useLoginGuard();
  const { handleLogin, isLoading, error, clearError } = useLogin();
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
    <AuthLayout compact={isPaymentSuccessView}>
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
        className="w-full"
      >
        <Card className="border border-border/60 bg-card/90 shadow-sm">
          <CardHeader className="space-y-2 px-6 pb-2 pt-6">
            <CardTitle className="text-2xl">{t("title")}</CardTitle>
            <CardDescription className="text-sm text-muted-foreground">
              {t("description")}
            </CardDescription>
            {isVerifying && (
              <div className="rounded-md border border-border/60 bg-muted/50 px-3 py-2 text-sm text-muted-foreground">
                {t("verifyingSession") || "Verifying session..."}
              </div>
            )}
          </CardHeader>
          <CardContent className="space-y-5 px-6 pb-6 pt-2">
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
              <FieldGroup className="space-y-4">
                <Field className="space-y-2">
                  <FieldLabel htmlFor="email">{t("emailLabel") || "Email"}</FieldLabel>
                  <Input
                    id="email"
                    type="email"
                    placeholder={t("emailPlaceholder") || "Enter your email"}
                    {...register("email")}
                    disabled={isFormLoading}
                    aria-invalid={!!errors.email}
                    className="h-11"
                  />
                  {errors.email && (
                    <FieldError>{errors.email.message}</FieldError>
                  )}
                </Field>

                <Field className="space-y-2">
                  <div className="flex items-center justify-between">
                    <FieldLabel htmlFor="password">
                      {t("passwordLabel")}
                    </FieldLabel>
                    <Link
                      href="/forgot-password"
                      className="text-xs font-medium text-primary hover:underline cursor-pointer"
                    >
                      {t("forgotPassword")}
                    </Link>
                  </div>
                  <div className="relative">
                    <Input
                      id="password"
                      type={showPassword ? "text" : "password"}
                      placeholder={t("passwordPlaceholder")}
                      {...register("password")}
                      disabled={isFormLoading}
                      aria-invalid={!!errors.password}
                      className="h-11 pr-10"
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      disabled={isFormLoading}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
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

                <Field>
                  <label className="flex cursor-pointer items-center gap-2">
                    <Checkbox
                      {...register("rememberMe")}
                      disabled={isFormLoading}
                    />
                    <span className="text-sm text-muted-foreground">
                      {t("rememberMe")}
                    </span>
                  </label>
                </Field>

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
                    className="h-11 w-full text-sm font-semibold tracking-wide"
                  >
                    <ButtonLoading loading={isFormLoading} loadingText={t("submitting")}>
                      {t("submit")}
                    </ButtonLoading>
                  </Button>
                </Field>
              </FieldGroup>
            </form>

            <div className="text-center text-sm text-muted-foreground">
              {t("noAccount")}{" "}
              <Link
                href="/register"
                className="font-medium text-primary hover:underline"
              >
                {t("createAccount")}
              </Link>
            </div>
          </CardContent>
        </Card>
      </motion.div>
    </AuthLayout>
  );
}
