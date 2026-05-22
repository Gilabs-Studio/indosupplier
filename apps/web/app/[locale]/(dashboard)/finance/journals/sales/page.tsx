import { PageMotion } from "@/components/motion";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { SalesJournalsList } from "@/features/finance/journals/components/sales-journals-list";

export default function FinanceSalesJournalsPage() {
  return (
    <PermissionGuard requiredPermission="sales_journal.read">
      <PageMotion>
        <SalesJournalsList />
      </PageMotion>
    </PermissionGuard>
  );
}
