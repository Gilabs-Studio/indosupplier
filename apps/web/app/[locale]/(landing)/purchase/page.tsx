import { Metadata } from "next";
import { setRequestLocale } from "next-intl/server";
import { FeatureLandingPage } from "@/components/landing/feature-landing-page";
import { buildLandingMetadata } from "@/lib/seo";

export async function generateMetadata({ params }: { params: Promise<{ locale: string }> }): Promise<Metadata> {
  const { locale } = await params;

  return buildLandingMetadata({
    locale,
    path: "/purchase",
    title:
      locale === "id"
        ? "Software Procurement Purchase Order & Hutang Supplier | SalesView"
        : "Procurement Software: Purchase Orders & Supplier Payables | SalesView",
    description:
      locale === "id"
        ? "Kelola permintaan pembelian, purchase order, penerimaan barang, faktur supplier, hingga rekapan hutang dalam satu alur procurement yang tersambung ke inventory dan finance."
        : "Manage purchase requisitions, purchase orders, goods receipt, supplier invoices, and payable recap in one procurement flow connected to inventory and finance.",
    keywords:
      locale === "id"
        ? [
            "software procurement",
            "sistem purchase order",
            "permintaan pembelian",
            "penerimaan barang",
            "hutang supplier",
            "rekapan hutang",
            "faktur supplier",
            "salesview purchase",
          ]
        : [
            "procurement software",
            "purchase order system",
            "purchase requisition",
            "goods receipt tracking",
            "supplier payables",
            "accounts payable aging",
            "supplier invoice management",
            "salesview purchase",
          ],
    imageAlt: "SalesView procurement purchase order overview",
  });
}

export default async function PurchaseLandingPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  setRequestLocale(locale);
  const isId = locale === "id";

  const schemaOrg = {
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    "@id": `https://salesview.id/${locale}/purchase#product`,
    name: isId ? "SalesView Purchase" : "SalesView Purchase",
    applicationCategory: "BusinessApplication",
    operatingSystem: "Web",
    description: isId
      ? "Procurement terintegrasi untuk B2B: dari permintaan pembelian sampai kontrol hutang supplier."
      : "Integrated procurement for B2B: from purchase requests through supplier payable control.",
    url: `https://salesview.id/${locale}/purchase`,
  };

  return (
    <>
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(schemaOrg) }} />
      <FeatureLandingPage
        locale={locale}
        label="SalesView Purchase"
        heroTitle={isId ? "Ubah Pengadaan Rumit" : "Turn Complex Procurement"}
        heroAccent={isId ? "Menjadi Kendali Otomatis" : "Into Automated Control"}
        heroDescription={
          isId
            ? "Tinggalkan proses pengadaan manual. Hubungkan permintaan, pesanan, dan tagihan supplier dalam satu sistem mulus yang mengamankan margin Anda."
            : "Leave manual procurement behind. Connect requests, orders, and supplier invoices in one seamless system that protects your margins."
        }
        heroImageAlt={isId ? "Tampilan procurement SalesView" : "SalesView procurement interface"}
        heroImageKey="purchaseHero"
        introTitle={isId ? "Sederhanakan pembelian bisnis Anda hari ini" : "Simplify your business purchasing today"}
        introDescription={
          isId
            ? "Nikmati alur kerja yang dirancang untuk mencegah pengeluaran berlebih, memastikan persetujuan cepat, dan mempererat hubungan dengan pemasok."
            : "Enjoy workflows designed to prevent overspending, ensure fast approvals, and strengthen relationships with your suppliers."
        }
        featureBlocks={[
          {
            title: isId ? "Dari permintaan ke purchase order" : "From requisition to purchase order",
            description: isId
              ? "Ajukan permintaan pembelian, dapatkan approval, lalu konversi langsung ke purchase order — tanpa proses berlapis."
              : "Submit purchase requisitions, get approval, and convert directly into purchase orders — without layered delays.",
            descriptionSecondary: isId
              ? "Hubungkan dengan supplier, payment terms, dan business unit dalam satu alur."
              : "Link with supplier, payment terms, and business unit in one flow.",
            screenshotKey: "purchaseRequisitionToPO",
            placeholder: "multiOutlet",
          },
          {
            title: isId ? "Penerimaan barang tersambung ke faktur" : "Goods receipt linked to supplier billing",
            description: isId
              ? "Catat barang masuk per purchase order, validasi kuantitas, lalu teruskan ke faktur supplier — jejak data tetap utuh."
              : "Record inbound items per purchase order, validate quantities, and pass to supplier invoice — the data trail stays complete.",
            screenshotKey: "purchaseGRToInvoice",
            placeholder: "warehouseFlow",
          },
          {
            title: isId ? "Pantau hutang supplier dengan aging jelas" : "Track supplier payables with clear aging",
            description: isId
              ? "Lihat total hutang per supplier dengan kategori aging: lancar, jatuh tempo 1-30 hari, 31-60 hari, hingga bad debt."
              : "View total payables per supplier with aging categories: current, 1-30 days overdue, 31-60 days, up to bad debt.",
            screenshotKey: "purchasePayableRecap",
            placeholder: "scorecardFlow",
          },
        ]}
        finalTitle={isId ? "Mulai kendalikan pengeluaran bisnis Anda sekarang" : "Start controlling your business spend now"}
        finalCtaLabel={isId ? "Mulai Uji Coba Gratis Anda" : "Start My Free Trial"}
      />
    </>
  );
}
