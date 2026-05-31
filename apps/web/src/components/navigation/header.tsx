"use client";

import React, { useState, useEffect } from "react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";
import LanguageSwitcher from "@/components/navigation/language-switcher";
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
        "fixed top-0 left-0 right-0 z-50 transition-all duration-300 ease-in-out py-5 px-6 md:px-16 lg:px-24 border-b border-transparent",
        isVisible ? "translate-y-0" : "-translate-y-full",
        !isAtTop && "border-border/10"
      )}
    >
      {/* Progressive Blur Background Layer */}
      <div
        className={cn(
          "absolute inset-0 -z-10 transition-opacity duration-300 pointer-events-none",
          isAtTop ? "opacity-0" : "opacity-100"
        )}
        style={{
          background: "linear-gradient(to bottom, hsl(var(--background) / 0.8) 0%, hsl(var(--background) / 0.3) 70%, transparent 100%)",
          backdropFilter: "blur(12px)",
          WebkitBackdropFilter: "blur(12px)",
          maskImage: "linear-gradient(to bottom, black 65%, transparent 100%)",
          WebkitMaskImage: "linear-gradient(to bottom, black 65%, transparent 100%)",
        }}
      />

      <div className="max-w-[1400px] w-full mx-auto flex items-center justify-between relative z-10">
        {/* Brand Logo */}
        <div className="flex items-center gap-3">
          <img src="/logo.png" alt="IndoSupplier Logo" className="h-6 w-auto object-contain" />
          <span className="font-normal text-[15px] tracking-widest uppercase text-foreground">
            IndoSupplier
          </span>
        </div>

        {/* Center menu links */}
        <div className="hidden md:flex items-center gap-12 text-[12px] tracking-widest uppercase font-light text-muted-foreground">
          <a href="#features" className="hover:text-foreground transition-colors">
            {t("features.badge")}
          </a>
          <a href="#about" className="hover:text-foreground transition-colors">
            {t("nav.about")}
          </a>
          <a href="#join" className="hover:text-foreground transition-colors">
            {t("nav.waitlist")}
          </a>
        </div>

        {/* Right menu actions */}
        <div className="flex items-center gap-6 text-[12px] font-medium tracking-widest uppercase">
          <LanguageSwitcher currentLocale={locale} />
          <Link
            href="/login"
            className="hover:underline text-foreground"
          >
            {t("nav.signIn")}
          </Link>
        </div>
      </div>
    </header>
  );
}
