import { Suspense } from "react";
import dynamic from "next/dynamic";

import { PermissionGuard } from "@/features/auth/components/permission-guard";

const ARAgingReportsView = dynamic(
  () =>
    import("@/features/finance/aging-reports/components/ar-aging-reports-view").then((mod) => ({
      default: mod.ARAgingReportsView,
    })),
  { loading: () => null },
);

export default function FinanceARAgingReportsPage() {
  return (
    <PermissionGuard requiredPermission="aging_report.read">
      <Suspense fallback={null}>
        <ARAgingReportsView />
      </Suspense>
    </PermissionGuard>
  );
}
