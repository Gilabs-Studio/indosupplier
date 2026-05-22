import { Suspense } from "react";
import dynamic from "next/dynamic";

// Lazy load list component for code splitting
const DivisionList = dynamic(
  () =>
    import("@/features/master-data/organization/components/division").then(
      (mod) => ({ default: mod.DivisionList }),
    ),
  {
    loading: () => null,
  },
);

import { PermissionGuard } from "@/features/auth/components/permission-guard";

export default function DivisionsPage() {
  return (
    <PermissionGuard requiredPermission="division.read">
      <Suspense fallback={null}>
        <DivisionList />
      </Suspense>
    </PermissionGuard>
  );
}
