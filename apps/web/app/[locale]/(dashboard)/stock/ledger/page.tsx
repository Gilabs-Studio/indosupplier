import { Suspense } from "react";
import dynamic from "next/dynamic";
import { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const StockLedgerPage = dynamic(
  () =>
    import("@/features/stock/stock-ledger/components/stock-ledger-page").then(
      (mod) => ({ default: mod.StockLedgerPage })
    ),
  { loading: () => null }
);

export async function generateMetadata({ params }: { params: Promise<{ locale: string }> }): Promise<Metadata> {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: "stock_ledger" });
  return {
    title: `${t("title")} | SalesView`,
  };
}

export default function StockLedgerRoutePage() {
  return (
    <PermissionGuard requiredPermission="ledger.read">
      <Suspense fallback={null}>
        <StockLedgerPage />
      </Suspense>
    </PermissionGuard>
  );
}
