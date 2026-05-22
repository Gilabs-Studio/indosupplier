import { useRouter } from "@/i18n/routing";
import { useAuthStore } from "../stores/use-auth-store";
import { authService } from "../services/auth-service";
import type { LoginFormData } from "../schemas/login.schema";
import type { AuthError } from "../types/errors";
import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";

function getSafeLoginErrorMessage(error: AuthError): string {
  const status = error.response?.status;
  if (status && status >= 500) {
    return "Login failed. Please try again.";
  }

  return (
    error.response?.data?.error?.message ||
    error.message ||
    "Login failed"
  );
}

export function useLogin() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const {
    setUser,
    setSessionVerified,
    isLoading: storeIsLoading,
    error: storeError,
    clearError,
  } = useAuthStore();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // NOTE: Do NOT prefetch CSRF here on mount.
  // useLoginGuard already fetches CSRF before its /auth/refresh-token probe,
  // guaranteeing a single sequential GET → POST flow. Adding a second
  // concurrent GET here races with that POST and can cause the backend to
  // issue two different CSRF tokens, breaking the Double-Submit Cookie match.
  //
  // handleLogin explicitly awaits prefetchCSRFToken() before every login
  // attempt, which is the authoritative refresh point.

  const handleLogin = async (data: LoginFormData) => {
    setIsLoading(true);
    setError(null);
    try {
      // Fetch a fresh CSRF token and pass it explicitly to the login call.
      // Using the return value (not the memoryCsrfToken cache) is the only
      // reliable approach in cross-origin staging/production where
      // document.cookie is inaccessible and module state can be null after
      // navigation or a JS-engine microtask reorder.
      const csrfToken = await authService.prefetchCSRFToken();

      const response = await authService.login(
        { email: data.email, password: data.password },
        csrfToken,
      );

      if (response.success && response.data) {
        const { user } = response.data;
        // Clear previous session cache before binding the new authenticated user.
        queryClient.clear();
        // setUser also sets isAuthenticated: true
        setUser(user);
        // Mark session as verified since we just logged in successfully
        setSessionVerified(true);
        useAuthStore.setState({
          error: null,
        });
        // Redirect to dashboard - don't set isLoading false, let redirect complete
        router.replace("/dashboard");
        return; // Exit early, keep loading state until redirect
      }
    } catch (err) {
      const authError = err as AuthError;
      const errorMessage = getSafeLoginErrorMessage(authError);
      setError(errorMessage);
      useAuthStore.setState({ isAuthenticated: false, error: errorMessage });
      setIsLoading(false);
      throw err;
    }
  };

  return {
    handleLogin,
    isLoading: isLoading || storeIsLoading,
    error: error || storeError,
    clearError: () => {
      setError(null);
      clearError();
    },
  };
}
