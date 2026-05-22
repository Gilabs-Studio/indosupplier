import { PageMotion } from "@/components/motion";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { CashBankTransactionsList } from "@/features/finance/cash-bank-transactions/components/cash-bank-transactions-list";

export default function FinanceCashBankJournalsPage() {
  return (
    <PermissionGuard requiredPermission="cash_bank_transaction.read">
      <PageMotion>
        <CashBankTransactionsList />
      </PageMotion>
    </PermissionGuard>
  );
}
