import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const ProductDetailPage = dynamic(
  () =>
    import(
      "@/features/reports/product-analysis/components/product-detail-page"
    ).then((mod) => ({ default: mod.ProductDetailPage })),
  { loading: () => null }
);

interface ProductDetailRouteProps {
  params: Promise<{ productId: string }>;
}

export default async function ProductDetailRoute({
  params,
}: ProductDetailRouteProps) {
  const { productId } = await params;

  return (
    <PermissionGuard requiredPermission="report_product_analysis.read">
      <Suspense fallback={null}>
        <ProductDetailPage productId={productId} />
      </Suspense>
    </PermissionGuard>
  );
}
