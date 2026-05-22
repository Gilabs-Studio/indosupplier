"use client";

import { useEffect, useRef } from "react";
import { Loader2 } from "lucide-react";

interface StageScrollLoaderProps {
  onLoadMore: () => void;
  hasMore: boolean;
  isLoading: boolean;
}

/**
 * IntersectionObserver-based infinite scroll trigger.
 * Place at the bottom of a scrollable column to progressively load more items.
 */
export function StageScrollLoader({
  onLoadMore,
  hasMore,
  isLoading,
}: StageScrollLoaderProps) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!hasMore || isLoading) return;

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting) {
          onLoadMore();
        }
      },
      { threshold: 0.1, rootMargin: "100px" }
    );

    const el = ref.current;
    if (el) observer.observe(el);

    return () => {
      if (el) observer.unobserve(el);
    };
  }, [hasMore, isLoading, onLoadMore]);

  if (!hasMore) return null;

  return (
    <div ref={ref} className="flex items-center justify-center py-3">
      {isLoading && (
        <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
      )}
    </div>
  );
}
