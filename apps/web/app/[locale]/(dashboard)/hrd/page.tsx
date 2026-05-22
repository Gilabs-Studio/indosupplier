import { PermissionGuard } from "@/features/auth/components/permission-guard";
import HrdDashboardClient from "./hrd-dashboard-client";

export default function HrdPage() {
  return (
    <PermissionGuard requiredPermission="hrd:read">
      <HrdDashboardClient />
    </PermissionGuard>
  );
}
