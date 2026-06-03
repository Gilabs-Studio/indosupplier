"use client";

import React from "react";
import Image from "next/image";
import { useTranslations } from "next-intl";
import { Link, usePathname } from "@/i18n/routing";
import LanguageSwitcher from "@/components/navigation/language-switcher";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface PublicNavbarProps {
  locale: string;
}

export function PublicNavbar({ locale }: PublicNavbarProps) {
  const t = useTranslations("public.navbar");
  const pathname = usePathname();

  const navLinks = [
    { href: "/demo/search", label: t("search") },
    { href: "/demo/categories", label: t("categories") },
    { href: "/demo/help", label: t("help") },
  ];

  return (
    <header className="sticky top-0 z-50 w-full border-b border-border bg-background/80 backdrop-blur-md">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
        {/* Brand Logo */}
        <div className="flex items-center gap-8">
          <Link href="/demo" className="flex items-center gap-3 transition-opacity hover:opacity-90">
            <Image
              src="/logo.png"
              alt="IndoSupplier Logo"
              width={110}
              height={22}
              className="h-5.5 w-auto object-contain brightness-0"
            />
            <span className="font-sans text-[15px] font-semibold tracking-wider uppercase text-foreground">
              IndoSupplier
            </span>
          </Link>

          {/* Navigation Links */}
          <nav className="hidden md:flex items-center gap-6">
            {navLinks.map((link) => {
              const isActive = pathname === link.href;
              return (
                <Link
                  key={link.href}
                  href={link.href}
                  className={cn(
                    "text-sm font-medium transition-colors hover:text-primary cursor-pointer",
                    isActive ? "text-primary" : "text-muted-foreground"
                  )}
                >
                  {link.label}
                </Link>
              );
            })}
          </nav>
        </div>

        {/* Right Section: Language Switcher and CTA Buttons */}
        <div className="flex items-center gap-4">
          <LanguageSwitcher currentLocale={locale} />
          
          <div className="hidden sm:flex items-center gap-3">
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
        </div>
      </div>
    </header>
  );
}
