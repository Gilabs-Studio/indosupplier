"use client";

import { useEffect, useRef } from "react";
import { usePathname } from "next/navigation";
import { authService } from "../services/auth-service";
import { useAuthStore } from "../stores/use-auth-store";

function shouldSkipBootstrap(pathname: string | null): boolean {
  if (!pathname) {
    return false;
  }

  return pathname.includes("/sysadmin") || pathname.endsWith("/login");
}

export function AuthSessionBootstrap() {
  const pathname = usePathname();
  const attemptedPathRef = useRef<string | null>(null);
  const { setUser, setSessionVerified, isAuthenticated, isSessionVerified } = useAuthStore();

  useEffect(() => {
    if (shouldSkipBootstrap(pathname)) {
      return;
    }

    if (pathname && attemptedPathRef.current === pathname) {
      return;
    }

    if (isAuthenticated && isSessionVerified) {
      if (pathname) {
        attemptedPathRef.current = pathname;
      }
      return;
    }

    if (pathname) {
      attemptedPathRef.current = pathname;
    }

    let cancelled = false;

    const syncSession = async () => {
      try {
        try {
          await authService.prefetchCSRFToken();
        } catch {
          // Non-fatal: refresh flow may still succeed with cookie-backed token state.
        }

        const response = await authService.getMe();
        if (cancelled) {
          return;
        }

        const user = response?.data?.user ?? null;
        if (user) {
          setUser(user);
          setSessionVerified(true);
          return;
        }
      } catch {
        if (cancelled) {
          return;
        }
      }

      useAuthStore.setState({
        user: null,
        isAuthenticated: false,
        isSessionVerified: true,
        error: null,
      });
    };

    void syncSession();

    return () => {
      cancelled = true;
    };
  }, [isAuthenticated, isSessionVerified, pathname, setSessionVerified, setUser]);

  return null;
}
