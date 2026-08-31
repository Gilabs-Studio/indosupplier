import { Skeleton } from "@/components/ui/skeleton";
import { PageMotion } from "@/components/motion";

export default function BuyerLoading() {
  return (
    <PageMotion className="p-6 space-y-6">
      {/* Title Header Skeleton */}
      <div className="flex flex-col gap-2">
        <Skeleton className="h-8 w-64 rounded-md" />
        <Skeleton className="h-4 w-96 rounded-md" />
      </div>

      {/* Filter Tabs & Search Skeleton */}
      <div className="space-y-4">
        <div className="flex items-center gap-3 border-b border-border pb-2">
          <Skeleton className="h-6 w-16 rounded-md" />
          <Skeleton className="h-6 w-20 rounded-md" />
          <Skeleton className="h-6 w-24 rounded-md" />
          <Skeleton className="h-6 w-20 rounded-md" />
        </div>
        <Skeleton className="h-10 w-full rounded-lg" />
      </div>

      {/* List Card Skeletons */}
      <div className="space-y-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="rounded-xl border border-border bg-card p-5 space-y-4 shadow-xs">
            <div className="flex items-center justify-between border-b border-border pb-3">
              <Skeleton className="h-4 w-40 rounded-md" />
              <Skeleton className="h-5 w-20 rounded-md" />
            </div>
            <div className="flex items-start gap-4">
              <Skeleton className="h-12 w-12 rounded-lg shrink-0" />
              <div className="flex-1 space-y-2">
                <Skeleton className="h-4 w-3/4 rounded-md" />
                <Skeleton className="h-3 w-1/2 rounded-md" />
                <Skeleton className="h-3 w-1/3 rounded-md" />
              </div>
              <Skeleton className="h-6 w-24 rounded-md shrink-0" />
            </div>
            <div className="flex items-center justify-between pt-2">
              <Skeleton className="h-4 w-32 rounded-md" />
              <div className="flex gap-2">
                <Skeleton className="h-8 w-20 rounded-md" />
                <Skeleton className="h-8 w-20 rounded-md" />
              </div>
            </div>
          </div>
        ))}
      </div>
    </PageMotion>
  );
}
