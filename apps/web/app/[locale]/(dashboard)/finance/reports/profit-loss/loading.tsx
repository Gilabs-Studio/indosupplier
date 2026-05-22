"use client";

import { Skeleton } from "@/components/ui/skeleton";
import { PageMotion } from "@/components/motion";

export default function ProfitLossLoading() {
  return (
    <PageMotion className="p-6 space-y-6">
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div className="space-y-2">
          <Skeleton className="h-9 w-64" />
          <Skeleton className="h-5 w-96" />
        </div>
        <Skeleton className="h-9 w-32" />
      </div>
      <div className="flex gap-4 items-end">
        <Skeleton className="h-9 w-48" />
        <Skeleton className="h-9 w-48" />
      </div>
      <div className="space-y-6">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="rounded-md border">
            <div className="p-3 bg-muted/50">
              <Skeleton className="h-5 w-32" />
            </div>
            <div className="p-4 space-y-3">
              {Array.from({ length: 3 }).map((_, j) => (
                <Skeleton key={j} className="h-10 w-full" />
              ))}
            </div>
          </div>
        ))}
      </div>
    </PageMotion>
  );
}
