import { Suspense } from "react";
import dynamic from "next/dynamic";
import { Skeleton } from "@/components/ui/skeleton";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const PurchaseOrdersContainer = dynamic(
  () =>
    import("@/features/purchase/orders/components").then((mod) => ({
      default: mod.PurchaseOrdersContainer,
    })),
  { loading: () => null },
);

function PurchaseOrdersSkeleton() {
  return (
    <div className="space-y-4">
      <Skeleton className="h-10 w-64" />
      <Skeleton className="h-10 w-80" />
      <div className="rounded-md border">
        <div className="p-4 space-y-4">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-12 w-full" />
          ))}
        </div>
      </div>
    </div>
  );
}

export default function PurchaseOrdersPage() {
  return (
    <PermissionGuard requiredPermission="purchase_order.read">
      <Suspense fallback={<PurchaseOrdersSkeleton />}>
        <PurchaseOrdersContainer />
      </Suspense>
    </PermissionGuard>
  );
}
