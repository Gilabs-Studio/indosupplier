import type { Metadata } from "next";
import { LandingHeader } from "@/components/landing/landing-header";
import { buildLandingMetadata } from "@/lib/seo";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string }>;
}): Promise<Metadata> {
  const { locale } = await params;

  const isId = locale === "id";
  return buildLandingMetadata({
    locale,
    path: "/",
    title: isId
      ? "SalesView: ERP, CRM, HRIS, POS & Finance untuk Bisnis Indonesia"
      : "SalesView: ERP, CRM, HRIS, POS & Finance for Modern Teams",
    description: isId
      ? "Platform all-in-one untuk penjualan, pembelian, stok, SDM, dan keuangan. Jalankan operasional bisnis lebih cepat dengan SalesView."
      : "An all-in-one platform for sales, purchasing, inventory, HR, and finance. Run your business operations faster with SalesView.",
    keywords: isId
      ? [
          "erp indonesia",
          "software crm",
          "hris indonesia",
          "aplikasi kasir pos",
          "software akuntansi",
          "salesview",
        ]
      : [
          "erp software",
          "crm software",
          "hris platform",
          "point of sale software",
          "accounting software",
          "salesview",
        ],
    imageAlt: "SalesView platform overview",
  });
}

export default function MarketingLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const schemaOrg = {
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": "Organization",
        "@id": "https://salesview.id/#organization",
        "name": "GILABS",
        "url": "https://salesview.id",
        "logo": "https://salesview.id/logo.png",
        "description": "Provider of SalesView: All-in-One Enterprise Platform ERP, CRM, HRIS, POS, and Finance.",
        "sameAs": [
          "https://www.linkedin.com/company/gilabs"
        ]
      },
      {
        "@type": "WebSite",
        "@id": "https://salesview.id/#website",
        "url": "https://salesview.id",
        "name": "SalesView",
        "publisher": {
          "@id": "https://salesview.id/#organization"
        }
      }
    ]
  };

  return (
    <>
      {/* JSON-LD rendered server-side so crawlers see it immediately */}
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(schemaOrg) }}
      />
      {/* Preload LCP image for faster perception of page load */}
      <link
        rel="preload"
        as="image"
        href="/screenshot/dashboard.webp"
        type="image/webp"
      />
      <LandingHeader />
      {children}
    </>
  );
}
