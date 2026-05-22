import type { Metadata } from "next";
import { setRequestLocale } from "next-intl/server";
import { PricingSection } from "@/components/landing/pricing-section";
import { buildLandingMetadata, SEO_BASE_URL } from "@/lib/seo";

export async function generateMetadata({ params }: { params: Promise<{ locale: string }> }): Promise<Metadata> {
  const { locale } = await params;
  const isId = locale === "id";

  const title = isId
    ? "Harga Paket ERP, CRM, HRIS, POS & Finance | SalesView"
    : "Pricing Plans for ERP, CRM, HRIS, POS & Finance | SalesView";
  const description = isId
    ? "Lihat harga SalesView untuk paket ERP, CRM, HRIS, POS, dan Finance. Bandingkan bundle vs modular lalu pilih paket yang sesuai kebutuhan dan skala bisnis Anda."
    : "See SalesView pricing for ERP, CRM, HRIS, POS, and Finance. Compare bundle vs modular options and choose the plan that fits your business size and needs.";

  return buildLandingMetadata({
    locale,
    path: "/pricing",
    title,
    description,
    keywords: isId
      ? [
          "harga erp salesview",
          "harga salesview",
          "pricing erp indonesia",
          "harga software crm",
          "harga aplikasi kasir pos",
          "harga hris",
          "paket software bisnis",
        ]
      : [
          "salesview pricing",
          "erp pricing indonesia",
          "crm pricing",
          "hris pricing",
          "pos pricing",
          "business software plans",
        ],
    imageAlt: "SalesView pricing overview",
  });
}

export default async function PricingPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  setRequestLocale(locale);

  const isId = locale === "id";
  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "Product",
    name: "SalesView Pricing",
    description: isId
      ? "Halaman pricing resmi SalesView untuk paket ERP, CRM, HRIS, POS, dan Finance."
      : "Official SalesView pricing page for ERP, CRM, HRIS, POS, and Finance plans.",
    brand: {
      "@type": "Brand",
      name: "SalesView",
    },
    category: "Business Software",
    offers: {
      "@type": "AggregateOffer",
      priceCurrency: "IDR",
      lowPrice: "79000",
      highPrice: "175000",
      offerCount: "6",
      url: `${SEO_BASE_URL}/${locale}/pricing`,
      availability: "https://schema.org/InStock",
    },
  };

  return (
    <>
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }} />
      <div className="relative z-10 overflow-x-hidden pt-16">
        <PricingSection />
      </div>
    </>
  );
}
