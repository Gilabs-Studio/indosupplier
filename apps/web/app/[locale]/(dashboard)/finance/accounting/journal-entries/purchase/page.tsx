import { PageMotion } from "@/components/motion";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { PurchaseJournalsList } from "@/features/finance/journals/components";

export default function FinanceAccountingPurchaseJournalPage() {
  return (
    <PermissionGuard requiredPermission="purchase_journal.read">
      <PageMotion>
        <PurchaseJournalsList />
      </PageMotion>
    </PermissionGuard>
  );
}