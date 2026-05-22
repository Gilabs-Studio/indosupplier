import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const OutletMapView = dynamic(
  () =>
    import("@/features/master-data/outlet/components/outlet").then((mod) => ({
      default: mod.OutletMapView,
    })),
  { loading: () => null },
);

export default function OutletPage() {
  return (
    <PermissionGuard requiredPermission="outlet.read">
      <Suspense fallback={null}>
        <OutletMapView />
      </Suspense>
    </PermissionGuard>
  );
}
