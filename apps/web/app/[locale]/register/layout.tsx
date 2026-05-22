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
    ? "Daftar SalesView Gratis — Coba ERP, CRM, POS & Finance Sekarang"
    : "Sign Up SalesView Free — Try ERP, CRM, POS & Finance Today";
  const description = isId
    ? "Buat akun SalesView gratis dan mulai kelola bisnis Anda. Platform ERP, CRM, HRIS, POS, dan Finance all-in-one untuk bisnis Indonesia."
    : "Create your free SalesView account and start managing your business. All-in-one ERP, CRM, HRIS, POS, and Finance platform.";
  const canonicalPath = `/${locale}/register`;

  return {
    title,
    description,
    keywords: isId
      ? [
          "daftar salesview",
          "coba salesview gratis",
          "register salesview",
          "buat akun salesview",
          "salesview free trial",
          "daftar erp indonesia gratis",
          "coba software kasir gratis",
          "software hrd gratis",
          "aplikasi pencatatan penjualan gratis",
        ]
      : [
          "sign up salesview",
          "salesview free trial",
          "salesview register",
          "try salesview free",
          "salesview erp free",
        ],
    alternates: {
      canonical: `${SEO_BASE_URL}${canonicalPath}`,
      languages: {
        id: `${SEO_BASE_URL}/id/register`,
        en: `${SEO_BASE_URL}/en/register`,
        "x-default": `${SEO_BASE_URL}/en/register`,
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

export default function RegisterLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const schemaOrg = {
    "@context": "https://schema.org",
    "@type": "WebPage",
    "@id": `${SEO_BASE_URL}/register`,
    "name": "SalesView - Daftar Gratis",
    "url": `${SEO_BASE_URL}/id/register`,
    "description": "Daftar akun SalesView gratis. Platform ERP, CRM, HRIS, POS, dan Finance all-in-one untuk bisnis Indonesia. Tidak perlu kartu kredit.",
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
