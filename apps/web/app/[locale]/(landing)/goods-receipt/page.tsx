import { Metadata } from "next";
import { setRequestLocale } from "next-intl/server";
import { FeatureLandingPage } from "@/components/landing/feature-landing-page";
import { buildLandingMetadata } from "@/lib/seo";

export async function generateMetadata({ params }: { params: Promise<{ locale: string }> }): Promise<Metadata> {
  const { locale } = await params;

  return buildLandingMetadata({
    locale,
    path: "/goods-receipt",
    title:
      locale === "id"
        ? "Sistem Penerimaan Barang (Goods Receipt) | Terhubung PO & Supplier Invoice | SalesView"
        : "Goods Receipt System | Linked to PO & Supplier Invoice | SalesView",
    description:
      locale === "id"
        ? "Catat penerimaan barang per purchase order, validasi kuantitas, lacak audit trail, dan teruskan ke faktur supplier — semua dalam satu alur goods receipt yang akurat dan siap diaudit."
        : "Record goods receipt per purchase order, validate quantities, track audit trails, and pass to supplier invoice — all in one accurate, audit-ready goods receipt workflow.",
    keywords:
      locale === "id"
        ? [
            "penerimaan barang",
            "goods receipt",
            "validasi barang masuk",
            "purchase order matching",
            "barang masuk gudang",
            "audit trail penerimaan",
            "konversi faktur supplier",
            "salesview goods receipt",
          ]
        : [
            "goods receipt",
            "inbound receiving",
            "purchase order matching",
            "warehouse receiving",
            "receipt audit trail",
            "supplier invoice conversion",
            "goods receipt software",
            "salesview goods receipt",
          ],
    imageAlt: "SalesView goods receipt management overview",
  });
}

export default async function GoodsReceiptLandingPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  setRequestLocale(locale);
  const isId = locale === "id";

  const schemaOrg = {
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    "@id": `https://salesview.id/${locale}/goods-receipt#product`,
    name: "SalesView Goods Receipt",
    applicationCategory: "BusinessApplication",
    operatingSystem: "Web",
    description: isId
      ? "Sistem penerimaan barang gudang yang terhubung ke purchase order, audit trail, dan konversi faktur supplier."
      : "Warehouse goods receipt system linked to purchase orders, audit trails, and supplier invoice conversion.",
    url: `https://salesview.id/${locale}/goods-receipt`,
  };

  return (
    <>
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(schemaOrg) }} />
      <FeatureLandingPage
        locale={locale}
        label="SalesView Goods Receipt"
        heroTitle={isId ? "Ubah Penerimaan Kacau" : "Turn Chaotic Receiving"}
        heroAccent={isId ? "Menjadi Pencatatan Presisi" : "Into Precise Records"}
        heroDescription={
          isId
            ? "Jangan biarkan barang hilang saat kedatangan. Validasi kuantitas secara akurat dan cocokkan dengan purchase order secara instan untuk mencegah kebocoran."
            : "Never let items slip through the cracks upon arrival. Validate quantities accurately and match them to purchase orders instantly to prevent leakage."
        }
        heroImageAlt={isId ? "Tampilan penerimaan barang SalesView" : "SalesView goods receipt interface"}
        heroImageKey="goodsReceiptHero"
        introTitle={isId ? "Kunci akurasi stok Anda sejak pintu depan" : "Lock in your stock accuracy from the front door"}
        introDescription={
          isId
            ? "Beri tim gudang Anda alat yang tepat untuk memproses penerimaan barang tanpa kesalahan, mempercepat pembayaran ke pemasok yang benar."
            : "Give your warehouse team the right tools to process inbound goods flawlessly, speeding up correct payments to suppliers."
        }
        featureBlocks={[
          {
            title: isId ? "Catat barang per purchase order" : "Record items per purchase order",
            description: isId
              ? "Hubungkan setiap penerimaan barang ke purchase order — pilih PO yang sudah approved, isi kuantitas per item, dan simpan dalam satu langkah."
              : "Link every goods receipt to a purchase order — select an approved PO, fill in quantities per item, and save in one step.",
            descriptionSecondary: isId
              ? "Setiap item divalidasi agar quantity received > 0 sebelum bisa disimpan."
              : "Each item is validated so quantity received > 0 before saving.",
            screenshotKey: "goodsReceiptPO",
            placeholder: "warehouseFlow",
          },
          {
            title: isId ? "Jejak audit dari draft sampai closed" : "Audit trail from draft to closed",
            description: isId
              ? "Setiap status — DRAFT, SUBMITTED, APPROVED, REJECTED, CLOSED — tercatat dengan timestamp dan user yang melakukan aksi."
              : "Every status — DRAFT, SUBMITTED, APPROVED, REJECTED, CLOSED — is logged with a timestamp and the user who performed the action.",
            screenshotKey: "goodsReceiptAudit",
            placeholder: "scorecardFlow",
          },
          {
            title: isId ? "Teruskan ke faktur supplier" : "Pass to supplier invoice",
            description: isId
              ? "Begitu barang diterima dan disetujui, konversi goods receipt ke supplier invoice dalam satu klik — data quantity dan harga langsung terbawa."
              : "Once goods are received and approved, convert the goods receipt to a supplier invoice in one click — quantity and cost data carry over automatically.",
            screenshotKey: "goodsReceiptToInvoice",
            placeholder: "invoiceFlow",
          },
        ]}
        finalTitle={isId ? "Tingkatkan standar penerimaan gudang Anda" : "Raise the standard of your warehouse receiving"}
        finalPoints={[
          {
            title: isId ? "Akurasi instan" : "Instant accuracy",
            description: isId
              ? "Cegah kesalahan sebelum barang masuk ke sistem Anda."
              : "Prevent errors before items even enter your system.",
          },
          {
            title: isId ? "Ketelusuran total" : "Total traceability",
            description: isId
              ? "Siap diaudit dengan jejak data yang tak terbantahkan."
              : "Audit-ready with an undeniable trail of evidence.",
          },
        ]}
        finalCtaLabel={isId ? "Mulai Uji Coba Gratis Anda" : "Start My Free Trial"}
      />
    </>
  );
}
