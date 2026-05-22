import { Suspense } from "react";
import dynamic from "next/dynamic";
import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";

import { Skeleton } from "@/components/ui/skeleton";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const TravelPlannerContainer = dynamic(
  () =>
    import("@/features/travel/travel-planner/components").then((mod) => ({
      default: mod.TravelPlannerPage,
    })),
  { loading: () => null },
);

type GenerateMetadataProps = {
  params: { locale: string } | Promise<{ locale: string }>;
};

export async function generateMetadata({
  params,
}: GenerateMetadataProps): Promise<Metadata> {
  const { locale } = await Promise.resolve(params);
  const t = await getTranslations({ locale, namespace: "travelPlanner" });
  return {
    title: `${t("title")} | SalesView`,
    description: t("description"),
  };
}

function TravelPlannerSkeleton() {
  return (
    <div className="space-y-4">
      <Skeleton className="h-10 w-72" />
      <Skeleton className="h-5 w-full max-w-3xl" />
      <div className="grid gap-4 md:grid-cols-2">
        <Skeleton className="h-44 w-full" />
        <Skeleton className="h-44 w-full" />
      </div>
      <Skeleton className="h-56 w-full" />
    </div>
  );
}

export default function TravelPlannerPageRoute() {
  return (
    <PermissionGuard requiredPermission="travel_planner.read">
      <Suspense fallback={<TravelPlannerSkeleton />}>
        <TravelPlannerContainer />
      </Suspense>
    </PermissionGuard>
  );
}
