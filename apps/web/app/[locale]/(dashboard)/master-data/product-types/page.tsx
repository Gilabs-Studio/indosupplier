import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { ProductTypeContainer } from "@/features/master-data/product/components/product-type";
import { getTranslations } from "next-intl/server";

export async function generateMetadata() {
  const t = await getTranslations("product.productType");
  return {
    title: t("title"),
    description: t("description"),
  };
}

export default function ProductTypePage() {
  return (
    <PermissionGuard requiredPermission="product_type.read">
      <ProductTypeContainer />
    </PermissionGuard>
  );
}
