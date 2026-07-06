"use client";

import { useEffect, useState } from "react";
import { usePathname, useRouter } from "@/i18n/routing";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import { useTranslations, useLocale } from "next-intl";
import { useSupplierProfile } from "@/features/supplier/profile/hooks/useProfile";
import {
  LayoutDashboard,
  Inbox,
  Building2,
  Megaphone,
  Gavel,
  User,
  Receipt,
  CreditCard,
  Headset,
  Star,
} from "lucide-react";

export function useSupplierLayout() {
  const t = useTranslations("supplier.layout");
  const locale = useLocale();
  const router = useRouter();
  const pathname = usePathname();
  const { user, isAuthenticated, isSessionVerified, setUser, setSessionVerified, logout } = useAuthStore();
  const [mounted, setMounted] = useState(false);
  const [isAuthorizing, setIsAuthorizing] = useState(true);
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
          // Ignore CSRF prefetch failures
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

  const handleLogout = () => {
    setShowLogoutConfirm(false);
    logout();
    router.push("/login");
  };

  return {
    t,
    locale,
    pathname,
    user,
    mounted,
    isAuthorizing,
    showLogoutConfirm,
    setShowLogoutConfirm,
    isVerified,
    menuGroups,
    handleLogout,
  };
}
