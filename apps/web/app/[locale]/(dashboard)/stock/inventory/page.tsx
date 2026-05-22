import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const InventoryList = dynamic(
  () =>
    import("@/features/stock/inventory/components/inventory-list").then(
      (mod) => ({ default: mod.InventoryList })
    ),
  { loading: () => null }
);

export default function InventoryPage() {
  return (
      <PermissionGuard requiredPermission="inventory.read">
        <Suspense fallback={null}>
          <InventoryList />
        </Suspense>
      </PermissionGuard>
  );
}
