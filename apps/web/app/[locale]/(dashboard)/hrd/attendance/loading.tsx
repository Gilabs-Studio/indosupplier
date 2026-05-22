import { Skeleton } from "@/components/ui/skeleton";
import { Card } from "@/components/ui/card";

export default function AttendanceLoading() {
  return (
    <div className="flex h-full flex-col gap-6">
      {/* Page Header Skeleton */}
      <div className="space-y-2">
        <Skeleton className="h-10 w-64" />
        <Skeleton className="h-4 w-96" />
      </div>

      {/* Calendar Skeleton */}
      <Card className="flex-1 overflow-hidden">
        <div className="flex h-full flex-col">
          {/* Calendar Header */}
          <div className="flex items-center justify-between border-b border-border px-6 py-4">
            <div className="flex items-center gap-4">
              <Skeleton className="h-8 w-48" />
              <Skeleton className="h-8 w-20" />
            </div>
            <div className="flex gap-2">
              <Skeleton className="h-8 w-8" />
              <Skeleton className="h-8 w-8" />
            </div>
          </div>

          {/* Calendar Grid */}
          <div className="flex-1 p-4">
            <div className="grid grid-cols-7 gap-4">
              {Array.from({ length: 35 }).map((_, i) => (
                <div key={i} className="space-y-2">
                  <Skeleton className="h-6 w-6 rounded-full" />
                  <Skeleton className="h-16 w-full" />
                </div>
              ))}
            </div>
          </div>
        </div>
      </Card>
    </div>
  );
}
