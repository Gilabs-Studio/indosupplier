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
  User,
  Heart,
  RefreshCw,
  LogOut,
  Bell,
  LayoutDashboard,
  Search,
  Mail,
  Package,
} from "lucide-react";

interface PublicNavbarProps {
  locale: string;
}

export function PublicNavbar({ locale }: PublicNavbarProps) {
  const t = useTranslations("public.navbar");
  const router = useRouter();
  const { user, isAuthenticated, logout } = useAuthStore();
  const [searchQuery, setSearchQuery] = useState("");

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
    <header className="sticky top-0 z-50 w-full border-b border-border bg-background/95 backdrop-blur-md">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8 gap-4">
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

        {/* Search Bar (Tokopedia-style) */}
        <div className="flex-1 max-w-xl">
          <form onSubmit={handleSearchSubmit} className="relative flex items-center w-full">
            <Search className="absolute left-3 h-4 w-4 text-muted-foreground pointer-events-none" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Cari di IndoSupplier..."
              className="w-full pl-9 pr-4 py-2 bg-secondary text-foreground placeholder:text-muted-foreground border border-border rounded-full text-sm outline-hidden focus:ring-1 focus:ring-primary focus:border-primary transition-all cursor-pointer"
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
                <Button variant="ghost" size="icon" className="text-muted-foreground hover:text-foreground cursor-pointer h-9 w-9 rounded-full relative" asChild>
                  <Link href="/rfq">
                    <Mail className="h-5 w-5" />
                    <span className="absolute top-1.5 right-1.5 h-2 w-2 rounded-full bg-primary" />
                  </Link>
                </Button>

                {/* Notifications */}
                <Button variant="ghost" size="icon" className="text-muted-foreground hover:text-foreground cursor-pointer h-9 w-9 rounded-full relative" asChild>
                  <Link href="/notifications">
                    <Bell className="h-5 w-5" />
                    <span className="absolute top-1.5 right-1.5 h-2 w-2 rounded-full bg-destructive" />
                  </Link>
                </Button>
              </div>

              {/* Vertical Divider */}
              <div className="hidden sm:block h-5 w-[1px] bg-border" />

              {/* Toko (Supplier Hub) Link/Dropdown */}
              <div className="hidden sm:block">
                {user?.role === "supplier" ? (
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <button className="flex items-center gap-1.5 text-sm font-medium text-foreground hover:text-primary cursor-pointer transition-colors p-1 rounded-lg">
                        <Store className="h-4.5 w-4.5 text-primary" />
                        <span>Toko</span>
                        <ChevronDown className="h-3.5 w-3.5 text-muted-foreground" />
                      </button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" className="w-56 p-2 bg-background border border-border rounded-xl shadow-lg animate-in fade-in-50 slide-in-from-top-1">
                      <DropdownMenuLabel className="font-bold px-2 py-1 text-[10px] text-muted-foreground uppercase tracking-wider">Supplier Hub</DropdownMenuLabel>
                      <DropdownMenuSeparator className="my-1 border-border" />
                      <DropdownMenuItem asChild className="focus:bg-secondary cursor-pointer rounded-lg">
                        <Link href="/supplier/dashboard" className="flex items-center gap-2 w-full px-2 py-1.5 text-sm">
                          <LayoutDashboard className="h-4 w-4 text-muted-foreground" />
                          <span>Dashboard Toko</span>
                        </Link>
                      </DropdownMenuItem>
                      <DropdownMenuItem asChild className="focus:bg-secondary cursor-pointer rounded-lg">
                        <Link href="/supplier/profile/products" className="flex items-center gap-2 w-full px-2 py-1.5 text-sm">
                          <Package className="h-4 w-4 text-muted-foreground" />
                          <span>Kelola Produk</span>
                        </Link>
                      </DropdownMenuItem>
                      <DropdownMenuItem asChild className="focus:bg-secondary cursor-pointer rounded-lg">
                        <Link href="/supplier/rfq" className="flex items-center gap-2 w-full px-2 py-1.5 text-sm">
                          <RefreshCw className="h-4 w-4 text-muted-foreground" />
                          <span>RFQ Masuk</span>
                        </Link>
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                ) : (
                  <Button
                    asChild
                    variant="outline"
                    size="sm"
                    className="border-primary text-primary hover:bg-primary/5 hover:-translate-y-0.5 active:translate-y-0 shadow-xs transition-all duration-300 cursor-pointer"
                  >
                    <Link href="/supplier/onboarding">Buka Toko Gratis</Link>
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
                <DropdownMenuContent align="end" className="w-64 p-2 bg-background border border-border rounded-xl shadow-lg animate-in fade-in-50 slide-in-from-top-1">
                  <DropdownMenuLabel className="font-normal px-2 py-1.5">
                    <div className="flex flex-col space-y-1">
                      <p className="text-sm font-semibold text-foreground leading-none">{user?.name}</p>
                      <p className="text-xs text-muted-foreground leading-none">{user?.email}</p>
                      <div className="pt-1.5">
                        <span className="inline-flex items-center rounded-full bg-primary/10 px-2 py-0.5 text-[10px] font-semibold text-primary">
                          {user?.role === "supplier" ? "Buyer & Supplier" : "Buyer Account"}
                        </span>
                      </div>
                    </div>
                  </DropdownMenuLabel>
                  
                  <DropdownMenuSeparator className="my-1 border-border" />
                  <div className="px-2 py-1 text-[10px] font-bold text-muted-foreground uppercase tracking-wider">Buyer Hub</div>
                  
                  <DropdownMenuItem asChild className="focus:bg-secondary focus:text-foreground cursor-pointer rounded-lg">
                    <Link href="/rfq" className="flex items-center gap-2 w-full px-2 py-1.5 text-sm">
                      <RefreshCw className="h-4 w-4 text-muted-foreground" />
                      <span>Pembelian (RFQ)</span>
                    </Link>
                  </DropdownMenuItem>
                  <DropdownMenuItem asChild className="focus:bg-secondary focus:text-foreground cursor-pointer rounded-lg">
                    <Link href="/bookmarks" className="flex items-center gap-2 w-full px-2 py-1.5 text-sm">
                      <Heart className="h-4 w-4 text-muted-foreground" />
                      <span>Wishlist</span>
                    </Link>
                  </DropdownMenuItem>
                  <DropdownMenuItem asChild className="focus:bg-secondary focus:text-foreground cursor-pointer rounded-lg">
                    <Link href="/compare" className="flex items-center gap-2 w-full px-2 py-1.5 text-sm">
                      <ChevronDown className="h-4 w-4 rotate-90 text-muted-foreground" />
                      <span>Bandingkan Supplier</span>
                    </Link>
                  </DropdownMenuItem>
                  <DropdownMenuItem asChild className="focus:bg-secondary focus:text-foreground cursor-pointer rounded-lg">
                    <Link href="/profile" className="flex items-center gap-2 w-full px-2 py-1.5 text-sm">
                      <User className="h-4 w-4 text-muted-foreground" />
                      <span>Pengaturan Profil</span>
                    </Link>
                  </DropdownMenuItem>

                  {/* Mobile Supplier Options */}
                  <div className="block sm:hidden">
                    <DropdownMenuSeparator className="my-1 border-border" />
                    <div className="px-2 py-1 text-[10px] font-bold text-muted-foreground uppercase tracking-wider">Supplier Hub</div>
                    {user?.role === "supplier" ? (
                      <>
                        <DropdownMenuItem asChild className="focus:bg-secondary focus:text-foreground cursor-pointer rounded-lg">
                          <Link href="/supplier/dashboard" className="flex items-center gap-2 w-full px-2 py-1.5 text-sm">
                            <LayoutDashboard className="h-4 w-4 text-muted-foreground" />
                            <span>Dashboard Toko</span>
                          </Link>
                        </DropdownMenuItem>
                        <DropdownMenuItem asChild className="focus:bg-secondary focus:text-foreground cursor-pointer rounded-lg">
                          <Link href="/supplier/profile/products" className="flex items-center gap-2 w-full px-2 py-1.5 text-sm">
                            <Package className="h-4 w-4 text-muted-foreground" />
                            <span>Kelola Produk</span>
                          </Link>
                        </DropdownMenuItem>
                      </>
                    ) : (
                      <DropdownMenuItem asChild className="focus:bg-primary/5 focus:text-primary cursor-pointer rounded-lg">
                        <Link href="/supplier/onboarding" className="flex items-center gap-2 w-full px-2 py-1.5 text-sm font-medium text-primary">
                          <Store className="h-4 w-4" />
                          <span>Buka Toko Gratis</span>
                        </Link>
                      </DropdownMenuItem>
                    )}
                  </div>

                  <DropdownMenuSeparator className="my-1 border-border" />
                  <DropdownMenuItem
                    onClick={handleLogout}
                    className="focus:bg-destructive/10 focus:text-destructive text-destructive cursor-pointer rounded-lg flex items-center gap-2 px-2 py-1.5 text-sm"
                  >
                    <LogOut className="h-4 w-4" />
                    <span>Log Out</span>
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          ) : (
            <div className="flex items-center gap-3">
              <Button asChild variant="ghost" size="sm" className="text-muted-foreground hover:text-foreground cursor-pointer">
                <Link href="/login">{t("signIn")}</Link>
              </Button>
              
              <Button
                asChild
                size="sm"
                className="bg-primary text-primary-foreground hover:bg-primary/90 hover:-translate-y-0.5 active:translate-y-0 shadow-xs transition-all duration-300 cursor-pointer"
              >
                <Link href="/demo/register">{t("register")}</Link>
              </Button>
            </div>
          )}
        </div>
      </div>
    </header>
  );
}
