import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const SalesRepDetailPage = dynamic(
  () =>
    import(
      "@/features/reports/sales-overview/components/sales-rep-detail-page"
    ).then((mod) => ({ default: mod.SalesRepDetailPage })),
  { loading: () => null }
);

interface SalesRepDetailRouteProps {
  params: Promise<{ employeeId: string }>;
}

export default async function SalesRepDetailRoute({
  params,
}: SalesRepDetailRouteProps) {
  const { employeeId } = await params;

  return (
    <PermissionGuard requiredPermission="report_sales_overview.read">
      <Suspense fallback={null}>
        <SalesRepDetailPage employeeId={employeeId} />
      </Suspense>
    </PermissionGuard>
  );
}
