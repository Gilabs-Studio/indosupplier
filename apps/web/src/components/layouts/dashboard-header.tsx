"use client";

import React, { useState, memo } from "react";
import { useLocale, useTranslations } from "next-intl";
import { Link, usePathname } from "@/i18n/routing";
import {
  Activity,
  Laptop,
  Menu as MenuIcon,
  Search,
  Settings,
  Smartphone,
} from "lucide-react";
import { AnimatePresence, motion } from "framer-motion";

import { Button } from "@/components/ui/button";
import { Avatar, AvatarImage } from "@/components/ui/avatar";
// Dropdown primitives are not used anymore; keep imports removed.
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Separator } from "@/components/ui/separator";
import { ThemeToggleButton as ThemeToggle } from "@/components/ui/theme-toggle";
// Tooltip primitives removed; downloads moved to popover.
import { NotificationBadge } from "@/features/notifications/components/notification-badge";
import { HeaderAttendanceButton } from "@/features/hrd/attendance-records/components/header-attendance-button";
import { useLogout } from "@/features/auth/hooks/use-logout";
import { useIsMobile } from "@/hooks/use-mobile";
import { ActivityFeedDialog } from "@/features/crm/activity/components/activity-feed-dialog";
import type { AttendanceDrawerTab } from "@/features/hrd/attendance-records/components/attendance-right-drawer";
import { useHasPermission } from "@/features/master-data/user-management/hooks/use-has-permission";

interface DashboardHeaderProps {
  userName: string;
  tenantName?: string;
  subscriptionPlan?: string;
  avatarUrl?: string;
  fallbackAvatarUrl: string;
  onMobileMenuClick: () => void;
  onSearchClick: () => void;
  showAttendanceIndicator?: boolean;
  onOpenAttendanceDrawer: (
    tab: AttendanceDrawerTab,
    openCreateLeave?: boolean,
  ) => void;
}

type POSDownloadPlatform = "windows" | "android" | "macos" | "linux";

interface POSDownloadLink {
  platform: POSDownloadPlatform;
  label: string;
  url: string;
  icon: typeof Laptop;
}

function isSafeExternalUrl(value: string | undefined): string | null {
  if (!value) return null;
  const trimmed = value.trim();
  if (!trimmed) return null;

  try {
    const parsed = new URL(trimmed);
    if (parsed.protocol !== "https:" && parsed.protocol !== "http:") {
      return null;
    }
    return parsed.toString();
  } catch {
    return null;
  }
}

function getPOSDownloadUrlMap(): Partial<Record<POSDownloadPlatform, string>> {
  const fromJson = process.env.NEXT_PUBLIC_POS_DOWNLOAD_LINKS;
  if (fromJson) {
    try {
      const parsed = JSON.parse(fromJson) as Partial<Record<POSDownloadPlatform, unknown>>;
      return {
        windows:
          typeof parsed.windows === "string" ? parsed.windows : undefined,
        android:
          typeof parsed.android === "string" ? parsed.android : undefined,
        macos: typeof parsed.macos === "string" ? parsed.macos : undefined,
        linux: typeof parsed.linux === "string" ? parsed.linux : undefined,
      };
    } catch {
      // Ignore invalid JSON and fallback to per-platform environment variables.
    }
  }

  return {
    windows: process.env.NEXT_PUBLIC_POS_DOWNLOAD_WINDOWS_URL,
    android: process.env.NEXT_PUBLIC_POS_DOWNLOAD_ANDROID_URL,
    macos: process.env.NEXT_PUBLIC_POS_DOWNLOAD_MACOS_URL,
    linux: process.env.NEXT_PUBLIC_POS_DOWNLOAD_LINUX_URL,
  };
}

export const DashboardHeader = memo(function DashboardHeader({
  userName,
  tenantName,
  subscriptionPlan,
  avatarUrl,
  fallbackAvatarUrl,
  onMobileMenuClick,
  onSearchClick,
  showAttendanceIndicator = false,
  onOpenAttendanceDrawer,
}: DashboardHeaderProps) {
  const locale = useLocale();
  const t = useTranslations("common");
  const logout = useLogout();
  const pathname = usePathname();
  const isMobile = useIsMobile();

  const normalizedAvatarUrl =
    avatarUrl && avatarUrl.trim() !== "" ? avatarUrl : undefined;
  const [avatarLoadFailed, setAvatarLoadFailed] = useState(false);
  const [activityFeedOpen, setActivityFeedOpen] = useState(false);
  const [userPopoverOpen, setUserPopoverOpen] = useState(false);
  const [posDownloadOpen, setPosDownloadOpen] = useState(false);
  const currentSrc = avatarLoadFailed
    ? fallbackAvatarUrl
    : (normalizedAvatarUrl ?? fallbackAvatarUrl);

  // Allow viewing activities when either explicit activity permission
  // exists or when user can view pipeline (reuses existing permission).
  // Call hooks unconditionally to satisfy React rules of hooks.
  const canViewActivityPerm = useHasPermission("crm_activity.read");
  const canViewPipelinePerm = useHasPermission("crm_deal.read");
  const canViewActivities = canViewActivityPerm || canViewPipelinePerm;
  const posDownloadUrls = getPOSDownloadUrlMap();
  const posDownloadLinksRaw: POSDownloadLink[] = [
    {
      platform: "windows",
      label: t("posDownloadWindows"),
      url: isSafeExternalUrl(posDownloadUrls.windows) ?? "",
      icon: Laptop,
    },
    {
      platform: "android",
      label: t("posDownloadAndroid"),
      url: isSafeExternalUrl(posDownloadUrls.android) ?? "",
      icon: Smartphone,
    },
    {
      platform: "macos",
      label: t("posDownloadMacos"),
      url: isSafeExternalUrl(posDownloadUrls.macos) ?? "",
      icon: Laptop,
    },
    {
      platform: "linux",
      label: t("posDownloadLinux"),
      url: isSafeExternalUrl(posDownloadUrls.linux) ?? "",
      icon: Laptop,
    },
  ];
  const posDownloadLinks = posDownloadLinksRaw.filter((item) => item.url !== "");
  const hasPOSDownloadLinks = posDownloadLinks.length > 0;

  return (
    <header className="sticky top-0 z-20 flex h-16 shrink-0 items-center gap-3 border-b bg-background px-4 md:rounded-tl-3xl">
      {isMobile && (
        <Button
          variant="ghost"
          size="icon"
          className="h-9 w-9 md:hidden"
          onClick={onMobileMenuClick}
          aria-label={t("openMenu")}
        >
          <MenuIcon className="h-5 w-5" />
        </Button>
      )}

      <div className="flex-1">
        <div className="relative hidden max-w-sm flex-1 lg:block">
          <Search
            className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-foreground/60"
            aria-hidden="true"
          />
          <input
            type="search"
            placeholder={t("searchPlaceholder")}
            className="border-input file:text-foreground placeholder:text-muted-foreground selection:bg-primary selection:text-primary-foreground h-9 w-full cursor-pointer rounded-md border bg-background/60 px-3 py-1 pl-10 pr-4 text-sm shadow-sm outline-none transition-[color,box-shadow] focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50"
            onClick={(event) => {
              event.preventDefault();
              event.currentTarget.blur();
              onSearchClick();
            }}
            readOnly
          />
          <div className="bg-muted text-muted-foreground absolute right-2 top-1/2 hidden -translate-y-1/2 items-center gap-0.5 rounded-sm px-1.5 py-0.5 font-mono text-[10px] font-medium sm:flex">
            <span className="text-[11px]">/</span>
          </div>
        </div>

        <div className="block lg:hidden">
          <Button
            variant="ghost"
            size="icon"
            className="size-9"
            type="button"
            onClick={onSearchClick}
          >
            <Search className="h-4 w-4" aria-hidden="true" />
            <span className="sr-only">{t("openSearch")}</span>
          </Button>
        </div>
      </div>

      <div className="ml-auto flex items-center gap-1 overflow-visible">
        <HeaderAttendanceButton onOpenDrawer={onOpenAttendanceDrawer} />

        {/* POS download moved into user popover (see below) */}

        <NotificationBadge />
        <ThemeToggle />

        <Link
          href={pathname || "/dashboard"}
          locale={locale === "en" ? "id" : "en"}
          scroll={false}
        >
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 rounded-full bg-background/80 text-xs font-semibold shadow-sm hover:bg-accent/60"
            type="button"
          >
            {locale === "en" ? "ID" : "EN"}
          </Button>
        </Link>

        <div className="mx-2 h-4 w-px shrink-0 bg-border data-[orientation=horizontal]:h-px data-[orientation=horizontal]:w-full data-[orientation=vertical]:h-1/2 data-[orientation=vertical]:w-px" />

        <Popover
          key={normalizedAvatarUrl ?? fallbackAvatarUrl}
          open={userPopoverOpen}
          onOpenChange={setUserPopoverOpen}
        >
          <PopoverTrigger asChild>
            <Button
              variant="ghost"
              className="flex h-8 w-8 items-center justify-center rounded-full p-0 transition-colors hover:bg-muted"
            >
              <div className="relative">
                <Avatar className="h-8 w-8">
                  <AvatarImage
                    key={currentSrc}
                    src={currentSrc}
                    alt={userName}
                    onError={() => {
                      setAvatarLoadFailed(true);
                    }}
                  />
                </Avatar>
                <AnimatePresence>
                  {showAttendanceIndicator && (
                    <motion.div
                      className="absolute -inset-1 rounded-full border-2 border-amber-500/60"
                      initial={{ opacity: 0, scale: 0.8 }}
                      animate={{
                        opacity: [0.4, 0.8, 0.4],
                        scale: [1, 1.15, 1],
                      }}
                      exit={{ opacity: 0, scale: 0.8 }}
                      transition={{
                        duration: 2,
                        repeat: Infinity,
                        ease: "easeInOut",
                      }}
                    />
                  )}
                </AnimatePresence>
              </div>
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-56 p-2" align="end">
            <div className="px-2 py-1.5 text-xs text-muted-foreground">
              <div className="text-sm font-medium text-foreground">{userName}</div>
              {tenantName && (
                <div className="mt-0.5 truncate text-xs text-muted-foreground">{tenantName}</div>
              )}
              {subscriptionPlan && (
                <div className="mt-1 inline-flex items-center rounded-md bg-primary/10 px-1.5 py-0.5 text-[10px] font-semibold capitalize text-primary">
                  {subscriptionPlan.replace(/_/g, " ")}
                </div>
              )}
            </div>
            <Separator className="my-1" />
            <div className="flex flex-col gap-1">
              <button
                type="button"
                className={
                  "flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm transition-colors " +
                  (canViewActivities
                    ? "cursor-pointer hover:bg-accent"
                    : "opacity-50 cursor-not-allowed")
                }
                onClick={() => {
                  if (!canViewActivities) return;
                  setUserPopoverOpen(false);
                  setActivityFeedOpen(true);
                }}
                disabled={!canViewActivities}
                aria-disabled={!canViewActivities}
              >
                <Activity className="h-4 w-4" />
                {t("myActivities")}
              </button>
              {hasPOSDownloadLinks && (
                <>
                  <button
                    type="button"
                    onClick={() => setPosDownloadOpen(!posDownloadOpen)}
                    className="flex w-full cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm transition-colors hover:bg-accent"
                  >
                    <Smartphone className="h-4 w-4" />
                    {t("posDownloadLabel")}
                  </button>
                  {posDownloadOpen && (
                    <div className="ml-4 flex flex-col gap-1 border-l border-muted pl-2">
                      <div className="px-2 py-1 text-xs text-muted-foreground">
                        {t("posDownloadTooltip")}
                      </div>
                      <div className="flex flex-col gap-1">
                        {posDownloadLinks.map((item) => {
                          const PlatformIcon = item.icon;
                          return (
                            <a
                              key={item.platform}
                              href={item.url}
                              target="_blank"
                              rel="noopener noreferrer"
                              onClick={() => setUserPopoverOpen(false)}
                              className="flex items-center gap-2 rounded-md px-2 py-1 text-left text-xs transition-colors hover:bg-accent"
                            >
                              <PlatformIcon className="h-3 w-3" />
                              {item.label}
                            </a>
                          );
                        })}
                      </div>
                    </div>
                  )}
                </>
              )}
              <Link
                href="/profile"
                locale={locale}
                className="flex w-full cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm transition-colors hover:bg-accent"
                onClick={() => setUserPopoverOpen(false)}
              >
                <Settings className="h-4 w-4" />
                {t("settings")}
              </Link>
              
              <button
                type="button"
                onClick={() => {
                  setUserPopoverOpen(false);
                  logout();
                }}
                className="flex w-full cursor-pointer items-center rounded-md px-2 py-1.5 text-left text-sm text-destructive transition-colors hover:bg-destructive/10"
              >
                {t("logout")}
              </button>
            </div>
          </PopoverContent>
        </Popover>
      </div>

      <ActivityFeedDialog
        open={activityFeedOpen}
        onOpenChange={setActivityFeedOpen}
      />
    </header>
  );
});
