import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { PaymentTermsContainer } from "@/features/master-data/payment-and-couriers/payment-terms/components";
import { getTranslations } from "next-intl/server";

export async function generateMetadata() {
  const t = await getTranslations("paymentTerm");
  return {
    title: t("title"),
    description: t("description"),
  };
}

export default function PaymentTermsPage() {
  return (
    <PermissionGuard requiredPermission="payment_term.read">
      <PaymentTermsContainer />
    </PermissionGuard>
  );
}
