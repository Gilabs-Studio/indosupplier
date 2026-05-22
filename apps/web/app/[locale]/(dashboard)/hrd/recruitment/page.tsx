import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const RecruitmentList = dynamic(
  () =>
    import("@/features/hrd/recruitment/components/recruitment-list").then(
      (mod) => ({ default: mod.RecruitmentList })
    ),
  { loading: () => null }
);

export default function RecruitmentPage() {
  return (
    <PermissionGuard requiredPermission="recruitment.read">
      <Suspense fallback={null}>
        <RecruitmentList />
      </Suspense>
    </PermissionGuard>
  );
}
