import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const EvaluationPage = dynamic(
  () =>
    import("@/features/hrd/evaluation/components/evaluation-page").then(
      (mod) => ({ default: mod.EvaluationPage })
    ),
  { loading: () => null }
);

export default function EvaluationRoute() {
  return (
    <PermissionGuard requiredPermission="evaluation.read">
      <Suspense fallback={null}>
        <EvaluationPage />
      </Suspense>
    </PermissionGuard>
  );
}
