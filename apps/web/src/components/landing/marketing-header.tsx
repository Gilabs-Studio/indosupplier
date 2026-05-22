"use client";

import { Link } from "@/i18n/routing";
import { Button } from "@/components/ui/button";
import { motion } from "framer-motion";

export function MarketingHeader() {
  const navItems = [
    { label: "Fitur", href: "#features" },
    { label: "Modul", href: "#modules" },
    { label: "Pricing", href: "#pricing" },
    { label: "Scenario", href: "#scenarios" },
  ];

  return (
    <motion.header
      initial={{ opacity: 0, y: -8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, ease: "easeOut" }}
      className="sticky top-0 z-50 w-full border-b border-border/50 bg-background/80 backdrop-blur-xl"
    >
      <div className="container flex h-14 items-center justify-between">
        {/* Brand */}
        <Link
          href="/"
          className="group flex items-center gap-2 transition-opacity hover:opacity-75"
        >
          <span className="text-lg font-semibold tracking-tight text-foreground">
            SalesView
          </span>
          <span className="hidden text-xs font-medium text-muted-foreground sm:inline-block">
          </span>
        </Link>

        {/* Nav */}
        <nav className="hidden md:flex items-center gap-1">
          {navItems.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className="rounded-md px-3 py-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
            >
              {item.label}
            </Link>
          ))}
        </nav>

        {/* Actions */}
        <div className="flex items-center gap-2">
          <Link href="/login" prefetch={false}>
            <Button
              variant="ghost"
              size="sm"
              className="cursor-pointer text-muted-foreground hover:text-foreground"
            >
              Masuk
            </Button>
          </Link>
          <Link href="#pricing">
            <Button size="sm" className="cursor-pointer">
              Lihat pricing
            </Button>
          </Link>
        </div>
      </div>
    </motion.header>
  );
}
