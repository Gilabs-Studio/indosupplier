import type { Metadata } from "next";
import { getLocale } from "next-intl/server";
import { routing } from "@/i18n/routing";
import { buildLandingMetadata } from "@/lib/seo";
import type { Locale } from "@/types/locale";

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

  return buildLandingMetadata({
    locale,
    path: "/",
    title: isId
      ? "Cari Supplier Indonesia Terverifikasi | IndoSupplier"
      : "Find Verified Indonesian Suppliers & Manufacturers | IndoSupplier",
    description: isId
      ? "Bingung cari supplier terpercaya di Indonesia? IndoSupplier menghubungkan buyer langsung dengan produsen dan supplier terverifikasi dari seluruh Indonesia. Tanpa broker, harga tangan pertama, aman dan transparan."
      : "Struggling to find trusted suppliers in Indonesia? IndoSupplier connects buyers directly with verified manufacturers and suppliers from across Indonesia. No brokers, first-hand pricing, safe and transparent.",
    keywords: isId
      ? [
          "cari supplier Indonesia",
          "supplier terverifikasi Indonesia",
          "marketplace B2B Indonesia",
          "platform supplier Indonesia",
          "beli langsung dari produsen",
          "direktori supplier Indonesia",
          "supplier tangan pertama",
          "grosir bahan baku Indonesia",
          "produsen Indonesia terverifikasi",
          "IndoSupplier",
          "sourcing supplier Indonesia",
          "marketplace jual beli supplier",
          "platform pengadaan barang Indonesia",
          "supplier ekspor Indonesia",
          "distributor Indonesia terpercaya",
        ]
      : [
          "find Indonesian suppliers",
          "verified Indonesian suppliers marketplace",
          "B2B marketplace Indonesia",
          "Indonesian supplier directory",
          "buy direct from Indonesian manufacturers",
          "first-hand pricing Indonesia",
          "Indonesia wholesale marketplace",
          "verified manufacturer Indonesia",
          "IndoSupplier",
          "Indonesia sourcing platform",
          "Indonesian supplier search",
          "Indonesia procurement marketplace",
          "export suppliers Indonesia",
          "trusted Indonesia distributors",
        ],
    imageAlt: isId
      ? "IndoSupplier — Platform marketplace menemukan supplier Indonesia terverifikasi"
      : "IndoSupplier — Find verified Indonesian suppliers on Indonesia's B2B marketplace",
    imageUrl: "/screenshot/dashboard.webp",
  });
}

const faqSchema = {
  "@context": "https://schema.org",
  "@type": "FAQPage",
  mainEntity: [
    {
      "@type": "Question",
      name: "Bagaimana cara mencari supplier di Indonesia?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "IndoSupplier menyediakan platform B2B untuk menemukan supplier dan produsen terverifikasi dari seluruh Indonesia. Cukup daftarkan bisnis Anda dan gunakan fitur pencarian untuk menemukan supplier yang sesuai dengan kebutuhan Anda.",
      },
    },
    {
      "@type": "Question",
      name: "Apakah supplier di IndoSupplier sudah terverifikasi?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Ya, semua supplier di IndoSupplier telah melalui proses verifikasi ketat meliputi verifikasi legalitas usaha, kapasitas produksi, dan rekam jejak bisnis sebelum dapat bergabung ke platform.",
      },
    },
    {
      "@type": "Question",
      name: "Apakah ada biaya untuk mendaftar sebagai buyer di IndoSupplier?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Pendaftaran awal di IndoSupplier gratis untuk buyer. Member VIP Waitlist awal mendapatkan gratis biaya komisi transaksi selama 6 bulan pertama.",
      },
    },
    {
      "@type": "Question",
      name: "Apa keuntungan membeli langsung dari produsen melalui IndoSupplier?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Dengan IndoSupplier, buyer mendapatkan harga tangan pertama tanpa markup broker, jaminan keaslian produk, transparansi proses transaksi, dan akses ke ribuan supplier terverifikasi dari seluruh Indonesia.",
      },
    },
    {
      "@type": "Question",
      name: "How can I find suppliers in Indonesia?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "IndoSupplier provides a B2B marketplace platform to find verified suppliers and manufacturers from across Indonesia. Register your business and use the search feature to find suppliers that match your requirements.",
      },
    },
  ],
};

export default function MarketingLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(faqSchema) }}
      />
      {children}
    </>
  );
}
