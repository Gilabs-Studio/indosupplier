import type { Metadata } from "next";
import { SEO_BASE_URL } from "@/lib/seo";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string }>;
}): Promise<Metadata> {
  const { locale } = await params;
  const isId = locale === "id";

  const title = isId
    ? "Login SalesView — Masuk ke Akun ERP, CRM & POS Anda"
    : "SalesView Login — Sign In to Your ERP, CRM & POS Account";
  const description = isId
    ? "Masuk ke platform SalesView untuk mengelola penjualan, stok, keuangan, dan SDM bisnis Anda dalam satu dasbor."
    : "Sign in to SalesView to manage your sales, inventory, finance, and HR operations from one unified dashboard.";
  const canonicalPath = `/${locale}/login`;

  return {
    title,
    description,
    keywords: isId
      ? [
          "salesview login",
          "login salesview",
          "masuk salesview",
          "login erp indonesia",
          "login software kasir",
          "salesview masuk akun",
        ]
      : [
          "salesview login",
          "sign in salesview",
          "salesview erp login",
          "salesview account",
        ],
    alternates: {
      canonical: `${SEO_BASE_URL}${canonicalPath}`,
      languages: {
        id: `${SEO_BASE_URL}/id/login`,
        en: `${SEO_BASE_URL}/en/login`,
        "x-default": `${SEO_BASE_URL}/en/login`,
      },
    },
    openGraph: {
      type: "website",
      locale: isId ? "id_ID" : "en_US",
      url: `${SEO_BASE_URL}${canonicalPath}`,
      title,
      description,
      siteName: "SalesView",
    },
    robots: {
      index: false,
      follow: true,
      googleBot: { index: false, follow: true },
    },
  };
}

export default function LoginLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const schemaOrg = {
    "@context": "https://schema.org",
    "@type": "WebPage",
    "@id": `${SEO_BASE_URL}/login`,
    "name": "SalesView Login",
    "url": `${SEO_BASE_URL}/id/login`,
    "description": "Halaman login SalesView — masuk ke platform ERP, CRM, HRIS, POS, dan Finance all-in-one untuk bisnis Indonesia.",
    "isPartOf": {
      "@type": "WebSite",
      "@id": `${SEO_BASE_URL}/#website`,
      "name": "SalesView",
      "url": SEO_BASE_URL,
    },
  };

  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(schemaOrg) }}
      />
      {children}
    </>
  );
}
