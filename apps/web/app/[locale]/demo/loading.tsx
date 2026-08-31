import { Skeleton } from "@/components/ui/skeleton";
import { PageMotion } from "@/components/motion";

export default function DemoLoading() {
  return (
    <PageMotion className="min-h-screen pb-16 space-y-8">
      {/* Hero Skeleton */}
      <div className="py-12 px-4 flex flex-col items-center space-y-4 text-center bg-muted/20">
        <Skeleton className="h-8 w-72 max-w-full rounded-md" />
        <Skeleton className="h-4 w-96 max-w-full rounded-md" />
        <Skeleton className="h-11 w-full max-w-2xl rounded-lg" />
        <div className="flex gap-4 pt-2">
          <Skeleton className="h-4 w-28 rounded-md" />
          <Skeleton className="h-4 w-28 rounded-md" />
          <Skeleton className="h-4 w-28 rounded-md" />
        </div>
      </div>

      {/* Filter Pills Bar Skeleton */}
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 flex items-center gap-2 overflow-hidden">
        {[1, 2, 3, 4, 5, 6, 7].map((i) => (
          <Skeleton key={i} className="h-8 w-24 rounded-full shrink-0" />
        ))}
      </div>

      {/* 6-Column Marketplace Grid Skeleton */}
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 space-y-4">
        <div className="flex items-center justify-between">
          <Skeleton className="h-6 w-36 rounded-md" />
          <Skeleton className="h-4 w-24 rounded-md" />
        </div>
        <div className="grid grid-cols-2 gap-2.5 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
          {Array.from({ length: 12 }).map((_, idx) => (
            <div key={idx} className="rounded-lg bg-card overflow-hidden space-y-2 border border-border/40 shadow-2xs">
              <Skeleton className="aspect-square w-full" />
              <div className="p-2 space-y-1.5">
                <Skeleton className="h-3.5 w-full rounded-xs" />
                <Skeleton className="h-4 w-3/4 rounded-xs" />
                <Skeleton className="h-3 w-1/2 rounded-xs" />
              </div>
            </div>
          ))}
        </div>
      </div>
    </PageMotion>
  );
}
