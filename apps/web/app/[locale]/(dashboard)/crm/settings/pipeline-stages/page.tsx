import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { PipelineStageContainer } from "@/features/crm/pipeline-stage/components";

export default function PipelineStagePage() {
  return (
    <PermissionGuard requiredPermission="crm_pipeline_stage.read">
      <PipelineStageContainer />
    </PermissionGuard>
  );
}
