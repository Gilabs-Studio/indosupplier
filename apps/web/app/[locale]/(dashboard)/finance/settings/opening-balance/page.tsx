import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { OpeningBalanceWizard } from "@/features/finance/settings/components/opening-balance-wizard";
import { PageMotion } from "@/components/motion";

export default function OpeningBalancePage() {
  return (
    <PermissionGuard requiredPermission="opening_balance.read">
      <PageMotion className="container mx-auto py-10 w-full max-w-6xl">
        <h1 className="text-3xl font-bold tracking-tight">Opening Balance</h1>
        <p className="text-muted-foreground mt-2">
          Setup opening journal balances before operational posting starts.
        </p>
        <div className="mt-6">
          <OpeningBalanceWizard />
        </div>
      </PageMotion>
    </PermissionGuard>
  );
}
