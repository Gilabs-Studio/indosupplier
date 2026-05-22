import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const SupplierDetailPage = dynamic(
  () =>
    import(
      "@/features/reports/supplier-research/components/supplier-detail-page"
    ).then((mod) => ({ default: mod.SupplierDetailPage })),
  { loading: () => null }
);

interface SupplierDetailRouteProps {
  params: Promise<{ supplierId: string }>;
}

export default async function SupplierDetailRoute({
  params,
}: SupplierDetailRouteProps) {
  const { supplierId } = await params;

  return (
    <PermissionGuard requiredPermission="report_supplier_research.read">
      <Suspense fallback={null}>
        <SupplierDetailPage supplierId={supplierId} />
      </Suspense>
    </PermissionGuard>
  );
}
