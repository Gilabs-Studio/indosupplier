import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

// Lazy load map component for code splitting
const CustomerMapView = dynamic(
  () =>
    import("@/features/master-data/customer/components/customer").then(
      (mod) => ({ default: mod.CustomerMapView }),
    ),
  {
    loading: () => null,
  },
);

export default function CustomersPage() {
  return (
    <PermissionGuard requiredPermission="customer.read">
      <Suspense fallback={null}>
        <CustomerMapView />
      </Suspense>
    </PermissionGuard>
  );
}
