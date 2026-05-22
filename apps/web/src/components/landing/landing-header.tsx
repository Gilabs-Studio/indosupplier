"use client";

import { useLocale, useTranslations } from "next-intl";
import { Link, usePathname } from "@/i18n/routing";
import { Button } from "@/components/ui/button";
import { ThemeToggleButton } from "@/components/ui/theme-toggle";
import { motion, useScroll, useTransform, AnimatePresence } from "framer-motion";
import { useState } from "react";
import { Menu, X, Globe, ChevronDown } from "lucide-react";
import { cn } from "@/lib/utils";
import { marketingApps } from "@/lib/marketing-apps";

export function LandingHeader() {
  const t = useTranslations("landing.nav");
  const locale = useLocale();
  const pathname = usePathname();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [megaMenuOpen, setMegaMenuOpen] = useState(false);

  const { scrollY } = useScroll();
  const headerBg = useTransform(scrollY, [0, 80], ["rgba(0,0,0,0)", "rgba(0,0,0,0.01)"]);

  const navLinks = [
    { href: "/templates", label: locale === "id" ? "Template" : "Templates" },
    { href: "/about", label: locale === "id" ? "Tentang" : "About" },
    { href: "/pricing", label: t("pricing") },
  ] as const;



  return (
    <motion.header
      style={{ backgroundColor: headerBg }}
      className="fixed top-0 left-0 right-0 z-50 bg-background/95 backdrop-blur-md border-b border-border/50"
      onMouseLeave={() => setMegaMenuOpen(false)}
    >
      <div className="mx-auto flex h-16 max-w-[1400px] items-center justify-between px-6 lg:px-12">
        {/* Logo */}
        <div className="flex items-center gap-8">
          <Link href="/" className="flex items-center gap-2.5 select-none">
            <span className="text-2xl font-bold tracking-tight text-primary">SalesView</span>
          </Link>

          {/* Desktop nav */}
          <nav className="hidden items-center gap-2 md:flex">
            <div 
              className="relative py-4"
              onMouseEnter={() => setMegaMenuOpen(true)}
            >
              <button 
                className={cn(
                  "flex items-center gap-1 rounded-lg px-3 py-2 text-sm font-medium transition-colors cursor-pointer",
                  megaMenuOpen ? "text-primary" : "text-muted-foreground hover:text-foreground"
                )}
              >
                {locale === "id" ? "Aplikasi" : "Apps"}
                <ChevronDown className={cn("h-4 w-4 transition-transform", megaMenuOpen && "rotate-180")} />
              </button>
            </div>
            
            {navLinks.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                className="rounded-lg px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground cursor-pointer"
                onMouseEnter={() => setMegaMenuOpen(false)}
              >
                {link.label}
              </Link>
            ))}
          </nav>
        </div>

        {/* Desktop actions */}
        <div className="hidden items-center gap-4 md:flex" onMouseEnter={() => setMegaMenuOpen(false)}>
          <ThemeToggleButton />

          {/* Language toggle */}
          <Link href={pathname || "/"} locale={locale === "en" ? "id" : "en"} scroll={false}>
            <Button
              variant="ghost"
              size="sm"
              className="h-8 gap-1.5 rounded-full px-3 text-xs font-medium text-muted-foreground hover:text-foreground cursor-pointer"
              aria-label="Switch language"
            >
              <Globe className="h-3.5 w-3.5" />
              {locale === "en" ? "ID" : "EN"}
            </Button>
          </Link>

          <Link href="/login" prefetch={false}>
            <Button
              variant="ghost"
              size="sm"
              className="cursor-pointer rounded-full px-5 text-sm font-medium"
            >
              {locale === "id" ? "Masuk" : "Login"}
            </Button>
          </Link>
          <Link href="/register" prefetch={false}>
            <Button
              size="sm"
              className="cursor-pointer rounded-full px-5 text-sm shadow-sm bg-primary hover:bg-primary/90 text-primary-foreground font-medium"
            >
              {locale === "id" ? "Uji coba gratis" : "Free trial"}
            </Button>
          </Link>
        </div>

        {/* Mobile menu button */}
        <div className="flex items-center gap-2 md:hidden">
          <ThemeToggleButton />
          <Link href={pathname || "/"} locale={locale === "en" ? "id" : "en"} scroll={false}>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8 rounded-full text-muted-foreground cursor-pointer"
              aria-label="Switch language"
            >
              <Globe className="h-4 w-4" />
            </Button>
          </Link>
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 cursor-pointer"
            onClick={() => setMobileOpen((v) => !v)}
            aria-label="Toggle menu"
          >
            {mobileOpen ? <X className="h-4 w-4" /> : <Menu className="h-4 w-4" />}
          </Button>
        </div>
      </div>

      {/* Mega Menu Dropdown */}
      <AnimatePresence>
        {megaMenuOpen && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.2 }}
            className="absolute top-full left-0 right-0 bg-background border-b border-border shadow-lg"
          >
            <div className="mx-auto max-w-[1400px] px-6 lg:px-12 py-8">
              <div className="grid grid-cols-2 md:grid-cols-5 lg:grid-cols-5 gap-8">
                {marketingApps.map((topMenu) => (
                  <div key={topMenu.category.en} className="flex flex-col gap-3">
                    <h3 className="text-sm font-semibold tracking-wider text-primary uppercase border-b border-primary/20 pb-2 mb-2">
                      {locale === 'id' ? topMenu.category.id : topMenu.category.en}
                    </h3>
                    <ul className="flex flex-col gap-2">
                      {topMenu.items.map((child) => (
                        <li key={child.name.en}>
                          <Link 
                            href={child.href} 
                            className="text-sm text-muted-foreground hover:text-primary transition-colors cursor-pointer"
                            onClick={() => setMegaMenuOpen(false)}
                          >
                            {locale === 'id' ? child.name.id : child.name.en}
                          </Link>
                        </li>
                      ))}
                    </ul>
                  </div>
                ))}
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Mobile nav drawer */}
      {mobileOpen && (
        <div className="border-b border-border bg-background/95 backdrop-blur-md md:hidden max-h-[80vh] overflow-y-auto">
          <nav className="mx-auto flex flex-col gap-1 px-6 py-4">
            <div className="py-2">
              <h3 className="text-sm font-semibold tracking-wider text-primary uppercase pb-2">{locale === "id" ? "Aplikasi" : "Apps"}</h3>
              <div className="grid grid-cols-2 gap-4 mt-2">
                {marketingApps.map((topMenu) => (
                  <div key={topMenu.category.en} className="flex flex-col gap-1">
                    <span className="text-xs font-semibold uppercase text-muted-foreground/80 mb-1">
                      {locale === 'id' ? topMenu.category.id : topMenu.category.en}
                    </span>
                    {topMenu.items.slice(0, 3).map((child) => (
                      <Link 
                        key={child.name.en}
                        href={child.href} 
                        className="text-xs text-muted-foreground hover:text-foreground cursor-pointer py-1"
                        onClick={() => setMobileOpen(false)}
                      >
                        {locale === 'id' ? child.name.id : child.name.en}
                      </Link>
                    ))}
                  </div>
                ))}
              </div>
            </div>
            
            <div className="h-px bg-border my-2" />
            
            {navLinks.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                onClick={() => setMobileOpen(false)}
                className="rounded-lg py-2.5 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground cursor-pointer"
              >
                {link.label}
              </Link>
            ))}
            
            <div className="mt-4 pt-4 border-t border-border flex flex-col gap-2">
              <Link href="/login" prefetch={false}>
                <Button variant="outline" className="w-full cursor-pointer rounded-full text-sm font-medium">
                  {locale === "id" ? "Masuk" : "Login"}
                </Button>
              </Link>
              <Link href="/register" prefetch={false}>
                <Button className="w-full cursor-pointer rounded-full bg-primary hover:bg-primary/90 text-primary-foreground text-sm font-medium">
                  {locale === "id" ? "Uji coba gratis" : "Free trial"}
                </Button>
              </Link>
            </div>
          </nav>
        </div>
      )}
    </motion.header>
  );
}
