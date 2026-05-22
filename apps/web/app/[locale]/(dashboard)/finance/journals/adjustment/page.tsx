import { PageMotion } from "@/components/motion";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { AdjustmentJournalsList } from "@/features/finance/journals/components";

export default function FinanceAdjustmentJournalsPage() {
  return (
    <PermissionGuard requiredPermission="adjustment_journal.read">
      <PageMotion>
        <AdjustmentJournalsList />
      </PageMotion>
    </PermissionGuard>
  );
}
