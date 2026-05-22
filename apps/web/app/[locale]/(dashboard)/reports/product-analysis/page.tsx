import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const ProductAnalysisPage = dynamic(
  () =>
    import(
      "@/features/reports/product-analysis/components/product-analysis-page"
    ).then((mod) => ({ default: mod.ProductAnalysisPage })),
  { loading: () => null }
);

export default function ReportsProductAnalysisPage() {
  return (
    <PermissionGuard requiredPermission="report_product_analysis.read">
      <Suspense fallback={null}>
        <ProductAnalysisPage />
      </Suspense>
    </PermissionGuard>
  );
}
