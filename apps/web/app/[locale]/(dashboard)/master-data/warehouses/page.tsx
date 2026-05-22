import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

// Lazy load map component for code splitting
const WarehouseMapView = dynamic(
  () =>
    import("@/features/master-data/warehouse/components/warehouse").then(
      (mod) => ({ default: mod.WarehouseMapView }),
    ),
  {
    loading: () => null,
  },
);

export default function WarehousesPage() {
  return (
    <PermissionGuard requiredPermission="warehouse.read">
      <Suspense fallback={null}>
        <WarehouseMapView />
      </Suspense>
    </PermissionGuard>
  );
}
