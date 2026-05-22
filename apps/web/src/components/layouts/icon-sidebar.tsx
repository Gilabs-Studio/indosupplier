"use client";

import React, { memo } from "react";
import Image from "next/image";
import { useLocale } from "next-intl";
import { Link, usePathname } from "@/i18n/routing";
import { Plus, Settings } from "lucide-react";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";

export interface IconSidebarItem {
  id: string;
  name: string;
  icon: React.ReactNode;
  href?: string;
  hasChildren: boolean;
}

interface IconSidebarProps {
  items: IconSidebarItem[];
  activeParentId: string | null;
  onSelectParent: (id: string) => void;
  logoSrc?: string;
  logoAlt?: string;
}

export const IconSidebar = memo(function IconSidebar({
  items,
  activeParentId,
  onSelectParent,
  logoSrc = "/logo.png",
  logoAlt = "Logo",
}: IconSidebarProps) {
  const pathname = usePathname();
  const locale = useLocale();

  const isItemActive = (item: IconSidebarItem) => {
    if (item.id === activeParentId) return true;
    if (item.href && (pathname === item.href || pathname.startsWith(`${item.href}/`))) {
      return true;
    }
    return false;
  };

  return (
    <aside
      className="fixed left-0 top-0 z-40 flex h-screen w-16 flex-col bg-sidebar-dark text-sidebar-dark-foreground"
      data-slot="icon-sidebar"
    >
      {/* Logo */}
      <div className="flex h-16 items-center justify-center">
        <Link href="/dashboard" className="flex items-center justify-center">
          <Image
            src={logoSrc}
            alt={logoAlt}
            width={36}
            height={36}
            className="object-contain rounded-lg"
          />
        </Link>
      </div>

      {/* Menu Items */}
      <nav className="flex flex-1 flex-col items-center gap-1 overflow-y-auto py-3 px-2">
        {items.map((item) => {
          const isActive = isItemActive(item);

          if (item.hasChildren) {
            return (
              <Tooltip key={item.id} delayDuration={0}>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className={cn(
                      "h-10 w-10 rounded-xl transition-all duration-200 text-sidebar-dark-foreground",
                      "hover:bg-white/10",
                      isActive && "bg-primary text-primary-foreground shadow-lg hover:bg-primary/90 hover:text-primary-foreground"
                    )}
                    onClick={() => onSelectParent(item.id)}
                    aria-label={item.name}
                    aria-pressed={isActive}
                  >
                    <span className="[&>svg]:h-5 [&>svg]:w-5">{item.icon}</span>
                  </Button>
                </TooltipTrigger>
                <TooltipContent side="right" sideOffset={8}>
                  {item.name}
                </TooltipContent>
              </Tooltip>
            );
          }

          return (
            <Tooltip key={item.id} delayDuration={0}>
              <TooltipTrigger asChild>
                <Link href={item.href || "/dashboard"} onClick={() => onSelectParent(item.id)}>
                  <Button
                    variant="ghost"
                    size="icon"
                    className={cn(
                      "h-10 w-10 rounded-xl transition-all duration-200 text-sidebar-dark-foreground",
                      "hover:bg-white/10",
                      isActive && "bg-primary text-primary-foreground shadow-lg hover:bg-primary/90 hover:text-primary-foreground"
                    )}
                    aria-label={item.name}
                  >
                    <span className="[&>svg]:h-5 [&>svg]:w-5">{item.icon}</span>
                  </Button>
                </Link>
              </TooltipTrigger>
              <TooltipContent side="right" sideOffset={8}>
                {item.name}
              </TooltipContent>
            </Tooltip>
          );
        })}
      </nav>

      {/* Bottom Actions */}
      <div className="flex flex-col items-center gap-2 py-3 px-2">
        <Tooltip delayDuration={0}>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className="h-10 w-10 rounded-xl hover:bg-white/10 text-sidebar-dark-foreground"
              aria-label="Add new"
            >
              <Plus className="h-5 w-5" />
            </Button>
          </TooltipTrigger>
          <TooltipContent side="right" sideOffset={8}>
            New Folder
          </TooltipContent>
        </Tooltip>

        <Tooltip delayDuration={0}>
          <TooltipTrigger asChild>
            <Link href="/profile" locale={locale}>
              <Button
                variant="ghost"
                size="icon"
                className="h-10 w-10 rounded-xl hover:bg-white/10 text-sidebar-dark-foreground"
                aria-label="Settings"
              >
                <Settings className="h-5 w-5" />
              </Button>
            </Link>
          </TooltipTrigger>
          <TooltipContent side="right" sideOffset={8}>
            Settings
          </TooltipContent>
        </Tooltip>
      </div>
    </aside>
  );
});
