import { useEffect, useState, useCallback, useRef, useSyncExternalStore } from "react";
import { useAuthStore } from "../stores/use-auth-store";
import { fullAuthCleanup } from "../utils/clear-auth-cookies";

const subscribeToHydration = (callback: () => void) => {
  callback();
  return () => {};
};
const getHydrationSnapshot = () => typeof window !== "undefined";
const getHydrationServerSnapshot = () => false;

/**
 * Auth guard hook that verifies session with backend before allowing access.
 *
 * CRITICAL: This hook ensures we don't trust localStorage alone.
 * It ALWAYS validates the session with the backend on every page load.
 * This handles cases where:
 * - Cookies exist but localStorage was cleared
 * - localStorage says authenticated but cookies are expired/invalid
 * - Backend restarted and invalidated all sessions
 *
 * States:
 * - isLoading: true while verifying session (show loading UI)
 * - isAuthenticated: true only after backend confirms session is valid
 * - isRateLimited: true when backend returns 429 (show retry UI)
 */
export function useAuthGuard() {
  const { user, setUser, logout, isSessionVerified, setSessionVerified } =
    useAuthStore();

  const isHydrated = useSyncExternalStore(
    subscribeToHydration,
    getHydrationSnapshot,
    getHydrationServerSnapshot,
  );

  const [isVerifying, setIsVerifying] = useState(false);
  const [verificationComplete, setVerificationComplete] = useState(false);
  const [isRateLimited, setIsRateLimited] = useState(false);
  const hasAttemptedVerification = useRef(false);

  // Deps: exclude `isVerifying` (guarded by ref) and `user` (set inside this fn)
  const verifySession = useCallback(async () => {
    if (isSessionVerified) {
      setVerificationComplete(true);
      return;
    }

    if (hasAttemptedVerification.current) return;
    hasAttemptedVerification.current = true;
    setIsVerifying(true);

    try {
      const { authService } = await import("../services/auth-service");
      const response = await authService.getMe();

      if (response?.data?.user) {
        setUser(response.data.user);
        setSessionVerified(true);
      } else {
        logout();
        await fullAuthCleanup();
      }
    } catch (error: unknown) {
      const axiosError = error as { response?: { status?: number } };
      const status = axiosError?.response?.status;

      if (status === 429) {
        // Don't logout — let user retry after cooldown
        setIsRateLimited(true);
      } else {
        logout();
        await fullAuthCleanup();
      }
    } finally {
      setIsVerifying(false);
      setVerificationComplete(true);
    }
  }, [isSessionVerified, setUser, setSessionVerified, logout]);
  //   ^ removed: isVerifying (ref-guarded), user (mutated inside)

  useEffect(() => {
    if (!isHydrated) return;

    const timeoutId = window.setTimeout(() => {
      if (isSessionVerified) {
        setVerificationComplete(true);
        return;
      }
      if (!verificationComplete) {
        verifySession();
      }
    }, 0);

    return () => window.clearTimeout(timeoutId);
  }, [isHydrated, isSessionVerified, verificationComplete, verifySession]);

  const isLoading = !isHydrated || isVerifying || !verificationComplete;
  const isAuthenticated = isHydrated && isSessionVerified && !!user;

  return {
    isAuthenticated,
    isLoading,
    isRateLimited,
    user,
    isSessionVerified,
    verificationComplete,
  };
}