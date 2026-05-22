import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { TaskContainer } from "@/features/crm/task/components";

export default function TasksPage() {
  return (
    <PermissionGuard requiredPermission="crm_task.read">
      <TaskContainer />
    </PermissionGuard>
  );
}
