import { Skeleton } from "@/components/ui/skeleton";
import { PageMotion } from "@/components/motion";

export default function SearchLoading() {
  return (
    <PageMotion className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-8 space-y-6">
      {/* Search Header Skeleton */}
      <div className="flex flex-col sm:flex-row items-center gap-4">
        <Skeleton className="h-11 w-full rounded-lg" />
        <Skeleton className="h-11 w-32 rounded-lg shrink-0" />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        {/* Filter Sidebar Skeleton */}
        <div className="space-y-4 rounded-xl border border-border bg-card p-4 shadow-xs">
          <Skeleton className="h-5 w-24 rounded-md" />
          <Skeleton className="h-4 w-full rounded-md" />
          <Skeleton className="h-4 w-3/4 rounded-md" />
          <Skeleton className="h-4 w-5/6 rounded-md" />
          <div className="h-px bg-border my-2" />
          <Skeleton className="h-5 w-28 rounded-md" />
          <Skeleton className="h-8 w-full rounded-md" />
        </div>

        {/* Results Grid Skeleton */}
        <div className="md:col-span-3 space-y-4">
          <div className="flex items-center justify-between">
            <Skeleton className="h-5 w-32 rounded-md" />
            <Skeleton className="h-8 w-40 rounded-md" />
          </div>
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3">
            {Array.from({ length: 8 }).map((_, idx) => (
              <div key={idx} className="rounded-lg bg-card overflow-hidden space-y-2 border border-border shadow-xs">
                <Skeleton className="aspect-square w-full" />
                <div className="p-2.5 space-y-2">
                  <Skeleton className="h-3.5 w-full rounded-xs" />
                  <Skeleton className="h-4 w-2/3 rounded-xs" />
                  <Skeleton className="h-3 w-1/2 rounded-xs" />
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </PageMotion>
  );
}
