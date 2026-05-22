import { Suspense } from "react";
import dynamic from "next/dynamic";
import { Skeleton } from "@/components/ui/skeleton";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";

const ReceivablesRecapContainer = dynamic(
  () =>
    import("@/features/sales/receivables-recap/components").then((mod) => ({
      default: mod.ReceivablesRecapContainer,
    })),
  { loading: () => null },
);

type GenerateMetadataProps = {
  params: { locale: string } | Promise<{ locale: string }>;
};

export async function generateMetadata({ params }: GenerateMetadataProps): Promise<Metadata> {
  const { locale } = await Promise.resolve(params);
  const t = await getTranslations({ locale, namespace: "receivablesRecap" });
  return {
    title: t("title") + " | SalesView",
    description: t("description"),
  };
}

function RecapSkeleton() {
  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-24 w-full rounded-xl" />
        ))}
      </div>
      <Skeleton className="h-10 w-80" />
      <div className="rounded-md border">
        <div className="p-4 space-y-4">
          {Array.from({ length: 8 }).map((_, i) => (
            <Skeleton key={i} className="h-10 w-full" />
          ))}
        </div>
      </div>
    </div>
  );
}

export default function ReceivablesRecapPage() {
  return (
    <PermissionGuard requiredPermission="sales_payment.read">
      <Suspense fallback={<RecapSkeleton />}>
        <ReceivablesRecapContainer />
      </Suspense>
    </PermissionGuard>
  );
}
