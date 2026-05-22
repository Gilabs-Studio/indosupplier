import { Suspense } from "react";
import dynamic from "next/dynamic";
import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { Skeleton } from "@/components/ui/skeleton";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const POSTerminalPageClient = dynamic(
  () =>
    import("@/features/pos/terminal/components/pos-terminal-page-client").then((mod) => ({
      default: mod.POSTerminalPageClient,
    })),
  { loading: () => null },
);

type GenerateMetadataProps = {
  params: { locale: string } | Promise<{ locale: string }>;
};

function OutletSelectionSkeleton() {
  return (
    <div className="space-y-6 max-w-5xl">
      <div>
        <Skeleton className="h-7 w-52" />
        <Skeleton className="h-4 w-md mt-2" />
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
        {Array.from({ length: 8 }).map((_, index) => (
          <div key={index} className="rounded-lg border p-4 space-y-3">
            <Skeleton className="h-4 w-3/4" />
            <Skeleton className="h-3 w-1/3" />
            <Skeleton className="h-3 w-2/3" />
            <Skeleton className="h-8 w-full" />
          </div>
        ))}
      </div>
    </div>
  );
}

export async function generateMetadata({
  params,
}: GenerateMetadataProps): Promise<Metadata> {
  const { locale } = await Promise.resolve(params);
  const t = await getTranslations({ locale, namespace: "posTerminal" });
  return {
    title: `${t("title")} | SalesView`,
    description: t("subtitle"),
  };
}

export default function POSTerminalPage() {
  return (
    <PermissionGuard requiredPermission="pos.order.read">
      <Suspense fallback={<OutletSelectionSkeleton />}>
        <POSTerminalPageClient />
      </Suspense>
    </PermissionGuard>
  );
}
