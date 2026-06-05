"use client";

import React, { useState } from "react";
import Image from "next/image";
import { useTranslations } from "next-intl";
import { Link, useRouter } from "@/i18n/routing";
import LanguageSwitcher from "@/components/navigation/language-switcher";
import { Button } from "@/components/ui/button";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  ChevronDown,
  Store,
  LogOut,
  Bell,
  LayoutDashboard,
  Search,
  Mail,
  FileText,
  Heart,
  LifeBuoy,
  RefreshCw,
  Scale,
  Settings2,
} from "lucide-react";

interface PublicNavbarProps {
  locale: string;
}

export function PublicNavbar({ locale }: Readonly<PublicNavbarProps>) {
  const t = useTranslations("public.navbar");
  const buyerLayoutT = useTranslations("buyer.layout");
  const buyerNotificationsT = useTranslations("buyer.notifications");
  const buyerProfileT = useTranslations("buyer.profile");
  const router = useRouter();
  const { user, isAuthenticated, logout } = useAuthStore();
  const [searchQuery, setSearchQuery] = useState("");
  const hasSupplierAccess =
    user?.capabilities.supplier === true || !!user?.supplier_profile;

  const handleLogout = async () => {
    try {
      const { authService } = await import("@/features/auth/services/auth-service");
      const { fullAuthCleanup } = await import("@/features/auth/utils/clear-auth-cookies");
      await authService.logout();
      await fullAuthCleanup();
      logout();
    } catch {
      logout();
    }
  };

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (searchQuery.trim()) {
      router.push(`/search?query=${encodeURIComponent(searchQuery.trim())}`);
    }
  };

  return (
    <header className="sticky top-0 z-50 w-full border-b border-border/40 bg-background/80 backdrop-blur-md">
      <div className="mx-auto flex h-14 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8 gap-4">
        {/* Brand Logo & Kategori */}
        <div className="flex items-center gap-6 shrink-0">
          <Link href="/" className="flex items-center gap-3 transition-opacity hover:opacity-90">
            <Image
              src="/logo.png"
              alt="IndoSupplier Logo"
              width={110}
              height={22}
              className="h-5.5 w-auto object-contain brightness-0"
            />
            <span className="font-sans text-[15px] font-semibold tracking-wider uppercase text-foreground hidden sm:inline">
              IndoSupplier
            </span>
          </Link>

          {/* Categories Dropdown */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button className="hidden md:flex items-center gap-1 text-sm font-medium text-muted-foreground hover:text-foreground cursor-pointer transition-colors">
                Kategori
                <ChevronDown className="h-3.5 w-3.5" />
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start" className="w-56 p-2 bg-background border border-border rounded-xl shadow-lg animate-in fade-in-50 slide-in-from-top-1">
              <DropdownMenuItem asChild className="focus:bg-secondary cursor-pointer rounded-lg">
                <Link href="/search?category=manufacturing" className="w-full px-2 py-1.5 text-sm">Manufaktur</Link>
              </DropdownMenuItem>
              <DropdownMenuItem asChild className="focus:bg-secondary cursor-pointer rounded-lg">
                <Link href="/search?category=agriculture" className="w-full px-2 py-1.5 text-sm">Pertanian</Link>
              </DropdownMenuItem>
              <DropdownMenuItem asChild className="focus:bg-secondary cursor-pointer rounded-lg">
                <Link href="/search?category=textile" className="w-full px-2 py-1.5 text-sm">Tekstil</Link>
              </DropdownMenuItem>
              <DropdownMenuItem asChild className="focus:bg-secondary cursor-pointer rounded-lg">
                <Link href="/search?category=furniture" className="w-full px-2 py-1.5 text-sm">Furnitur</Link>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        {/* Search Bar (Modern Integrated Style) */}
        <div className="flex-1 max-w-sm hidden md:block">
          <form onSubmit={handleSearchSubmit} className="relative flex items-center w-full">
            <Search className="absolute left-3 h-3.5 w-3.5 text-muted-foreground pointer-events-none" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Cari di IndoSupplier..."
              className="w-full pl-8.5 pr-4 py-1.5 bg-secondary/60 text-foreground placeholder:text-muted-foreground/50 border border-border/70 rounded-full text-[11px] outline-hidden focus:ring-1 focus:ring-primary/50 focus:border-primary/50 transition-all cursor-pointer font-light"
            />
          </form>
        </div>

        {/* Right Section: Icons, Divider, Persona Controls */}
        <div className="flex items-center gap-4 shrink-0">
          <LanguageSwitcher currentLocale={locale} />
          
          {isAuthenticated ? (
            <div className="flex items-center gap-3">
              {/* Message & Notification Icons */}
              <div className="flex items-center gap-1.5">
                {/* Inbox/Messages */}
                <Button
                  variant="ghost"
                  size="icon"
                  title={buyerLayoutT("rfqList")}
                  className="text-muted-foreground hover:text-foreground cursor-pointer h-9 w-9 rounded-full relative"
                  asChild
                >
                  <Link href="/rfq">
                    <Mail className="h-5 w-5" />
                    <span className="absolute top-1.5 right-1.5 h-2 w-2 rounded-full bg-primary" />
                  </Link>
                </Button>

                {/* Notifications */}
                <Button
                  variant="ghost"
                  size="icon"
                  title={buyerNotificationsT("title")}
                  className="text-muted-foreground hover:text-foreground cursor-pointer h-9 w-9 rounded-full relative"
                  asChild
                >
                  <Link href="/notifications">
                    <Bell className="h-5 w-5" />
                    <span className="absolute top-1.5 right-1.5 h-2 w-2 rounded-full bg-destructive" />
                  </Link>
                </Button>
              </div>

              {/* Vertical Divider */}
              <div className="hidden sm:block h-5 w-[1px] bg-border" />

              {/* Toko (Supplier Hub) Link */}
              <div className="hidden sm:block">
                {hasSupplierAccess ? (
                  <Button
                    asChild
                    variant="outline"
                    size="sm"
                    className="border-primary text-primary hover:bg-primary/5 hover:-translate-y-0.5 active:translate-y-0 shadow-xs transition-all duration-300 cursor-pointer"
                  >
                    <Link href="/supplier/dashboard">Dashboard Toko</Link>
                  </Button>
                ) : (
                  <Button
                    asChild
                    variant="outline"
                    size="sm"
                    className="border-primary text-primary hover:bg-primary/5 hover:-translate-y-0.5 active:translate-y-0 shadow-xs transition-all duration-300 cursor-pointer"
                  >
                    <Link href="/supplier/register">Daftar Supplier</Link>
                  </Button>
                )}
              </div>

              {/* Profile Dropdown ("yohanes" - Buyer Hub) */}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button className="flex items-center gap-1.5 rounded-full focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 p-1 cursor-pointer transition-all hover:bg-secondary">
                    <Avatar className="h-8 w-8 border border-border">
                      <AvatarImage src={`https://api.dicebear.com/7.x/lorelei/svg?seed=${user?.email}`} alt={user?.name} />
                      <AvatarFallback className="bg-primary/10 text-primary font-semibold text-xs">
                        {user?.name?.slice(0, 2).toUpperCase() ?? "US"}
                      </AvatarFallback>
                    </Avatar>
                    <span className="hidden md:inline text-sm font-medium text-foreground max-w-[100px] truncate">
                      {user?.name}
                    </span>
                    <ChevronDown className="h-3.5 w-3.5 text-muted-foreground" />
                  </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-[calc(100vw-2rem)] max-w-[420px] p-0 bg-background border border-border rounded-xl shadow-xl animate-in fade-in-50 slide-in-from-top-1 overflow-hidden">
                  <DropdownMenuLabel className="font-normal p-4">
                    <div className="flex items-center gap-3 rounded-lg border border-border bg-card px-3 py-3 shadow-xs">
                      <Avatar className="h-12 w-12 border border-border">
                        <AvatarImage src={`https://api.dicebear.com/7.x/lorelei/svg?seed=${user?.email}`} alt={user?.name} />
                        <AvatarFallback className="bg-primary/10 text-primary font-semibold text-sm">
                          {user?.name?.slice(0, 2).toUpperCase() ?? "US"}
                        </AvatarFallback>
                      </Avatar>
                      <div className="min-w-0 flex-1">
                        <p className="truncate text-base font-bold text-foreground">{user?.name}</p>
                        <p className="truncate text-xs text-muted-foreground">{user?.email}</p>
                      </div>
                    </div>
                  </DropdownMenuLabel>

                  <div className="grid grid-cols-1 gap-0 px-4 pb-4 sm:grid-cols-[1.1fr_0.9fr]">
                    <div className="space-y-3 border-b border-border pb-3 sm:border-b-0 sm:border-r sm:pb-0 sm:pr-4">
                      <div className="rounded-lg border border-border bg-card px-3 py-3 text-sm">
                        <p className="font-semibold text-foreground">{buyerLayoutT("buyerAccount")}</p>
                        <p className="mt-1 text-xs text-muted-foreground">
                          {user?.buyer_profile?.status || user?.email}
                        </p>
                        <div className="mt-3 grid grid-cols-2 gap-2 text-xs">
                          <Link href="/dashboard" className="rounded-lg border border-border px-3 py-2 text-muted-foreground hover:bg-secondary">
                            {buyerLayoutT("dashboard")}
                          </Link>
                          <Link href="/notifications" className="rounded-lg border border-border px-3 py-2 text-muted-foreground hover:bg-secondary">
                            {buyerNotificationsT("title")}
                          </Link>
                        </div>
                      </div>
                    </div>

                    <div className="space-y-1 pt-3 sm:pl-4 sm:pt-0">
                      <DropdownMenuItem asChild className="focus:bg-secondary focus:text-foreground cursor-pointer rounded-lg">
                        <Link href="/dashboard" className="flex items-center gap-2 w-full px-2 py-2 text-sm">
                          <LayoutDashboard className="h-4 w-4 text-muted-foreground" />
                          <span>{buyerLayoutT("dashboard")}</span>
                        </Link>
                      </DropdownMenuItem>
                      <DropdownMenuItem asChild className="focus:bg-secondary focus:text-foreground cursor-pointer rounded-lg">
                        <Link href="/rfq" className="flex items-center gap-2 w-full px-2 py-2 text-sm">
                          <RefreshCw className="h-4 w-4 text-muted-foreground" />
                          <span>{buyerLayoutT("rfqList")}</span>
                        </Link>
                      </DropdownMenuItem>
                      <DropdownMenuItem asChild className="focus:bg-secondary focus:text-foreground cursor-pointer rounded-lg">
                        <Link href="/bookmarks" className="flex items-center gap-2 w-full px-2 py-2 text-sm">
                          <Heart className="h-4 w-4 text-muted-foreground" />
                          <span>{buyerLayoutT("wishlist")}</span>
                        </Link>
                      </DropdownMenuItem>
                      <DropdownMenuItem asChild className="focus:bg-secondary focus:text-foreground cursor-pointer rounded-lg">
                        <Link href="/compare" className="flex items-center gap-2 w-full px-2 py-2 text-sm">
                          <Scale className="h-4 w-4 text-muted-foreground" />
                          <span>{buyerLayoutT("compare")}</span>
                        </Link>
                      </DropdownMenuItem>
                      <DropdownMenuItem asChild className="focus:bg-secondary focus:text-foreground cursor-pointer rounded-lg">
                        <Link href="/profile/documents" className="flex items-center gap-2 w-full px-2 py-2 text-sm">
                          <FileText className="h-4 w-4 text-muted-foreground" />
                          <span>{buyerProfileT("tabDocuments")}</span>
                        </Link>
                      </DropdownMenuItem>
                      <DropdownMenuItem asChild className="focus:bg-secondary focus:text-foreground cursor-pointer rounded-lg">
                        <Link href="/profile" className="flex items-center gap-2 w-full px-2 py-2 text-sm">
                          <Settings2 className="h-4 w-4 text-muted-foreground" />
                          <span>{buyerLayoutT("profile")}</span>
                        </Link>
                      </DropdownMenuItem>
                      <DropdownMenuItem asChild className="focus:bg-secondary focus:text-foreground cursor-pointer rounded-lg">
                        <Link href="/support" className="flex items-center gap-2 w-full px-2 py-2 text-sm">
                          <LifeBuoy className="h-4 w-4 text-muted-foreground" />
                          <span>{buyerLayoutT("support")}</span>
                        </Link>
                      </DropdownMenuItem>

                      <div className="block sm:hidden">
                        <DropdownMenuSeparator className="my-1 border-border" />
                        {hasSupplierAccess ? (
                          <DropdownMenuItem asChild className="focus:bg-secondary focus:text-foreground cursor-pointer rounded-lg">
                            <Link href="/supplier/dashboard" className="flex items-center gap-2 w-full px-2 py-2 text-sm">
                              <Store className="h-4 w-4 text-muted-foreground" />
                              <span>Dashboard Toko</span>
                            </Link>
                          </DropdownMenuItem>
                        ) : (
                          <DropdownMenuItem asChild className="focus:bg-primary/5 focus:text-primary cursor-pointer rounded-lg">
                            <Link href="/supplier/register" className="flex items-center gap-2 w-full px-2 py-2 text-sm font-medium text-primary">
                              <Store className="h-4 w-4" />
                              <span>Daftar Supplier</span>
                            </Link>
                          </DropdownMenuItem>
                        )}
                      </div>

                      <DropdownMenuSeparator className="my-2 border-border" />
                      <DropdownMenuItem
                        onClick={handleLogout}
                        className="focus:bg-destructive/10 focus:text-destructive text-muted-foreground cursor-pointer rounded-lg flex items-center gap-2 px-2 py-2 text-sm"
                      >
                        <LogOut className="h-4 w-4" />
                        <span>Keluar</span>
                      </DropdownMenuItem>
                    </div>
                  </div>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          ) : (
            <div className="flex items-center gap-3">
              <Link
                href="/login"
                className="text-xs font-semibold text-muted-foreground hover:text-foreground transition-all cursor-pointer px-2.5 py-1.5 rounded-lg hover:bg-secondary"
              >
                {t("signIn")}
              </Link>
              
              <Link
                href="/register"
                className="bg-neutral-900 hover:bg-neutral-800 text-white dark:bg-white dark:hover:bg-neutral-100 dark:text-neutral-900 shadow-sm border border-neutral-800/10 dark:border-neutral-200/10 hover:-translate-y-0.5 active:translate-y-0 active:scale-95 transition-all text-xs font-semibold px-4.5 py-1.5 rounded-lg cursor-pointer"
              >
                {t("register")}
              </Link>
            </div>
          )}
        </div>
      </div>
    </header>
  );
}
