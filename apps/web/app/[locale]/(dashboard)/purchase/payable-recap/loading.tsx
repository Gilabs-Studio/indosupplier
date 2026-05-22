import { Skeleton } from "@/components/ui/skeleton";
import { PageMotion } from "@/components/motion";

export default function PayableRecapLoading() {
  return (
    <PageMotion className="space-y-8 flex flex-col p-6 animate-in fade-in duration-500">
      {/* Header Skeleton */}
      <div className="space-y-3">
        <Skeleton className="h-10 w-72 rounded-full" />
        <Skeleton className="h-5 w-96 rounded-full opacity-60" />
      </div>

      {/* Summary Cards Skeleton */}
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="rounded-2xl border bg-card/50 p-5 h-28 flex flex-col justify-between shadow-sm border-border/50">
            <div className="flex justify-between w-full">
              <Skeleton className="h-4 w-28 rounded-full" />
              <Skeleton className="h-10 w-10 rounded-xl" />
            </div>
            <Skeleton className="h-9 w-36 rounded-full mt-auto" />
          </div>
        ))}
      </div>

      {/* Toolbar Skeleton */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 w-full">
        <Skeleton className="h-11 w-full sm:max-w-md rounded-xl" />
        <div className="flex gap-3">
          <Skeleton className="h-11 w-32 rounded-xl" />
          <Skeleton className="h-11 w-11 rounded-xl" />
        </div>
      </div>

      {/* Table Skeleton */}
      <div className="rounded-2xl border border-border/50 bg-card/30 overflow-hidden">
        <div className="border-b border-border/50 px-6 py-4 flex gap-6 bg-muted/30">
           {Array.from({ length: 6 }).map((_, i) => (
             <Skeleton key={i} className="h-5 flex-1 rounded-full opacity-70" />
           ))}
        </div>
        <div className="p-6 space-y-5">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="flex gap-6 items-center">
              {Array.from({ length: 6 }).map((_, j) => (
                <Skeleton key={j} className="h-9 flex-1 rounded-lg" />
              ))}
            </div>
          ))}
        </div>
      </div>
    </PageMotion>
  );
}
