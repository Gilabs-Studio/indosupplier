"use client";

import React from "react";
import Image from "next/image";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";

export function PublicFooter() {
  const t = useTranslations("public.navbar");
  const year = new Date().getFullYear();

  return (
    <footer className="w-full border-t border-border bg-muted py-12 text-muted-foreground">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
          {/* Logo & Description */}
          <div className="md:col-span-2 space-y-4">
            <div className="flex items-center gap-3">
              <Image
                src="/logo.png"
                alt="IndoSupplier Logo"
                width={100}
                height={20}
                className="h-5 w-auto object-contain brightness-0"
              />
              <span className="font-sans text-[14px] font-semibold tracking-wider uppercase text-foreground">
                IndoSupplier
              </span>
            </div>
            <p className="text-xs text-muted-foreground max-w-xs leading-relaxed">
              The premier B2B marketplace platform connecting verified suppliers and business buyers across Indonesia.
            </p>
          </div>

          {/* Quick Links */}
          <div>
            <h3 className="text-xs font-semibold uppercase tracking-wider text-foreground mb-4">
              Discovery
            </h3>
            <ul className="space-y-2 text-xs">
              <li>
                <Link href="/demo/search" className="hover:text-primary transition-colors cursor-pointer">
                  {t("search")}
                </Link>
              </li>
              <li>
                <Link href="/demo/categories" className="hover:text-primary transition-colors cursor-pointer">
                  {t("categories")}
                </Link>
              </li>
            </ul>
          </div>

          {/* Support Links */}
          <div>
            <h3 className="text-xs font-semibold uppercase tracking-wider text-foreground mb-4">
              Support
            </h3>
            <ul className="space-y-2 text-xs">
              <li>
                <Link href="/demo/help" className="hover:text-primary transition-colors cursor-pointer">
                  {t("help")}
                </Link>
              </li>
              <li>
                <Link href="/demo/faq" className="hover:text-primary transition-colors cursor-pointer">
                  {t("faq")}
                </Link>
              </li>
            </ul>
          </div>
        </div>

        <div className="mt-8 border-t border-border pt-6 flex flex-col sm:flex-row items-center justify-between gap-4 text-xs text-muted-foreground/60">
          <p>© {year} IndoSupplier. All rights reserved.</p>
          <div className="flex gap-4">
            <Link href="/terms" className="hover:text-foreground transition-colors cursor-pointer">Terms of Service</Link>
            <Link href="/privacy" className="hover:text-foreground transition-colors cursor-pointer">Privacy Policy</Link>
          </div>
        </div>
      </div>
    </footer>
  );
}
