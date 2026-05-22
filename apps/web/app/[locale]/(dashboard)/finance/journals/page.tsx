import { Suspense } from "react";
import dynamic from "next/dynamic";
import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";

import { Skeleton } from "@/components/ui/skeleton";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { PageMotion } from "@/components/motion";

const FinanceJournalsContainer = dynamic(
  () =>
    import("@/features/finance/journals/components").then((mod) => ({
      default: mod.FinanceJournalsContainer,
    })),
  { loading: () => null },
);

type GenerateMetadataProps = {
  params: { locale: string } | Promise<{ locale: string }>;
};

export async function generateMetadata({ params }: GenerateMetadataProps): Promise<Metadata> {
  const { locale } = await Promise.resolve(params);
  const t = await getTranslations({ locale, namespace: "financeJournals" });
  return {
    title: `${t("title")} | SalesView`,
    description: t("description"),
  };
}

function JournalsSkeleton() {
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

export default function FinanceJournalsPage() {
  return (
    <PermissionGuard requiredPermission="journal.read">
      <PageMotion>
        <Suspense fallback={<JournalsSkeleton />}>
          <FinanceJournalsContainer />
        </Suspense>
      </PageMotion>
    </PermissionGuard>
  );
}
