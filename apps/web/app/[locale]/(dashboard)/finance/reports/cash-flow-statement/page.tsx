import { Suspense } from "react";
import dynamic from "next/dynamic";
import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";

import { PageMotion } from "@/components/motion";
import { Skeleton } from "@/components/ui/skeleton";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const CashFlowStatementView = dynamic(
  () =>
    import("@/features/finance/reports/cash-flow-statement/components/cash-flow-statement-view").then((mod) => ({
      default: mod.CashFlowStatementView,
    })),
  { loading: () => null },
);

type GenerateMetadataProps = {
  params: { locale: string } | Promise<{ locale: string }>;
};

export async function generateMetadata({ params }: GenerateMetadataProps): Promise<Metadata> {
  const { locale } = await Promise.resolve(params);
  const t = await getTranslations({ locale, namespace: "financeReports" });
  return {
    title: `${t("cf_title")} | SalesView`,
    description: t("cf_description"),
  };
}

function ReportSkeleton() {
  return (
    <div className="space-y-4">
      <Skeleton className="h-10 w-64" />
      <Skeleton className="h-10 w-80" />
      <div className="rounded-md border">
        <div className="p-4 space-y-4">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-12 w-full" />
          ))}
        </div>
      </div>
    </div>
  );
}

export default function CashFlowStatementPage() {
  return (
    <PermissionGuard requiredPermission="cash_flow_statement.read">
      <PageMotion>
        <Suspense fallback={<ReportSkeleton />}>
          <CashFlowStatementView />
        </Suspense>
      </PageMotion>
    </PermissionGuard>
  );
}
