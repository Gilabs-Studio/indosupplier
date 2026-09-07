import React from "react";

export function QuickActionsSkeleton() {
  return (
    <section className="w-full">
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
        {Array.from({ length: 4 }).map((_, idx) => (
          <div
            key={idx}
            className="flex items-center justify-between rounded-xl border border-border bg-card p-3.5 shadow-2xs animate-pulse"
          >
            <div className="flex items-center gap-3.5 min-w-0 flex-1">
              <div className="h-11 w-11 rounded-lg bg-muted/40 shrink-0" />
              <div className="min-w-0 flex-1 space-y-2">
                <div className="h-3.5 w-28 rounded-xs bg-muted/50" />
                <div className="h-2.5 w-36 rounded-xs bg-muted/30" />
              </div>
            </div>
            <div className="h-4 w-4 rounded-xs bg-muted/30 shrink-0 ml-2" />
          </div>
        ))}
      </div>
    </section>
  );
}
