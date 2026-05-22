import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { VisitReportContainer } from "@/features/crm/visit-report/components";

export default function VisitReportsPage() {
  return (
    <PermissionGuard requiredPermission="crm_visit.read">
      <VisitReportContainer />
    </PermissionGuard>
  );
}
