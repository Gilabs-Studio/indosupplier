import { Suspense } from "react";
import dynamic from "next/dynamic";

// Lazy load list component for code splitting
const BusinessUnitList = dynamic(
  () =>
    import("@/features/master-data/organization/components/business-unit").then(
      (mod) => ({ default: mod.BusinessUnitList }),
    ),
  {
    loading: () => null,
  },
);

import { PermissionGuard } from "@/features/auth/components/permission-guard";

export default function BusinessUnitsPage() {
  return (
    <PermissionGuard requiredPermission="business_unit.read">
      <Suspense fallback={null}>
        <BusinessUnitList />
      </Suspense>
    </PermissionGuard>
  );
}
