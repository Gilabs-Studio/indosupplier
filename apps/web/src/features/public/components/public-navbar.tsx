"use client";
/* eslint-disable @next/next/no-img-element */

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
  Search,
  Settings,
  FileText,
  Heart,
  LifeBuoy,
  RefreshCw,
  Scale,
  Settings2,
  Wallet,
  Package,
  UserPlus,
} from "lucide-react";
import { useBuyerBookmarks } from "@/features/buyer/bookmarks/hooks/useBuyerBookmarks";
import { useBuyerFollowing } from "@/features/buyer/following/hooks/useBuyerFollowing";
import { getDicebearUrl } from "@/lib/utils";

interface PublicNavbarProps {
  locale: string;
}

export function PublicNavbar({ locale }: Readonly<PublicNavbarProps>) {
  const t = useTranslations("public.navbar");
  const buyerLayoutT = useTranslations("buyer.layout");
  const buyerProfileT = useTranslations("buyer.profile");
  const router = useRouter();
  const { user, isAuthenticated, logout } = useAuthStore();
  const { bookmarks } = useBuyerBookmarks();
  const { following } = useBuyerFollowing();
  const productBookmarks = bookmarks.filter((item) => item.type === "product" || Boolean(item.supplierProductId));
  const [searchQuery, setSearchQuery] = useState("");
  const [isSavedOpen, setIsSavedOpen] = useState(false);
  const [isHeartHovered, setIsHeartHovered] = useState(false);
  const [isNotifOpen, setIsNotifOpen] = useState(false);
  const [isProfileOpen, setIsProfileOpen] = useState(false);
  const [activeNotifTab, setActiveNotifTab] = useState<"transaksi" | "update">("transaksi");
  const hasSupplierAccess =
    user?.capabilities.supplier === true || !!user?.supplier_profile;

  const closeAllDropdowns = () => {
    setIsSavedOpen(false);
    setIsNotifOpen(false);
    setIsProfileOpen(false);
  };

  const handleLogout = async () => {
    closeAllDropdowns();

    try {
      const { authService } = await import("@/features/auth/services/auth-service");
      const { fullAuthCleanup } = await import("@/features/auth/utils/clear-auth-cookies");
      await authService.logout();
      await fullAuthCleanup();
      logout();
    } catch {
      logout();
    } finally {
      closeAllDropdowns();
    }
  };

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (searchQuery.trim()) {
      router.push(`/search?query=${encodeURIComponent(searchQuery.trim())}`);
    }
  };

  const isAnyDropdownOpen = isSavedOpen || isNotifOpen || isProfileOpen;

function resolveBookmarkThumbnail(item: {
  productImage?: string;
  productName?: string;
  companyName?: string;
  category?: string;
}): string {
  let img = item.productImage?.trim() || "";
  if (img) {
    if (img.startsWith("/images/") && img.endsWith(".png")) {
      img = img.replace(/\.png$/, ".webp");
    }
    return img;
  }

  const combined = `${item.productName || ""} ${item.companyName || ""} ${item.category || ""}`.toLowerCase();
  if (
    combined.includes("steel") ||
    combined.includes("baja") ||
    combined.includes("rebar") ||
    combined.includes("plate") ||
    combined.includes("denim") ||
    combined.includes("yarn") ||
    combined.includes("fiber")
  ) {
    return "/images/categories/cat-bahan-baku.webp";
  }
  if (
    combined.includes("coffee") ||
    combined.includes("kopi") ||
    combined.includes("sugar") ||
    combined.includes("gula") ||
    combined.includes("makanan") ||
    combined.includes("ginger")
  ) {
    return "/images/categories/cat-makanan-minuman.webp";
  }
  if (
    combined.includes("sand") ||
    combined.includes("powder") ||
    combined.includes("bentonite") ||
    combined.includes("garnet")
  ) {
    return "/images/products/prod-mineral-powder.webp";
  }
  if (combined.includes("masker") || combined.includes("medis")) {
    return "/images/products/prod-masker.webp";
  }
  if (combined.includes("helm") || combined.includes("safety") || combined.includes("k3")) {
    return "/images/products/prod-helmet.webp";
  }
  if (combined.includes("laptop") || combined.includes("elektronik") || combined.includes("computer")) {
    return "/images/products/prod-laptop.webp";
  }
  if (combined.includes("pompa") || combined.includes("pump") || combined.includes("mesin")) {
    return "/images/products/prod-water-pump.webp";
  }
  if (combined.includes("kursi") || combined.includes("chair") || combined.includes("furniture")) {
    return "/images/products/prod-office-chair.webp";
  }
  if (combined.includes("karton") || combined.includes("box") || combined.includes("packaging")) {
    return "/images/products/prod-carton-boxes.webp";
  }
  return "/images/categories/cat-bahan-baku.webp";
}

  const renderSavedDropdown = () => (
    <div
      onMouseEnter={() => {
        setIsSavedOpen(true);
        setIsNotifOpen(false);
        setIsProfileOpen(false);
      }}
      onMouseLeave={() => {
        setIsSavedOpen(false);
        setIsHeartHovered(false);
      }}
      className="relative"
    >
      <div
        onMouseEnter={() => setIsHeartHovered(true)}
        onMouseLeave={() => setIsHeartHovered(false)}
        className="relative inline-flex items-center justify-center"
      >
        <Button
          variant="ghost"
          size="icon"
          aria-label={buyerLayoutT("wishlist") || "Disimpan"}
          className="text-muted-foreground hover:text-foreground hover:bg-secondary cursor-pointer h-9 w-9 rounded-full relative overflow-visible"
          asChild
        >
          <Link href="/bookmarks" className="overflow-visible flex items-center justify-center">
            <Heart className="h-5 w-5" />
          </Link>
        </Button>

        {productBookmarks.length > 0 && (
          <span className="absolute -top-1 -right-1 h-4.5 min-w-4.5 px-1 flex items-center justify-center rounded-full bg-amber-500 text-[10px] font-bold text-white shadow-xs leading-none pointer-events-none z-10 select-none">
            {productBookmarks.length}
          </span>
        )}

        {/* Dark Tooltip "Disimpan" matching user screenshot */}
        {isHeartHovered && !isSavedOpen && (
          <div className="absolute top-full left-1/2 -translate-x-1/2 mt-1.5 px-2.5 py-1 bg-neutral-900 text-white text-[11px] font-medium rounded-md shadow-lg pointer-events-none whitespace-nowrap z-50 animate-in fade-in duration-150">
            Disimpan
            <div className="absolute -top-1 left-1/2 -translate-x-1/2 border-4 border-transparent border-b-neutral-900" />
          </div>
        )}
      </div>

      {isSavedOpen && (
        <div className="absolute right-0 top-full pt-2 z-50">
          <div className="w-84 p-4 bg-background border border-border rounded-xl shadow-xl animate-in fade-in slide-in-from-top-1 duration-150 text-left font-sans">
            <div className="flex items-center justify-between border-b border-border/60 pb-2.5 mb-3">
              <span className="text-xs font-bold text-foreground">
                {buyerLayoutT("wishlist")} ({productBookmarks.length})
              </span>
              <Link
                href="/bookmarks"
                onClick={() => setIsSavedOpen(false)}
                className="text-xs font-semibold text-amber-600 dark:text-amber-400 hover:text-amber-700 hover:underline cursor-pointer"
              >
                {t("view")}
              </Link>
            </div>
            {productBookmarks.length === 0 ? (
              <div className="text-center py-5 text-xs text-muted-foreground">
                {locale === "id" ? "Belum ada produk disimpan" : "No saved products yet"}
              </div>
            ) : (
              <div className="space-y-2.5 max-h-72 overflow-y-auto pr-0.5">
                {productBookmarks.slice(0, 6).map((item) => {
                  const detailUrl = item.supplierProductId
                    ? `/demo/products/${item.supplierProductId}`
                    : item.supplierSlug
                    ? `/demo/suppliers/${item.supplierSlug}`
                    : "/bookmarks";

                  const thumbUrl = resolveBookmarkThumbnail(item);

                  return (
                    <Link
                      key={item.id}
                      href={detailUrl}
                      onClick={() => setIsSavedOpen(false)}
                      className="flex gap-3 items-center border-b border-border/30 pb-2.5 last:border-0 last:pb-0 hover:bg-secondary/60 p-1.5 rounded-lg transition-all duration-200 cursor-pointer block group"
                    >
                      <div className="h-10 w-10 rounded-lg bg-muted border border-border/80 flex items-center justify-center shrink-0 overflow-hidden relative">
                        <img
                          src={thumbUrl}
                          alt=""
                          className="object-cover w-full h-full group-hover:scale-105 transition-transform duration-200"
                          onError={(e) => {
                            const target = e.currentTarget;
                            if (target.src !== "/images/categories/cat-bahan-baku.webp") {
                              target.src = "/images/categories/cat-bahan-baku.webp";
                            }
                          }}
                        />
                      </div>
                      <div className="min-w-0 flex-1">
                        <p className="truncate text-xs font-bold text-foreground leading-tight group-hover:text-primary transition-colors">
                          {item.productName || item.companyName}
                        </p>
                        <p className="text-[11px] text-muted-foreground truncate mt-0.5">
                          {item.companyName || (item.location || "Indonesia")}
                        </p>
                      </div>
                    </Link>
                  );
                })}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );

  return (
    <>
      <header className="sticky top-0 z-50 w-full border-b border-border/60 bg-background/95 backdrop-blur-md shadow-xs">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8 gap-3 sm:gap-6">
        {/* Brand Logo & Kategori */}
        <div className="flex items-center gap-4 sm:gap-6 shrink-0">
          <Link href="/" className="flex items-center gap-2.5 transition-opacity hover:opacity-90">
            <Image
              src="/logo.png"
              alt="IndoSupplier Logo"
              width={110}
              height={22}
              className="h-6 w-auto object-contain brightness-0"
            />
            <span className="font-sans text-[15px] font-black tracking-wider uppercase text-foreground hidden sm:inline">
              IndoSupplier
            </span>
          </Link>

          {/* Categories Dropdown */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button className="hidden md:flex items-center gap-1 text-xs font-semibold text-muted-foreground hover:text-foreground cursor-pointer transition-colors">
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

        {/* Search Bar (Prominent Wide Tokopedia/Shopee Style) */}
        <div className="flex-1 max-w-3xl mx-2 sm:mx-4">
          <form onSubmit={handleSearchSubmit} className="relative flex items-center w-full">
            <Search className="absolute left-3.5 h-4 w-4 text-muted-foreground pointer-events-none" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Cari di IndoSupplier..."
              className="w-full h-10 pl-10 pr-4 bg-secondary/80 text-foreground placeholder:text-muted-foreground/70 border border-border/80 rounded-lg text-xs sm:text-sm outline-hidden focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all font-normal"
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
                {renderSavedDropdown()}

                {/* Notifications */}
                <div
                  onMouseEnter={() => {
                    setIsNotifOpen(true);
                    setIsSavedOpen(false);
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
                  setIsSavedOpen(false);
                  setIsNotifOpen(false);
                }}
                onMouseLeave={() => setIsProfileOpen(false)}
                className="relative"
              >
                <button className="flex items-center gap-1.5 rounded-full focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 p-1 cursor-pointer transition-all hover:bg-secondary">
                  <Avatar className="h-8 w-8 border border-border">
                    <AvatarImage src={getDicebearUrl(user?.email || "", "lorelei")} alt={user?.name} />
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
                          <AvatarImage src={getDicebearUrl(user?.email || "", "lorelei")} alt={user?.name} />
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
                                <span className="h-2 w-2 rounded-full bg-success" />
                                <span className="text-muted-foreground font-light">{buyerLayoutT("following")}</span>
                              </div>
                              <span className="font-semibold text-foreground">
                                {locale === "id" ? `${following.length} Toko` : `${following.length} Shops`}
                              </span>
                            </div>
                          </div>
                        </div>

                        {/* Right Column: Menu Actions */}
                        <div className="p-3 flex flex-col justify-between min-h-[260px]">
                          <div className="space-y-0.5">
                            {[
                              { href: "/transactions", label: buyerLayoutT("transactions"), icon: Wallet },
                              { href: "/rfq", label: buyerLayoutT("rfqList"), icon: RefreshCw },
                              { href: "/following", label: buyerLayoutT("following"), icon: UserPlus },
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
              {renderSavedDropdown()}
              <Link
                href="/login"
                className="text-xs font-semibold text-muted-foreground hover:text-foreground transition-all cursor-pointer px-2.5 py-1.5 rounded-lg hover:bg-secondary"
              >
                {t("signIn")}
              </Link>
              
              <Link
                href="/register"
                className="bg-foreground hover:bg-foreground/90 text-background shadow-xs border border-border/10 hover:-translate-y-0.5 active:translate-y-0 active:scale-95 transition-all text-xs font-semibold px-4.5 py-1.5 rounded-lg cursor-pointer"
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
