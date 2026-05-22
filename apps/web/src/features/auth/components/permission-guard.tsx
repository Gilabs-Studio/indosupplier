"use client";

import { useEffect } from "react";
import { Loader2 } from "lucide-react";
import { useRouter, usePathname } from "@/i18n/routing";
import { useUserPermissions } from "@/features/master-data/user-management/hooks/use-user-permissions";
import { useHasPermission } from "@/features/master-data/user-management/hooks/use-has-permission";
import { useValidateRole } from "../hooks/use-validate-role";
import { getLocaleFromPathname } from "@/lib/i18n/get-locale";

interface PermissionGuardProps {
  readonly children: React.ReactNode;
  readonly requiredPermission: string;
  readonly fallbackUrl?: string;
}

/**
 * PermissionGuard component that checks if user has required permission
 * - Real-time role validation
 * - Permission checking
 * - Auto redirect to block page if permission revoked
 */
export function PermissionGuard({
  children,
  requiredPermission,
  fallbackUrl = "/block",
}: PermissionGuardProps) {
  const router = useRouter();
  const pathname = usePathname();
  const { data: permissionsData, isLoading } = useUserPermissions();
  const hasPermission = useHasPermission(requiredPermission);
  
  // Real-time role validation
  const { isValid: isRoleValid, isLoading: isValidatingRole } = useValidateRole();

  useEffect(() => {
    // Don't redirect while loading permissions or validating role
    if (isLoading || isValidatingRole) {
      return;
    }

    // If role is invalid, redirect to block page
    // Only redirect if we have explicit invalid result (not during initial load)
    if (isRoleValid === false) {
      const locale = getLocaleFromPathname(pathname);
      const blockPath = `/${locale}${fallbackUrl}`;

      if (typeof window !== "undefined") {
        window.location.href = blockPath;
      } else {
        router.replace(blockPath);
      }
      return;
    }

    // If permissions loaded but user doesn't have permission, redirect
    if (permissionsData && !hasPermission) {
      // Get current locale from pathname
      const locale = getLocaleFromPathname(pathname);

      // Ensure fallbackUrl is absolute (starts with /)
      const absoluteBlockUrl = fallbackUrl.startsWith("/")
        ? fallbackUrl
        : `/${fallbackUrl}`;

      // Construct absolute path with locale
      const blockPath = `/${locale}${absoluteBlockUrl}`;

      // Use window.location for absolute redirect to avoid routing issues
      if (typeof window !== "undefined") {
        window.location.href = blockPath;
      } else {
        router.replace(blockPath);
      }
    }
  }, [
    hasPermission,
    isLoading,
    isValidatingRole,
    permissionsData,
    isRoleValid,
    router,
    pathname,
    fallbackUrl,
    requiredPermission,
  ]);

  // Show nothing while checking permissions or validating role
  if (isLoading || isValidatingRole) {
    return (
      <div className="flex h-full min-h-60 w-full items-center justify-center">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  // If role invalid or no permission, don't render children (redirect will happen)
  // Only check if we have explicit results (not during initial load)
  if (isRoleValid === false || (permissionsData && !hasPermission)) {
    return null;
  }

  return <>{children}</>;
}
