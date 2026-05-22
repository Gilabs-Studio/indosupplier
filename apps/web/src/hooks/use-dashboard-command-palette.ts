import { useCallback, useEffect, useMemo, useState } from "react";
import { useRouter } from "@/i18n/routing";
import type { MenuWithActions } from "@/features/master-data/user-management/types";
import { isValidRoute } from "@/lib/route-validator";

export interface CommandMenuItem {
  readonly id: string;
  readonly name: string;
  readonly href: string;
  readonly icon: string;
  readonly group: string;
}

interface UseDashboardCommandPaletteOptions {
  readonly menus: MenuWithActions[] | undefined;
}

interface UseDashboardCommandPaletteResult {
  readonly isOpen: boolean;
  readonly open: () => void;
  readonly close: () => void;
  readonly toggle: () => void;
  readonly items: CommandMenuItem[];
  readonly onSelectItem: (href: string) => void;
}

export function useDashboardCommandPalette(
  options: UseDashboardCommandPaletteOptions
): UseDashboardCommandPaletteResult {
  const { menus } = options;
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);

  const items: CommandMenuItem[] = useMemo(() => {
    if (!menus || menus.length === 0) {
      return [];
    }

    const allItems: CommandMenuItem[] = [];

    const walkChildren = (children: MenuWithActions[], group: string) => {
      children.forEach((child) => {
        // Only add menu item if URL exists and is a valid route
        if (child.url && isValidRoute(child.url)) {
          allItems.push({
            id: String(child.id),
            name: child.name,
            href: child.url,
            icon: child.icon,
            group,
          });
        }

        if (child.children && child.children.length > 0) {
          walkChildren(child.children, group);
        }
      });
    };

    menus.forEach((menu) => {
      const group = menu.url === "/dashboard" ? "Dashboards" : menu.name;

      // Only add menu item if URL exists and is a valid route
      if (menu.url && isValidRoute(menu.url)) {
        allItems.push({
          id: String(menu.id),
          name: menu.name,
          href: menu.url,
          icon: menu.icon,
          group,
        });
      }

      if (menu.children && menu.children.length > 0) {
        walkChildren(menu.children, group);
      }
    });

    return allItems;
  }, [menus]);

  const open = useCallback(() => setIsOpen(true), []);
  const close = useCallback(() => setIsOpen(false), []);
  const toggle = useCallback(() => {
    setIsOpen((prev) => !prev);
  }, []);

  const onSelectItem = useCallback(
    (href: string) => {
      // Validate route before navigation
      if (!href || !isValidRoute(href)) {
        console.warn(`Invalid route: ${href}`);
        return;
      }
      router.push(href);
      setIsOpen(false);
    },
    [router]
  );

  // Global keyboard shortcuts: "/" and "Ctrl+K" / "Meta+K"
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      const target = event.target as HTMLElement | null;
      const isInputLike =
        target?.tagName === "INPUT" ||
        target?.tagName === "TEXTAREA" ||
        target?.isContentEditable;

      // Ignore when typing in inputs
      if (isInputLike) {
        return;
      }

      // "/" opens palette
      if (event.key === "/" && !event.ctrlKey && !event.metaKey && !event.altKey) {
        event.preventDefault();
        setIsOpen(true);
        return;
      }

      // "Ctrl+K" or "Meta+K"
      if (
        (event.key === "k" || event.key === "K") &&
        (event.ctrlKey || event.metaKey)
      ) {
        event.preventDefault();
        setIsOpen((prev) => !prev);
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []);

  return {
    isOpen,
    open,
    close,
    toggle,
    items,
    onSelectItem,
  };
}



