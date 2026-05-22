import { Suspense } from "react";
import dynamic from "next/dynamic";
import type { Metadata } from "next";
import { Skeleton } from "@/components/ui/skeleton";
import { PermissionGuard } from "@/features/auth/components/permission-guard";

const LiveTablePageClient = dynamic(
  () =>
    import("@/features/pos/fb/live-table/components/live-table-page-client").then((mod) => ({
      default: mod.LiveTablePageClient,
    })),
  { loading: () => null }
);

export const metadata: Metadata = {
  title: "Live Table | SalesView",
  description: "Real-time floor and table management for F&B outlets.",
};

function PageSkeleton() {
  return (
    <div className="space-y-4 p-4">
      <div className="flex items-center justify-between">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-8 w-32" />
      </div>
      <div className="grid grid-cols-2 sm:grid-cols-4 md:grid-cols-5 gap-3">
        {Array.from({ length: 10 }).map((_, i) => (
          <Skeleton key={i} className="h-32 rounded-xl" />
        ))}
      </div>
    </div>
  );
}

export default function LiveTablePage() {
  return (
    <PermissionGuard requiredPermission="pos.order.create">
      <Suspense fallback={<PageSkeleton />}>
        <LiveTablePageClient />
      </Suspense>
    </PermissionGuard>
  );
}
