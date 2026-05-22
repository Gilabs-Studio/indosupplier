import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const SalesReturnsContainer = dynamic(
  () =>
    import("@/features/sales/returns/components/sales-returns-container").then((mod) => ({
      default: mod.SalesReturnsContainer,
    })),
  { loading: () => null },
);

export default function SalesReturnsPage() {
  return (
    <PermissionGuard requiredPermission="sales_return.read">
      <Suspense fallback={null}>
        <SalesReturnsContainer />
      </Suspense>
    </PermissionGuard>
  );
}
