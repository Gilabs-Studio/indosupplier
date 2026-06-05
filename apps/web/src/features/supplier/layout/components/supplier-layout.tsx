"use client";

import React, { useEffect, useState } from "react";
import { usePathname, useRouter, Link } from "@/i18n/routing";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import { useTranslations, useLocale } from "next-intl";
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
  Building2,
  Megaphone,
  Gavel,
  CreditCard,
  Receipt,
  BadgeCheck,
  Headset,
  Star
} from "lucide-react";
import { toast } from "sonner";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

interface SupplierLayoutProps {
  children: React.ReactNode;
}

export default function SupplierLayoutComponent({ children }: SupplierLayoutProps) {
  const t = useTranslations("supplier.layout");
  const locale = useLocale();
  const router = useRouter();
  const pathname = usePathname();
  const { user, logout } = useAuthStore();
  const [mounted, setMounted] = useState(false);
  const [isSidebarExpanded, setIsSidebarExpanded] = useState(true);

  useEffect(() => {
    const timer = setTimeout(() => {
      setMounted(true);
    }, 0);
    return () => clearTimeout(timer);
  }, []);

  if (!mounted) {
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
          url: "/supplier/profile/products",
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
          name: t("menu.verification"),
          icon: BadgeCheck,
          url: "/supplier/verification",
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
      {/* ── Left Sidebar (Tokopedia Seller style) ── */}
      <aside
        className={`fixed top-0 bottom-0 left-0 z-40 bg-card border-r border-border/80 flex flex-col justify-between transition-all duration-300 ease-in-out ${
          isSidebarExpanded ? "w-64" : "w-[72px]"
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
                <span className="text-[10px] text-primary font-bold mt-0.5 tracking-wider uppercase">
                  seller
                </span>
              </div>
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
                      className={`flex items-center gap-3 px-3 py-2.5 rounded-lg relative overflow-hidden transition-all duration-200 group cursor-pointer ${
                        isActive
                          ? "text-primary bg-primary/10 font-semibold"
                          : "text-muted-foreground hover:text-foreground hover:bg-muted/40"
                      }`}
                    >
                      {/* Left indicator line for active page */}
                      {isActive && (
                        <div className="absolute left-0 top-0 bottom-0 w-1 bg-primary rounded-r-md" />
                      )}
                      
                      <Icon className={`h-4.5 w-4.5 transition-transform duration-200 group-hover:scale-105 ${isActive ? "text-primary" : ""}`} />
                      
                      {isSidebarExpanded && (
                        <span className="text-sm select-none truncate transition-opacity duration-200">
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

        {/* Footer Actions inside Sidebar */}
        <div className="p-3 border-t border-border/80 space-y-1">
          {/* Sign Out Button */}
          <button
            onClick={() => {
              if (confirm(t("signOutConfirm"))) {
                logout();
                router.push("/login");
              }
            }}
            className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-all duration-200 group cursor-pointer ${
              isSidebarExpanded ? "justify-start" : "justify-center"
            }`}
          >
            <LogOut className="h-4.5 w-4.5 shrink-0 group-hover:translate-x-0.5 transition-transform" />
            {isSidebarExpanded && (
              <span className="text-sm font-medium select-none truncate">
                {t("signOut")}
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
          isSidebarExpanded ? "pl-64" : "pl-[72px]"
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
            <div className="flex items-center gap-2.5">
              {/* Avatar circle */}
              <Avatar className="h-9 w-9 border border-border">
                <AvatarImage src={`https://api.dicebear.com/7.x/lorelei/svg?seed=${user?.email || "supplier"}`} alt={user?.name} />
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
