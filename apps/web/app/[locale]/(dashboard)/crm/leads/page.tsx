import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { LeadContainer } from "@/features/crm/lead/components";

export default function LeadsPage() {
  return (
    <PermissionGuard requiredPermission="crm_lead.read">
      <LeadContainer />
    </PermissionGuard>
  );
}
