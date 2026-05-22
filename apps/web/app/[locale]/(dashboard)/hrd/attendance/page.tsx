import AttendancePageClient from "@/features/hrd/attendance-records/components/attendance-page-client";

import { PermissionGuard } from "@/features/auth/components/permission-guard";

export default function AttendancePage() {
  return (
    <PermissionGuard requiredPermission="attendance.read">
      <AttendancePageClient />
    </PermissionGuard>
  );
}
