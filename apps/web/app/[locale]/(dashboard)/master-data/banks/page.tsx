import { Suspense } from "react";
import dynamic from "next/dynamic";

// Lazy load list component for code splitting
const BankList = dynamic(
  () =>
    import("@/features/master-data/supplier/components/bank/bank-list").then(
      (mod) => ({ default: mod.BankList }),
    ),
  {
    loading: () => null,
  },
);

import { PermissionGuard } from "@/features/auth/components/permission-guard";

export default function BanksPage() {
  return (
    <PermissionGuard requiredPermission="bank.read">
      <Suspense fallback={null}>
        <BankList />
      </Suspense>
    </PermissionGuard>
  );
}
