"use client";

import { Skeleton } from "@/components/ui/skeleton";

export default function AIChatbotLoading() {
  return (
    <div className="flex h-full w-full overflow-hidden">
      {/* Sidebar skeleton */}
      <div className="w-[280px] shrink-0 border-r border-border p-3">
        <Skeleton className="mb-4 h-6 w-32" />
        <div className="space-y-2">
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} className="h-12 w-full rounded-md" />
          ))}
        </div>
      </div>
      {/* Chat area skeleton */}
      <div className="flex flex-1 flex-col">
        <div className="border-b border-border p-4">
          <Skeleton className="h-6 w-40" />
          <Skeleton className="mt-1 h-3 w-56" />
        </div>
        <div className="flex-1 space-y-4 p-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <div
              key={i}
              className={`flex gap-3 ${i % 2 === 0 ? "justify-start" : "justify-end"}`}
            >
              <Skeleton className="h-8 w-8 rounded-full" />
              <Skeleton
                className={`h-16 rounded-2xl ${i % 2 === 0 ? "w-3/5" : "w-2/5"}`}
              />
            </div>
          ))}
        </div>
        <div className="border-t border-border p-3">
          <Skeleton className="h-10 w-full rounded-lg" />
        </div>
      </div>
    </div>
  );
}
