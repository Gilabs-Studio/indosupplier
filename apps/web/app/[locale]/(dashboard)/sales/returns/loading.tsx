"use client";

import { PageMotion } from "@/components/motion";
import { Skeleton } from "@/components/ui/skeleton";

export default function SalesReturnsLoading() {
  return (
    <PageMotion className="space-y-6 p-6">
      <div className="space-y-2">
        <Skeleton className="h-9 w-56" />
        <Skeleton className="h-5 w-80" />
      </div>
      <Skeleton className="h-9 w-80" />
      <div className="rounded-md border p-4 space-y-4">
        {Array.from({ length: 6 }).map((_, idx) => (
          <Skeleton key={idx} className="h-10 w-full" />
        ))}
      </div>
    </PageMotion>
  );
}
