"use client";

import React from "react";
import { useLocale, useTranslations } from "next-intl";
import { Link, usePathname } from "@/i18n/routing";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import { PublicNavbar } from "@/features/public/components/public-navbar";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  RefreshCw,
  Heart,
  Columns3,
  User,
  Headset,
  Wallet,
  MessageSquare,
  Star,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useBuyerBookmarks } from "@/features/buyer/bookmarks/hooks/useBuyerBookmarks";

interface BuyerLayoutProps {
  readonly children: React.ReactNode;
}

export function BuyerLayout({ children }: BuyerLayoutProps) {
  const locale = useLocale();
  const pathname = usePathname();
  const { user } = useAuthStore();
  const t = useTranslations("buyer.layout");
  const { bookmarks } = useBuyerBookmarks();

  const menuItems = [
    {
      name: t("transactions"),
      href: "/transactions",
      icon: Wallet,
    },
    {
      name: t("chat"),
      href: "/chat",
      icon: MessageSquare,
    },
    {
      name: t("reviews"),
      href: "/reviews",
      icon: Star,
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
            {/* Sourcing Access & Balances Card */}
            <div className="bg-card rounded-xl border border-border p-4 shadow-xs space-y-4">
              {/* Premium Sourcing Info */}
              <div className="rounded-lg border border-primary/20 bg-primary/5 p-3 space-y-1.5">
                <div className="flex items-center justify-between">
                  <span className="text-[9px] font-bold tracking-wider uppercase bg-primary/15 text-primary px-2 py-0.5 rounded-md">
                    {t("buyerAccount")}
                  </span>
                  <span className="text-[10px] text-muted-foreground font-light font-sans">
                    {user?.buyer_profile?.status === "active" ? t("active") : (user?.buyer_profile?.status || "Aktif")}
                  </span>
                </div>
                <h3 className="text-xs font-semibold text-foreground font-sans">
                  {t("premiumSourcingTitle")}
                </h3>
                <p className="text-[10px] text-muted-foreground font-light leading-normal font-sans">
                  {t("premiumSourcingDesc")}
                </p>
              </div>

              {/* Balances / Stats List */}
              <div className="space-y-2.5">
                <Link
                  href="/transactions"
                  className="flex items-center justify-between text-xs py-1 border-b border-border/30 pb-2 hover:bg-secondary/40 px-1.5 rounded-lg transition-all duration-300 cursor-pointer group hover:-translate-y-0.5 active:translate-y-0"
                >
                  <div className="flex items-center gap-2">
                    <span className="h-2 w-2 rounded-full bg-primary" />
                    <span className="text-muted-foreground font-light font-sans group-hover:text-foreground transition-colors">{t("gimsPay")}</span>
                  </div>
                  <span className="font-semibold text-foreground font-sans">Rp 0</span>
                </Link>

                <Link
                  href="/rfq"
                  className="flex items-center justify-between text-xs py-1 border-b border-border/30 pb-2 hover:bg-secondary/40 px-1.5 rounded-lg transition-all duration-300 cursor-pointer group hover:-translate-y-0.5 active:translate-y-0"
                >
                  <div className="flex items-center gap-2">
                    <span className="h-2 w-2 rounded-full bg-cyan" />
                    <span className="text-muted-foreground font-light font-sans group-hover:text-foreground transition-colors">{t("rfqSent")}</span>
                  </div>
                  <span className="font-semibold text-foreground font-sans">6 RFQ</span>
                </Link>

                <Link
                  href="/bookmarks"
                  className="flex items-center justify-between text-xs py-1 hover:bg-secondary/40 px-1.5 rounded-lg transition-all duration-300 cursor-pointer group hover:-translate-y-0.5 active:translate-y-0"
                >
                  <div className="flex items-center gap-2">
                    <span className="h-2 w-2 rounded-full bg-success" />
                    <span className="text-muted-foreground font-light font-sans group-hover:text-foreground transition-colors">{t("supplierSaved")}</span>
                  </div>
                  <span className="font-semibold text-foreground font-sans">
                    {locale === "id" ? `${bookmarks.length} Toko` : `${bookmarks.length} Saved`}
                  </span>
                </Link>
              </div>
            </div>

            {/* Navigation Menus */}
            <nav className="bg-card rounded-xl border border-border p-2.5 shadow-xs space-y-1">
              {menuItems.map((item) => {
                const IconComponent = item.icon;
                const isActive = pathname === item.href || (item.href !== "/transactions" && pathname.startsWith(item.href));

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
