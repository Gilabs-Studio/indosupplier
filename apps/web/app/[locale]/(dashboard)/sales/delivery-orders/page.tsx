import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const DeliveryList = dynamic(
  () =>
    import("@/features/sales/delivery/components/delivery-list").then(
      (mod) => ({ default: mod.DeliveryList })
    ),
  { loading: () => null }
);

export default function DeliveriesPage() {
  return (
    <PermissionGuard requiredPermission="delivery_order.read">
      <Suspense fallback={null}>
        <DeliveryList />
      </Suspense>
    </PermissionGuard>
  );
}
