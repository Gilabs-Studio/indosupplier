"use client";
/* eslint-disable @next/next/no-img-element */

import React from "react";
import Image from "next/image";
import { Link } from "@/i18n/routing";
import {
  ArrowRight,
  Facebook,
  Instagram,
  Linkedin,
  ShieldCheck,
  Twitter,
  Smartphone,
  Percent,
  Truck,
  Zap,
} from "lucide-react";

export function PublicFooter() {
  const year = new Date().getFullYear();

  return (
    <footer className="w-full border-t border-border/50 bg-background text-foreground pt-12 pb-8">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 space-y-10">
        {/* Main 4-Column Layout (Inspired by Tokopedia Marketplace Footer) */}
        <div className="grid grid-cols-1 gap-8 sm:grid-cols-2 lg:grid-cols-4">
          {/* Column 1: IndoSupplier Brand & Company Links */}
          <div className="space-y-3">
            <div className="flex items-center gap-2.5">
              <Image
                src="/logo.png"
                alt="IndoSupplier Logo"
                width={100}
                height={20}
                className="h-5 w-auto object-contain brightness-0"
              />
              <span className="font-sans text-sm font-bold tracking-wider uppercase text-foreground">
                IndoSupplier
              </span>
            </div>
            <ul className="space-y-1.5 text-xs text-muted-foreground">
              {[
                { label: "Tentang IndoSupplier", href: "/demo/help" },
                { label: "Hak Kekayaan Intelektual", href: "/terms" },
                { label: "Karir & Talent B2B", href: "/demo/help" },
                { label: "Blog & Insight Industri", href: "#berita" },
                { label: "IndoSupplier Affiliate Program", href: "/demo/help" },
                { label: "IndoSupplier B2B Digital", href: "/demo/search" },
                { label: "IndoSupplier Marketing Solutions", href: "/demo/help" },
                { label: "Direktori Supplier Terverifikasi", href: "/demo/search?verified=true" },
                { label: "Promo Hari Ini", href: "/demo/search" },
                { label: "B2B Lokal Indonesia", href: "/demo/search" },
                { label: "Promo Guncang B2B", href: "/demo/search" },
              ].map((link) => (
                <li key={link.label}>
                  <Link href={link.href} className="hover:text-primary transition-colors cursor-pointer">
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          {/* Column 2: Beli, Jual, Bantuan & Panduan */}
          <div className="space-y-5">
            {/* Beli */}
            <div className="space-y-2">
              <h3 className="text-sm font-bold text-foreground">Beli</h3>
              <ul className="space-y-1.5 text-xs text-muted-foreground">
                <li><Link href="/rfq" className="hover:text-primary transition-colors cursor-pointer">Pengajuan RFQ (Minta Penawaran)</Link></li>
                <li><Link href="/demo/search" className="hover:text-primary transition-colors cursor-pointer">IndoSupplier COD</Link></li>
                <li><Link href="/demo/search" className="hover:text-primary transition-colors cursor-pointer">Bebas Ongkir Ekspor</Link></li>
              </ul>
            </div>

            {/* Jual */}
            <div className="space-y-2">
              <h3 className="text-sm font-bold text-foreground">Jual</h3>
              <ul className="space-y-1.5 text-xs text-muted-foreground">
                <li><Link href="/supplier/register" className="hover:text-primary transition-colors cursor-pointer">Pusat Edukasi Seller</Link></li>
                <li><Link href="/supplier/register" className="hover:text-primary transition-colors cursor-pointer">Daftar Sebagai Supplier</Link></li>
                <li><Link href="/demo/search?verified=true" className="hover:text-primary transition-colors cursor-pointer">Daftar Verified Mall</Link></li>
              </ul>
            </div>

            {/* Bantuan dan Panduan */}
            <div className="space-y-2">
              <h3 className="text-sm font-bold text-foreground">Bantuan dan Panduan</h3>
              <ul className="space-y-1.5 text-xs text-muted-foreground">
                <li><Link href="/demo/help" className="hover:text-primary transition-colors cursor-pointer">IndoSupplier Care</Link></li>
                <li><Link href="/terms" className="hover:text-primary transition-colors cursor-pointer">Syarat dan Ketentuan</Link></li>
                <li><Link href="/privacy" className="hover:text-primary transition-colors cursor-pointer">Kebijakan Privasi</Link></li>
              </ul>
            </div>
          </div>

          {/* Column 3: Keamanan & Privasi & Social Icons */}
          <div className="space-y-6">
            <div className="space-y-3">
              <h3 className="text-sm font-bold text-foreground">Keamanan & Privasi</h3>
              <div className="space-y-2">
                {/* Badge 1: PCI DSS COMPLIANT */}
                <div className="flex items-center gap-2.5 rounded-lg border border-border/60 bg-card p-2 shadow-2xs">
                  <div className="flex h-7 w-7 items-center justify-center rounded-xs bg-primary/10 text-primary font-black text-[10px]">
                    PCI
                  </div>
                  <div>
                    <p className="text-[11px] font-bold text-foreground leading-none">PCI DSS COMPLIANT</p>
                    <p className="text-[9px] text-muted-foreground mt-0.5">Assessed by ControlCase</p>
                  </div>
                </div>

                {/* Badge 2: ISO 27001 SECURITY */}
                <div className="flex items-center gap-2.5 rounded-lg border border-border/60 bg-card p-2 shadow-2xs">
                  <ShieldCheck className="h-6 w-6 text-success shrink-0" />
                  <div>
                    <p className="text-[11px] font-bold text-foreground leading-none">ISO/IEC 27001</p>
                    <p className="text-[9px] text-muted-foreground mt-0.5">Information Security Management</p>
                  </div>
                </div>
              </div>
            </div>

            {/* Social Media Links */}
            <div className="space-y-2.5">
              <h3 className="text-sm font-bold text-foreground">Ikuti Kami</h3>
              <div className="flex items-center gap-2">
                <a href="#" className="flex h-8 w-8 items-center justify-center rounded-full bg-blue-600 text-white transition-transform hover:scale-105" aria-label="Facebook">
                  <Facebook className="h-4 w-4 fill-current" />
                </a>
                <a href="#" className="flex h-8 w-8 items-center justify-center rounded-full bg-sky-500 text-white transition-transform hover:scale-105" aria-label="Twitter">
                  <Twitter className="h-4 w-4 fill-current" />
                </a>
                <a href="#" className="flex h-8 w-8 items-center justify-center rounded-full bg-blue-700 text-white transition-transform hover:scale-105" aria-label="LinkedIn">
                  <Linkedin className="h-4 w-4 fill-current" />
                </a>
                <a href="#" className="flex h-8 w-8 items-center justify-center rounded-full bg-gradient-to-tr from-amber-500 via-rose-500 to-purple-600 text-white transition-transform hover:scale-105" aria-label="Instagram">
                  <Instagram className="h-4 w-4" />
                </a>
              </div>
            </div>
          </div>

          {/* Column 4: Mobile App App Section & Scan QR (Tokopedia Reference Style) */}
          <div className="space-y-4">
            <h3 className="text-sm font-bold leading-snug text-foreground">
              Nikmati keuntungan spesial di aplikasi:
            </h3>

            <ul className="space-y-2 text-xs text-muted-foreground">
              <li className="flex items-center gap-2">
                <Percent className="h-4 w-4 text-success shrink-0" />
                <span>Diskon 70%* khusus pengajuan RFQ aplikasi</span>
              </li>
              <li className="flex items-center gap-2">
                <Zap className="h-4 w-4 text-warning shrink-0" />
                <span>Notifikasi real-time penawaran supplier</span>
              </li>
              <li className="flex items-center gap-2">
                <Truck className="h-4 w-4 text-primary shrink-0" />
                <span>Gratis Ongkir & Audit Pabrik tiap hari</span>
              </li>
            </ul>

            <div className="space-y-2 pt-1">
              <p className="text-[11px] text-muted-foreground">Buka aplikasi dengan scan QR atau klik tombol:</p>
              <div className="flex items-center gap-3">
                {/* QR Code Container */}
                <div className="flex h-20 w-20 shrink-0 items-center justify-center rounded-lg border border-border bg-card p-1 shadow-2xs">
                  <div className="relative flex h-full w-full items-center justify-center rounded bg-foreground text-background">
                    {/* Stylized QR Code SVG mockup */}
                    <svg viewBox="0 0 24 24" fill="currentColor" className="h-14 w-14">
                      <path d="M2 2h8v8H2V2zm2 2v4h4V4H4zm9-2h8v8h-8V2zm2 2v4h4V4h-4zM2 14h8v8H2v-8zm2 2v4h4v-4H4zm13-2h2v2h-2v-2zm-4 0h2v2h-2v-2zm2 2h2v2h-2v-2zm-2 2h2v2h-2v-2zm4 0h2v2h-2v-2zm0-4h2v2h-2v-2zm2 4h2v2h-2v-2z" />
                    </svg>
                  </div>
                </div>

                {/* App Download Store Buttons */}
                <div className="flex flex-col gap-1.5 min-w-[130px]">
                  <a href="#" className="flex items-center gap-2 rounded-md bg-neutral-900 px-2.5 py-1 text-white transition-opacity hover:opacity-90">
                    <Smartphone className="h-3.5 w-3.5" />
                    <div className="text-[9px] leading-tight text-left">
                      <p className="opacity-80">GET IT ON</p>
                      <p className="font-bold text-[10px]">Google Play</p>
                    </div>
                  </a>

                  <a href="#" className="flex items-center gap-2 rounded-md bg-neutral-900 px-2.5 py-1 text-white transition-opacity hover:opacity-90">
                    <Smartphone className="h-3.5 w-3.5" />
                    <div className="text-[9px] leading-tight text-left">
                      <p className="opacity-80">Download on the</p>
                      <p className="font-bold text-[10px]">App Store</p>
                    </div>
                  </a>
                </div>
              </div>

              <Link href="/demo/help" className="inline-flex items-center gap-1 pt-1 text-xs font-bold text-primary hover:underline cursor-pointer">
                <span>Pelajari Selengkapnya</span>
                <ArrowRight className="h-3.5 w-3.5" />
              </Link>
            </div>
          </div>
        </div>

        {/* Bottom Copyright & Language Switcher Bar */}
        <div className="flex flex-col items-center justify-between gap-4 border-t border-border/40 pt-6 sm:flex-row text-xs text-muted-foreground">
          <p>© 2009 - {year}, PT. IndoSupplier Global. All Rights Reserved.</p>

          <div className="flex items-center gap-1 rounded-lg border border-border/60 bg-muted/30 p-1">
            <button type="button" className="rounded-md bg-primary px-3 py-1 text-[11px] font-semibold text-primary-foreground">
              Indonesia
            </button>
            <button type="button" className="rounded-md px-3 py-1 text-[11px] font-medium text-muted-foreground hover:text-foreground">
              English
            </button>
          </div>
        </div>
      </div>
    </footer>
  );
}
