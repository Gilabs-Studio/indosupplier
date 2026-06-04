"use client";

import React from "react";
import { useLocale, useTranslations } from "next-intl";
import { Link, usePathname } from "@/i18n/routing";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import { PublicNavbar } from "@/features/public/components/public-navbar";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  LayoutDashboard,
  RefreshCw,
  Heart,
  Columns3,
  User,
  Headset,
  Wallet,
} from "lucide-react";
import { cn } from "@/lib/utils";

interface BuyerLayoutProps {
  readonly children: React.ReactNode;
}

export function BuyerLayout({ children }: BuyerLayoutProps) {
  const locale = useLocale();
  const pathname = usePathname();
  const { user } = useAuthStore();
  const t = useTranslations("buyer.layout");

  const menuItems = [
    {
      name: t("dashboard"),
      href: "/dashboard",
      icon: LayoutDashboard,
    },
    {
      name: t("rfqList"),
      href: "/rfq",
      icon: RefreshCw,
    },
    {
      name: t("wishlist"),
      href: "/bookmarks",
      icon: Heart,
    },
    {
      name: t("compare"),
      href: "/compare",
      icon: Columns3,
    },
    {
      name: t("profile"),
      href: "/profile",
      icon: User,
    },
    {
      name: t("support"),
      href: "/support",
      icon: Headset,
    },
  ];

  return (
    <div className="min-h-screen bg-background text-foreground flex flex-col font-sans antialiased">
      {/* Top Header */}
      <PublicNavbar locale={locale} />

      {/* Page Content Layout */}
      <div className="flex-1 w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-[260px_1fr] gap-8 items-start">
          {/* Sidebar */}
          <aside className="space-y-6 lg:sticky lg:top-24">
            {/* Profile Brief Card */}
            <div className="bg-card rounded-xl border border-border p-4 shadow-xs flex items-center gap-3">
              <Avatar className="h-12 w-12 border border-border">
                <AvatarImage src={`https://api.dicebear.com/7.x/lorelei/svg?seed=${user?.email}`} alt={user?.name} />
                <AvatarFallback className="bg-primary/10 text-primary font-semibold text-sm">
                  {user?.name?.slice(0, 2).toUpperCase() ?? "US"}
                </AvatarFallback>
              </Avatar>
              <div className="min-w-0 flex-1">
                <h2 className="text-sm font-semibold text-foreground truncate">{user?.name ?? "Guest User"}</h2>
                <div className="flex items-center gap-1.5 mt-0.5 text-xs text-muted-foreground">
                  <span className="inline-flex h-1.5 w-1.5 rounded-full bg-emerald-500" />
                  <span>{t("buyerAccount")}</span>
                </div>
              </div>
            </div>

            {/* B2B Balance/Credit Card */}
            <div className="bg-linear-to-br from-neutral-800 to-neutral-950 text-white rounded-xl p-4 shadow-xs space-y-3 relative overflow-hidden">
              <div className="absolute right-0 bottom-0 opacity-10 pointer-events-none">
                <Wallet className="h-24 w-24 translate-x-4 translate-y-4" />
              </div>
              <div className="flex items-center gap-2 text-white/85 text-xs">
                <Wallet className="h-4 w-4" />
                <span>{t("creditLimit")}</span>
              </div>
              <div className="space-y-0.5">
                <p className="text-lg font-bold tracking-tight">Rp 50.000.000</p>
                <p className="text-[10px] text-emerald-400 font-medium">{t("active")}</p>
              </div>
            </div>

            {/* Navigation Menus */}
            <nav className="bg-card rounded-xl border border-border p-2.5 shadow-xs space-y-1">
              {menuItems.map((item) => {
                const IconComponent = item.icon;
                const isActive = pathname === item.href || (item.href !== "/dashboard" && pathname.startsWith(item.href));

                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    className={cn(
                      "flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-all duration-300 cursor-pointer hover:-translate-y-0.5 active:translate-y-0",
                      isActive
                        ? "bg-primary text-primary-foreground hover:bg-primary/95 hover:shadow-lg hover:shadow-primary/20"
                        : "text-muted-foreground hover:text-foreground hover:bg-secondary"
                    )}
                  >
                    <IconComponent className="h-4.5 w-4.5 shrink-0" />
                    <span>{item.name}</span>
                  </Link>
                );
              })}
            </nav>
          </aside>

          {/* Main Content Area */}
          <main className="min-w-0">
            {children}
          </main>
        </div>
      </div>
    </div>
  );
}
