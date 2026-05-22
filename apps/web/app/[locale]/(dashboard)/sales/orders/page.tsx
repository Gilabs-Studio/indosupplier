import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const OrderList = dynamic(
  () =>
    import("@/features/sales/order/components/order-list").then(
      (mod) => ({ default: mod.OrderList })
    ),
  { loading: () => null }
);

export default function OrdersPage() {
  return (
    <PermissionGuard requiredPermission="sales_order.read">
      <Suspense fallback={null}>
        <OrderList />
      </Suspense>
    </PermissionGuard>
  );
}
