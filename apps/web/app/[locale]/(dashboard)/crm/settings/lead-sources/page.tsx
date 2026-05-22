import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { LeadSourceContainer } from "@/features/crm/lead-source/components";

export default function LeadSourcePage() {
  return (
    <PermissionGuard requiredPermission="crm_lead_source.read">
      <LeadSourceContainer />
    </PermissionGuard>
  );
}
