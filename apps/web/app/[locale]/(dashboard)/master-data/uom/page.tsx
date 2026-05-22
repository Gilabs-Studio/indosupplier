import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { UnitOfMeasureContainer } from "@/features/master-data/product/components/unit-of-measure";
import { getTranslations } from "next-intl/server";

export async function generateMetadata() {
  const t = await getTranslations("product.unitOfMeasure");
  return {
    title: t("title"),
    description: t("description"),
  };
}

export default function UnitOfMeasurePage() {
  return (
    <PermissionGuard requiredPermission="uom.read">
      <UnitOfMeasureContainer />
    </PermissionGuard>
  );
}
