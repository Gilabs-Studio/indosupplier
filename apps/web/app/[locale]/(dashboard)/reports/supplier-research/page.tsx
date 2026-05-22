import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const SupplierResearchPage = dynamic(
  () =>
    import(
      "@/features/reports/supplier-research/components/supplier-research-page"
    ).then((mod) => ({ default: mod.SupplierResearchPage })),
  { loading: () => null }
);

export default function ReportsSupplierResearchPage() {
  return (
    <PermissionGuard requiredPermission="report_supplier_research.read">
      <Suspense fallback={null}>
        <SupplierResearchPage />
      </Suspense>
    </PermissionGuard>
  );
}
