import { Metadata } from "next";
import { setRequestLocale } from "next-intl/server";
import { FeatureLandingPage } from "@/components/landing/feature-landing-page";
import { buildLandingMetadata } from "@/lib/seo";

export async function generateMetadata({ params }: { params: Promise<{ locale: string }> }): Promise<Metadata> {
  const { locale } = await params;

  return buildLandingMetadata({
    locale,
    path: "/employees",
    title:
      locale === "id"
        ? "Manajemen Data Karyawan: Kontrak, Riwayat Pendidikan & Sertifikasi | SalesView"
        : "Employee Data Management: Contracts, Education & Certifications | SalesView",
    description:
      locale === "id"
        ? "Kelola profil karyawan terpusat dengan kontrak PKWTT/PKWT, riwayat pendidikan, sertifikasi, aset, tanda tangan digital, dan assignment area/outlet/warehouse — semua dalam satu master data yang terhubung ke attendance dan evaluation."
        : "Manage centralized employee profiles with PKWTT/PKWT contracts, education history, certifications, assets, digital signatures, and area/outlet/warehouse assignment — all in one master record connected to attendance and evaluation.",
    keywords:
      locale === "id"
        ? [
            "data karyawan",
            "master data karyawan",
            "kontrak PKWTT PKWT",
            "riwayat pendidikan karyawan",
            "sertifikasi karyawan",
            "aset karyawan",
            "profil karyawan terpusat",
            "salesview employees",
          ]
        : [
            "employee data",
            "employee master data",
            "PKWTT PKWT contract",
            "employee education history",
            "employee certifications",
            "employee assets",
            "employee profile management",
            "salesview employees",
          ],
    imageAlt: "SalesView employee management overview",
  });
}

export default async function EmployeesLandingPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  setRequestLocale(locale);
  const isId = locale === "id";

  const schemaOrg = {
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    "@id": `https://salesview.id/${locale}/employees#product`,
    name: "SalesView Employees",
    applicationCategory: "BusinessApplication",
    operatingSystem: "Web",
    description: isId
      ? "Sistem manajemen data karyawan terpusat dengan kontrak, riwayat pendidikan, sertifikasi, dan aset."
      : "Centralized employee data management with contracts, education history, certifications, and assets.",
    url: `https://salesview.id/${locale}/employees`,
  };

  return (
    <>
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(schemaOrg) }} />
      <FeatureLandingPage
        locale={locale}
        label="SalesView Employees"
        heroTitle={isId ? "Ubah Tumpukan Dokumen" : "Turn Piles of Documents"}
        heroAccent={isId ? "Menjadi Profil Digital Cerdas" : "Into Smart Digital Profiles"}
        heroDescription={
          isId
            ? "Pusatkan seluruh data karyawan Anda. Kelola kontrak, riwayat pendidikan, aset, dan sertifikasi dalam satu sumber kebenaran yang aman dan mudah diakses."
            : "Centralize all your employee data. Manage contracts, education history, assets, and certifications in one secure, easily accessible single source of truth."
        }
        heroImageAlt={isId ? "Tampilan data karyawan SalesView" : "SalesView employee data interface"}
        heroImageKey="employeeHero"
        introTitle={isId ? "Bangun fondasi HR yang kuat dan bebas repot" : "Build a strong, hassle-free HR foundation"}
        introDescription={
          isId
            ? "Data karyawan yang terstruktur sempurna menghilangkan kebingungan administratif, sehingga Anda bisa fokus mengembangkan talenta terbaik Anda."
            : "Perfectly structured employee data eliminates administrative confusion, so you can focus on developing your best talent."
        }
        featureBlocks={[
          {
            title: isId ? "Profil karyawan terstruktur per tab" : "Structured employee profiles with tabs",
            description: isId
              ? "Lihat dan kelola profil karyawan dalam tampilan tab: Overview, Employment, Contract History, Signature, Education, Certifications, Assets, Areas, dan Scope Outlet & Warehouse."
              : "View and manage employee profiles across tabs: Overview, Employment, Contract History, Signature, Education, Certifications, Assets, Areas, and Outlet & Warehouse Scope.",
            screenshotKey: "employeeList",
            placeholder: "multiOutlet",
          },
          {
            title: isId ? "Kontrak PKWTT, PKWT, dan Intern" : "PKWTT, PKWT, and Intern contracts",
            description: isId
              ? "Buat kontrak dengan tipe PKWTT, PKWT, atau Intern — lengkap dengan start date, end date, status, dan aksi renew, terminate, atau correct jika ada perubahan."
              : "Create contracts with PKWTT, PKWT, or Intern types — complete with start date, end date, status, and renew, terminate, or correct actions if changes occur.",
            screenshotKey: "employeeContract",
            placeholder: "timelineFlow",
          },
          {
            title: isId ? "Riwayat pendidikan, sertifikasi, dan aset" : "Education, certification, and asset history",
            description: isId
              ? "Tambahkan riwayat pendidikan, sertifikasi, dan aset yang diberikan ke karyawan — semua tersimpan dalam timeline yang bisa ditelusuri kapan saja."
              : "Add education history, certifications, and assets assigned to employees — all stored in a traceable timeline.",
            descriptionSecondary: isId
              ? "Data ini langsung tersedia untuk proses evaluation dan kebutuhan audit SDM."
              : "This data is directly available for evaluation processes and HR audit needs.",
            screenshotKey: "employeeHistory",
            placeholder: "scorecardFlow",
          },
        ]}
        finalTitle={isId ? "Kelola tim Anda dengan kepastian data yang absolut" : "Manage your team with absolute data certainty"}
        finalPoints={[
          {
            title: isId ? "Satu sumber kebenaran" : "Single source of truth",
            description: isId
              ? "Semua data dari kontrak hingga aset di tempat yang sama."
              : "All data from contracts to assets in one place.",
          },
          {
            title: isId ? "Otomatisasi tanpa batas" : "Limitless automation",
            description: isId
              ? "Kelola pembaruan dengan cepat dan tanpa stres."
              : "Handle renewals quickly and without the stress.",
          },
        ]}
        finalCtaLabel={isId ? "Mulai Uji Coba Gratis Anda" : "Start My Free Trial"}
      />
    </>
  );
}
