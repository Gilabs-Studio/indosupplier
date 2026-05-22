import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { ProductSegmentContainer } from "@/features/master-data/product/components/product-segment";
import { getTranslations } from "next-intl/server";

export async function generateMetadata() {
  const t = await getTranslations("product.productSegment");
  return {
    title: t("title"),
    description: t("description"),
  };
}

export default function ProductSegmentPage() {
  return (
    <PermissionGuard requiredPermission="product_segment.read">
      <ProductSegmentContainer />
    </PermissionGuard>
  );
}
