import type { Metadata, Viewport } from "next";
import localFont from "next/font/local";
import { getLocale } from "next-intl/server";
import { routing } from "@/i18n/routing";
import { getLanguageAlternates } from "@/lib/seo";
import type { Locale } from "@/types/locale";
import "./globals.css";

const tinos = localFont({
  src: [
    {
      path: "./fonts/tinos-400.woff2",
      weight: "400",
      style: "normal",
    },
    {
      path: "./fonts/tinos-400-italic.woff2",
      weight: "400",
      style: "italic",
    },
    {
      path: "./fonts/tinos-700.woff2",
      weight: "700",
      style: "normal",
    },
    {
      path: "./fonts/tinos-700-italic.woff2",
      weight: "700",
      style: "italic",
    },
  ],
  variable: "--font-tinos",
  display: "swap",
});

const inter = localFont({
  src: "./fonts/inter.woff2",
  variable: "--font-inter",
  weight: "100 900",
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

  const isId = locale === "id";

  return {
    metadataBase: new URL("https://indosupplier.id"),
    title: {
      template: "%s | IndoSupplier",
      default: isId
        ? "IndoSupplier — Platform Marketplace Supplier Indonesia Terverifikasi"
        : "IndoSupplier — Find Verified Indonesian Suppliers & Manufacturers",
    },
    description: isId
      ? "IndoSupplier adalah marketplace B2B Indonesia yang menghubungkan buyer dengan supplier dan produsen terverifikasi. Temukan supplier terpercaya, bandingkan harga langsung dari produsen, dan beli tanpa broker."
      : "IndoSupplier is Indonesia's B2B marketplace connecting buyers with verified suppliers and manufacturers. Find trusted suppliers, compare prices directly from producers, and buy without middlemen.",
    keywords: isId
      ? [
          "cari supplier Indonesia",
          "marketplace supplier Indonesia",
          "supplier terverifikasi",
          "platform B2B Indonesia",
          "beli langsung dari produsen",
          "supplier terpercaya",
          "marketplace produsen Indonesia",
          "direktori supplier Indonesia",
          "grosir Indonesia",
          "ekspor impor Indonesia",
          "supplier bahan baku",
          "vendor Indonesia",
          "IndoSupplier",
          "platform pengadaan Indonesia",
          "supplier manufaktur Indonesia",
          "jual beli supplier",
          "B2B marketplace Indonesia",
          "sourcing supplier Indonesia",
          "cari produsen Indonesia",
          "distributor Indonesia",
        ]
      : [
          "Indonesian suppliers marketplace",
          "find suppliers Indonesia",
          "verified Indonesian manufacturers",
          "B2B marketplace Indonesia",
          "buy direct from Indonesian producers",
          "Indonesian supplier directory",
          "Indonesia wholesale suppliers",
          "export suppliers Indonesia",
          "raw material suppliers Indonesia",
          "IndoSupplier",
          "Indonesia procurement platform",
          "Indonesian manufacturer directory",
          "sourcing Indonesia",
          "Indonesia B2B trading",
          "trusted Indonesian suppliers",
        ],
    authors: [{ name: "IndoSupplier" }],
    creator: "IndoSupplier",
    publisher: "IndoSupplier",
    openGraph: {
      type: "website",
      locale: isId ? "id_ID" : "en_US",
      alternateLocale: isId ? "en_US" : "id_ID",
      url: "https://indosupplier.id",
      title: isId
        ? "IndoSupplier — Platform Marketplace Supplier Indonesia Terverifikasi"
        : "IndoSupplier — Find Verified Indonesian Suppliers & Manufacturers",
      description: isId
        ? "Marketplace B2B Indonesia terpercaya. Temukan dan hubungi supplier, produsen, dan eksportir terverifikasi dari seluruh Indonesia. Tanpa broker, harga langsung dari produsen."
        : "Indonesia's trusted B2B marketplace. Discover and connect with verified suppliers, manufacturers, and exporters from across Indonesia. No middlemen, direct producer pricing.",
      siteName: "IndoSupplier",
      images: [
        {
          url: "/screenshot/dashboard.webp",
          width: 1920,
          height: 1080,
          alt: isId
            ? "IndoSupplier — Platform marketplace supplier Indonesia"
            : "IndoSupplier — Indonesian supplier marketplace platform",
        },
      ],
    },
    twitter: {
      card: "summary_large_image",
      title: isId
        ? "IndoSupplier — Platform Marketplace Supplier Indonesia Terverifikasi"
        : "IndoSupplier — Find Verified Indonesian Suppliers & Manufacturers",
      description: isId
        ? "Marketplace B2B Indonesia terpercaya. Temukan supplier, produsen, dan eksportir terverifikasi dari seluruh Indonesia."
        : "Indonesia's trusted B2B marketplace. Find verified suppliers, manufacturers and exporters from across Indonesia.",
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
    category: "business",
  };
}

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  maximumScale: 5,
  userScalable: true,
};

const organizationSchema = {
  "@context": "https://schema.org",
  "@type": "Organization",
  "@id": "https://indosupplier.id/#organization",
  name: "IndoSupplier",
  url: "https://indosupplier.id",
  logo: {
    "@type": "ImageObject",
    url: "https://indosupplier.id/logo.png",
    width: 200,
    height: 60,
  },
  description:
    "Platform marketplace B2B Indonesia yang menghubungkan buyer dengan supplier dan produsen terverifikasi dari seluruh Indonesia.",
  areaServed: {
    "@type": "Country",
    name: "Indonesia",
  },
  knowsAbout: [
    "Supplier Indonesia",
    "Manufaktur Indonesia",
    "B2B Marketplace",
    "Pengadaan Barang",
    "Ekspor Indonesia",
  ],
  sameAs: ["https://www.linkedin.com/company/indosupplier"],
};

const websiteSchema = {
  "@context": "https://schema.org",
  "@type": "WebSite",
  "@id": "https://indosupplier.id/#website",
  url: "https://indosupplier.id",
  name: "IndoSupplier",
  description:
    "Platform marketplace B2B untuk menemukan supplier dan produsen terverifikasi di Indonesia",
  publisher: {
    "@id": "https://indosupplier.id/#organization",
  },
  inLanguage: ["id-ID", "en-US"],
  potentialAction: {
    "@type": "SearchAction",
    target: {
      "@type": "EntryPoint",
      urlTemplate: "https://indosupplier.id/id?q={search_term_string}",
    },
    "query-input": {
      "@type": "PropertyValueSpecification",
      valueRequired: true,
      valueName: "search_term_string",
    },
  },
};

const softwareApplicationSchema = {
  "@context": "https://schema.org",
  "@type": "SoftwareApplication",
  name: "IndoSupplier",
  applicationCategory: "BusinessApplication",
  operatingSystem: "Web",
  url: "https://indosupplier.id",
  description:
    "Platform marketplace B2B Indonesia: temukan supplier, produsen, dan eksportir terverifikasi dengan mudah dan cepat.",
  offers: {
    "@type": "Offer",
    price: "0",
    priceCurrency: "IDR",
    description: "Pendaftaran gratis untuk buyer dan supplier",
  },
  publisher: {
    "@id": "https://indosupplier.id/#organization",
  },
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
    <html lang={locale} suppressHydrationWarning className={`${tinos.variable} ${inter.variable}`}>
      <head>
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(organizationSchema) }}
        />
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(websiteSchema) }}
        />
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(softwareApplicationSchema) }}
        />
      </head>
      <body
        className="antialiased"
      >
        {children}
      </body>
    </html>
  );
}
