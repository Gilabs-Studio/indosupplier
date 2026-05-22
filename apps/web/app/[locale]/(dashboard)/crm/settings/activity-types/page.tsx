import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { ActivityTypeContainer } from "@/features/crm/activity-type/components";

export default function ActivityTypePage() {
  return (
    <PermissionGuard requiredPermission="crm_activity_type.read">
      <ActivityTypeContainer />
    </PermissionGuard>
  );
}
