"use client";

import React, { useState, useEffect } from "react";
import Image from "next/image";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";
import LanguageSwitcher from "@/components/navigation/language-switcher";
import { Button } from "@/components/ui/button";
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
        "fixed left-0 right-0 top-0 z-50 w-full transition-all duration-300 ease-in-out border-b",
        isAtTop
          ? "border-transparent bg-transparent"
          : "border-[#333333]/40 bg-[#1a1a1a]/95 backdrop-blur-md",
        isVisible ? "translate-y-0" : "-translate-y-full"
      )}
    >
      <div className="mx-auto max-w-[1400px] px-6 py-4 flex items-center justify-between gap-6">
        {/* Brand Logo */}
        <Link href="/" className="flex items-center gap-3 hover:opacity-90 transition-opacity">
          <Image src="/logo.png" alt="IndoSupplier Logo" width={120} height={24} className="h-6 w-auto object-contain" />
          <span className="font-sans text-[15px] font-semibold tracking-widest uppercase text-[#ffffff]">
            IndoSupplier
          </span>
        </Link>

        {/* Center menu links */}
        <div className="hidden md:flex items-center gap-10 text-[13px] tracking-wider uppercase font-medium text-[#ffffff]/70">
          <a href="#join" className="transition-colors hover:text-[#ffffff] flex items-center gap-1">
            {locale === "id" ? "Cari Supplier" : "Find Supplier"}
            <svg className="h-3 w-3 opacity-60" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
            </svg>
          </a>
          <a href="#features" className="transition-colors hover:text-[#ffffff]">
            {locale === "id" ? "Kategori" : "Categories"}
          </a>
          <a href="#join" className="transition-colors hover:text-[#ffffff] flex items-center gap-1">
            {locale === "id" ? "Layanan" : "Services"}
            <svg className="h-3 w-3 opacity-60" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
            </svg>
          </a>
          <a href="#about" className="transition-colors hover:text-[#ffffff]">
            {t("nav.about")}
          </a>
          <a href="#join" className="transition-colors hover:text-[#ffffff]">
            {locale === "id" ? "Untuk Bisnis" : "For Business"}
          </a>
        </div>

        {/* Right menu actions */}
        <div className="flex items-center gap-4 text-[13px] font-medium tracking-wider uppercase">
          <LanguageSwitcher currentLocale={locale} />
          <Button
            asChild
            size="sm"
            className="px-5 py-2.5 bg-white hover:bg-neutral-200 text-neutral-900 border border-neutral-200 rounded-none text-[12px] font-semibold tracking-wider transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 cursor-pointer"
          >
            <a href="#join">
              {t("nav.waitlist")}
            </a>
          </Button>
        </div>
      </div>
    </header>
  );
}
