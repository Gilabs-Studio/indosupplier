import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { PipelineContainer } from "@/features/crm/deal/components/pipeline-container";

export default function PipelinePage() {
  return (
    <PermissionGuard requiredPermission="crm_deal.read">
      <PipelineContainer />
    </PermissionGuard>
  );
}
