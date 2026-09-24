import React from "react";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

export function ReviewsSkeleton() {
  return (
    <div className="space-y-4">
      {[1, 2, 3].map((i) => (
        <Card key={i} className="border border-border bg-card rounded-xl p-5 space-y-4">
          <div className="flex items-start justify-between">
            <div className="flex items-center gap-3">
              <Skeleton className="h-10 w-10 rounded-full" />
              <div className="space-y-1.5">
                <Skeleton className="h-4 w-40" />
                <Skeleton className="h-3 w-24" />
              </div>
            </div>
            <Skeleton className="h-5 w-20 rounded-md" />
          </div>
          <div className="flex gap-2">
            <Skeleton className="h-6 w-36 rounded-md" />
            <Skeleton className="h-6 w-28 rounded-md" />
          </div>
          <Skeleton className="h-14 w-full rounded-lg" />
        </Card>
      ))}
    </div>
  );
}
