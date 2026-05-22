"use client";
'use no memo';

import { X } from "lucide-react";
import { Button } from "../button";
import { cn } from "@/lib/utils";
import { useIsMobile } from "@/hooks/use-mobile";
import { useRef } from "react";
import { useVirtualizer } from "@tanstack/react-virtual";

interface MapSidebarProps<T extends { id: number | string }> {
  readonly items: T[];
  readonly selectedItemId?: number | string | null;
  readonly onItemClick?: (item: T) => void;
  readonly onViewDetail?: (item: T) => void;
  readonly renderItem: (item: T) => React.ReactNode;
  readonly emptyMessage?: string;
  readonly title?: string;
  readonly className?: string;
  readonly isOpen?: boolean;
  readonly onClose?: () => void;
  readonly estimateItemHeight?: number; // Height per item in pixels
}

export function MapSidebar<T extends { id: number | string }>({
  items,
  selectedItemId,
  onItemClick,
  onViewDetail,
  renderItem,
  emptyMessage = "No items found",
  title,
  className,
  isOpen = true,
  onClose,
  estimateItemHeight = 140, // Default height including padding and button
}: MapSidebarProps<T>) {
  const isMobile = useIsMobile();
  const parentRef = useRef<HTMLDivElement>(null);

  // Virtual scrolling for large lists
  const virtualizer = useVirtualizer({
    count: items.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => estimateItemHeight,
    overscan: 5, // Render 5 extra items outside viewport
    measureElement:
      typeof window !== "undefined" && navigator.userAgent.indexOf("Firefox") === -1
        ? (element) => element?.getBoundingClientRect().height
        : undefined, // Firefox has issues with measureElement
  });

  // Mobile: Use Sheet/Drawer pattern
  if (isMobile) {
    return (
      <>
        {/* Backdrop */}
        {isOpen && (
          <div
            className="fixed inset-0 bg-black/50 z-[999] md:hidden"
            onClick={onClose}
            aria-hidden="true"
          />
        )}
        {/* Sidebar */}
        <div
          className={cn(
            "fixed left-0 top-0 h-full w-80 bg-background border-r z-[1000] transition-transform duration-300 ease-in-out md:relative md:z-auto",
            isOpen ? "translate-x-0" : "-translate-x-full md:translate-x-0",
            className
          )}
        >
          <div className="h-full flex flex-col">
            {/* Mobile Header */}
            {title && (
              <div className="flex items-center justify-between p-4 border-b md:hidden shrink-0">
                <h2 className="font-semibold text-sm">{title}</h2>
                {onClose && (
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={onClose}
                    className="cursor-pointer"
                    aria-label="Close sidebar"
                  >
                    <X className="h-4 w-4" />
                  </Button>
                )}
              </div>
            )}
            <div className="flex-1 min-h-0 overflow-hidden">
              {items.length === 0 ? (
                <div className="p-4 text-center text-muted-foreground text-sm">
                  {emptyMessage}
                </div>
              ) : (
                <div ref={parentRef} className="h-full overflow-auto">
                  <div
                    style={{
                      height: `${virtualizer.getTotalSize()}px`,
                      width: "100%",
                      position: "relative",
                    }}
                  >
                    {virtualizer.getVirtualItems().map((virtualItem) => {
                      const item = items[virtualItem.index];
                      return (
                        <div
                          key={virtualItem.key}
                          data-index={virtualItem.index}
                          ref={virtualizer.measureElement}
                          style={{
                            position: "absolute",
                            top: 0,
                            left: 0,
                            width: "100%",
                            transform: `translateY(${virtualItem.start}px)`,
                          }}
                        >
                          <div
                            className={cn(
                              "w-full border-b last:border-0",
                              selectedItemId === item.id &&
                                "bg-accent border-l-4 border-l-primary"
                            )}
                          >
                            <div
                              onClick={() => {
                                onItemClick?.(item);
                                // Close sidebar on mobile after selection
                                if (isMobile && onClose) {
                                  onClose();
                                }
                              }}
                              className={cn(
                                "w-full text-left p-4 hover:bg-accent/50 transition-colors cursor-pointer",
                                selectedItemId === item.id && "bg-transparent"
                              )}
                            >
                              {renderItem(item)}
                            </div>
                            {onViewDetail && (
                              <div className="px-4 pb-3">
                                <Button
                                  size="sm"
                                  variant="outline"
                                  className="w-full cursor-pointer"
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    onViewDetail(item);
                                    // Close sidebar on mobile after viewing detail
                                    if (isMobile && onClose) {
                                      onClose();
                                    }
                                  }}
                                >
                                  View Details
                                </Button>
                              </div>
                            )}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </>
    );
  }

  // Desktop: Normal sidebar with virtual scrolling
  return (
    <div className={cn("h-full border-r bg-background flex flex-col", className)}>
      {title && (
        <div className="flex items-center justify-between p-4 border-b shrink-0">
          <h2 className="font-semibold text-sm">{title}</h2>
        </div>
      )}
      <div className="flex-1 min-h-0 overflow-hidden">
        {items.length === 0 ? (
          <div className="p-4 text-center text-muted-foreground text-sm">
            {emptyMessage}
          </div>
        ) : (
          <div ref={parentRef} className="h-full overflow-auto">
            <div
              style={{
                height: `${virtualizer.getTotalSize()}px`,
                width: "100%",
                position: "relative",
              }}
            >
              {virtualizer.getVirtualItems().map((virtualItem) => {
                const item = items[virtualItem.index];
                return (
                  <div
                    key={virtualItem.key}
                    data-index={virtualItem.index}
                    ref={virtualizer.measureElement}
                    style={{
                      position: "absolute",
                      top: 0,
                      left: 0,
                      width: "100%",
                      transform: `translateY(${virtualItem.start}px)`,
                    }}
                  >
                    <div
                      className={cn(
                        "w-full border-b last:border-0",
                        selectedItemId === item.id &&
                          "bg-accent border-l-4 border-l-primary"
                      )}
                    >
                      <div
                        onClick={() => onItemClick?.(item)}
                        className={cn(
                          "w-full text-left p-4 hover:bg-accent/50 transition-colors cursor-pointer",
                          selectedItemId === item.id && "bg-transparent"
                        )}
                      >
                        {renderItem(item)}
                      </div>
                      {onViewDetail && (
                        <div className="px-4 pb-3">
                          <Button
                            size="sm"
                            variant="outline"
                            className="w-full cursor-pointer"
                            onClick={(e) => {
                              e.stopPropagation();
                              onViewDetail(item);
                            }}
                          >
                            View Details
                          </Button>
                        </div>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

