import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { LeaveTypeContainer } from "@/features/master-data/leave-type/components";
import { getTranslations } from "next-intl/server";

export async function generateMetadata() {
  const t = await getTranslations("leaveType");
  return { title: t("title"), description: t("description") };
}

export default function LeaveTypePage() {
  return (
    <PermissionGuard requiredPermission="leave_type.read">
      <LeaveTypeContainer />
    </PermissionGuard>
  );
}
