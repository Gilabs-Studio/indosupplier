"use client";

import { useEffect, useRef } from "react";
import { useQuery } from "@tanstack/react-query";
import { useRouter, usePathname } from "@/i18n/routing";
import { useAuthStore } from "../stores/use-auth-store";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import { getLocaleFromPathname } from "@/lib/i18n/get-locale";

const ROLE_VALIDATION_CACHE_TTL_MS = 5 * 60 * 1000;

/**
 * Hook untuk real-time validation role user
 * - Cek apakah role user masih valid setiap interval
 * - Auto logout jika role terhapus atau tidak valid
 * - Redirect ke block page jika permission di-revoke
 */
export function useValidateRole() {
  const router = useRouter();
  const pathname = usePathname();
  const { user } = useAuthStore();
  const t = useTranslations("auth");
  const lastValidatedRef = useRef<boolean | null>(null); // null = not yet validated, true = was valid, false = was invalid

  const { data: validationData } = useQuery({
    queryKey: ["validate-role", user?.id],
    queryFn: async () => {
      if (!user?.id) {
        return { is_valid: false };
      }

      // Validate by checking permissions endpoint
      // If it returns 404 or 401, role is invalid
      try {
        const { userService } = await import("@/features/master-data/user-management/services/user-service");
        await userService.getPermissions(user.id);
        return { is_valid: true };
      } catch (error: unknown) {
        // Check if error is 404 or 401
        const axiosError = error as { 
          response?: { 
            status?: number; 
            data?: { error?: { code?: string; message?: string } } 
          };
          message?: string;
          code?: string;
        };

        // Only treat as invalid if we have a clear auth error response
        if (
          axiosError.response?.status === 404 ||
          axiosError.response?.status === 401 ||
          axiosError.response?.data?.error?.code === "USER_NOT_FOUND" ||
          axiosError.response?.data?.error?.code === "ROLE_NOT_FOUND"
        ) {
          return { is_valid: false };
        }
        
        // For errors without response (network errors, CORS, etc.) or other errors, assume valid
        // This prevents false positives from temporary API issues
        return { is_valid: true };
      }
    },
    enabled: !!user?.id && !!user?.role,
    staleTime: ROLE_VALIDATION_CACHE_TTL_MS,
    gcTime: ROLE_VALIDATION_CACHE_TTL_MS * 2,
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
    refetchOnMount: true,
    retry: (failureCount, error) => {
      // Don't retry on auth errors
      const axiosError = error as { response?: { status?: number; data?: { error?: { code?: string } } } };
      if (
        axiosError.response?.status === 401 ||
        axiosError.response?.status === 403 ||
        axiosError.response?.status === 404 ||
        axiosError.response?.data?.error?.code === "USER_NOT_FOUND" ||
        axiosError.response?.data?.error?.code === "ROLE_NOT_FOUND"
      ) {
        return false;
      }
      // Retry once for network errors
      return failureCount < 1;
    },
    retryDelay: 2000, // Wait 2 seconds before retry
  });

  useEffect(() => {
    // Don't redirect on initial load or if still loading.
    // Only redirect if we have a definitive invalid result AND we previously had a valid result.
    if (
      validationData &&
      !validationData.is_valid &&
      lastValidatedRef.current === true &&
      user?.role
    ) {
      lastValidatedRef.current = false;

      toast.error(
        t("roleInvalid.title", { defaultValue: "Access Revoked" }),
        {
          description: t("roleInvalid.description", {
            defaultValue:
              "Your role has been removed or permissions have been revoked. You will be redirected.",
          }),
          duration: 5000,
        }
      );

      const locale = getLocaleFromPathname(pathname);
      const blockPath = `/${locale}/block`;

      if (typeof window !== "undefined") {
        window.location.href = blockPath;
      } else {
        router.replace(blockPath);
      }
      return;
    }

    if (validationData?.is_valid) {
      lastValidatedRef.current = true;
    }
  }, [validationData, pathname, router, t, user?.role]);

  // Return isValid based on validationData
  // If validationData is undefined (still loading), return true to prevent false positives
  // Only return false if we have explicit invalid result
  const isValid = validationData === undefined ? true : (validationData.is_valid ?? true);

  return {
    isValid,
    isLoading: !validationData,
  };
}

