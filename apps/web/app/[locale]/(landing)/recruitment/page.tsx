import { Metadata } from "next";
import { setRequestLocale } from "next-intl/server";
import { FeatureLandingPage } from "@/components/landing/feature-landing-page";
import { buildLandingMetadata } from "@/lib/seo";

export async function generateMetadata({ params }: { params: Promise<{ locale: string }> }): Promise<Metadata> {
  const { locale } = await params;

  return buildLandingMetadata({
    locale,
    path: "/recruitment",
    title:
      locale === "id"
        ? "Sistem Recruitment Pipeline: Kanban Kandidat, Approval & Convert | SalesView"
        : "Recruitment Pipeline System: Candidate Kanban, Approval & Convert | SalesView",
    description:
      locale === "id"
        ? "Kelola recruitment request dengan approval, lihat kandidat dalam kanban board per tahapan, pindahkan antar stage, dan convert applicant yang diterima menjadi employee — semua terhubung ke master data karyawan."
        : "Manage recruitment requests with approval, view candidates in a kanban board per stage, move between stages, and convert accepted applicants into employees — all connected to employee master data.",
    keywords:
      locale === "id"
        ? [
            "sistem recruitment",
            "kanban kandidat",
            "pipeline recruitment",
            "convert applicant karyawan",
            "approval hiring",
            "rekrutmen HR",
            "tahapan seleksi",
            "salesview recruitment",
          ]
        : [
            "recruitment system",
            "candidate kanban",
            "recruitment pipeline",
            "convert applicant to employee",
            "hiring approval",
            "HR recruitment",
            "candidate selection stages",
            "salesview recruitment",
          ],
    imageAlt: "SalesView recruitment pipeline overview",
  });
}

export default async function RecruitmentLandingPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  setRequestLocale(locale);
  const isId = locale === "id";

  const schemaOrg = {
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    "@id": `https://salesview.id/${locale}/recruitment#product`,
    name: "SalesView Recruitment",
    applicationCategory: "BusinessApplication",
    operatingSystem: "Web",
    description: isId
      ? "Sistem recruitment dengan kanban kandidat, approval request, dan konversi applicant ke employee."
      : "Recruitment system with candidate kanban, request approval, and applicant-to-employee conversion.",
    url: `https://salesview.id/${locale}/recruitment`,
  };

  return (
    <>
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(schemaOrg) }} />
      <FeatureLandingPage
        locale={locale}
        label="SalesView Recruitment"
        heroTitle={isId ? "Ubah Rekrutmen Lambat" : "Turn Slow Recruiting"}
        heroAccent={isId ? "Menjadi Mesin Akuisisi Bakat" : "Into Talent Acquisition Machines"}
        heroDescription={
          isId
            ? "Jangan biarkan kandidat terbaik lari ke kompetitor. Kelola permintaan, visualisasikan pelamar di papan Kanban, dan otomatisasi onboarding dengan sekali klik."
            : "Don't let top candidates run to competitors. Manage requests, visualize applicants on a Kanban board, and automate onboarding with a single click."
        }
        heroImageAlt={isId ? "Tampilan recruitment SalesView" : "SalesView recruitment interface"}
        heroImageKey="recruitmentHero"
        introTitle={isId ? "Dapatkan talenta terbaik lebih cepat dari sebelumnya" : "Secure top talent faster than ever before"}
        introDescription={
          isId
            ? "Sistem rekrutmen kami yang ramping memangkas birokrasi, sehingga Anda bisa fokus menemukan, mengevaluasi, dan merekrut orang yang tepat."
            : "Our streamlined recruitment system cuts the red tape, so you can focus on finding, evaluating, and hiring the right people."
        }
        featureBlocks={[
          {
            title: isId ? "Kanban board kandidat per tahapan" : "Candidate kanban board per stage",
            description: isId
              ? "Lihat semua applicant dalam tampilan kanban — setiap kolom mewakili satu tahapan seleksi. Pindahkan kandidat antar stage dengan drag atau klik move."
              : "View all applicants in a kanban board — each column represents a selection stage. Move candidates between stages via drag or click move.",
            descriptionSecondary: isId
              ? "Detail kandidat tampil di sheet samping tanpa meninggalkan kanban."
              : "Applicant detail shows in a side sheet without leaving the kanban.",
            screenshotKey: "recruitmentKanban",
            placeholder: "multiOutlet",
          },
          {
            title: isId ? "Recruitment request dengan approval" : "Recruitment request with approval",
            description: isId
              ? "Buat permintaan rekrutmen — pilih divisi, posisi, jumlah, employment type (FULL_TIME/PART_TIME/CONTRACT/INTERN), salary range, priority, dan job description."
              : "Create recruitment requests — select division, position, count, employment type (FULL_TIME/PART_TIME/CONTRACT/INTERN), salary range, priority, and job description.",
            descriptionSecondary: isId
              ? "Status mengalir dari DRAFT → PENDING → APPROVED → OPEN → CLOSED dengan audit lengkap."
              : "Status flows from DRAFT → PENDING → APPROVED → OPEN → CLOSED with full audit.",
            screenshotKey: "recruitmentRequest",
            placeholder: "invoiceFlow",
          },
          {
            title: isId ? "Convert applicant ke data karyawan" : "Convert applicant to employee",
            description: isId
              ? "Begitu kandidat diterima, convert langsung ke employee — data applicant dijembatani ke form karyawan tanpa input ulang."
              : "Once a candidate is accepted, convert directly to employee — applicant data bridges to the employee form without re-entry.",
            screenshotKey: "recruitmentConvert",
            placeholder: "timelineFlow",
          },
        ]}
        finalTitle={isId ? "Bangun tim impian Anda tanpa hambatan administratif" : "Build your dream team without the administrative roadblocks"}
        finalPoints={[
          {
            title: isId ? "Hiring visual" : "Visual hiring",
            description: isId
              ? "Lihat seluruh pipeline rekrutmen dalam satu lirikan."
              : "See your entire recruitment pipeline at a glance.",
          },
          {
            title: isId ? "Transisi instan" : "Instant transition",
            description: isId
              ? "Ubah kandidat menjadi karyawan tanpa ketik ulang data."
              : "Turn candidates into employees without retyping data.",
          },
        ]}
        finalCtaLabel={isId ? "Mulai Uji Coba Gratis Anda" : "Start My Free Trial"}
      />
    </>
  );
}
