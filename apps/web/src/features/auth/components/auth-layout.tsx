"use client";

import { useMemo } from "react";
import type { ReactNode } from "react";
import { useTheme } from "next-themes";
import Image from "next/image";

interface AuthLayoutProps {
  readonly children: ReactNode;
  readonly compact?: boolean;
}

export function AuthLayout({ children, compact = false }: AuthLayoutProps) {
  const { resolvedTheme } = useTheme();

  const bgImageSrc = useMemo(() => {
    if (resolvedTheme === "dark") {
      return "/login2.png";
    }
    return "/login.png";
  }, [resolvedTheme]);

  if (compact) {
    return (
      <div className="flex min-h-screen items-center justify-center p-6">
        <div className="w-full max-w-md space-y-6">{children}</div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen">
      {/* Left Side - Full Image (2/3) */}
      <div className="hidden lg:block lg:w-2/3 p-6">
        <div className="relative h-full w-full overflow-hidden rounded-lg shadow-lg">
          <Image
            src={bgImageSrc}
            alt="SalesView Platform"
            fill
            className="object-cover"
            priority
            suppressHydrationWarning
          />
        </div>
      </div>

      {/* Right Side - Form (1/3) */}
      <div className="flex w-full items-center justify-center p-8 lg:w-1/3">
        <div className="w-full max-w-md space-y-8">
          {/* Form Content */}
          {children}
        </div>
      </div>
    </div>
  );
}
