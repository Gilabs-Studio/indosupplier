"use client";

import React, {
  memo,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { usePathname } from "@/i18n/routing";

import { TooltipProvider } from "@/components/ui/tooltip";
import { Breadcrumb } from "@/components/navigation/breadcrumb";
import { CommandPalette } from "@/features/command-palette";
import { AIChatWidget } from "@/features/ai-chat/components/ai-chat-widget";
import { SubscriptionLifecycleGuard } from "@/features/auth/components/subscription-lifecycle-guard";
import { NotificationDrawer } from "@/features/notifications/components/notification-drawer";
import { useNotificationStore } from "@/features/notifications/stores/use-notification-store";
import { useWebSocket as useNotificationWebSocket } from "@/features/notifications/hooks/use-web-socket";
import { useOnboardingWebSocket } from "@/features/general/dashboard/hooks/use-onboarding-ws";
import { useAttendanceTodayWebSocket } from "@/features/hrd/attendance-records/hooks/use-attendance-today-ws";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import { useNavigation } from "@/hooks/use-navigation";
import { useValidateRole } from "@/features/auth/hooks/use-validate-role";
import { useTodayAttendance } from "@/features/hrd/attendance-records/hooks/use-attendance-records";
import type { MenuWithActions } from "@/features/master-data/user-management/types";
import { useIsMobile } from "@/hooks/use-mobile";
import { getMenuIcon } from "@/lib/menu-icons";
import {
  AttendanceRightDrawer,
  type AttendanceDrawerTab,
} from "@/features/hrd/attendance-records/components/attendance-right-drawer";
import { LeaveRequestDrawer } from "@/features/hrd/leave-request/components/leave-request-drawer";
import { OvertimeDrawer } from "@/features/hrd/overtime/components/overtime-drawer";
import { useLocationPermission } from "@/features/hrd/attendance-records/hooks/use-location-permission";
import { usePOSUIStore } from "@/features/pos/stores/use-pos-ui-store";
import { useOnboardingState } from "@/features/general/dashboard/hooks/use-onboarding";
import { IconSidebar, type IconSidebarItem } from "./icon-sidebar";
import { DetailSidebar, type DetailSidebarItem } from "./detail-sidebar";
import { cn } from "@/lib/utils";
import { DashboardHeader } from "./dashboard-header";
import { DashboardMobileSidebar } from "./dashboard-mobile-sidebar";
import { useQuery } from "@tanstack/react-query";
import { billingService } from "@/features/settings/services/billing-service";

const DETAIL_SIDEBAR_STORAGE_KEY = "detail_sidebar_state";
const ACTIVE_PARENT_STORAGE_KEY = "active_parent_id";

function checkViewPermission(menu: MenuWithActions): boolean {
  if (menu.actions && menu.actions.length > 0) {
    const viewAction = menu.actions.find((action) => {
      if (action.action === "VIEW") return action.access;
      return (
        (action.code === "VIEW" || action.code.startsWith("VIEW_")) &&
        action.access
      );
    });

    if (viewAction) return true;
  }

  if (menu.children && menu.children.length > 0) {
    return menu.children.some((child) => checkViewPermission(child));
  }

  return false;
}

function mapFinanceMenuUrl(url?: string): string | undefined {
  if (!url) return undefined;

  const mappings: Record<string, string> = {
    "/finance/accounting/journal-entries": "/finance/journals",
    "/finance/accounting/journal-entries/purchase": "/finance/journals/purchase",
    "/finance/accounting/journal-entries/adjustment": "/finance/journals/adjustment",
    "/finance/accounting/closing": "/finance/closing",
    "/finance/ar/customer-invoices": "/finance/journals/sales",
    "/finance/ar/customer-payments": "/finance/ap/payments",
    "/finance/ar/credit-notes": "/finance/journals/sales",
    "/finance/ap/supplier-invoices": "/finance/journals/purchase",
    "/finance/ap/debit-notes": "/finance/journals/purchase",
    "/finance/fixed-assets/assets": "/finance/assets",
    "/finance/fixed-assets/categories": "/finance/asset-categories",
    "/finance/fixed-assets/locations": "/finance/asset-locations",
    "/finance/fixed-assets/depreciation-schedule": "/finance/asset-batch-depreciation",
    "/finance/fixed-assets/disposal": "/finance/asset-disposal",
    "/finance/fixed-assets/revaluation": "/finance/asset-revaluation",
    "/finance/reports/trial-balance": "/finance/reports/general-ledger",
  };

  return mappings[url] ?? url;
}

function isPathMatch(pathname: string, url: string): boolean {
  return pathname === url || pathname.startsWith(`${url}/`);
}

function hasMatchingChildPath(
  children: MenuWithActions[],
  pathname: string,
): boolean {
  return children.some((child) => {
    const normalizedUrl = mapFinanceMenuUrl(child.url);
    if (normalizedUrl && isPathMatch(pathname, normalizedUrl)) return true;
    if (child.children) return hasMatchingChildPath(child.children, pathname);
    return false;
  });
}

function findParentMenuByPath(
  menus: MenuWithActions[],
  pathname: string,
): string | null {
  for (const menu of menus) {
    if (!menu.children || menu.children.length === 0) continue;

    const hasMatch = menu.children.some((child) => {
      const normalizedUrl = mapFinanceMenuUrl(child.url);
      if (normalizedUrl && isPathMatch(pathname, normalizedUrl)) return true;
      if (child.children) return hasMatchingChildPath(child.children, pathname);
      return false;
    });

    if (hasMatch) return String(menu.id);
  }

  return null;
}

interface DashboardLayoutProps {
  readonly children: React.ReactNode;
}

export const DashboardLayout = memo(function DashboardLayout({
  children,
}: DashboardLayoutProps) {
  const { user } = useAuthStore();
  const { menus } = useNavigation();
  useValidateRole();

  const { isDrawerOpen, closeDrawer } = useNotificationStore();
  useNotificationWebSocket();
  useOnboardingWebSocket();
  useAttendanceTodayWebSocket();
  const pathname = usePathname();
  const isMobile = useIsMobile();
  const isGeneralDashboardPage = Boolean(
    pathname && /(^|\/)dashboard$/.test(pathname) && !pathname.includes("/system/"),
  );
  const isTenantOwner = user?.role?.is_owner === true;
  const { data: onboardingState } = useOnboardingState({
    enabled: isGeneralDashboardPage && isTenantOwner,
  });

  const userName = user?.name ?? "User";
  const tenantName = user?.tenant_name ?? "";
  // Prefer authoritative subscription from billing API so that cancelled-but-not-expired
  // subscriptions still show until `expires_at`. Fall back to `user.subscription_plan`.
  const { data: mySubscription } = useQuery({
    queryKey: ["my-subscription"],
    queryFn: billingService.getMySubscription,
  });

  let subscriptionPlan = user?.subscription_plan;
  if (mySubscription) {
    const expiresAt = mySubscription.expires_at ? new Date(mySubscription.expires_at) : null;
    const notExpired = !expiresAt || expiresAt.getTime() > Date.now();
    if (notExpired) subscriptionPlan = mySubscription.plan;
    else subscriptionPlan = undefined;
  }
  const primaryAvatarUrl =
    user?.avatar_url && user.avatar_url.trim() !== ""
      ? user.avatar_url
      : undefined;
  const fallbackAvatarUrl = "/avatar-placeholder.svg";

  const { data: todayData } = useTodayAttendance();
  const today = todayData?.data;

  const {
    isPrompt: isLocationPrompt,
    requestPermission: requestLocationPermission,
  } = useLocationPermission();
  useEffect(() => {
    if (isLocationPrompt) {
      requestLocationPermission();
    }
  }, [isLocationPrompt, requestLocationPermission]);

  const showAttendanceIndicator = useMemo(() => {
    if (!today || !today.is_working_day || today.has_checked_in) return false;

    const schedule = today.work_schedule;
    if (!schedule?.start_time) return false;

    try {
      const serverTime = new Date(today.current_server_time);
      const [startH, startM] = schedule.start_time.split(":").map(Number);

      const startTimeDate = new Date(serverTime);
      startTimeDate.setHours(startH, startM, 0, 0);

      return serverTime >= startTimeDate;
    } catch {
      return today.is_working_day && !today.has_checked_in;
    }
  }, [today]);

  const [activeParentId, setActiveParentId] = useState<string | null>(null);
  const [isDetailSidebarOpen, setIsDetailSidebarOpen] = useState(true);
  const [isMounted, setIsMounted] = useState(false);
  const [isMobileSidebarOpen, setIsMobileSidebarOpen] = useState(false);
  const [isAttendanceDrawerOpen, setIsAttendanceDrawerOpen] = useState(false);
  const [, setAttendanceDrawerTab] = useState<AttendanceDrawerTab>("calendar");

  // Separate drawer states for Leave Request and Overtime
  const [isLeaveRequestDrawerOpen, setIsLeaveRequestDrawerOpen] =
    useState(false);
  const [isOvertimeDrawerOpen, setIsOvertimeDrawerOpen] = useState(false);
  const [openCreateLeaveInDrawer, setOpenCreateLeaveInDrawer] = useState(0);

  // Command palette control (opened by header search)
  const [isCommandPaletteOpen, setIsCommandPaletteOpen] = useState(false);
  const handleSearchClick = useCallback(() => {
    setIsCommandPaletteOpen(true);
  }, []);

  useEffect(() => {
    React.startTransition(() => {
      setIsMounted(true);
    });
  }, []);

  useEffect(() => {
    if (typeof window !== "undefined") {
      localStorage.setItem(
        DETAIL_SIDEBAR_STORAGE_KEY,
        String(isDetailSidebarOpen),
      );
    }
  }, [isDetailSidebarOpen]);

  useEffect(() => {
    if (typeof window !== "undefined" && activeParentId) {
      localStorage.setItem(ACTIVE_PARENT_STORAGE_KEY, activeParentId);
    }
  }, [activeParentId]);

  const parentItems: IconSidebarItem[] = useMemo(() => {
    const items: IconSidebarItem[] = [];

    if (menus && menus.length > 0) {
      menus.forEach((menu) => {
        if (!checkViewPermission(menu)) return;

        items.push({
          id: String(menu.id),
          name: menu.name,
          icon: getMenuIcon(menu.icon),
          href: mapFinanceMenuUrl(menu.url),
          hasChildren: Boolean(menu.children && menu.children.length > 0),
        });
      });
    }

    return items;
  }, [menus]);

  const detailItems: DetailSidebarItem[] = useMemo(() => {
    if (!activeParentId || !menus) return [];

    const parentMenu = menus.find((menu) => String(menu.id) === activeParentId);
    if (!parentMenu?.children) return [];

    const buildDetailItems = (
      menuItems: MenuWithActions[],
      parentKey = "root",
    ): DetailSidebarItem[] => {
      return menuItems
        .filter((item) => checkViewPermission(item))
        .map((item, index) => {
          const href = mapFinanceMenuUrl(item.url);
          const rawId = String(item.id ?? "").trim();
          // Ensure unique, stable IDs per sibling even when backend IDs/URLs collide.
          const idBase = rawId || href || item.name || "menu-item";
          const detailId = `${parentKey}:${idBase}:${index}`;

          return {
            id: detailId,
            name: item.name,
            href,
            icon: getMenuIcon(item.icon),
            children: item.children ? buildDetailItems(item.children, detailId) : undefined,
          };
        });
    };

    return buildDetailItems(parentMenu.children);
  }, [activeParentId, menus]);

  const activeParentTitle = useMemo(() => {
    if (!activeParentId) return "Menu";
    const parent = parentItems.find((item) => item.id === activeParentId);
    return parent?.name || "Menu";
  }, [activeParentId, parentItems]);

  const manualSelectionRef = useRef(false);
  const previousPathnameRef = useRef<string | null>(null);
  const isInitialMountRef = useRef(true);

  useEffect(() => {
    if (!menus || !isMounted) return;

    const currentPathname = pathname;
    const previousPathname = previousPathnameRef.current;
    const isInitialMount = isInitialMountRef.current;

    if (isInitialMount) {
      isInitialMountRef.current = false;
      const storedSidebarState = localStorage.getItem(
        DETAIL_SIDEBAR_STORAGE_KEY,
      );
      const detectedParent = findParentMenuByPath(menus, currentPathname);

      if (detectedParent) {
        const parentMenu = menus.find(
          (menu) => String(menu.id) === detectedParent,
        );
        const hasChildren = Boolean(
          parentMenu?.children && parentMenu.children.length > 0,
        );

        React.startTransition(() => {
          setActiveParentId(detectedParent);

          if (hasChildren) {
            if (storedSidebarState !== null) {
              setIsDetailSidebarOpen(storedSidebarState !== "false");
            } else {
              setIsDetailSidebarOpen(true);
            }
          } else {
            setIsDetailSidebarOpen(false);
          }
        });
      } else {
        React.startTransition(() => {
          setActiveParentId(null);
          setIsDetailSidebarOpen(false);
        });
      }

      previousPathnameRef.current = currentPathname;
      return;
    }

    const pathnameChanged = previousPathname !== currentPathname;

    if (pathnameChanged && !manualSelectionRef.current) {
      const detectedParent = findParentMenuByPath(menus, currentPathname);

      if (detectedParent !== activeParentId) {
        const parentMenu = menus.find(
          (menu) => String(menu.id) === detectedParent,
        );
        const hasChildren = Boolean(
          parentMenu?.children && parentMenu.children.length > 0,
        );

        React.startTransition(() => {
          setActiveParentId(detectedParent);

          if (!hasChildren) {
            setIsDetailSidebarOpen(false);
          } else if (isDetailSidebarOpen) {
            setIsDetailSidebarOpen(true);
          }
        });
      }
    }

    previousPathnameRef.current = currentPathname;

    if (manualSelectionRef.current && pathnameChanged) {
      const timer = setTimeout(() => {
        manualSelectionRef.current = false;
      }, 200);
      return () => clearTimeout(timer);
    }
  }, [pathname, menus, isMounted, activeParentId, isDetailSidebarOpen]);

  // Allow pages (e.g. full-screen map pages) to open the mobile sidebar
  // by dispatching a global event `openMobileSidebar` (window.dispatchEvent).
  useEffect(() => {
    const handler = () => setIsMobileSidebarOpen(true);
    if (typeof window !== "undefined") {
      window.addEventListener("openMobileSidebar", handler as EventListener);
      return () => window.removeEventListener("openMobileSidebar", handler as EventListener);
    }
    return undefined;
  }, []);

  const handleSelectParent = useCallback(
    (id: string) => {
      const item = parentItems.find((parent) => parent.id === id);
      if (!item) return;

      manualSelectionRef.current = true;
      setTimeout(() => {
        manualSelectionRef.current = false;
      }, 500);

      if (item.hasChildren) {
        setActiveParentId(id);
        setIsDetailSidebarOpen(true);
      } else if (item.href) {
        setActiveParentId(id);
        setIsDetailSidebarOpen(false);
      }
    },
    [parentItems],
  );

  const handleToggleDetailSidebar = useCallback(() => {
    setIsDetailSidebarOpen((prev) => !prev);
  }, []);

  const isAIChatbotPage = pathname?.includes("/ai-chatbot");
  // POS terminal is full-screen when an outlet is selected (tracked via store).
  // Live-table is always full-screen (URL-based, as it always has outlet context).
  const posTerminalActive = usePOSUIStore((s) => s.isFullScreen);
  const isPOSPage =
    posTerminalActive || pathname?.includes("/pos/fb/live-table");
  const isFullScreenMapPage =
    isPOSPage ||
    pathname?.includes("/master-data/company") ||
    pathname?.includes("/master-data/outlet") ||
    pathname?.includes("/master-data/suppliers") ||
    pathname?.includes("/master-data/warehouses") ||
    pathname?.includes("/master-data/customers") ||
    pathname?.includes("/master-data/areas") ||
    pathname?.includes("/master-data/geographic") ||
    pathname?.includes("/crm/area-mapping") ||
    pathname?.includes("/travel-planner") ||
    pathname?.includes("/visit-planner");
  const isDashboardOnboardingMode =
    isGeneralDashboardPage && isTenantOwner && onboardingState?.completed === false;
  const shouldShowDashboardHeader = !isAIChatbotPage && !isFullScreenMapPage;

  const handleOpenAttendanceDrawer = useCallback(
    (tab: AttendanceDrawerTab, openCreateLeave?: boolean) => {
      if (tab === "leave") {
        setIsLeaveRequestDrawerOpen(true);
        if (openCreateLeave) {
          setOpenCreateLeaveInDrawer((prev) => prev + 1);
        }
      } else if (tab === "overtime") {
        setIsOvertimeDrawerOpen(true);
      } else {
        setAttendanceDrawerTab(tab);
        setIsAttendanceDrawerOpen(true);
      }
    },
    [],
  );

  const shouldShowDetailSidebar = useMemo(() => {
    if (!activeParentId || !isMounted) return false;
    const parent = parentItems.find((item) => item.id === activeParentId);
    return parent?.hasChildren === true;
  }, [activeParentId, parentItems, isMounted]);

  const contentMarginLeft = useMemo(() => {
    if (!isMounted) return undefined;
    if (isMobile) return "0";
    if (isDetailSidebarOpen && shouldShowDetailSidebar)
      return "calc(4rem + 14rem)";
    return "4rem";
  }, [isMobile, isDetailSidebarOpen, shouldShowDetailSidebar, isMounted]);

  return (
    <TooltipProvider delayDuration={0}>
      <div className="min-h-screen bg-sidebar">
        {!isMobile && (
          <>
            <IconSidebar
              items={parentItems}
              activeParentId={activeParentId}
              onSelectParent={handleSelectParent}
            />
            {shouldShowDetailSidebar && (
              <DetailSidebar
                title={activeParentTitle}
                items={detailItems}
                isOpen={isDetailSidebarOpen}
                onToggle={handleToggleDetailSidebar}
              />
            )}
            {shouldShowDetailSidebar && !isDetailSidebarOpen && (
              <button
                type="button"
                onClick={handleToggleDetailSidebar}
                className="fixed left-16 top-1/2 z-30 flex h-8 w-5 -translate-y-1/2 items-center justify-center rounded-l-none rounded-r-md bg-sidebar/90 transition-colors hover:bg-accent"
                aria-label="Open detail sidebar"
              >
                <span className="text-lg leading-none">&rsaquo;</span>
              </button>
            )}
          </>
        )}

        <DashboardMobileSidebar
          isOpen={isMobileSidebarOpen}
          onClose={() => setIsMobileSidebarOpen(false)}
          parentItems={parentItems}
          activeParentId={activeParentId}
          onSelectParent={handleSelectParent}
          detailItems={detailItems}
          detailTitle={activeParentTitle}
        />

        <main
          className={cn(
            "relative min-h-screen md:ml-16",
            isMounted && "transition-[margin] duration-300 ease-out",
          )}
          style={contentMarginLeft ? { marginLeft: contentMarginLeft } : undefined}
        >
          <div
            className="relative min-h-full bg-background md:rounded-none md:shadow-none"
          >
            {shouldShowDashboardHeader &&
              (isDashboardOnboardingMode ? (
                <div className="absolute inset-x-0 top-0 z-20">
                  <DashboardHeader
                    userName={userName}
                    tenantName={tenantName}
                    subscriptionPlan={subscriptionPlan}
                    avatarUrl={primaryAvatarUrl}
                    fallbackAvatarUrl={fallbackAvatarUrl}
                    onMobileMenuClick={() => setIsMobileSidebarOpen(true)}
                    onSearchClick={handleSearchClick}
                    showAttendanceIndicator={showAttendanceIndicator}
                    onOpenAttendanceDrawer={handleOpenAttendanceDrawer}
                  />
                </div>
              ) : (
                <DashboardHeader
                  userName={userName}
                  tenantName={tenantName}
                  subscriptionPlan={subscriptionPlan}
                  avatarUrl={primaryAvatarUrl}
                  fallbackAvatarUrl={fallbackAvatarUrl}
                  onMobileMenuClick={() => setIsMobileSidebarOpen(true)}
                  onSearchClick={handleSearchClick}
                  showAttendanceIndicator={showAttendanceIndicator}
                  onOpenAttendanceDrawer={handleOpenAttendanceDrawer}
                />
              ))}

            <div
              className={cn(
                "flex flex-1 flex-col",
                isAIChatbotPage || isFullScreenMapPage
                  ? "h-[calc(100vh-1px)] gap-0 p-0"
                  : isDashboardOnboardingMode
                      ? "min-h-screen items-center justify-center gap-0 overflow-y-auto px-4 py-6"
                    : "gap-4 p-4 md:p-6",
              )}
            >
              {!isAIChatbotPage && !isFullScreenMapPage && (
                <SubscriptionLifecycleGuard />
              )}
              {!isAIChatbotPage &&
                !isFullScreenMapPage &&
                !isDashboardOnboardingMode && <Breadcrumb />}
              {children}
            </div>
          </div>
        </main>

        <NotificationDrawer open={isDrawerOpen} onOpenChange={closeDrawer} />

        <AttendanceRightDrawer
          open={isAttendanceDrawerOpen}
          onOpenChange={setIsAttendanceDrawerOpen}
        />

        <LeaveRequestDrawer
          open={isLeaveRequestDrawerOpen}
          onOpenChange={setIsLeaveRequestDrawerOpen}
          openCreateSignal={openCreateLeaveInDrawer}
        />

        <OvertimeDrawer
          open={isOvertimeDrawerOpen}
          onOpenChange={setIsOvertimeDrawerOpen}
        />

        <CommandPalette
          open={isCommandPaletteOpen}
          onOpenChange={(open) => setIsCommandPaletteOpen(open)}
          menus={menus}
        />
        {!isAIChatbotPage && !isPOSPage && <AIChatWidget />}
      </div>
    </TooltipProvider>
  );
});
