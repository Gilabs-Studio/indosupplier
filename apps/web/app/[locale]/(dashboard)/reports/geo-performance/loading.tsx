import { Skeleton } from "@/components/ui/skeleton";

export default function ReportsGeoPerformanceLoading() {
  return (
    <div className="space-y-6 p-6">
      <div className="space-y-2">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-4 w-96" />
      </div>
      <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
        <Skeleton className="h-20" />
        <Skeleton className="h-20" />
        <Skeleton className="h-20" />
        <Skeleton className="h-20" />
      </div>
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-[280px_1fr]">
        <Skeleton className="h-[500px]" />
        <Skeleton className="h-[500px]" />
      </div>
      <Skeleton className="h-[300px] w-full" />
    </div>
  );
}
