import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { CourierAgencyContainer } from "@/features/master-data/payment-and-couriers/courier-agency/components";
import { getTranslations } from "next-intl/server";

export async function generateMetadata() {
  const t = await getTranslations("courierAgency");
  return { title: t("title"), description: t("description") };
}

export default function CourierAgencyPage() {
  return (
    <PermissionGuard requiredPermission="courier_agency.read">
      <CourierAgencyContainer />
    </PermissionGuard>
  );
}
