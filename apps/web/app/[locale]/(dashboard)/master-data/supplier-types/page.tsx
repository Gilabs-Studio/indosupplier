import { Suspense } from "react";
import dynamic from "next/dynamic";

// Lazy load list component for code splitting
const SupplierTypeList = dynamic(
  () =>
    import("@/features/master-data/supplier/components/supplier-type/supplier-type-list").then(
      (mod) => ({ default: mod.SupplierTypeList }),
    ),
  {
    loading: () => null,
  },
);

import { PermissionGuard } from "@/features/auth/components/permission-guard";

export default function SupplierTypesPage() {
  return (
    <PermissionGuard requiredPermission="supplier_type.read">
      <Suspense fallback={null}>
        <SupplierTypeList />
      </Suspense>
    </PermissionGuard>
  );
}
