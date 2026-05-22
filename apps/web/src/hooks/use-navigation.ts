"use client";

import { useMemo } from "react";
import { useTranslations } from "next-intl";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import { navigationConfig, type NavItem } from "@/lib/navigation-config";
import { hasPermissionCode } from "@/lib/permission-utils";
import type { MenuWithActions, Action } from "@/features/master-data/user-management/types";

function normalizeSegment(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/\s+/g, "-")
    .replace(/[^a-z0-9\-_/]/g, "");
}

function buildMenuId(item: NavItem, parentId: string, index: number): string {
  if (item.id && item.id.trim() !== "") {
    return item.id;
  }

  if (item.url && item.url.trim() !== "") {
    return item.url;
  }

  const keySource = item.i18nKey && item.i18nKey.trim() !== "" ? item.i18nKey : item.name;
  const keySegment = normalizeSegment(keySource);
  return `${parentId}/${keySegment || "item"}-${index}`;
}

function isItemVisible(item: NavItem, permissionMap: Record<string, string>): boolean {
  if (item.permission) {
    return hasPermissionCode(permissionMap, item.permission);
  }
  if (item.children) {
    return item.children.some((child) => isItemVisible(child, permissionMap));
  }
  return true; 
}

function shouldHideForPlan(item: NavItem, planSlug?: string): boolean {
  const normalizedPlan = (planSlug ?? "").trim().toLowerCase();
  if (normalizedPlan !== "pos_growth") {
    return false;
  }

  return item.i18nKey === "accountsReceivable" || item.i18nKey === "accountsPayable";
}

function transformItem(
  item: NavItem,
  permissionMap: Record<string, string>,
  t: (key: string) => string,
  planSlug?: string,
  parentId = "root",
  index = 0,
): MenuWithActions | null {
  if (shouldHideForPlan(item, planSlug)) return null;
  if (!isItemVisible(item, permissionMap)) return null;

  const itemId = buildMenuId(item, parentId, index);

  const children = item.children
    ? item.children
        .map((child, childIndex) => transformItem(child, permissionMap, t, planSlug, itemId, childIndex))
        .filter((c): c is MenuWithActions => c !== null)
    : undefined;

  // If item had children definition but result is empty, hide parent
  if (item.children && (!children || children.length === 0)) {
    return null;
  }

  // Synthesize a generic VIEW action to satisfy dashboard-layout check
  const viewAction: Action = {
    id: "view",
    code: "VIEW",
    name: "View",
    action: "VIEW",
    access: true,
  };

  return {
    id: itemId,
    name: item.i18nKey ? t(item.i18nKey) || item.name : item.name,
    icon: item.icon,
    url: item.url,
    children,
    actions: [viewAction],
  };
}

export function useNavigation() {
  const { user } = useAuthStore();
  const t = useTranslations("navigation");
  
  const menus = useMemo(() => {
    const permissionMap = user?.permissions ?? {};
    const planSlug = user?.subscription_plan;
    return navigationConfig
      .map((item) => transformItem(item, permissionMap, t, planSlug))
      .filter((item): item is MenuWithActions => item !== null);
  }, [user?.permissions, user?.subscription_plan, t]);

  return { menus };
}
