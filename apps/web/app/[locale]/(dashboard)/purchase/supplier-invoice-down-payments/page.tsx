import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { SupplierInvoiceDPContainer } from "@/features/purchase/supplier-invoice-down-payments/components/supplier-invoice-dp-container";

export default function SupplierInvoiceDownPaymentsPage() {
  return (
    <PermissionGuard requiredPermission="supplier_invoice_dp.read">
      <SupplierInvoiceDPContainer />
    </PermissionGuard>
  );
}
