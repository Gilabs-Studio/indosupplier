import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { Metadata } from "next";
import { getTranslations } from "next-intl/server";

const StockOpnameDetailPage = dynamic(
  () =>
    import(
      "@/features/stock/stock-opname/components/stock-opname-detail-page"
    ).then((mod) => ({ default: mod.StockOpnameDetailPage })),
  { loading: () => null }
);

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string; id: string }>;
}): Promise<Metadata> {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: "stock_opname" });
  return {
    title: t("detail.title") + " | SalesView",
  };
}

export default async function StockOpnameDetailRoute({
  params,
}: {
  params: Promise<{ locale: string; id: string }>;
}) {
  const { id } = await params;

  return (
    <PermissionGuard requiredPermission="stock_opname.read">
      <Suspense fallback={null}>
        <StockOpnameDetailPage id={id} />
      </Suspense>
    </PermissionGuard>
  );
}
