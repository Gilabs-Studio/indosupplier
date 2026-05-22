import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const LeaveRequestList = dynamic(
  () =>
    import("@/features/hrd/leave-request/components/leave-request-list").then(
      (mod) => ({ default: mod.LeaveRequestList })
    ),
  { loading: () => null }
);

export default function LeaveRequestsPage() {
  return (
    <PermissionGuard requiredPermission="leave_request.read">
      <Suspense fallback={null}>
        <LeaveRequestList />
      </Suspense>
    </PermissionGuard>
  );
}
