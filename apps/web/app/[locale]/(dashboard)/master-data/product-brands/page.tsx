import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { ProductBrandContainer } from "@/features/master-data/product/components/product-brand";
import { getTranslations } from "next-intl/server";

export async function generateMetadata() {
  const t = await getTranslations("product.productBrand");
  return {
    title: t("title"),
    description: t("description"),
  };
}

export default function ProductBrandPage() {
  return (
    <PermissionGuard requiredPermission="product_brand.read">
      <ProductBrandContainer />
    </PermissionGuard>
  );
}
