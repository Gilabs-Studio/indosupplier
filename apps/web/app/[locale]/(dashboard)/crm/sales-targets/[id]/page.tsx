import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { SalesTargetDetailPageClient } from "./client";

interface SalesTargetDetailPageProps {
  params: Promise<{ id: string }>;
}

export default async function SalesTargetDetailPage({
  params,
}: SalesTargetDetailPageProps) {
  const { id } = await params;

  return (
    <PermissionGuard requiredPermission="sales_target.read">
      <SalesTargetDetailPageClient targetId={id} />
    </PermissionGuard>
  );
}
