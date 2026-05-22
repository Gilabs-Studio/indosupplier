import { CustomerDetailPage } from "@/features/reports/customer-research/components/customer-detail-page";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

interface Props {
  readonly params: Promise<{ customerId: string }>;
}

export default async function CustomerResearchDetailPage({ params }: Props) {
  const { customerId } = await params;

  return (
    <PermissionGuard requiredPermission="report_customer_research.read">
      <CustomerDetailPage customerId={customerId} />
    </PermissionGuard>
  );
}
