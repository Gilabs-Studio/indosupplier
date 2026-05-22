import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { DealDetailClient } from "./client";

interface DealDetailPageProps {
  params: Promise<{ id: string }>;
}

export default async function DealDetailPage({ params }: DealDetailPageProps) {
  const { id } = await params;

  return (
    <PermissionGuard requiredPermission="crm_deal.read">
      <DealDetailClient dealId={id} />
    </PermissionGuard>
  );
}
