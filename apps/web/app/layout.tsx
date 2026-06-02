import type { Metadata, Viewport } from "next";
import { Geist, Geist_Mono, Sora, Newsreader, Jost, Macondo } from "next/font/google";
import { getLocale } from "next-intl/server";
import { routing } from "@/i18n/routing";
import { getLanguageAlternates } from "@/lib/seo";
import type { Locale } from "@/types/locale";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

const headingFont = Sora({
  variable: "--font-heading",
  subsets: ["latin"],
  weight: ["500", "600", "700"],
  display: "swap",
});

const accentFont = Newsreader({
  variable: "--font-accent",
  subsets: ["latin"],
  weight: ["400", "500", "600"],
  style: ["italic", "normal"],
  display: "swap",
});

const jostFont = Jost({
  variable: "--font-jost",
  subsets: ["latin"],
  weight: ["300", "400", "500", "600"],
  display: "swap",
});

const macondoFont = Macondo({
  variable: "--font-macondo",
  subsets: ["latin"],
  weight: ["400"],
  display: "swap",
});

export async function generateMetadata(): Promise<Metadata> {
  let locale: Locale;

  try {
    const localeValue = await getLocale();
    locale = routing.locales.includes(localeValue as Locale)
      ? (localeValue as Locale)
      : routing.defaultLocale;
  } catch {
    locale = routing.defaultLocale;
  }

  return {
    metadataBase: new URL("https://indosupplier.id"),
    title: {
      template: "%s | Indosupplier",
      default: "Indosupplier - ERP, CRM, HRIS, POS & Finance",
    },
    description:
      "Indosupplier adalah software all-in-one ERP, CRM, HRIS, POS, dan Finance untuk bisnis Indonesia. Satu platform untuk operasional, penjualan, stok, HR, dan laporan keuangan.",
    keywords: [
      "ERP Indonesia",
      "Aplikasi CRM",
      "Sistem HRIS",
      "Aplikasi Kasir POS",
      "Software Finance",
      "Indosupplier",
      "Manajemen Bisnis Terintegrasi",
      "All-in-one Software",
      "Vendor ERP",
      "Custom ERP",
      "Custom ERP Indonesia",
      "ERP Indonesia Murah",
      "point of sales",
      "pos kasir",
      "software hrd",
      "inventory software",
      "sales erp",
      "aplikasi pencatatan penjualan",
      "sales management software",
      "software manajemen penjualan",
    ],
    authors: [{ name: "Indosupplier" }],
    creator: "Indosupplier",
    publisher: "Indosupplier",
    openGraph: {
      type: "website",
      locale: locale === "id" ? "id_ID" : "en_US",
      url: "https://indosupplier.id",
      title: "Indosupplier - ERP, CRM, HRIS, POS & Finance",
      description:
        "Indosupplier adalah platform bisnis all-in-one untuk ERP, CRM, HRIS, POS, dan Finance di Indonesia.",
      siteName: "Indosupplier",
      images: [
        {
          url: "/screenshot/dashboard.webp",
          width: 1920,
          height: 1080,
          alt: "Indosupplier dashboard preview",
        },
      ],
    },
    twitter: {
      card: "summary_large_image",
      title: "Indosupplier - ERP, CRM, HRIS, POS & Finance",
      description:
        "Indosupplier adalah platform bisnis all-in-one untuk ERP, CRM, HRIS, POS, dan Finance di Indonesia.",
      creator: "@indosupplier",
      images: ["/screenshot/dashboard.webp"],
    },
    robots: {
      index: true,
      follow: true,
      googleBot: {
        index: true,
        follow: true,
        "max-video-preview": -1,
        "max-image-preview": "large",
        "max-snippet": -1,
      },
    },
    alternates: {
      canonical: "https://indosupplier.id",
      languages: getLanguageAlternates("/"),
    },
  };
}

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  maximumScale: 5,
  userScalable: true,
};

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  let locale: Locale;
  try {
    const localeValue = await getLocale();
    locale = routing.locales.includes(localeValue as Locale)
      ? (localeValue as Locale)
      : routing.defaultLocale;
  } catch {
    locale = routing.defaultLocale;
  }

  return (
    <html lang={locale} suppressHydrationWarning>
      <head />
      <body
        className={`${geistSans.variable} ${geistMono.variable} ${headingFont.variable} ${accentFont.variable} ${jostFont.variable} ${macondoFont.variable} antialiased`}
      >
        {children}
      </body>
    </html>
  );
}
