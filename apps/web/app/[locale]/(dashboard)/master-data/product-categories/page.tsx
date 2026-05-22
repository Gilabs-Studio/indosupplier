import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { ProductCategoryContainer } from "@/features/master-data/product/components/product-category";
import { getTranslations } from "next-intl/server";

export async function generateMetadata() {
  const t = await getTranslations("product.productCategory");
  return {
    title: t("title"),
    description: t("description"),
  };
}

export default function ProductCategoryPage() {
  return (
    <PermissionGuard requiredPermission="product_category.read">
      <ProductCategoryContainer />
    </PermissionGuard>
  );
}
