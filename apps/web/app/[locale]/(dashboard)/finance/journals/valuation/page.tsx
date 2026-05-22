import { PageMotion } from "@/components/motion";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { ValuationJournalsList } from "@/features/finance/journals/components";

export default function FinanceValuationJournalsPage() {
  return (
    <PermissionGuard requiredPermission="journal_valuation.read">
      <PageMotion>
        <ValuationJournalsList />
      </PageMotion>
    </PermissionGuard>
  );
}
