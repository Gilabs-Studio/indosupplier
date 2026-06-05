"use client";

import React, { useEffect, useState } from "react";
import { usePathname, useRouter, Link } from "@/i18n/routing";
import { useSysadminStore } from "@/features/sysadmin/auth/stores/use-sysadmin-store";
import { useTranslations } from "next-intl";
import {
  LayoutDashboard,
  Inbox,
  LogOut,
  ChevronLeft,
  ChevronRight,
  Loader2,
  Search,
  Bell,
  User,
  ShieldCheck,
  Building,
  Users,
  Megaphone,
  Gavel,
  FolderTree,
  CreditCard,
  Headset,
  BookOpen,
  Star,
  Flag,
  ScrollText
} from "lucide-react";
import { toast } from "sonner";

interface SysadminLayoutProps {
  children: React.ReactNode;
}

export default function SysadminLayoutComponent({ children }: SysadminLayoutProps) {
  const t = useTranslations("sysadminDashboard");
  const router = useRouter();
  const pathname = usePathname();
  const { isAuthenticated, isSessionVerified, isLoading, checkSession, admin, logout } = useSysadminStore();
  const [mounted, setMounted] = useState(false);
  const [isSidebarExpanded, setIsSidebarExpanded] = useState(true);

  useEffect(() => {
    const timer = setTimeout(() => {
      setMounted(true);
      checkSession();
    }, 0);
    return () => clearTimeout(timer);
  }, [checkSession]);

  useEffect(() => {
    if (!mounted || isLoading || !isSessionVerified) return;

    const isLoginPage = pathname.endsWith("/sysadmin/login") || pathname.includes("/sysadmin/login");

    if (isAuthenticated) {
      if (isLoginPage) {
        router.push("/sysadmin");
      }
    } else {
      if (!isLoginPage) {
        router.push("/sysadmin/login");
      }
    }
  }, [mounted, isAuthenticated, isSessionVerified, isLoading, pathname, router]);

  if (!mounted || isLoading || !isSessionVerified) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-background text-foreground">
        <Loader2 className="h-10 w-10 animate-spin text-primary mb-4" />
        <p className="text-sm font-semibold tracking-wide text-muted-foreground">
          {t("layout.loading")}
        </p>
      </div>
    );
  }

  const isLoginPage = pathname.endsWith("/sysadmin/login") || pathname.includes("/sysadmin/login");
  if (!isAuthenticated && !isLoginPage) {
    return null;
  }
  if (isAuthenticated && isLoginPage) {
    return null;
  }

  // If on login page, just render children without sidebar/header
  if (isLoginPage) {
    return <>{children}</>;
  }

  // Navigation Items
  const menuItems = [
    {
      name: t("layout.menu.dashboard"),
      icon: LayoutDashboard,
      url: "/sysadmin",
    },
    {
      name: t("layout.menu.waitingList"),
      icon: Inbox,
      url: "/sysadmin/waiting-list",
    },
    {
      name: t("layout.menu.suppliers"),
      icon: Building,
      url: "/sysadmin/suppliers",
    },
    {
      name: t("layout.menu.buyers"),
      icon: Users,
      url: "/sysadmin/buyers",
    },
    {
      name: t("layout.menu.adReviews"),
      icon: Megaphone,
      url: "/sysadmin/ads",
    },
    {
      name: t("layout.menu.auctions"),
      icon: Gavel,
      url: "/sysadmin/auctions",
    },
    {
      name: t("layout.menu.categories"),
      icon: FolderTree,
      url: "/sysadmin/categories",
    },
    {
      name: t("layout.menu.subscriptionPlans"),
      icon: CreditCard,
      url: "/sysadmin/subscription-plans",
    },
    {
      name: t("layout.menu.supportTickets"),
      icon: Headset,
      url: "/sysadmin/support",
    },
    {
      name: t("layout.menu.faqManagement"),
      icon: BookOpen,
      url: "/sysadmin/faq",
    },
    {
      name: t("layout.menu.reviewsModeration"),
      icon: Star,
      url: "/sysadmin/reviews",
    },
    {
      name: t("layout.menu.abuseReports"),
      icon: Flag,
      url: "/sysadmin/abuse-reports",
    },
    {
      name: t("layout.menu.auditLogs"),
      icon: ScrollText,
      url: "/sysadmin/audit-logs",
    },
  ];

  return (
    <div className="min-h-screen flex bg-background text-foreground transition-colors duration-300">
      {/* ── Left Sidebar (Tokopedia Seller style) ── */}
      <aside
        className={`fixed top-0 bottom-0 left-0 z-40 bg-card border-r border-border/80 flex flex-col justify-between transition-all duration-300 ease-in-out ${
          isSidebarExpanded ? "w-60" : "w-[72px]"
        }`}
      >
        <div className="flex flex-col flex-1 overflow-y-auto">
          {/* Logo Section */}
          <div className="h-16 flex items-center px-5 border-b border-border/80 gap-3 overflow-hidden">
            <div className="flex items-center justify-center h-9 w-9 rounded-lg bg-primary/10 text-primary shrink-0">
              <ShieldCheck className="h-5 w-5" />
            </div>
            {isSidebarExpanded && (
              <div className="flex flex-col select-none leading-none">
                <span className="font-extrabold text-foreground tracking-tight text-base">
                  indosupplier
                </span>
                <span className="text-[10px] text-muted-foreground font-normal mt-0.5 tracking-wider uppercase">
                  sysadmin
                </span>
              </div>
            )}
          </div>

          {/* Navigation Menus */}
          <nav className="p-3 space-y-1">
            {menuItems.map((item) => {
              const isActive = pathname === item.url;
              const Icon = item.icon;

              return (
                <Link
                  key={item.name}
                  href={item.url}
                  className={`flex items-center gap-3 px-3 py-3 rounded-lg relative overflow-hidden transition-all duration-200 group cursor-pointer ${
                    isActive
                      ? "text-primary bg-primary/10 font-semibold"
                      : "text-muted-foreground hover:text-foreground hover:bg-muted/40"
                  }`}
                >
                  {/* Left indicator line for active page */}
                  {isActive && (
                    <div className="absolute left-0 top-0 bottom-0 w-1 bg-primary rounded-r-md" />
                  )}
                  
                  <Icon className={`h-5 w-5 transition-transform duration-200 group-hover:scale-105 ${isActive ? "text-primary" : ""}`} />
                  
                  {isSidebarExpanded && (
                    <span className="text-sm select-none truncate transition-opacity duration-200">
                      {item.name}
                    </span>
                  )}
                </Link>
              );
            })}
          </nav>
        </div>

        {/* Footer Actions inside Sidebar */}
        <div className="p-3 border-t border-border/80 space-y-1">
          {/* Sign Out Button */}
          <button
            onClick={() => {
              if (confirm(t("layout.signOutConfirm"))) logout();
            }}
            className={`w-full flex items-center gap-3 px-3 py-3 rounded-lg text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-all duration-200 group cursor-pointer ${
              isSidebarExpanded ? "justify-start" : "justify-center"
            }`}
          >
            <LogOut className="h-5 w-5 shrink-0 group-hover:translate-x-0.5 transition-transform" />
            {isSidebarExpanded && (
              <span className="text-sm font-medium select-none truncate">
                {t("layout.signOut")}
              </span>
            )}
          </button>

          {/* Collapse Toggle Button */}
          <button
            onClick={() => setIsSidebarExpanded(!isSidebarExpanded)}
            className={`w-full flex items-center gap-3 px-3 py-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted/40 transition-all duration-200 cursor-pointer ${
              isSidebarExpanded ? "justify-end" : "justify-center"
            }`}
          >
            {isSidebarExpanded ? (
              <ChevronLeft className="h-4 w-4" />
            ) : (
              <ChevronRight className="h-4 w-4" />
            )}
          </button>
        </div>
      </aside>

      {/* ── Main Content Area ── */}
      <div
        className={`flex-1 flex flex-col min-h-screen transition-all duration-300 ${
          isSidebarExpanded ? "pl-60" : "pl-[72px]"
        }`}
      >
        {/* Header (Top navigation) */}
        <header className="sticky top-0 z-30 h-16 bg-card border-b border-border/80 flex items-center justify-between px-6">
          {/* Search bar section */}
          <div className="max-w-md w-full relative">
            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
              <Search className="h-4 w-4 text-muted-foreground" />
            </div>
            <input
              type="text"
              placeholder={t("layout.searchPlaceholder")}
              className="w-full pl-9 pr-4 py-2 text-sm bg-muted/30 border border-border rounded-lg placeholder-muted-foreground focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary transition-all"
            />
          </div>

          {/* User profile & actions */}
          <div className="flex items-center gap-4">
            {/* Notification bell */}
            <button
              onClick={() => toast.info(t("layout.notificationAlert"))}
              className="relative p-2 text-muted-foreground hover:text-foreground rounded-lg hover:bg-muted/40 transition-colors cursor-pointer"
            >
              <Bell className="h-5 w-5" />
              <span className="absolute top-1 right-1 flex h-4 w-4 items-center justify-center rounded-full bg-destructive text-[9px] font-bold text-destructive-foreground">
                3
              </span>
            </button>

            {/* Separator line */}
            <div className="h-6 w-px bg-border/80" />

            {/* Admin Info Profile */}
            <div className="flex items-center gap-2.5">
              {/* Avatar circle */}
              <div className="h-9 w-9 rounded-full bg-primary/10 flex items-center justify-center text-primary border border-border select-none">
                <User className="h-4 w-4" />
              </div>
              <div className="hidden sm:flex flex-col text-left">
                <span className="text-xs font-semibold text-foreground leading-none">{admin?.name || "System Admin"}</span>
                <span className="text-[10px] text-muted-foreground font-normal capitalize mt-0.5 leading-none">
                  {admin?.role?.replace("_", " ") || "Super Admin"}
                </span>
              </div>
            </div>
          </div>
        </header>

        {/* Content body wrapper with smooth fade transition */}
        <main className="flex-1 bg-muted/10 p-6 md:p-8 animate-fade-in overflow-y-auto">
          <div className="max-w-6xl mx-auto">
            {children}
          </div>
        </main>
      </div>
    </div>
  );
}
