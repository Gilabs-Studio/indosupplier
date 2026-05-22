import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { OvertimePageClient } from "@/features/hrd/overtime/components/overtime-page-client";

export default function OvertimePage() {
  return (
    <PermissionGuard requiredPermission="overtime.read">
      <OvertimePageClient />
    </PermissionGuard>
  );
}
