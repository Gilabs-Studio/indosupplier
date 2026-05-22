import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

// Lazy load map component for code splitting
const SupplierMapView = dynamic(
  () =>
    import("@/features/master-data/supplier/components/supplier").then(
      (mod) => ({ default: mod.SupplierMapView }),
    ),
  {
    loading: () => null,
  },
);

export default function SuppliersPage() {
  return (
    <PermissionGuard requiredPermission="supplier.read">
      <Suspense fallback={null}>
        <SupplierMapView />
      </Suspense>
    </PermissionGuard>
  );
}

