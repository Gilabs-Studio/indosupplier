import { Suspense } from "react";
import dynamic from "next/dynamic";

import { PermissionGuard } from "@/features/auth/components/permission-guard";

const APAgingReportsView = dynamic(
  () =>
    import("@/features/finance/aging-reports/components/ap-aging-reports-view").then((mod) => ({
      default: mod.APAgingReportsView,
    })),
  { loading: () => null },
);

export default function FinanceAPAgingReportsPage() {
  return (
    <PermissionGuard requiredPermission="aging_report.read">
      <Suspense fallback={null}>
        <APAgingReportsView />
      </Suspense>
    </PermissionGuard>
  );
}
