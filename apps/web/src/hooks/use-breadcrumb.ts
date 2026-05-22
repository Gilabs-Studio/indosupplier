"use client";

import { useMemo } from "react";
import { usePathname } from "@/i18n/routing";
import type { MenuWithActions } from "@/features/master-data/user-management/types";

export interface BreadcrumbItem {
  readonly label: string;
  readonly href: string;
  readonly icon?: string;
  readonly isActive?: boolean;
}

/**
 * Check if a path matches a menu item's URL
 */
function isPathMatch(pathname: string, url: string): boolean {
  if (!url) return false;
  return pathname === url || pathname.startsWith(`${url}/`);
}

/**
 * Recursively find menu item and its complete breadcrumb path (including parent menus)
 * This function builds the breadcrumb path by including parent menus even if they don't match the pathname
 */
function findMenuBreadcrumb(
  menus: MenuWithActions[],
  pathname: string,
  breadcrumbPath: BreadcrumbItem[] = []
): BreadcrumbItem[] | null {
  for (const menu of menus) {
    // Always add current menu to breadcrumb path (parent menu)
    const currentBreadcrumb: BreadcrumbItem = {
      label: menu.name,
      href: menu.url || "#",
      icon: menu.icon,
      isActive: menu.url ? isPathMatch(pathname, menu.url) : false,
    };

    const newBreadcrumbPath = [...breadcrumbPath, currentBreadcrumb];

    // Check if current menu matches exactly (exact match, not just prefix)
    const isExactMatch = menu.url && pathname === menu.url;
    
    // If exact match and no children, return immediately
    if (isExactMatch && (!menu.children || menu.children.length === 0)) {
      return newBreadcrumbPath;
    }

    // Check children recursively - parent is already in breadcrumbPath
    // Even if parent matches, we still need to check children for more specific matches
    if (menu.children && menu.children.length > 0) {
      const found = findMenuBreadcrumb(
        menu.children,
        pathname,
        newBreadcrumbPath
      );
      if (found) return found;
    }

    // If no children found and this is an exact match, return this menu
    if (isExactMatch) {
      return newBreadcrumbPath;
    }
  }
  return null;
}

/**
 * Normalize pathname by removing locale prefix
 */
function normalizePathname(pathname: string): string {
  // Remove locale prefix if present (e.g., /en/purchase/order -> /purchase/order)
  const segments = pathname.split("/").filter(Boolean);
  if (segments[0] === "id" || segments[0] === "en") {
    return "/" + segments.slice(1).join("/");
  }
  return pathname;
}

/**
 * Generate breadcrumb items from pathname and menu structure
 */
export function useBreadcrumb(menus?: MenuWithActions[]): BreadcrumbItem[] {
  const pathname = usePathname();
  const normalizedPathname = normalizePathname(pathname);

  return useMemo(() => {
    const items: BreadcrumbItem[] = [];

    // Always start with Dashboard
    const isDashboard = normalizedPathname === "/dashboard";
    items.push({
      label: "Dashboard",
      href: "/dashboard",
      isActive: isDashboard,
    });

    // If on dashboard, return early
    if (isDashboard) {
      return items;
    }

    // If no menus, try to parse pathname segments
    if (!menus || menus.length === 0) {
      const segments = normalizedPathname.split("/").filter(Boolean);

      // Skip dashboard segment (already added)
      if (segments[0] === "dashboard") {
        segments.shift();
      }

      // Build breadcrumb from path segments
      let currentPath = "";
      segments.forEach((segment, index) => {
        currentPath += `/${segment}`;
        const isLast = index === segments.length - 1;

        // Format segment name (capitalize, replace hyphens)
        const label = segment
          .split("-")
          .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
          .join(" ");

        items.push({
          label,
          href: currentPath,
          isActive: isLast,
        });
      });

      return items;
    }

    // Find matching menu item
    const found = findMenuBreadcrumb(menus, normalizedPathname);

    if (found && found.length > 0) {
      // Add found breadcrumb items (skip dashboard if it's the first item)
      // This includes parent menus in the path
      found.forEach((item) => {
        if (item.href !== "/dashboard") {
          items.push(item);
        }
      });
    } else {
      // If no menu match found, try to parse pathname segments
      const segments = normalizedPathname.split("/").filter(Boolean);

      // Skip dashboard segment (already added)
      if (segments[0] === "dashboard") {
        segments.shift();
      }

      // Build breadcrumb from path segments
      let currentPath = "";
      segments.forEach((segment, index) => {
        currentPath += `/${segment}`;
        const isLast = index === segments.length - 1;

        // Format segment name (capitalize, replace hyphens)
        const label = segment
          .split("-")
          .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
          .join(" ");

        items.push({
          label,
          href: currentPath,
          isActive: isLast,
        });
      });
    }

    return items;
  }, [normalizedPathname, menus]);
}

