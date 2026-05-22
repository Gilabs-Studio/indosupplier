import { Suspense } from "react";
import dynamic from "next/dynamic";

// Lazy load catalog component (new layout with category tree)
const ProductCatalog = dynamic(
  () =>
    import("@/features/master-data/product/components/product/product-catalog").then(
      (mod) => ({ default: mod.ProductCatalog }),
    ),
  {
    loading: () => null,
  },
);

import { PermissionGuard } from "@/features/auth/components/permission-guard";

export default function ProductsPage() {
  return (
    <PermissionGuard requiredPermission="product.read">
      <div className="h-[calc(100vh-10rem)]">
        <Suspense fallback={null}>
          <ProductCatalog />
        </Suspense>
      </div>
    </PermissionGuard>
  );
}

