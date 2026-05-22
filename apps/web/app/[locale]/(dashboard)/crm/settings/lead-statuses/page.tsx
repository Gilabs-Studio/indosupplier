import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { LeadStatusContainer } from "@/features/crm/lead-status/components";

export default function LeadStatusPage() {
  return (
    <PermissionGuard requiredPermission="crm_lead_status.read">
      <LeadStatusContainer />
    </PermissionGuard>
  );
}
