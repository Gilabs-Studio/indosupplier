import { Metadata } from "next";
import { setRequestLocale } from "next-intl/server";
import { FeatureLandingPage } from "@/components/landing/feature-landing-page";
import { buildLandingMetadata } from "@/lib/seo";

export async function generateMetadata({ params }: { params: Promise<{ locale: string }> }): Promise<Metadata> {
  const { locale } = await params;

  return buildLandingMetadata({
    locale,
    path: "/stock",
    title:
      locale === "id"
        ? "Software Inventory Multi Gudang, Batch & Stock Opname | SalesView"
        : "Multi-Warehouse Inventory Software: Batch, Stock Opname & Ledger | SalesView",
    description:
      locale === "id"
        ? "Pantau stok real-time lintas gudang, lacak batch kadaluarsa, catat pergerakan barang, dan jalankan stock opname dengan pencocokan variance — semua dalam satu sistem inventory yang tersambung ke purchase dan finance."
        : "Monitor real-time stock across warehouses, track expiry batches, record stock movements, and run stock opname with variance matching — all in one inventory system connected to purchase and finance.",
    keywords:
      locale === "id"
        ? [
            "software inventory",
            "stok gudang multi lokasi",
            "batch kadaluarsa",
            "stock opname",
            "pergerakan stok",
            "ledger stok",
            "inventory real-time",
            "salesview inventory",
          ]
        : [
            "inventory software",
            "multi-warehouse stock",
            "batch expiry tracking",
            "stock opname",
            "stock movement",
            "stock ledger",
            "real-time inventory",
            "salesview inventory",
          ],
    imageAlt: "SalesView multi-warehouse inventory overview",
  });
}

export default async function StockLandingPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  setRequestLocale(locale);
  const isId = locale === "id";

  const schemaOrg = {
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    "@id": `https://salesview.id/${locale}/stock#product`,
    name: "SalesView Inventory",
    applicationCategory: "BusinessApplication",
    operatingSystem: "Web",
    description: isId
      ? "Sistem inventory multi gudang dengan batch tracking, stock opname, dan ledger stok terintegrasi."
      : "Multi-warehouse inventory system with batch tracking, stock opname, and integrated stock ledger.",
    url: `https://salesview.id/${locale}/stock`,
  };

  return (
    <>
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(schemaOrg) }} />
      <FeatureLandingPage
        locale={locale}
        label="SalesView Inventory"
        heroTitle={isId ? "Ubah Stok Berantakan" : "Turn Messy Inventory"}
        heroAccent={isId ? "Menjadi Akurasi Sempurna" : "Into Perfect Accuracy"}
        heroDescription={
          isId
            ? "Akhiri mimpi buruk kehabisan barang. Pantau stok multi-gudang Anda, lacak tanggal kedaluwarsa, dan lakukan opname tanpa pusing."
            : "End stockout nightmares. Monitor your multi-warehouse inventory, track expiry dates, and perform stock opnames without the headache."
        }
        heroImageAlt={isId ? "Tampilan inventory SalesView" : "SalesView inventory interface"}
        heroImageKey="stockInventoryList"
        introTitle={isId ? "Amankan aset terbesar bisnis Anda" : "Secure your business's biggest asset"}
        introDescription={
          isId
            ? "Ubah gudang yang kacau menjadi pusat operasi yang rapi dengan pembaruan real-time, sehingga Anda selalu tahu apa yang Anda miliki."
            : "Transform chaotic warehouses into streamlined hubs with real-time updates, so you always know exactly what you have."
        }
        featureBlocks={[
          {
            title: isId ? "Stok real-time lintas gudang" : "Real-time stock across warehouses",
            description: isId
              ? "Pantau ketersediaan barang di setiap gudang dengan status ok, low stock, overstock, dan out of stock — tanpa rekap manual."
              : "Track item availability per warehouse with status badges — ok, low stock, overstock, out of stock — without manual recaps.",
            descriptionSecondary: isId
              ? "Filter berdasarkan gudang, kategori, atau status stok untuk fokus pada area yang perlu tindakan."
              : "Filter by warehouse, category, or stock status to focus on areas needing action.",
            screenshotKey: "stockInventoryList",
            placeholder: "multiOutlet",
          },
          {
            title: isId ? "Lacak batch dan tanggal kadaluarsa" : "Track batches and expiry dates",
            description: isId
              ? "Lihat batch per produk — nomor batch, kuantitas, dan tanggal kadaluarsa — agar barang expired tidak terlewat."
              : "View batches per product — batch number, quantity, and expiry date — so expired goods never slip through.",
            screenshotKey: "stockInventoryBatch",
            placeholder: "warehouseFlow",
          },
          {
            title: isId ? "Stock opname dengan pencocokan variance" : "Stock opname with variance matching",
            description: isId
              ? "Jalankan stock opname terjadwal, bandingkan system quantity vs physical quantity, dan catat variance langsung — semua dalam satu alur audit."
              : "Run scheduled stock opname, compare system qty vs physical qty, and record variance directly — all in one auditable flow.",
            descriptionSecondary: isId
              ? "Hasil opname tersambung ke ledger stok untuk jejak mutasi yang lengkap."
              : "Opname results connect to stock ledger for a complete mutation trail.",
            screenshotKey: "stockOpnameVariance",
            placeholder: "scorecardFlow",
          },
        ]}
        finalTitle={isId ? "Skalakan operasi gudang Anda dengan percaya diri" : "Scale your warehouse operations with confidence"}
        finalPoints={[
          {
            title: isId ? "Pertumbuhan tanpa batas" : "Limitless growth",
            description: isId
              ? "Kelola satu atau seratus lokasi tanpa melambat."
              : "Manage one or a hundred locations without slowing down.",
          },
          {
            title: isId ? "Kepercayaan penuh" : "Full confidence",
            description: isId
              ? "Audit trail otomatis menjaga bisnis Anda tetap aman."
              : "Automated audit trails keep your business secure.",
          },
        ]}
        finalCtaLabel={isId ? "Coba Inventory Gratis" : "Start My Free Trial"}
      />
    </>
  );
}
