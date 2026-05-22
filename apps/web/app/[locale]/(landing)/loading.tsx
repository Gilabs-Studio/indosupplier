import { Skeleton } from "@/components/ui/skeleton";
import { PageMotion } from "@/components/motion";

export default function MarketingLoading() {
  return (
    <PageMotion className="p-6 space-y-12">
      {/* Hero Skeleton */}
      <div className="flex flex-col items-center space-y-4 py-20 text-center">
        <Skeleton className="h-6 w-32 rounded-full" />
        <Skeleton className="h-16 w-3/4 max-w-2xl" />
        <Skeleton className="h-20 w-1/2 max-w-xl" />
        <div className="flex gap-4">
          <Skeleton className="h-12 w-40 rounded-full" />
          <Skeleton className="h-12 w-40 rounded-full" />
        </div>
      </div>

      {/* Grid Skeleton */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-8 container">
        <Skeleton className="h-80 rounded-3xl" />
        <Skeleton className="h-80 rounded-3xl" />
      </div>

      {/* Rows Skeleton */}
      <div className="space-y-8 container">
        <Skeleton className="h-40 rounded-3xl" />
        <Skeleton className="h-40 rounded-3xl" />
      </div>
    </PageMotion>
  );
}
