"use client";

import { Skeleton } from "@/components/ui/skeleton";

export default function AreaMappingLoading() {
  return (
    <div className="relative w-full h-[calc(100vh-64px)]">
      <Skeleton className="absolute inset-0 rounded-none" />
      <div className="absolute top-6 left-6 z-10">
        <Skeleton className="h-16 w-64 rounded-2xl" />
      </div>
      <div className="absolute top-6 right-6 z-10 flex gap-3">
        <Skeleton className="h-20 w-[100px] rounded-xl" />
        <Skeleton className="h-20 w-[100px] rounded-xl" />
        <Skeleton className="h-20 w-[100px] rounded-xl" />
      </div>
    </div>
  );
}
