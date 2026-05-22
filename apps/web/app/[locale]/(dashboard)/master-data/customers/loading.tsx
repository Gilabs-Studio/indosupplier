"use client";

import { Skeleton } from "@/components/ui/skeleton";

export default function CustomersLoading() {
  return (
    <div className="h-full w-full flex overflow-hidden">
      {/* Sidebar Skeleton */}
      <div className="w-80 border-r shrink-0 hidden lg:flex flex-col">
        <div className="border-b p-4 space-y-4">
          <div className="flex items-center gap-2">
            <Skeleton className="h-5 w-5" />
            <Skeleton className="h-5 w-32" />
          </div>
          <Skeleton className="h-9 w-full" />
          <div className="flex items-center gap-2">
            <Skeleton className="h-9 flex-1" />
            <Skeleton className="h-9 w-24" />
          </div>
        </div>
        <div className="flex-1 p-4 space-y-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="p-4 border-b space-y-2">
              <Skeleton className="h-4 w-3/4" />
              <Skeleton className="h-3 w-1/2" />
              <Skeleton className="h-5 w-20" />
            </div>
          ))}
        </div>
      </div>

      {/* Map Skeleton */}
      <div className="flex-1 bg-muted flex items-center justify-center">
        <div className="text-center">
          <Skeleton className="h-8 w-8 mx-auto mb-2 rounded-full" />
          <Skeleton className="h-4 w-24 mx-auto" />
        </div>
      </div>
    </div>
  );
}
