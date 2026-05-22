import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const SalesOverviewPage = dynamic(
  () =>
    import(
      "@/features/reports/sales-overview/components/sales-overview-page"
    ).then((mod) => ({ default: mod.SalesOverviewPage })),
  { loading: () => null }
);

export default function ReportsSalesOverviewPage() {
  return (
    <PermissionGuard requiredPermission="report_sales_overview.read">
      <Suspense fallback={null}>
        <SalesOverviewPage />
      </Suspense>
    </PermissionGuard>
  );
}
