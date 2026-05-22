import { Suspense } from "react";

import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { DashboardGrid } from "@/features/general/dashboard/components/dashboard-grid";

export default function DashboardPage() {
  return (
    <PermissionGuard requiredPermission="dashboard.view">
      <Suspense fallback={null}>
        <DashboardGrid />
      </Suspense>
    </PermissionGuard>
  );
}

