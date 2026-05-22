import { Suspense } from "react";
import dynamic from "next/dynamic";
import { Skeleton } from "@/components/ui/skeleton";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const PurchaseRequisitionsContainer = dynamic(
  () =>
    import("@/features/purchase/requisitions/components").then((mod) => ({
      default: mod.PurchaseRequisitionsContainer,
    })),
  { loading: () => null },
);

function PurchaseRequisitionsSkeleton() {
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

export default function PurchaseRequisitionsPage() {
  return (
    <PermissionGuard requiredPermission="purchase_requisition.read">
      <Suspense fallback={<PurchaseRequisitionsSkeleton />}>
        <PurchaseRequisitionsContainer />
      </Suspense>
    </PermissionGuard>
  );
}
