"use client";

import { cn } from "@/lib/utils";
import { useIsMobile } from "@/hooks/use-mobile";

interface SplitViewContainerProps {
  readonly children: React.ReactNode;
  readonly className?: string;
}

export function SplitViewContainer({
  children,
  className,
}: SplitViewContainerProps) {
  const isMobile = useIsMobile();

  return (
    <div
      className={cn(
        "relative h-[calc(100vh-12rem)] min-h-[400px] sm:min-h-[600px] border rounded-lg overflow-hidden bg-background",
        !isMobile && "flex",
        className
      )}
    >
      {children}
    </div>
  );
}

