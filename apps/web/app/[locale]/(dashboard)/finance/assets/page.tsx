import { Suspense } from "react";
import dynamic from "next/dynamic";
import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";

import { Skeleton } from "@/components/ui/skeleton";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const FinanceAssetsContainer = dynamic(
  () =>
    import("@/features/finance/assets/components").then((mod) => ({
      default: mod.FinanceAssetsContainer,
    })),
  { loading: () => null },
);

type GenerateMetadataProps = {
  params: { locale: string } | Promise<{ locale: string }>;
};

export async function generateMetadata({ params }: GenerateMetadataProps): Promise<Metadata> {
  const { locale } = await Promise.resolve(params);
  const t = await getTranslations({ locale, namespace: "financeAssets" });
  return {
    title: `${t("title")} | SalesView`,
    description: t("description"),
  };
}

function PageSkeleton() {
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

export default function FinanceAssetsPage() {
  return (
    <PermissionGuard requiredPermission="asset.read">
      <Suspense fallback={<PageSkeleton />}>
        <FinanceAssetsContainer />
      </Suspense>
    </PermissionGuard>
  );
}
