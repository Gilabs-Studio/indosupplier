import { Suspense } from "react";
import dynamic from "next/dynamic";
import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";

import { Skeleton } from "@/components/ui/skeleton";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const FinanceBankReconciliationContainer = dynamic(
  () =>
    import("@/features/finance/bank-reconciliation/components").then((mod) => ({
      default: mod.FinanceBankReconciliationContainer,
    })),
  { loading: () => null },
);

type GenerateMetadataProps = {
  params: { locale: string } | Promise<{ locale: string }>;
};

export async function generateMetadata({ params }: GenerateMetadataProps): Promise<Metadata> {
  const { locale } = await Promise.resolve(params);
  const t = await getTranslations({ locale, namespace: "financeBankReconciliation" });

  return {
    title: `${t("title")} | SalesView`,
    description: t("description"),
  };
}

function LoadingSkeleton() {
  return (
    <div className="space-y-4">
      <Skeleton className="h-10 w-64" />
      <Skeleton className="h-10 w-full" />
      <Skeleton className="h-64 w-full" />
    </div>
  );
}

export default function FinanceBankReconciliationPage() {
  return (
    <PermissionGuard requiredPermission="bank_reconciliation.read">
      <Suspense fallback={<LoadingSkeleton />}>
        <FinanceBankReconciliationContainer />
      </Suspense>
    </PermissionGuard>
  );
}
