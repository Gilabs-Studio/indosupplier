"use client";

import React, { useEffect, useState } from "react";
import { usePathname, useRouter, Link } from "@/i18n/routing";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import { useTranslations, useLocale } from "next-intl";
import { useSupplierProfile } from "@/features/supplier/profile/hooks/useProfile";
import {
  LayoutDashboard,
  Inbox,
  LogOut,
  Loader2,
  Search,
  Bell,
  User,
  Building2,
  Megaphone,
  Gavel,
  CreditCard,
  Receipt,
  Headset,
  Star,
  AlertCircle
} from "lucide-react";
import { toast } from "sonner";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { getDicebearUrl } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogClose,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

interface SupplierLayoutProps {
  children: React.ReactNode;
}

export default function SupplierLayoutComponent({ children }: SupplierLayoutProps) {
  const t = useTranslations("supplier.layout");
  const locale = useLocale();
  const router = useRouter();
  const pathname = usePathname();
  const { user, isAuthenticated, isSessionVerified, setUser, setSessionVerified, logout } = useAuthStore();
  const [mounted, setMounted] = useState(false);
  const [isAuthorizing, setIsAuthorizing] = useState(true);
  const [isSidebarExpanded] = useState(true);
  const [showLogoutConfirm, setShowLogoutConfirm] = useState(false);
  const { data: profile } = useSupplierProfile();
  const isLocalVerified = typeof window !== "undefined" && localStorage.getItem("supplier_verified") === "true";
  const isVerified = profile ? (profile.status === "active" || isLocalVerified) : isLocalVerified;

  const hasSupplierAccess =
    user?.capabilities.supplier === true || !!user?.supplier_profile;

  useEffect(() => {
    const timer = setTimeout(() => {
      setMounted(true);
    }, 0);
    return () => clearTimeout(timer);
  }, []);

  useEffect(() => {
    if (!mounted) {
      return;
    }

    let isCancelled = false;

    const authorizeSupplierPortal = async () => {
      if (isSessionVerified && isAuthenticated) {
        if (!hasSupplierAccess && pathname !== "/supplier/register") {
          router.replace("/supplier/register");
          return;
        }

        if (hasSupplierAccess && (pathname === "/supplier/onboarding" || pathname === "/supplier/register")) {
          router.replace("/supplier/dashboard");
          return;
        }

        setIsAuthorizing(false);
        return;
      }

      try {
        const { authService } = await import("@/features/auth/services/auth-service");

        try {
          await authService.prefetchCSRFToken();
        } catch {
          // Ignore CSRF prefetch failures and let refresh-token decide.
        }

        const response = await authService.getMe();
        const authenticatedUser = response?.data?.user;

        if (!authenticatedUser) {
          throw new Error("UNAUTHENTICATED_SUPPLIER_PORTAL");
        }

        if (isCancelled) {
          return;
        }

        setUser(authenticatedUser);
        setSessionVerified(true);

        const authenticatedHasSupplierAccess =
          authenticatedUser.capabilities.supplier === true || !!authenticatedUser.supplier_profile;

        if (!authenticatedHasSupplierAccess && pathname !== "/supplier/register") {
          router.replace("/supplier/register");
          return;
        }

        if (authenticatedHasSupplierAccess && (pathname === "/supplier/onboarding" || pathname === "/supplier/register")) {
          router.replace("/supplier/dashboard");
          return;
        }

        setIsAuthorizing(false);
      } catch {
        if (isCancelled) {
          return;
        }

        logout();
        const { fullAuthCleanup } = await import("@/features/auth/utils/clear-auth-cookies");
        await fullAuthCleanup();
        router.replace(`/login?redirectTo=${encodeURIComponent(pathname)}`);
      }
    };

    void authorizeSupplierPortal();

    return () => {
      isCancelled = true;
    };
  }, [
    hasSupplierAccess,
    isAuthenticated,
    isSessionVerified,
    logout,
    mounted,
    pathname,
    router,
    setSessionVerified,
    setUser,
  ]);

  if (!mounted || isAuthorizing) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-background text-foreground">
        <Loader2 className="h-10 w-10 animate-spin text-primary mb-4" />
        <p className="text-sm font-semibold tracking-wide text-muted-foreground">
          {t("loading")}
        </p>
      </div>
    );
  }

  // Navigation Groups matching Tokopedia Seller style
  const menuGroups = [
    {
      title: locale === "id" ? "Menu Utama" : "Main Menu",
      items: [
        {
          name: t("menu.dashboard"),
          icon: LayoutDashboard,
          url: "/supplier/dashboard",
        },
        {
          name: t("menu.products"),
          icon: Building2,
          url: "/supplier/products",
        },
        {
          name: t("menu.rfqs"),
          icon: Inbox,
          url: "/supplier/rfq",
        },
      ],
    },
    {
      title: locale === "id" ? "Pemasaran" : "Marketing",
      items: [
        {
          name: t("menu.ads"),
          icon: Megaphone,
          url: "/supplier/ads",
        },
        {
          name: t("menu.auctions"),
          icon: Gavel,
          url: "/supplier/auction",
        },
      ],
    },
    {
      title: locale === "id" ? "Profil & Verifikasi" : "Profile & Verification",
      items: [
        {
          name: t("menu.profile"),
          icon: User,
          url: "/supplier/profile",
        },
        {
          name: t("menu.subscription"),
          icon: Receipt,
          url: "/supplier/subscription",
        },
      ],
    },
    {
      title: locale === "id" ? "Keuangan & Bantuan" : "Finance & Help",
      items: [
        {
          name: t("menu.billing"),
          icon: CreditCard,
          url: "/supplier/billing",
        },
        {
          name: t("menu.support"),
          icon: Headset,
          url: "/supplier/support",
        },
        {
          name: t("menu.reviews"),
          icon: Star,
          url: "/supplier/reviews",
        },
      ],
    },
  ];

  return (
    <div className="min-h-screen flex bg-background text-foreground transition-colors duration-300">
      {/* ── Left Sidebar (Floating style) ── */}
      <aside
        className={`fixed top-4 bottom-4 left-4 z-40 bg-card border border-border/50 shadow-[0_8px_30px_rgb(0,0,0,0.04)] dark:shadow-[0_8px_30px_rgb(0,0,0,0.2)] flex flex-col justify-between transition-all duration-300 ease-in-out rounded-lg ${
          isSidebarExpanded ? "w-64" : "w-[72px]"
        }`}
      >
        <div className="flex flex-col flex-1 overflow-y-auto">
          {/* Logo Section */}
          <div className="h-16 flex items-center px-6 border-b border-border/50 overflow-hidden shrink-0">
            {isSidebarExpanded ? (
              <span className="font-extrabold text-foreground tracking-tight text-lg select-none leading-none">
                indosupplier
              </span>
            ) : (
              <span className="font-extrabold text-primary tracking-tight text-lg select-none leading-none mx-auto">
                is
              </span>
            )}
          </div>

          {/* Navigation Menus */}
          <div className="p-3 space-y-4">
            {menuGroups.map((group) => (
              <div key={group.title} className="space-y-1">
                {isSidebarExpanded && (
                  <p className="px-3 text-[10px] font-bold text-muted-foreground uppercase tracking-wider mb-1">
                    {group.title}
                  </p>
                )}
                {group.items.map((item) => {
                  const isActive = pathname === item.url || (item.url !== "/supplier/dashboard" && pathname.startsWith(item.url));
                  const Icon = item.icon;

                  return (
                    <Link
                      key={item.name}
                      href={item.url}
                      className={`flex items-center gap-3 px-3.5 py-2.5 rounded-lg relative overflow-hidden transition-all duration-300 group cursor-pointer hover:-translate-y-0.5 active:translate-y-0 ${
                        isActive
                          ? "text-primary bg-primary/8 font-semibold shadow-xs shadow-primary/10"
                          : "text-muted-foreground hover:text-foreground hover:bg-muted/40"
                      }`}
                    >
                      <Icon className={`h-4.5 w-4.5 transition-transform duration-300 group-hover:scale-105 ${isActive ? "text-primary" : ""}`} />
                      
                      {isSidebarExpanded && (
                        <span className="text-sm select-none truncate transition-opacity duration-300">
                          {item.name}
                        </span>
                      )}
                    </Link>
                  );
                })}
              </div>
            ))}
          </div>
        </div>
      </aside>

      {/* ── Main Content Area ── */}
      <div
        className={`flex-1 flex flex-col min-h-screen transition-all duration-300 ${
          isSidebarExpanded ? "pl-[288px]" : "pl-[104px]"
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
              placeholder={t("searchPlaceholder")}
              className="w-full pl-9 pr-4 py-2 text-sm bg-muted/30 border border-border rounded-lg placeholder-muted-foreground focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary transition-all"
            />
          </div>

          {/* User profile & actions */}
          <div className="flex items-center gap-4">
            {/* Notification bell */}
            <button
              onClick={() => toast.info(t("notificationAlert"))}
              className="relative p-2 text-muted-foreground hover:text-foreground rounded-lg hover:bg-muted/40 transition-colors cursor-pointer"
            >
              <Bell className="h-5 w-5" />
              <span className="absolute top-1 right-1 flex h-4 w-4 items-center justify-center rounded-full bg-destructive text-[9px] font-bold text-destructive-foreground">
                2
              </span>
            </button>

            {/* Separator line */}
            <div className="h-6 w-px bg-border/80" />

            {/* Supplier Info Profile */}
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <div className="flex items-center gap-2.5 cursor-pointer hover:bg-muted/40 p-1.5 rounded-lg transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 select-none">
                  {/* Avatar circle */}
                  <Avatar className="h-9 w-9 border border-border">
                    <AvatarImage src={getDicebearUrl(user?.email || "supplier", "lorelei")} alt={user?.name} />
                    <AvatarFallback className="bg-primary/10 text-primary font-semibold text-xs">
                      {user?.name?.slice(0, 2).toUpperCase() || "SP"}
                    </AvatarFallback>
                  </Avatar>
                  <div className="hidden sm:flex flex-col text-left">
                    <span className="text-xs font-semibold text-foreground leading-none">{user?.name || "PT Nusantara Supplier"}</span>
                    <span className="text-[10px] text-success font-semibold flex items-center gap-1 mt-1 leading-none">
                      <span className="h-1.5 w-1.5 rounded-full bg-success inline-block" />
                      {t("online")}
                    </span>
                  </div>
                </div>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-56 mt-1 rounded-lg">
                <DropdownMenuLabel className="font-semibold text-xs text-muted-foreground uppercase tracking-wider px-3 py-2">
                  {locale === "id" ? "Portal Supplier" : "Supplier Portal"}
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem asChild>
                  <Link href="/supplier/profile" className="flex items-center gap-2 px-3 py-2 cursor-pointer w-full text-sm">
                    <User className="h-4 w-4" />
                    <span>{t("menu.profile")}</span>
                  </Link>
                </DropdownMenuItem>
                <DropdownMenuItem asChild>
                  <Link href="/supplier/subscription" className="flex items-center gap-2 px-3 py-2 cursor-pointer w-full text-sm">
                    <Receipt className="h-4 w-4" />
                    <span>{t("menu.subscription")}</span>
                  </Link>
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  onClick={() => setShowLogoutConfirm(true)}
                  className="flex items-center gap-2 px-3 py-2 text-destructive focus:text-destructive cursor-pointer"
                >
                  <LogOut className="h-4 w-4" />
                  <span>{t("signOut")}</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </header>

        {/* Content body wrapper with smooth fade transition */}
        <main className="flex-1 bg-muted/10 p-6 md:p-8 animate-fade-in overflow-y-auto">
          <div className="max-w-6xl mx-auto space-y-6">
            {!isVerified && (
              <div className="bg-card border border-border/80 border-l-4 border-l-primary/90 p-4.5 rounded-lg flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 select-none shadow-xs transition-all duration-300 hover:shadow-md">
                <div className="flex items-center gap-3.5 text-left">
                  <div className="h-9 w-9 rounded-lg bg-primary/10 text-primary flex items-center justify-center shrink-0">
                    <AlertCircle className="h-5 w-5" />
                  </div>
                  <div className="flex flex-col">
                    <span className="font-extrabold text-[10px] uppercase tracking-wider text-primary">
                      {t("warningBannerBadge")}
                    </span>
                    <span className="text-xs text-muted-foreground font-medium mt-1 leading-relaxed">
                      {t("warningBannerMessage")}
                    </span>
                  </div>
                </div>
                <Link
                  href="/supplier/verification"
                  className="border border-primary text-primary hover:bg-primary hover:text-primary-foreground px-4 py-2 rounded-lg text-xs font-bold transition-all duration-300 cursor-pointer shrink-0 self-start sm:self-auto hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/20 flex items-center gap-1.5"
                >
                  {t("warningBannerButton")}
                </Link>
              </div>
            )}
            {children}
          </div>
        </main>
      </div>

      <Dialog open={showLogoutConfirm} onOpenChange={setShowLogoutConfirm}>
        <DialogContent size="sm">
          <DialogHeader>
            <DialogTitle className="font-heading">{t("signOut")}</DialogTitle>
            <DialogDescription>{t("signOutConfirm")}</DialogDescription>
          </DialogHeader>
          <DialogFooter className="gap-2 sm:gap-0">
            <DialogClose asChild>
              <Button variant="outline" className="cursor-pointer">
                Cancel
              </Button>
            </DialogClose>
            <Button
              variant="destructive"
              className="cursor-pointer"
              onClick={() => {
                setShowLogoutConfirm(false);
                logout();
                router.push("/login");
              }}
            >
              {t("signOut")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
