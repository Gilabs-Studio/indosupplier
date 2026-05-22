import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

// Lazy load map component for code splitting
const CompanyMapView = dynamic(
  () =>
    import("@/features/master-data/organization/components/company").then(
      (mod) => ({ default: mod.CompanyMapView }),
    ),
  {
    loading: () => null,
  },
);

export default function CompaniesPage() {
  return (
    <PermissionGuard requiredPermission="company.read">
      <Suspense fallback={null}>
        <CompanyMapView />
      </Suspense>
    </PermissionGuard>
  );
}
