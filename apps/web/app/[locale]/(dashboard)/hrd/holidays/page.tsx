import HolidayPageClient from "@/features/hrd/holidays/components/holiday-page-client";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

export default function HolidaysPage() {
  return (
    <PermissionGuard requiredPermission="holiday.read">
      <HolidayPageClient />
    </PermissionGuard>
  );
}
