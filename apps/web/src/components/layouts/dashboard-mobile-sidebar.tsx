"use client";

import React, { memo, useState } from "react";
import Image from "next/image";
import { Link, usePathname } from "@/i18n/routing";

import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import type { IconSidebarItem } from "./icon-sidebar";
import type { DetailSidebarItem } from "./detail-sidebar";

interface DashboardMobileSidebarProps {
  isOpen: boolean;
  onClose: () => void;
  parentItems: IconSidebarItem[];
  activeParentId: string | null;
  onSelectParent: (id: string) => void;
  detailItems: DetailSidebarItem[];
  detailTitle: string;
}

interface MobileMenuItemProps {
  item: DetailSidebarItem;
  pathname: string;
  onClose: () => void;
  level?: number;
}

const MobileMenuItem = memo(function MobileMenuItem({
  item,
  pathname,
  onClose,
  level = 0,
}: MobileMenuItemProps) {
  const [isExpanded, setIsExpanded] = useState(true);
  const hasChildren = Boolean(item.children && item.children.length > 0);
  const isActive = item.href && (pathname === item.href || pathname.startsWith(`${item.href}/`));

  if (hasChildren) {
    return (
      <div>
        <button
          type="button"
          className="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-sm text-left transition-colors hover:bg-sidebar-accent"
          style={{ paddingLeft: `${level * 12 + 12}px` }}
          onClick={() => setIsExpanded((prev) => !prev)}
        >
          {item.icon && <span className="[&>svg]:h-4 [&>svg]:w-4">{item.icon}</span>}
          <span className="flex-1 truncate font-medium">{item.name}</span>
        </button>
        {isExpanded &&
          item.children?.map((child) => (
            <MobileMenuItem
              key={child.id}
              item={child}
              pathname={pathname}
              onClose={onClose}
              level={level + 1}
            />
          ))}
      </div>
    );
  }

  return (
    <Link
      href={item.href || "#"}
      className={cn(
        "flex w-full items-center gap-2 rounded-lg px-3 py-2 text-sm text-left transition-colors hover:bg-sidebar-accent",
        isActive && "bg-primary/10 font-medium text-primary"
      )}
      style={{ paddingLeft: `${level * 12 + 12}px` }}
      onClick={onClose}
    >
      {item.icon && <span className="[&>svg]:h-4 [&>svg]:w-4">{item.icon}</span>}
      <span className="flex-1 truncate">{item.name}</span>
    </Link>
  );
});

export const DashboardMobileSidebar = memo(function DashboardMobileSidebar({
  isOpen,
  onClose,
  parentItems,
  activeParentId,
  onSelectParent,
  detailItems,
  detailTitle,
}: DashboardMobileSidebarProps) {
  const pathname = usePathname();

  const activeParent = parentItems.find((parent) => parent.id === activeParentId);
  const showDetailColumn = Boolean(activeParent?.hasChildren && detailItems.length > 0);

  return (
    <Sheet
      open={isOpen}
      onOpenChange={(open) => {
        if (!open) {
          onClose();
        }
      }}
    >
      <SheetContent side="left" className="w-80 p-0">
        <SheetHeader className="sr-only">
          <SheetTitle>Navigation Menu</SheetTitle>
          <SheetDescription>Main navigation menu</SheetDescription>
        </SheetHeader>
        <div className="flex h-full">
          <div
            className={cn(
              "flex flex-col bg-sidebar-dark text-sidebar-dark-foreground transition-all duration-300",
              showDetailColumn ? "w-16" : "w-full"
            )}
          >
            <div className="flex h-16 items-center justify-center">
              <Image src="/logo.png" alt="Logo" width={36} height={36} className="rounded-lg object-contain" />
            </div>
            <nav
              className={cn(
                "flex flex-1 flex-col items-center gap-1 overflow-y-auto px-2 py-3",
                !showDetailColumn && "items-stretch px-4"
              )}
            >
              {parentItems.map((item) => {
                const isActive = item.id === activeParentId;
                const isCurrentPath = item.href && (pathname === item.href || pathname.startsWith(`${item.href}/`));

                if (!item.hasChildren && item.href) {
                  return (
                    <Link
                      key={item.id}
                      href={item.href}
                      onClick={() => {
                        onSelectParent(item.id);
                        onClose();
                      }}
                      className={cn(
                        "flex items-center gap-3 rounded-xl transition-all duration-200",
                        showDetailColumn ? "h-10 w-10 justify-center" : "h-11 px-4",
                        isActive || isCurrentPath
                          ? "bg-primary text-primary-foreground shadow-lg"
                          : "text-sidebar-dark-foreground hover:bg-white/10"
                      )}
                    >
                      <span className="[&>svg]:h-5 [&>svg]:w-5">{item.icon}</span>
                      {!showDetailColumn && <span className="text-sm font-medium">{item.name}</span>}
                    </Link>
                  );
                }

                return (
                  <Button
                    key={item.id}
                    variant="ghost"
                    size={showDetailColumn ? "icon" : "default"}
                    className={cn(
                      "rounded-xl text-sidebar-dark-foreground transition-all duration-200",
                      showDetailColumn ? "h-10 w-10" : "h-11 w-full justify-start gap-3 px-4",
                      isActive
                        ? "bg-primary text-primary-foreground shadow-lg hover:bg-primary/90 hover:text-primary-foreground"
                        : "hover:bg-white/10"
                    )}
                    onClick={() => onSelectParent(item.id)}
                  >
                    <span className="[&>svg]:h-5 [&>svg]:w-5">{item.icon}</span>
                    {!showDetailColumn && <span className="text-sm font-medium">{item.name}</span>}
                  </Button>
                );
              })}
            </nav>
          </div>

          {showDetailColumn && (
            <div className="flex flex-1 flex-col bg-sidebar">
              <div className="flex h-16 items-center border-b border-sidebar-border px-4">
                <h2 className="text-sm font-semibold">{detailTitle}</h2>
              </div>
              <nav className="flex-1 overflow-y-auto p-2">
                {detailItems.map((item) => (
                  <MobileMenuItem key={item.id} item={item} pathname={pathname} onClose={onClose} />
                ))}
              </nav>
            </div>
          )}
        </div>
      </SheetContent>
    </Sheet>
  );
});