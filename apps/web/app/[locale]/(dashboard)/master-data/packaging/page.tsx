import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { PackagingContainer } from "@/features/master-data/product/components/packaging";
import { getTranslations } from "next-intl/server";

export async function generateMetadata() {
  const t = await getTranslations("product.packaging");
  return {
    title: t("title"),
    description: t("description"),
  };
}

export default function PackagingPage() {
  return (
    <PermissionGuard requiredPermission="packaging.read">
      <PackagingContainer />
    </PermissionGuard>
  );
}
