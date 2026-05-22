import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const InvoiceList = dynamic(
  () =>
    import("@/features/sales/invoice/components/invoice-list").then(
      (mod) => ({ default: mod.InvoiceList })
    ),
  { loading: () => null }
);

export default function InvoicePage() {
  return (
    <PermissionGuard requiredPermission="customer_invoice.read">
      <Suspense fallback={null}>
        <InvoiceList />
      </Suspense>
    </PermissionGuard>
  );
}
