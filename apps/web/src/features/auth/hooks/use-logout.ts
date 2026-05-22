"use client";

import { useCallback } from "react";
import { useRouter } from "@/i18n/routing";
import { useAuthStore } from "../stores/use-auth-store";
import { authService } from "../services/auth-service";
import { useQueryClient } from "@tanstack/react-query";

export function useLogout() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { logout } = useAuthStore();

  const handleLogout = useCallback(async () => {
    try {
      await authService.logout();
    } catch {
      // Ignore logout errors - still clear local state
    } finally {
      queryClient.clear();
      // Use logout() to properly clear all auth state including isSessionVerified
      logout();
      router.push("/login");
    }
  }, [queryClient, router, logout]);

  return handleLogout;
}
