import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const GeoPerformancePage = dynamic(
  () =>
    import(
      "@/features/reports/geo-performance/components/geo-performance-page"
    ).then((mod) => ({ default: mod.GeoPerformancePage })),
  { loading: () => null }
);

export default function ReportsGeoPerformancePage() {
  return (
    <PermissionGuard requiredPermission="report_geo_performance.read">
      <Suspense fallback={null}>
        <GeoPerformancePage />
      </Suspense>
    </PermissionGuard>
  );
}
