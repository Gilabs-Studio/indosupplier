import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const PurchaseReturnsContainer = dynamic(
  () =>
    import("@/features/purchase/returns/components/purchase-returns-container").then((mod) => ({
      default: mod.PurchaseReturnsContainer,
    })),
  { loading: () => null },
);

export default function PurchaseReturnsPage() {
  return (
    <PermissionGuard requiredPermission="purchase_return.read">
      <Suspense fallback={null}>
        <PurchaseReturnsContainer />
      </Suspense>
    </PermissionGuard>
  );
}
