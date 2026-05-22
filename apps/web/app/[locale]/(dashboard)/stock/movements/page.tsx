import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { Metadata } from "next";
import { getTranslations } from "next-intl/server";

const MovementList = dynamic(
  () =>
    import("@/features/stock/stock-movement/components/movement-list").then(
      (mod) => ({ default: mod.MovementList })
    ),
  { loading: () => null }
);

export async function generateMetadata({ params }: { params: Promise<{ locale: string }> }): Promise<Metadata> {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: "stock_movement" });
  return {
    title: t("title") + " | SalesView",
  };
}

export default function StockMovementsPage() {
  return (
    <PermissionGuard requiredPermission="stock_movement.read">
      <Suspense fallback={null}>
        <MovementList />
      </Suspense>
    </PermissionGuard>
  );
}
