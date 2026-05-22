"use client";

import { useUserPermission } from "@/hooks/use-user-permission";

interface CanProps {
  readonly permission: string;
  readonly children: React.ReactNode;
  /** When set and user lacks permission, this is rendered instead (e.g. disabled control with tooltip). */
  readonly fallback?: React.ReactNode;
}

/**
 * Renders children only if the user has the given permission (from auth store / backend).
 * Use for buttons and actions; API must still enforce authorization.
 */
export function Can({ permission, children, fallback = null }: CanProps) {
  const allowed = useUserPermission(permission);
  if (allowed) {
    return <>{children}</>;
  }
  return <>{fallback}</>;
}
