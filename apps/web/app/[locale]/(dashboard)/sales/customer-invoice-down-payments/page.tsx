import { PageMotion } from "@/components/motion";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { CustomerInvoiceDPContainer } from "@/features/sales/customer-invoice-down-payments/components/customer-invoice-dp-container";

export default function CustomerInvoiceDownPaymentsPage() {
  return (
    <PermissionGuard requiredPermission="customer_invoice_dp.read">
      <PageMotion>
        <CustomerInvoiceDPContainer />
      </PageMotion>
    </PermissionGuard>
  );
}
