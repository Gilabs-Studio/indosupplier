import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const CustomerResearchPage = dynamic(
  () =>
    import(
      "@/features/reports/customer-research/components/customer-research-page"
    ).then((mod) => ({ default: mod.CustomerResearchPage })),
  { loading: () => null }
);

export default function ReportsCustomerResearchPage() {
  return (
    <PermissionGuard requiredPermission="report_customer_research.read">
      <Suspense fallback={null}>
        <CustomerResearchPage />
      </Suspense>
    </PermissionGuard>
  );
}
