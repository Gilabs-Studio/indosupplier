import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const QuotationList = dynamic(
  () =>
    import("@/features/sales/quotation/components/quotation-list").then(
      (mod) => ({ default: mod.QuotationList })
    ),
  { loading: () => null }
);

export default function QuotationPage() {
  return (
    <PermissionGuard requiredPermission="sales_quotation.read">
      <Suspense fallback={null}>
        <QuotationList />
      </Suspense>
    </PermissionGuard>
  );
}
