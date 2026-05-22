import { Metadata } from "next";
import { Suspense } from "react";
import { PageMotion } from "@/components/motion";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { StockMovementForm } from "@/features/stock/stock-movement/components/stock-movement-form";

export const metadata: Metadata = {
  title: "Create Stock Movement | SalesView",
  description: "Create a new manual stock movement",
};

export default function CreateStockMovementPage() {
  return (
    <PermissionGuard requiredPermission="stock_movement.create">
      <PageMotion className="p-6">
        <div className="mx-auto max-w-2xl">
          <Suspense fallback={null}>
            <StockMovementForm />
          </Suspense>
        </div>
      </PageMotion>
    </PermissionGuard>
  );
}
