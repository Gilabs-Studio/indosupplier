import { Skeleton } from "@/components/ui/skeleton";
import { PageMotion } from "@/components/motion";

export default function SupplierReviewsLoading() {
  return (
    <PageMotion className="space-y-6 text-left pb-10">
      {/* Header Skeleton */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border/80 pb-5">
        <div className="space-y-2">
          <Skeleton className="h-8 w-64 rounded-md" />
          <Skeleton className="h-4 w-96 rounded-md" />
        </div>
        <Skeleton className="h-9 w-24 rounded-lg" />
      </div>

      {/* Metric Cards Skeleton */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {[1, 2, 3, 4].map((i) => (
          <div
            key={i}
            className="rounded-xl border border-border bg-card p-5 flex items-center justify-between shadow-xs"
          >
            <div className="space-y-2">
              <Skeleton className="h-3 w-24 rounded-md" />
              <Skeleton className="h-7 w-20 rounded-md" />
            </div>
            <Skeleton className="h-10 w-10 rounded-xl" />
          </div>
        ))}
      </div>

      {/* Filter Bar Skeleton */}
      <div className="bg-card border border-border p-4 rounded-xl space-y-3 shadow-xs">
        <Skeleton className="h-9 max-w-md w-full rounded-lg" />
        <div className="flex gap-2 pt-2 border-t border-border">
          <Skeleton className="h-7 w-24 rounded-lg" />
          <Skeleton className="h-7 w-28 rounded-lg" />
          <Skeleton className="h-7 w-24 rounded-lg" />
        </div>
      </div>

      {/* Reviews List Skeleton */}
      <div className="space-y-4">
        {[1, 2, 3].map((i) => (
          <div
            key={i}
            className="rounded-xl border border-border bg-card p-5 space-y-4 shadow-xs"
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <Skeleton className="h-10 w-10 rounded-full" />
                <div className="space-y-1.5">
                  <Skeleton className="h-4 w-44 rounded-md" />
                  <Skeleton className="h-3 w-28 rounded-md" />
                </div>
              </div>
              <Skeleton className="h-5 w-20 rounded-md" />
            </div>
            <div className="flex gap-2">
              <Skeleton className="h-6 w-36 rounded-md" />
            </div>
            <Skeleton className="h-16 w-full rounded-lg" />
          </div>
        ))}
      </div>
    </PageMotion>
  );
}
