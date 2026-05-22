import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { SupplierInvoiceContainer } from "@/features/purchase/supplier-invoices/components/supplier-invoice-container";

export default function SupplierInvoicesPage() {
  return (
    <PermissionGuard requiredPermission="supplier_invoice.read">
      <SupplierInvoiceContainer />
    </PermissionGuard>
  );
}
