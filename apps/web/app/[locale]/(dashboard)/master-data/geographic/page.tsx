import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const GeographicMapView = dynamic(
  () =>
    import("@/features/master-data/geographic/components/geographic-map-view").then((mod) => ({
      default: mod.GeographicMapView,
    })),
  { loading: () => null }
);

export default function GeographicPage() {
  return (
    <PermissionGuard requiredPermission="geographic.read">
      <Suspense fallback={null}>
        <GeographicMapView />
      </Suspense>
    </PermissionGuard>
  );
}
