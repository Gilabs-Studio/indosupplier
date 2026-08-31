import { Skeleton } from "@/components/ui/skeleton";
import { PageMotion } from "@/components/motion";

export default function SysadminLoading() {
  return (
    <PageMotion className="space-y-8 p-6">
      {/* Welcome Banner Skeleton */}
      <div className="rounded-lg bg-muted/40 p-6 flex flex-col md:flex-row items-start md:items-center justify-between gap-4 border border-border/80">
        <div className="space-y-2 flex-1">
          <Skeleton className="h-7 w-72 rounded-md" />
          <Skeleton className="h-4 w-96 rounded-md" />
        </div>
        <Skeleton className="h-8 w-24 rounded-lg" />
      </div>

      {/* 4 Stats Cards Skeleton */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="rounded-xl border border-border/80 bg-card p-5 space-y-3 shadow-xs">
            <div className="flex items-center justify-between">
              <Skeleton className="h-4 w-28 rounded-md" />
              <Skeleton className="h-10 w-10 rounded-lg" />
            </div>
            <Skeleton className="h-8 w-16 rounded-md" />
            <Skeleton className="h-3 w-36 rounded-md" />
          </div>
        ))}
      </div>

      {/* 2 Column Main Content Grid Skeleton */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        <div className="lg:col-span-8 rounded-xl border border-border/80 bg-card overflow-hidden shadow-xs">
          <div className="p-5 border-b border-border bg-muted/10 flex items-center justify-between">
            <div className="space-y-1">
              <Skeleton className="h-5 w-44 rounded-md" />
              <Skeleton className="h-3 w-56 rounded-md" />
            </div>
            <Skeleton className="h-4 w-20 rounded-md" />
          </div>
          <div className="divide-y divide-border">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="p-4 flex items-center justify-between">
                <div className="space-y-1.5">
                  <Skeleton className="h-4 w-48 rounded-md" />
                  <Skeleton className="h-3 w-64 rounded-md" />
                </div>
                <div className="flex gap-2">
                  <Skeleton className="h-6 w-20 rounded-md" />
                  <Skeleton className="h-6 w-20 rounded-md" />
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="lg:col-span-4 rounded-xl border border-border/80 bg-card p-6 space-y-6 shadow-xs">
          <Skeleton className="h-5 w-36 rounded-md" />
          <div className="flex items-center gap-3">
            <Skeleton className="h-12 w-12 rounded-lg shrink-0" />
            <div className="space-y-1.5 flex-1">
              <Skeleton className="h-4 w-32 rounded-md" />
              <Skeleton className="h-3 w-40 rounded-md" />
            </div>
          </div>
          <div className="space-y-3 pt-2">
            <Skeleton className="h-4 w-28 rounded-md" />
            <Skeleton className="h-10 w-full rounded-lg" />
          </div>
        </div>
      </div>
    </PageMotion>
  );
}
