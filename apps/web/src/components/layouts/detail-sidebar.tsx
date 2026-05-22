"use client";

import React, { memo, useState, useEffect } from "react";
import { Link, usePathname } from "@/i18n/routing";
import { ChevronLeft, Folder, FolderOpen } from "lucide-react";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";

export interface DetailSidebarItem {
  id: string;
  name: string;
  href?: string;
  icon?: React.ReactNode;
  children?: DetailSidebarItem[];
}

interface DetailSidebarProps {
  title: string;
  items: DetailSidebarItem[];
  isOpen: boolean;
  onToggle: () => void;
}

interface TreeItemProps {
  item: DetailSidebarItem;
  level?: number;
  pathname: string;
  isLast?: boolean;
  parentPath?: boolean[];
}

/**
 * Check if the current item or any of its children contains the active path
 */
function hasActiveChild(item: DetailSidebarItem, pathname: string): boolean {
  if (item.href && (pathname === item.href || pathname.startsWith(`${item.href}/`))) {
    return true;
  }
  
  if (item.children) {
    return item.children.some(child => hasActiveChild(child, pathname));
  }
  
  return false;
}

const TreeItem = memo(function TreeItem({
  item,
  level = 0,
  pathname,
  isLast = false,
  parentPath = [],
}: TreeItemProps) {
  const hasChildren = item.children && item.children.length > 0;
  const isActive = item.href && (pathname === item.href || pathname.startsWith(`${item.href}/`));
  const shouldBeExpanded = hasChildren && hasActiveChild(item, pathname);
  
  const [isExpanded, setIsExpanded] = useState(shouldBeExpanded);
  
  // Auto-expand folders that contain the active item
  useEffect(() => {
    let timeoutId: ReturnType<typeof setTimeout> | null = null;
    if (shouldBeExpanded && !isExpanded) {
      timeoutId = setTimeout(() => {
        setIsExpanded(true);
      }, 0);
    }
    return () => {
      if (timeoutId) clearTimeout(timeoutId);
    };
  }, [shouldBeExpanded, isExpanded]);

  // Calculate indent based on level
  const indentBase = 20; // Base indent in pixels
  const indent = level * indentBase;
  const lineOffset = indent - indentBase + 8; // Position for connecting lines

  if (hasChildren) {
    const children = item.children ?? [];
    const lastChildIndex = children.length - 1;

    return (
      <Collapsible open={isExpanded} onOpenChange={setIsExpanded}>
        <div className="group/line relative">
          {/* Vertical line from parent (if not root level) */}
          {level > 0 && (
            <>
              {/* Vertical line going up from this item */}
              <div
                className="sidebar-tree-line absolute w-px transition-opacity duration-200 z-20"
                style={{
                  left: `${lineOffset}px`,
                  top: 0,
                  height: "50%",
                  backgroundColor: "hsl(var(--primary) / 0.5)",
                }}
              />
              {/* Horizontal connector line */}
              <div
                className="sidebar-tree-line absolute h-px transition-opacity duration-200 z-20"
                style={{
                  left: `${lineOffset}px`,
                  top: "50%",
                  width: `${indentBase - 8}px`,
                  transform: "translateY(-50%)",
                  backgroundColor: "hsl(var(--primary) / 0.5)",
                }}
              />
              {/* Vertical line continuation below (if not last) */}
              {!isLast && (
                <div
                  className="sidebar-tree-line absolute w-px transition-opacity duration-200 z-20"
                  style={{
                    left: `${lineOffset}px`,
                    top: "50%",
                    height: "100%",
                    backgroundColor: "hsl(var(--primary) / 0.5)",
                  }}
                />
              )}
            </>
          )}

          <CollapsibleTrigger asChild>
            <button
              type="button"
              className={cn(
                "group relative z-10 flex w-full min-w-0 items-center justify-start gap-2.5 rounded-md px-3 py-2 text-sm transition-all duration-200 text-left sidebar-item-hover-gradient",
                "hover:bg-accent hover:text-accent-foreground",
                "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
                isActive && "bg-primary/10 text-primary font-medium"
              )}
              style={{ paddingLeft: `${indent + 12}px` }}
            >
              {/* Folder icon instead of chevron */}
              {isExpanded ? (
                <FolderOpen
                  className={cn(
                    "h-4 w-4 shrink-0 transition-colors",
                    "text-primary" // Always primary blue when expanded
                  )}
                />
              ) : (
                <Folder
                  className={cn(
                    "h-4 w-4 shrink-0 transition-colors",
                    "text-primary/50" // Primary blue with 50% opacity when collapsed
                  )}
                />
              )}
              {item.icon && (
                <span
                  className={cn(
                    "shrink-0 [&>svg]:h-4 [&>svg]:w-4 transition-colors",
                    isActive ? "text-primary" : "text-muted-foreground group-hover:text-foreground"
                  )}
                >
                  {item.icon}
                </span>
              )}
              <span
                className={cn(
                  "flex-1 min-w-0 max-w-full wrap-break-word transition-colors",
                  isActive ? "text-primary font-medium" : "text-foreground/80 group-hover:text-foreground"
                )}
              >
                {item.name}
              </span>
            </button>
          </CollapsibleTrigger>

          <CollapsibleContent className="data-[state=open]:animate-collapsible-down data-[state=closed]:animate-collapsible-up overflow-hidden">
            <div className="relative">
              {/* Vertical line for children group - stops at last child's center */}
              {children.length > 0 && isExpanded && (
                <div
                  className="sidebar-tree-line absolute w-px transition-opacity duration-200 z-20"
                  style={{
                    left: `${indent + 8}px`,
                    top: 0,
                    backgroundColor: "hsl(var(--primary) / 0.5)",
                    // Stop at the center of the last child
                    // Each tree item has approximately 40px height (py-2 = 16px + content)
                    // We want the line to stop at 50% of the last item's height
                    height: children.length > 1 
                      ? `calc(100% - 20px)` // Approximate: stop before last item's center
                      : "20px", // If only one child, stop at its center (half of ~40px)
                  }}
                />
              )}

              {children.map((child, index) => {
                const isChildLast = index === lastChildIndex;
                return (
                  <TreeItem
                    key={child.id}
                    item={child}
                    level={level + 1}
                    pathname={pathname}
                    isLast={isChildLast}
                    parentPath={[...parentPath, !isLast]}
                  />
                );
              })}
            </div>
          </CollapsibleContent>
        </div>
      </Collapsible>
    );
  }

  return (
    <div className="group/line relative">
      {/* Connecting lines for leaf nodes */}
      {level > 0 && (
        <>
          {/* Vertical line from parent (going up) */}
          <div
            className="sidebar-tree-line absolute w-px transition-opacity duration-200 z-20"
            style={{
              left: `${lineOffset}px`,
              top: 0,
              height: "50%",
              backgroundColor: "hsl(var(--primary) / 0.5)",
            }}
          />
          {/* Horizontal connector line (L-shape) */}
          <div
            className="sidebar-tree-line absolute h-px transition-opacity duration-200 z-20"
            style={{
              left: `${lineOffset}px`,
              top: "50%",
              width: `${indentBase - 8}px`,
              transform: "translateY(-50%)",
              backgroundColor: "hsl(var(--primary) / 0.5)",
            }}
          />
          {/* Vertical line continuation below (only if NOT last - continues to next sibling) */}
          {!isLast && (
            <div
              className="sidebar-tree-line absolute w-px transition-opacity duration-200 z-20"
              style={{
                left: `${lineOffset}px`,
                top: "50%",
                height: "100%",
                backgroundColor: "hsl(var(--primary) / 0.5)",
              }}
            />
          )}
        </>
      )}

      <Link
        href={item.href || "#"}
        className={cn(
          "group relative z-10 flex w-full min-w-0 items-center justify-start gap-2.5 rounded-md px-3 py-2 text-sm transition-all duration-200 text-left sidebar-item-hover-gradient",
          "hover:bg-accent hover:text-accent-foreground",
          "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
          isActive
            ? "bg-primary/10 text-primary font-medium"
            : "text-foreground/70"
        )}
        style={{ paddingLeft: `${indent + 12}px` }}
      >
        {item.icon && (
          <span
            className={cn(
              "shrink-0 [&>svg]:h-4 [&>svg]:w-4 transition-colors",
              isActive
                ? "text-primary"
                : "text-muted-foreground group-hover:text-foreground"
            )}
          >
            {item.icon}
          </span>
        )}
        <span
          className={cn(
            "flex-1 min-w-0 max-w-full wrap-break-word transition-colors",
            isActive
              ? "text-primary font-medium"
              : "text-foreground/70 group-hover:text-foreground"
          )}
        >
          {item.name}
        </span>
      </Link>
    </div>
  );
});

export const DetailSidebar = memo(function DetailSidebar({
  title,
  items,
  isOpen,
  onToggle,
}: DetailSidebarProps) {
  const pathname = usePathname();

  return (
    <aside
      className={cn(
        "fixed left-16 top-0 z-30 h-screen w-56 bg-sidebar transition-transform duration-300 ease-out",
        isOpen ? "translate-x-0" : "-translate-x-full"
      )}
      data-slot="detail-sidebar"
    >
      <div className="flex h-full flex-col">
        {/* Header with collapse button */}
        <div className="flex h-16 shrink-0 items-center justify-between border-b border-sidebar-border px-4 bg-sidebar/95 backdrop-blur supports-backdrop-filter:bg-sidebar/80">
          <h2 className="text-sm font-semibold text-foreground/90 truncate">{title}</h2>
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7 shrink-0 rounded-md hover:bg-accent transition-colors"
            onClick={onToggle}
            aria-label="Collapse sidebar"
          >
            <ChevronLeft className="h-4 w-4" />
          </Button>
        </div>

        {/* Menu Items with Professional Scroll Area */}
        <div className="flex-1 min-h-0 overflow-hidden">
          <ScrollArea className="h-full">
            <div className="pl-1 pr-3 py-4 space-y-0.5">
              {items.length === 0 ? (
                <div className="flex items-center justify-center py-8 text-center">
                  <p className="text-sm text-muted-foreground">No items available</p>
                </div>
              ) : (
                items.map((item, index) => {
                  const isLast = index === items.length - 1;
                  return (
                    <TreeItem
                      key={item.id}
                      item={item}
                      pathname={pathname}
                      isLast={isLast}
                    />
                  );
                })
              )}
            </div>
          </ScrollArea>
        </div>
      </div>
    </aside>
  );
});
