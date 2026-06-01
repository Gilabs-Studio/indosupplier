"use client";

import React, { useState, useEffect } from "react";
import Image from "next/image";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";
import LanguageSwitcher from "@/components/navigation/language-switcher";
import { RainbowButton } from "@/components/ui/rainbow-button";
import { cn } from "@/lib/utils";

interface HeaderProps {
  locale: string;
}

export function Header({ locale }: HeaderProps) {
  const t = useTranslations("landing");
  const [isVisible, setIsVisible] = useState(true);
  const [lastScrollY, setLastScrollY] = useState(0);
  const [isAtTop, setIsAtTop] = useState(true);

  useEffect(() => {
    const handleScroll = () => {
      const currentScrollY = window.scrollY;

      // Handle transparent state at top
      setIsAtTop(currentScrollY < 10);

      // Handle hide/show header on scroll
      if (currentScrollY > lastScrollY && currentScrollY > 80) {
        setIsVisible(false); // Scrolling down -> hide
      } else {
        setIsVisible(true); // Scrolling up or near top -> show
      }

      setLastScrollY(currentScrollY);
    };

    window.addEventListener("scroll", handleScroll, { passive: true });
    return () => window.removeEventListener("scroll", handleScroll);
  }, [lastScrollY]);

  return (
    <header
      className={cn(
        "fixed left-0 right-0 top-0 z-50 px-3 pt-3 transition-all duration-300 ease-in-out sm:px-4 md:px-6 lg:px-8",
        isVisible ? "translate-y-0" : "-translate-y-full"
      )}
    >
      <div
        className={cn(
          "mx-auto max-w-[1400px] overflow-hidden transition-all duration-300 px-5 py-3 md:px-6 lg:px-8",
          isAtTop
            ? "border border-transparent bg-transparent shadow-none backdrop-blur-none"
            : "rounded-[1.75rem] border border-border/25 bg-background/58 shadow-[0_20px_48px_-28px_hsl(var(--foreground)/0.48)] backdrop-blur-2xl"
        )}
      >
        {!isAtTop && (
          <div
            className="pointer-events-none absolute inset-0 -z-10"
            style={{
              background:
                "linear-gradient(to bottom, hsl(var(--background) / 0.95) 0%, hsl(var(--background) / 0.72) 38%, hsl(var(--background) / 0.28) 100%)",
            }}
          />
        )}

        <div className="flex items-center justify-between gap-6">
          {/* Brand Logo */}
          <Link href="/" className="flex items-center gap-3 hover:opacity-90 transition-opacity">
            <Image src="/logo.png" alt="IndoSupplier Logo" width={120} height={24} className="h-6 w-auto object-contain" />
            <span className="font-normal text-[15px] tracking-widest uppercase text-foreground">
              IndoSupplier
            </span>
          </Link>

          {/* Center menu links */}
          <div className="hidden md:flex items-center gap-10 text-[12px] tracking-widest uppercase font-semibold text-foreground/70">
            <a href="#join" className="transition-colors hover:text-foreground flex items-center gap-1">
              {locale === "id" ? "Cari Supplier" : "Find Supplier"}
              <svg className="h-3 w-3 opacity-60" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
              </svg>
            </a>
            <a href="#features" className="transition-colors hover:text-foreground">
              {locale === "id" ? "Kategori" : "Categories"}
            </a>
            <a href="#join" className="transition-colors hover:text-foreground flex items-center gap-1">
              {locale === "id" ? "Layanan" : "Services"}
              <svg className="h-3 w-3 opacity-60" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
              </svg>
            </a>
            <a href="#about" className="transition-colors hover:text-foreground">
              {t("nav.about")}
            </a>
            <a href="#join" className="transition-colors hover:text-foreground">
              {locale === "id" ? "Untuk Bisnis" : "For Business"}
            </a>
          </div>

          {/* Right menu actions */}
          <div className="flex items-center gap-4 text-[12px] font-medium tracking-widest uppercase">
            <LanguageSwitcher currentLocale={locale} />
            <RainbowButton asChild size="sm" className="text-[11px] font-semibold tracking-wider uppercase transition-all duration-300 hover:scale-[1.02] active:scale-[0.98]">
              <a href="#join">
                {t("nav.waitlist")}
              </a>
            </RainbowButton>
          </div>
        </div>
      </div>
    </header>
  );
}
