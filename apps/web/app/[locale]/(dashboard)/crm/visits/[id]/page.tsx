import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { VisitReportDetailPageClient } from "./client";

interface VisitReportDetailPageProps {
  params: Promise<{ id: string }>;
}

export default async function VisitReportDetailPage({ params }: VisitReportDetailPageProps) {
  const { id } = await params;

  return (
    <PermissionGuard requiredPermission="crm_visit.read">
      <VisitReportDetailPageClient visitId={id} />
    </PermissionGuard>
  );
}
