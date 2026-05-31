"use client";

import React, { useEffect, useState } from "react";
import { usePathname, useRouter } from "@/i18n/routing";
import { useSysadminStore } from "@/features/sysadmin/stores/use-sysadmin-store";
import { Loader2 } from "lucide-react";

export default function SysadminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const pathname = usePathname();
  const { admin, isAuthenticated, isSessionVerified, isLoading, checkSession } = useSysadminStore();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
    checkSession();
  }, [checkSession]);

  useEffect(() => {
    if (!mounted || isLoading || !isSessionVerified) return;

    const isLoginPage = pathname.endsWith("/sysadmin/login") || pathname.includes("/sysadmin/login");

    if (isAuthenticated) {
      if (isLoginPage) {
        router.push("/sysadmin");
      }
    } else {
      if (!isLoginPage) {
        router.push("/sysadmin/login");
      }
    }
  }, [mounted, isAuthenticated, isSessionVerified, isLoading, pathname, router]);

  if (!mounted || isLoading || !isSessionVerified) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-background text-foreground">
        <Loader2 className="h-10 w-10 animate-spin text-primary mb-4" />
        <p className="text-sm font-semibold tracking-wide text-muted-foreground">
          Loading Admin Portal...
        </p>
      </div>
    );
  }

  // Prevent flash of content if auth state is invalid
  const isLoginPage = pathname.endsWith("/sysadmin/login") || pathname.includes("/sysadmin/login");
  if (!isAuthenticated && !isLoginPage) {
    return null;
  }
  if (isAuthenticated && isLoginPage) {
    return null;
  }

  return <>{children}</>;
}
