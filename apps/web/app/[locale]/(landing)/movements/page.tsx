import { Metadata } from "next";
import { setRequestLocale } from "next-intl/server";
import { FeatureLandingPage } from "@/components/landing/feature-landing-page";
import { buildLandingMetadata } from "@/lib/seo";

export async function generateMetadata({ params }: { params: Promise<{ locale: string }> }): Promise<Metadata> {
  const { locale } = await params;

  return buildLandingMetadata({
    locale,
    path: "/movements",
    title:
      locale === "id"
        ? "Sistem Pergerakan Stok (IN, OUT, Transfer, Adjust) | Audit Mutasi Gudang | SalesView"
        : "Stock Movement System (IN, OUT, Transfer, Adjust) | Warehouse Mutation Audit | SalesView",
    description:
      locale === "id"
        ? "Catat empat jenis mutasi stok — IN, OUT, ADJUST, dan TRANSFER antar gudang — dengan filter lengkap, detail qty in/out/balance, dan tautan referensi ke delivery order, goods receipt, dan stock opname."
        : "Record four stock movement types — IN, OUT, ADJUST, and inter-warehouse TRANSFER — with full filters, qty in/out/balance detail, and reference links to delivery orders, goods receipt, and stock opname.",
    keywords:
      locale === "id"
        ? [
            "pergerakan stok",
            "stock movement",
            "mutasi gudang",
            "transfer antar gudang",
            "IN OUT stok",
            "adjustment stok",
            "audit pergerakan barang",
            "salesview movements",
          ]
        : [
            "stock movement",
            "warehouse transfer",
            "inventory IN OUT",
            "stock adjustment",
            "mutation audit",
            "inter-warehouse movement",
            "inventory tracking",
            "salesview movements",
          ],
    imageAlt: "SalesView stock movements tracking overview",
  });
}

export default async function MovementsLandingPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  setRequestLocale(locale);
  const isId = locale === "id";

  const schemaOrg = {
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    "@id": `https://salesview.id/${locale}/movements#product`,
    name: "SalesView Movements",
    applicationCategory: "BusinessApplication",
    operatingSystem: "Web",
    description: isId
      ? "Sistem pencatatan pergerakan stok: IN, OUT, ADJUST, dan TRANSFER antar gudang dengan audit trail dan referensi dokumen."
      : "Stock movement tracking system: IN, OUT, ADJUST, and inter-warehouse TRANSFER with audit trail and document references.",
    url: `https://salesview.id/${locale}/movements`,
  };

  return (
    <>
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(schemaOrg) }} />
      <FeatureLandingPage
        locale={locale}
        label="SalesView Movements"
        heroTitle={isId ? "Ubah Mutasi Misterius" : "Turn Mysterious Mutations"}
        heroAccent={isId ? "Menjadi Jejak Transparan" : "Into Transparent Trails"}
        heroDescription={
          isId
            ? "Ketahui persis ke mana barang Anda pergi. Lacak setiap perpindahan, penyesuaian, dan transfer antar gudang dengan visibilitas tak tertandingi."
            : "Know exactly where your inventory goes. Track every movement, adjustment, and inter-warehouse transfer with unmatched visibility."
        }
        heroImageAlt={isId ? "Tampilan pergerakan stok SalesView" : "SalesView stock movements interface"}
        heroImageKey="movementHero"
        introTitle={isId ? "Dapatkan visibilitas penuh atas perjalanan stok Anda" : "Get full visibility over your stock's journey"}
        introDescription={
          isId
            ? "Setiap pergerakan barang Anda kini memiliki cerita yang jelas. Cegah kehilangan dan lacak kembali setiap mutasi ke dokumen aslinya dengan satu klik."
            : "Every item movement now tells a clear story. Prevent losses and trace any mutation back to its original document in just one click."
        }
        featureBlocks={[
          {
            title: isId ? "Empat jenis mutasi stok" : "Four movement types",
            description: isId
              ? "IN untuk barang masuk, OUT untuk keluar, ADJUST untuk koreksi stok, dan TRANSFER untuk pindah antar gudang — semua dalam satu tabel yang bisa difilter."
              : "IN for inbound, OUT for outbound, ADJUST for corrections, TRANSFER for inter-warehouse moves — all in one filterable table.",
            screenshotKey: "movementList",
            placeholder: "multiOutlet",
          },
          {
            title: isId ? "Filter lengkap per gudang dan rentang" : "Full filtering by warehouse and date range",
            description: isId
              ? "Filter berdasarkan gudang, produk, tipe mutasi, dan rentang tanggal — semua langsung tanpa reload. Ekspor CSV sekali klik."
              : "Filter by warehouse, product, movement type, and date range — all live without reload. Export CSV in one click.",
            screenshotKey: "movementFilter",
            placeholder: "terminalFlow",
          },
          {
            title: isId ? "Detail mutasi dengan referensi dokumen" : "Movement detail with document references",
            description: isId
              ? "Klik baris mana pun untuk melihat qty in, qty out, balance, unit cost, dan total value — plus tautan ke dokumen asal: delivery order, goods receipt, atau stock opname."
              : "Click any row to see qty in, qty out, balance, unit cost, and total value — plus links to source documents: delivery order, goods receipt, or stock opname.",
            descriptionSecondary: isId
              ? "Setiap mutasi tercatat di ledger dan bisa ditelusuri ke journal entry."
              : "Every movement is recorded in the ledger and traceable to its journal entry.",
            screenshotKey: "movementDetail",
            placeholder: "scorecardFlow",
          },
        ]}
        finalTitle={isId ? "Amankan pergerakan barang Anda sekarang juga" : "Secure your stock movements right now"}
        finalPoints={[
          {
            title: isId ? "Kendali penuh" : "Full control",
            description: isId
              ? "Pantau barang keluar masuk tanpa tebak-tebakan."
              : "Monitor items moving in and out without the guesswork.",
          },
          {
            title: isId ? "Transparansi" : "Transparency",
            description: isId
              ? "Buktikan setiap perubahan stok dengan mudah."
              : "Easily justify every single stock adjustment.",
          },
        ]}
        finalCtaLabel={isId ? "Mulai Uji Coba Gratis Anda" : "Start My Free Trial"}
      />
    </>
  );
}
