import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { ProcurementTypeContainer } from "@/features/master-data/product/components/procurement-type";
import { getTranslations } from "next-intl/server";

export async function generateMetadata() {
  const t = await getTranslations("product.procurementType");
  return {
    title: t("title"),
    description: t("description"),
  };
}

export default function ProcurementTypePage() {
  return (
    <PermissionGuard requiredPermission="procurement_type.read">
      <ProcurementTypeContainer />
    </PermissionGuard>
  );
}
