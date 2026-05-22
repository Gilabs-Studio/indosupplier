"use client";

import { useCallback, useEffect } from "react";
import { usePathname, useRouter } from "@/i18n/routing";
import { AlertCircle, CreditCard } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
const BILLING_TARGET_PATH = "/profile?tab=billing&billing_tab=seat";

export function SubscriptionLifecycleGuard() {
  const router = useRouter();
  const pathname = usePathname();
  const user = useAuthStore((state) => state.user);
  const effectiveAccess = user?.subscription_access ?? null;

  const canManageBilling = user?.role?.is_owner === true;

  useEffect(() => {
    if (!effectiveAccess || !canManageBilling) {
      return;
    }

    const isBillingPage = pathname.startsWith("/profile");
    if (effectiveAccess.force_billing_redirect && !isBillingPage) {
      router.replace(BILLING_TARGET_PATH);
    }
  }, [canManageBilling, effectiveAccess, pathname, router]);

  const handleOpenBilling = useCallback(() => {
    router.push(BILLING_TARGET_PATH);
  }, [router]);

  if (!effectiveAccess || effectiveAccess.state === "active") {
    return null;
  }

  const isGrace = effectiveAccess.state === "grace_period";
  const isSuspended = effectiveAccess.state === "suspended";

  return (
    <>
      <Alert
        className={[
          "border-l-4",
          isSuspended
            ? "border-l-destructive border-destructive/60 bg-destructive/5"
            : "border-l-amber-500 border-amber-300/60 bg-amber-50/50 dark:bg-amber-900/10",
        ].join(" ")}
      >
        <AlertCircle className="h-4 w-4" />
        <AlertTitle>
          {isSuspended && "Account Suspended"}
          {isGrace && "Subscription Grace Period"}
        </AlertTitle>
        <AlertDescription className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <span>
            {effectiveAccess.message ?? "Subscription status requires billing action."} Overdue {effectiveAccess.days_overdue} day(s).
          </span>
          <Button
            type="button"
            size="sm"
            className="cursor-pointer gap-2"
            onClick={handleOpenBilling}
          >
            <CreditCard className="h-4 w-4" />
            Pay Now
          </Button>
        </AlertDescription>
      </Alert>
    </>
  );
}
