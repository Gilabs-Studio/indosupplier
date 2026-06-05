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
  ShoppingCart,
  Settings,
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
  const buyerProfileT = useTranslations("buyer.profile");
  const router = useRouter();
  const { user, isAuthenticated, logout } = useAuthStore();
  const [searchQuery, setSearchQuery] = useState("");
  const [isCartOpen, setIsCartOpen] = useState(false);
  const [isNotifOpen, setIsNotifOpen] = useState(false);
  const [isProfileOpen, setIsProfileOpen] = useState(false);
  const [activeNotifTab, setActiveNotifTab] = useState<"transaksi" | "update">("transaksi");
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

  const isAnyDropdownOpen = isCartOpen || isNotifOpen || isProfileOpen;

  return (
    <>
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
              className="w-full pl-8.5 pr-4 py-1.5 bg-secondary/80 text-foreground placeholder:text-muted-foreground border border-border/70 rounded-full text-xs outline-hidden focus:ring-1 focus:ring-primary/50 focus:border-primary/50 transition-all cursor-pointer font-normal"
            />
          </form>
        </div>

        {/* Right Section: Icons, Divider, Persona Controls */}
        <div className="flex items-center gap-4 shrink-0">
          <LanguageSwitcher currentLocale={locale} />
          
          {isAuthenticated ? (
            <div className="flex items-center gap-3">
              {/* Message & Notification Icons */}
              <div className="flex items-center gap-1">
                {/* Cart/Keranjang */}
                <div
                  onMouseEnter={() => {
                    setIsCartOpen(true);
                    setIsNotifOpen(false);
                    setIsProfileOpen(false);
                  }}
                  onMouseLeave={() => setIsCartOpen(false)}
                  className="relative"
                >
                  <Button
                    variant="ghost"
                    size="icon"
                    title={t("cart")}
                    className="text-muted-foreground hover:text-foreground hover:bg-secondary cursor-pointer h-9 w-9 rounded-full relative"
                    asChild
                  >
                    <Link href="/cart">
                      <ShoppingCart className="h-5 w-5" />
                      <span className="absolute top-1 right-1 h-4 w-4 flex items-center justify-center rounded-full bg-primary text-[9px] font-bold text-white">
                        1
                      </span>
                    </Link>
                  </Button>

                  {isCartOpen && (
                    <div className="absolute right-0 top-full pt-2 z-50">
                      <div className="w-80 p-4 bg-background border border-border rounded-xl shadow-lg">
                        <div className="flex items-center justify-between border-b border-border/60 pb-2 mb-3">
                          <span className="text-xs font-bold text-foreground">
                            {t("cart")} (1)
                          </span>
                          <Link href="/cart" className="text-xs font-semibold text-primary hover:underline">
                            {t("view")}
                          </Link>
                        </div>
                        <div className="flex gap-3 items-start">
                          <div className="h-12 w-12 rounded-lg border border-border bg-card overflow-hidden shrink-0 flex items-center justify-center text-muted-foreground">
                            <ShoppingCart className="h-5 w-5" />
                          </div>
                          <div className="min-w-0 flex-1">
                            <p className="truncate text-xs font-bold text-foreground">
                              MSI PRO DESKTOP PC XII I7
                            </p>
                            <p className="text-[10px] text-muted-foreground mt-0.5">
                              1 x Rp14.920.000
                            </p>
                          </div>
                        </div>
                      </div>
                    </div>
                  )}
                </div>

                {/* Notifications */}
                <div
                  onMouseEnter={() => {
                    setIsNotifOpen(true);
                    setIsCartOpen(false);
                    setIsProfileOpen(false);
                  }}
                  onMouseLeave={() => setIsNotifOpen(false)}
                  className="relative"
                >
                  <Button
                    variant="ghost"
                    size="icon"
                    title={t("notification")}
                    className="text-muted-foreground hover:text-foreground hover:bg-secondary cursor-pointer h-9 w-9 rounded-full relative"
                    asChild
                  >
                    <Link href="/notifications">
                      <Bell className="h-5 w-5" />
                      <span className="absolute top-1.5 right-1.5 h-2 w-2 rounded-full bg-destructive" />
                    </Link>
                  </Button>

                  {isNotifOpen && (
                    <div className="absolute right-0 top-full pt-2 z-50">
                      <div className="w-96 bg-background border border-border rounded-xl shadow-lg overflow-hidden">
                        {/* Header */}
                        <div className="flex items-center justify-between px-4 py-3 border-b border-border/60 bg-muted/10">
                          <span className="text-sm font-bold text-foreground">
                            {t("notification")}
                          </span>
                          <Button variant="ghost" size="icon" className="h-7 w-7 rounded-full text-muted-foreground hover:text-foreground cursor-pointer">
                            <Settings className="h-4 w-4" />
                          </Button>
                        </div>

                        {/* Tabs */}
                        <div className="flex border-b border-border/60 text-xs font-semibold">
                          <button
                            onClick={() => setActiveNotifTab("transaksi")}
                            className={`flex-1 py-2 text-center border-b-2 transition-all cursor-pointer ${
                              activeNotifTab === "transaksi"
                                ? "border-primary text-primary font-bold"
                                : "border-transparent text-muted-foreground hover:text-foreground"
                            }`}
                          >
                            {t("transaction")}
                          </button>
                          <button
                            onClick={() => setActiveNotifTab("update")}
                            className={`flex-1 py-2 text-center border-b-2 transition-all cursor-pointer ${
                              activeNotifTab === "update"
                                ? "border-primary text-primary font-bold"
                                : "border-transparent text-muted-foreground hover:text-foreground"
                            }`}
                          >
                            {t("update")}
                          </button>
                        </div>

                        {/* Content Area */}
                        <div className="max-h-60 overflow-y-auto divide-y divide-border/40">
                          {activeNotifTab === "transaksi" ? (
                            <>
                              {/* RFQ 1 */}
                              <Link
                                href="/rfq"
                                className="block p-3 hover:bg-secondary/40 transition-colors text-left font-sans"
                              >
                                <p className="text-xs font-bold text-foreground">
                                  RFQ-2026-004 Garnet Sand Mesh 80
                                </p>
                                <p className="text-[10px] text-muted-foreground mt-0.5 leading-relaxed font-light">
                                  {locale === "id"
                                    ? "3 penawaran baru telah masuk dari supplier terverifikasi."
                                    : "3 new quotes received from verified suppliers."}
                                </p>
                                <p className="text-[9px] text-muted-foreground/60 mt-1 font-light">
                                  2026-06-01
                                </p>
                              </Link>

                              {/* RFQ 2 */}
                              <Link
                                href="/rfq"
                                className="block p-3 hover:bg-secondary/40 transition-colors text-left font-sans"
                              >
                                <p className="text-xs font-bold text-foreground">
                                  RFQ-2026-003 Bentonite Clay Powder
                                </p>
                                <p className="text-[10px] text-muted-foreground mt-0.5 leading-relaxed font-light">
                                  {locale === "id"
                                    ? "8 penawaran baru masuk. Bandingkan spesifikasi sekarang."
                                    : "8 new quotes received. Compare specifications now."}
                                </p>
                                <p className="text-[9px] text-muted-foreground/60 mt-1 font-light">
                                  2026-05-28
                                </p>
                              </Link>

                              {/* RFQ 3 */}
                              <Link
                                href="/rfq"
                                className="block p-3 hover:bg-secondary/40 transition-colors text-left font-sans"
                              >
                                <p className="text-xs font-bold text-foreground">
                                  RFQ-2026-002 Quartz Powder 325 Mesh
                                </p>
                                <p className="text-[10px] text-muted-foreground mt-0.5 leading-relaxed font-light">
                                  {locale === "id"
                                    ? "5 penawaran baru masuk."
                                    : "5 new quotes received."}
                                </p>
                                <p className="text-[9px] text-muted-foreground/60 mt-1 font-light">
                                  2026-05-15
                                </p>
                              </Link>
                            </>
                          ) : (
                            <>
                              <div className="p-4 text-center text-xs text-muted-foreground font-light font-sans">
                                {locale === "id"
                                  ? "Tidak ada pembaruan sistem baru."
                                  : "No new system updates."}
                              </div>
                            </>
                          )}
                        </div>

                        {/* Footer */}
                        <div className="flex items-center justify-between px-4 py-2 border-t border-border/60 bg-muted/5 text-[10px] font-bold">
                          <button className="text-muted-foreground hover:text-foreground cursor-pointer">
                            {locale === "id" ? "Tandai semua dibaca" : "Mark all as read"}
                          </button>
                          <Link href="/notifications" className="text-primary hover:underline">
                            {t("viewAll")}
                          </Link>
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              </div>

              {/* Vertical Divider */}
              <div className="hidden sm:block h-5 w-[1px] bg-border" />

              {/* Toko (Supplier Hub) Link */}
              <div className="hidden sm:block">
                <Link
                  href={hasSupplierAccess ? "/supplier/dashboard" : "/supplier/register"}
                  className="flex items-center gap-1.5 text-xs font-semibold text-muted-foreground hover:text-foreground cursor-pointer transition-colors px-2 py-1.5 rounded-lg hover:bg-secondary shrink-0"
                >
                  <Store className="h-4.5 w-4.5 text-muted-foreground" />
                  <span>{t("shop")}</span>
                </Link>
              </div>

              {/* Profile Dropdown ("yohanes" - Buyer Hub) */}
              <div
                onMouseEnter={() => {
                  setIsProfileOpen(true);
                  setIsCartOpen(false);
                  setIsNotifOpen(false);
                }}
                onMouseLeave={() => setIsProfileOpen(false)}
                className="relative"
              >
                <button className="flex items-center gap-1.5 rounded-full focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 p-1 cursor-pointer transition-all hover:bg-secondary">
                  <Avatar className="h-8 w-8 border border-border">
                    <AvatarImage src={`https://api.dicebear.com/7.x/lorelei/svg?seed=${user?.email}`} alt={user?.name} />
                    <AvatarFallback className="bg-primary/10 text-primary font-semibold text-xs">
                      {user?.name?.slice(0, 2).toUpperCase() ?? "US"}
                    </AvatarFallback>
                  </Avatar>
                  <span className="hidden md:inline text-sm font-medium text-foreground max-w-[100px] truncate font-sans">
                    {user?.name}
                  </span>
                  <ChevronDown className="h-3.5 w-3.5 text-muted-foreground" />
                </button>

                {isProfileOpen && (
                  <div className="absolute right-0 top-full pt-2 z-50">
                    <div className="w-[calc(100vw-2rem)] sm:w-[440px] p-0 bg-background border border-border rounded-xl shadow-xl overflow-hidden text-left font-sans">
                      <div className="flex items-center gap-3 p-4 border-b border-border/60 bg-muted/10">
                        <Avatar className="h-10 w-10 border border-border">
                          <AvatarImage src={`https://api.dicebear.com/7.x/lorelei/svg?seed=${user?.email}`} alt={user?.name} />
                          <AvatarFallback className="bg-primary/10 text-primary font-semibold text-sm">
                            {user?.name?.slice(0, 2).toUpperCase() ?? "US"}
                          </AvatarFallback>
                        </Avatar>
                        <div className="min-w-0 flex-1">
                          <p className="truncate text-sm font-bold text-foreground leading-tight font-sans">{user?.name}</p>
                          <p className="truncate text-xs text-muted-foreground font-sans">{user?.email}</p>
                        </div>
                      </div>

                      <div className="grid grid-cols-1 sm:grid-cols-[1.15fr_0.85fr] divide-y sm:divide-y-0 sm:divide-x divide-border/60">
                        {/* Left Column: Account Details & Membership */}
                        <div className="p-4 space-y-4">
                          <div className="rounded-lg border border-border bg-card p-3 shadow-xs space-y-2">
                            <div className="flex items-center justify-between">
                              <span className="text-[9px] font-bold tracking-wider uppercase bg-primary/15 text-primary px-2 py-0.5 rounded-md">
                                {buyerLayoutT("buyerAccount")}
                              </span>
                              <span className="text-[10px] text-muted-foreground font-light">{user?.buyer_profile?.status || "Aktif"}</span>
                            </div>
                            <p className="text-xs font-semibold text-foreground">Akses Premium Sourcing</p>
                            <p className="text-[10px] text-muted-foreground font-light leading-normal">
                              Nikmati akses tanpa batas untuk mengirim RFQ & berdiskusi dengan supplier terverifikasi.
                            </p>
                          </div>

                          <div className="space-y-2.5 pt-1">
                            <div className="flex items-center justify-between text-xs py-0.5 border-b border-border/30 pb-2">
                              <div className="flex items-center gap-2">
                                <span className="h-2 w-2 rounded-full bg-primary" />
                                <span className="text-muted-foreground font-light">GIMS Pay</span>
                              </div>
                              <span className="font-semibold text-foreground">Rp 0</span>
                            </div>

                            <div className="flex items-center justify-between text-xs py-0.5 border-b border-border/30 pb-2">
                              <div className="flex items-center gap-2">
                                <span className="h-2 w-2 rounded-full bg-cyan" />
                                <span className="text-muted-foreground font-light">RFQ Terkirim</span>
                              </div>
                              <span className="font-semibold text-foreground">6 RFQ</span>
                            </div>

                            <div className="flex items-center justify-between text-xs py-0.5">
                              <div className="flex items-center gap-2">
                                <span className="h-2 w-2 rounded-full bg-emerald-500" />
                                <span className="text-muted-foreground font-light">Supplier Saved</span>
                              </div>
                              <span className="font-semibold text-foreground">24 Toko</span>
                            </div>
                          </div>
                        </div>

                        {/* Right Column: Menu Actions */}
                        <div className="p-3 flex flex-col justify-between min-h-[260px]">
                          <div className="space-y-0.5">
                            {[
                              { href: "/dashboard", label: buyerLayoutT("dashboard"), icon: LayoutDashboard },
                              { href: "/rfq", label: buyerLayoutT("rfqList"), icon: RefreshCw },
                              { href: "/bookmarks", label: buyerLayoutT("wishlist"), icon: Heart },
                              { href: "/compare", label: buyerLayoutT("compare"), icon: Scale },
                              { href: "/profile/documents", label: buyerProfileT("tabDocuments"), icon: FileText },
                              { href: "/profile", label: buyerLayoutT("profile"), icon: Settings2 },
                              { href: "/support", label: buyerLayoutT("support"), icon: LifeBuoy }
                            ].map((item) => (
                              <Link
                                key={item.href}
                                href={item.href}
                                className="flex items-center gap-2 w-full px-2.5 py-1.5 text-xs text-foreground font-light hover:bg-secondary rounded-lg transition-colors cursor-pointer"
                              >
                                <item.icon className="h-4 w-4 text-muted-foreground shrink-0" />
                                <span>{item.label}</span>
                              </Link>
                            ))}

                            <div className="block sm:hidden border-t border-border/60 my-1 pt-1">
                              {hasSupplierAccess ? (
                                <Link
                                  href="/supplier/dashboard"
                                  className="flex items-center gap-2 w-full px-2.5 py-1.5 text-xs text-foreground font-light hover:bg-secondary rounded-lg transition-colors cursor-pointer"
                                >
                                  <Store className="h-4 w-4 text-muted-foreground shrink-0" />
                                  <span>Dashboard Toko</span>
                                </Link>
                              ) : (
                                <Link
                                  href="/supplier/register"
                                  className="flex items-center gap-2 w-full px-2.5 py-1.5 text-xs text-primary font-semibold hover:bg-primary/5 rounded-lg transition-colors cursor-pointer"
                                >
                                  <Store className="h-4 w-4 shrink-0" />
                                  <span>Daftar Supplier</span>
                                </Link>
                              )}
                            </div>
                          </div>

                          <div className="border-t border-border/60 pt-2 mt-2">
                            <button
                              onClick={handleLogout}
                              className="w-full text-muted-foreground hover:bg-destructive/10 hover:text-destructive cursor-pointer rounded-lg flex items-center justify-between px-2.5 py-1.5 text-xs font-medium transition-colors"
                            >
                              <span>Keluar</span>
                              <LogOut className="h-4 w-4 shrink-0" />
                            </button>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                )}
              </div>
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
    {/* Dark overlay backdrop */}
    {isAnyDropdownOpen && (
      <div className="fixed inset-0 top-14 bg-black/45 z-40 transition-all duration-200 animate-in fade-in" />
    )}
    </>
  );
}
