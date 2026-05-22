import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { LeadDetailPageClient } from "./client";

interface LeadDetailPageProps {
  params: Promise<{ id: string }>;
}

export default async function LeadDetailPage({ params }: LeadDetailPageProps) {
  const { id } = await params;

  return (
    <PermissionGuard requiredPermission="crm_lead.read">
      <LeadDetailPageClient leadId={id} />
    </PermissionGuard>
  );
}
