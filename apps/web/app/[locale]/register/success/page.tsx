"use client";

import { useEffect, useState, useRef, useCallback } from "react";
import { useSearchParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { useRouter } from "@/i18n/routing";
import { motion } from "framer-motion";
import { CheckCircle2, Loader2, XCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Confetti } from "@/components/ui/confetti";
import type { ConfettiRef } from "@/components/ui/confetti";
import { AuthLayout } from "@/features/auth/components/auth-layout";
import { authService } from "@/features/auth/services/auth-service";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";

/**
 * /register/success?token=<UUID>
 *
 * Xendit redirects here after a successful payment.
 * Polls verifySession() until the backend webhook fires and the session cookie
 * is set, then redirects to the dashboard.
 */
export default function RegisterSuccessPage() {
  const t = useTranslations("auth.registerSuccess");
  const searchParams = useSearchParams();
  const router = useRouter();
  const { setUser, setSessionVerified } = useAuthStore();
  const confettiRef = useRef<ConfettiRef>(null);
  const token = searchParams.get("token");
  const isTokenValid =
    typeof token === "string" &&
    /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(token);

  const [status, setStatus] = useState<"loading" | "success" | "failed">(
    isTokenValid ? "loading" : "failed",
  );
  const maxConfirmAttempts = 8;
  const baseRetryDelayMs = 1200;

  const triggerConfetti = useCallback(() => {
    const fire = confettiRef.current?.fire;
    if (!fire) {
      return;
    }

    fire({
      particleCount: 50,
      spread: 180,
      startVelocity: 42,
      decay: 0.92,
      origin: { x: 0.5, y: 0.5 },
    });

    window.setTimeout(() => {
      fire({
        particleCount: 40,
        spread: 360,
        startVelocity: 52,
        decay: 0.9,
        scalar: 0.95,
        origin: { x: 0.5, y: 0.5 },
      });

      window.setTimeout(() => {
        fire({
          particleCount: 20,
          spread: 360,
          startVelocity: 30,
          decay: 0.94,
          scalar: 1.1,
          origin: { x: 0.5, y: 0.5 },
        });
      }, 130);
    }, 140);
  }, []);

  useEffect(() => {
    if (!isTokenValid || !token) {
      return;
    }

    let isCancelled = false;
    let retryTimer: number | undefined;

    const confirmAndLogin = async (attempt: number) => {
      try {
        const csrfToken = await authService.prefetchCSRFToken();
        const response = await authService.confirmPaidRegistration(token, csrfToken);

        if (!response.success || !response.data?.user) {
          throw new Error("confirmation failed");
        }

        if (isCancelled) {
          return;
        }

        setUser(response.data.user);
        setSessionVerified(true);
        setStatus("success");
      } catch (error) {
        const maybeAxios = error as {
          response?: {
            status?: number;
            data?: {
              error?: {
                code?: string;
                message?: string;
              };
            };
          };
          message?: string;
        };

        const statusCode = maybeAxios?.response?.status;

        // Only retry transient failures. Business conflicts/validation errors
        // (4xx, including SLUG_ALREADY_TAKEN/EMAIL_ALREADY_TAKEN) should fail fast.
        const isRetryable =
          typeof statusCode !== "number" || statusCode >= 500 || statusCode === 429;

        if (isCancelled) {
          return;
        }

        if (isRetryable && attempt < maxConfirmAttempts) {
          const delay = baseRetryDelayMs * attempt;
          retryTimer = window.setTimeout(() => {
            void confirmAndLogin(attempt + 1);
          }, delay);
          return;
        }

        if (!isRetryable) {
          setStatus("failed");
          return;
        }

        if (!isCancelled) {
          setStatus("failed");
        }
      }
    };

    void confirmAndLogin(1);

    return () => {
      isCancelled = true;
      if (retryTimer) {
        window.clearTimeout(retryTimer);
      }
    };
  }, [isTokenValid, router, setSessionVerified, setUser, token]);

  useEffect(() => {
    if (status !== "success") {
      return;
    }

    const id = window.requestAnimationFrame(() => {
      triggerConfetti();
    });

    // Redirect after confetti animation completes
    const redirectTimer = window.setTimeout(() => {
      router.replace("/dashboard");
    }, 700);

    return () => {
      window.cancelAnimationFrame(id);
      window.clearTimeout(redirectTimer);
    };
  }, [status, triggerConfetti, router]);

  return (
    <AuthLayout compact>
      {status === "success" && (
        <Confetti
          ref={confettiRef}
          manualstart={true}
          options={{
            particleCount: 150,
            spread: 90,
            origin: { y: 0.6 },
          }}
          className="fixed inset-0 z-5 h-full w-full pointer-events-none"
        />
      )}
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: 0.4 }}
        className="flex flex-col items-center gap-6 text-center"
      >
        {status === "loading" && (
          <>
            <Loader2 className="h-12 w-12 animate-spin text-primary" />
            <div>
              <h2 className="text-xl font-bold">{t("confirmingTitle")}</h2>
              <p className="mt-2 text-sm text-muted-foreground">
                {t("confirmingDescription")}
              </p>
            </div>
          </>
        )}

        {status === "success" && (
          <>
            <CheckCircle2 className="h-12 w-12 text-green-500" />
            <div>
              <h2 className="text-xl font-bold">{t("successTitle")}</h2>
              <p className="mt-2 text-sm text-muted-foreground">
                {t("successDescription")}
              </p>
            </div>
          </>
        )}

        {status === "failed" && (
          <>
            <XCircle className="h-12 w-12 text-destructive" />
            <div>
              <h2 className="text-xl font-bold">{t("failedTitle")}</h2>
              <p className="mt-2 text-sm text-muted-foreground">
                {t("failedDescription")}
              </p>
              <p className="mt-1 text-xs text-muted-foreground break-all">
                {t("referenceLabel")}: {token ?? "n/a"}
              </p>
            </div>
            <Button
              variant="outline"
              onClick={() => router.push("/register")}
            >
              {t("backToRegister")}
            </Button>
          </>
        )}
      </motion.div>
    </AuthLayout>
  );
}
