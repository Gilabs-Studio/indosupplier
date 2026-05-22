import { WorkSchedulePageClient } from "@/features/hrd/work-schedules/components/work-schedule-page-client";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

export default function WorkSchedulesPage() {
  return (
    <PermissionGuard requiredPermission="work_schedule.read">
      <WorkSchedulePageClient />
    </PermissionGuard>
  );
}
