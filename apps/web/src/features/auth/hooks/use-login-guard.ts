import { useEffect, useState, useCallback, useRef, useSyncExternalStore } from "react";
import { useRouter } from "@/i18n/routing";
import { useAuthStore } from "../stores/use-auth-store";

interface UseLoginGuardOptions {
  redirectTo?: string;
}

const subscribeToHydration = (callback: () => void) => {
  callback();
  return () => {};
};
const getHydrationSnapshot = () => typeof window !== "undefined";
const getHydrationServerSnapshot = () => false;

/**
 * Authentication guard hook specifically for the login page.
 *
 * Does NOT block rendering of the login form. Verification runs in the
 * background so users can start typing immediately while the browser
 * checks whether an existing session is still valid.
 *
 * States:
 * - isLoading: true while verifying or redirecting
 * - shouldShowLoginForm: true once hydrated and not redirecting
 * - isRateLimited: true when backend returns 429
 */
export function useLoginGuard(options: UseLoginGuardOptions = {}) {
  const router = useRouter();
  const redirectTo = options.redirectTo ?? "/dashboard";
  const {
    isAuthenticated: localStorageAuth,
    isSessionVerified,
    setUser,
    setSessionVerified,
    logout,
  } = useAuthStore();

  const isHydrated = useSyncExternalStore(
    subscribeToHydration,
    getHydrationSnapshot,
    getHydrationServerSnapshot,
  );

  const [isVerifying, setIsVerifying] = useState(false);
  const [isRedirecting, setIsRedirecting] = useState(false);
  const [isRateLimited, setIsRateLimited] = useState(false);
  const hasAttemptedVerification = useRef(false);

  // Deps: exclude `isRedirecting` (guarded by ref)
  const verifyAndRedirect = useCallback(async () => {
    if (hasAttemptedVerification.current) return;
    hasAttemptedVerification.current = true;

    // Fast path: already verified this page load, skip round-trip
    if (isSessionVerified && localStorageAuth) {
      setIsRedirecting(true);
      router.push(redirectTo);
      return;
    }

    setIsVerifying(true);

    try {
      const { authService } = await import("../services/auth-service");

      try {
        await authService.prefetchCSRFToken();
      } catch {
        // Non-fatal — proceed without CSRF prefetch
      }

      const response = await authService.getMe();

      if (response?.data?.user) {
        setUser(response.data.user);
        setSessionVerified(true);
        setIsRedirecting(true);
        router.push(redirectTo);
        return;
      }

      logout();
    } catch (error: unknown) {
      const axiosError = error as {
        response?: { status?: number; data?: { error?: { code?: string } } };
      };
      const status = axiosError?.response?.status;
      const errorCode = axiosError?.response?.data?.error?.code;

      if (errorCode === "CSRF_INVALID") {
        // Stale token — allow user to retry naturally
        return;
      }

      if (status === 401) {
        logout();
        const { fullAuthCleanup } = await import("../utils/clear-auth-cookies");
        await fullAuthCleanup();
      } else if (status === 403) {
        // Session exists but lacks permissions — keep cookies for re-login
        logout();
      } else if (status === 429) {
        setIsRateLimited(true);
      } else {
        logout();
      }
    } finally {
      setIsVerifying(false);
    }
  }, [isSessionVerified, localStorageAuth, redirectTo, router, setUser, setSessionVerified, logout]);
  //   ^ removed: isRedirecting (guarded by ref)

  useEffect(() => {
    if (!isHydrated) return;

    const timeoutId = window.setTimeout(() => {
      verifyAndRedirect();
    }, 0);

    return () => window.clearTimeout(timeoutId);
  }, [isHydrated, verifyAndRedirect]);

  return {
    isLoading: isVerifying || isRedirecting,
    isRateLimited,
    shouldShowLoginForm: isHydrated && !isRedirecting,
  };
}
