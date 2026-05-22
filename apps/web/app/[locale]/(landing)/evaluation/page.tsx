import { Metadata } from "next";
import { setRequestLocale } from "next-intl/server";
import { FeatureLandingPage } from "@/components/landing/feature-landing-page";
import { buildLandingMetadata } from "@/lib/seo";

export async function generateMetadata({ params }: { params: Promise<{ locale: string }> }): Promise<Metadata> {
  const { locale } = await params;

  return buildLandingMetadata({
    locale,
    path: "/evaluation",
    title:
      locale === "id"
        ? "Sistem Evaluasi Kinerja: Group, Kriteria & Penilaian Karyawan | SalesView"
        : "Performance Evaluation System: Groups, Criteria & Employee Appraisal | SalesView",
    description:
      locale === "id"
        ? "Bangun group evaluasi dengan kriteria berbobot, lakukan penilaian SELF dan MANAGER per periode, lihat skor tertimbang per karyawan, dan lacak perubahan lewat audit trail — sistem appraisal yang terstruktur dan terukur."
        : "Build evaluation groups with weighted criteria, perform SELF and MANAGER appraisals per period, see weighted scores per employee, and track changes via audit trail — a structured and measurable appraisal system.",
    keywords:
      locale === "id"
        ? [
            "sistem evaluasi kinerja",
            "kriteria penilaian berbobot",
            "appraisal karyawan",
            "evaluasi SELF MANAGER",
            "skor KPI karyawan",
            "audit trail evaluasi",
            "group evaluasi HR",
            "salesview evaluation",
          ]
        : [
            "performance evaluation system",
            "weighted appraisal criteria",
            "employee appraisal",
            "SELF MANAGER evaluation",
            "employee KPI score",
            "evaluation audit trail",
            "HR evaluation group",
            "salesview evaluation",
          ],
    imageAlt: "SalesView evaluation employee appraisal overview",
  });
}

export default async function EvaluationLandingPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  setRequestLocale(locale);
  const isId = locale === "id";

  const schemaOrg = {
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    "@id": `https://salesview.id/${locale}/evaluation#product`,
    name: "SalesView Evaluation",
    applicationCategory: "BusinessApplication",
    operatingSystem: "Web",
    description: isId
      ? "Sistem evaluasi kinerja dengan group kriteria berbobot, penilaian SELF & MANAGER, dan audit trail."
      : "Performance evaluation system with weighted criteria groups, SELF & MANAGER appraisals, and audit trail.",
    url: `https://salesview.id/${locale}/evaluation`,
  };

  return (
    <>
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(schemaOrg) }} />
      <FeatureLandingPage
        locale={locale}
        label="SalesView Evaluation"
        heroTitle={isId ? "Ubah Penilaian Subjektif" : "Turn Subjective Appraisals"}
        heroAccent={isId ? "Menjadi Pertumbuhan Terukur" : "Into Measurable Growth"}
        heroDescription={
          isId
            ? "Hilangkan bias dalam evaluasi kinerja. Bangun kriteria berbobot, libatkan penilaian mandiri & atasan, dan dorong produktivitas dengan matriks yang jelas."
            : "Eliminate bias in performance reviews. Build weighted criteria, involve self & manager appraisals, and drive productivity with clear metrics."
        }
        heroImageAlt={isId ? "Tampilan evaluasi SalesView" : "SalesView evaluation interface"}
        heroImageKey="evaluationHero"
        introTitle={isId ? "Ubah ulasan kinerja menjadi pendorong semangat" : "Turn performance reviews into morale boosters"}
        introDescription={
          isId
            ? "Kerangka kerja evaluasi kami yang objektif mengubah diskusi tahunan yang ditakuti menjadi sesi perencanaan karier yang konstruktif dan berpusat pada data."
            : "Our objective evaluation framework turns dreaded annual discussions into constructive, data-centric career planning sessions."
        }
        featureBlocks={[
          {
            title: isId ? "Group & kriteria evaluasi berbobot" : "Weighted evaluation groups & criteria",
            description: isId
              ? "Buat group evaluasi (misal: Technical Skills, Soft Skills) lalu tambahkan kriteria dengan bobot (weight) dan skor maksimal (max_score). Total weight per group divalidasi agar tidak kelebihan."
              : "Create evaluation groups (e.g. Technical Skills, Soft Skills) then add criteria with weight and max score. Total weight per group is validated to prevent overflow.",
            descriptionSecondary: isId
              ? "Group bisa diaktifkan/nonaktifkan (is_active) tanpa menghapus data."
              : "Groups can be toggled active/inactive without deleting data.",
            screenshotKey: "evaluationCriteria",
            placeholder: "scorecardFlow",
          },
          {
            title: isId ? "Penilaian SELF & MANAGER per periode" : "SELF & MANAGER appraisal per period",
            description: isId
              ? "Jalankan evaluasi per karyawan dalam periode tertentu (period_start–period_end). Evaluator memberikan skor per kriteria — sistem otomatis menghitung weighted_score dan overall_score."
              : "Run evaluations per employee in a specific period (period_start–period_end). Evaluator scores each criterion — system auto-calculates weighted_score and overall_score.",
            descriptionSecondary: isId
              ? "Dua tipe evaluasi: SELF (karyawan menilai diri sendiri) dan MANAGER (atasan menilai)."
              : "Two evaluation types: SELF (employee self-rating) and MANAGER (supervisor rating).",
            screenshotKey: "evaluationEmployee",
            placeholder: "invoiceFlow",
          },
          {
            title: isId ? "Group evaluasi siap pakai" : "Ready-to-use evaluation groups",
            description: isId
              ? "Buat dan kelola group evaluasi — atur nama, deskripsi, dan status aktif/nonaktif. Setiap group menampung kriteria berbobot yang siap digunakan untuk penilaian karyawan."
              : "Create and manage evaluation groups — set name, description, and active/inactive status. Each group holds weighted criteria ready for employee appraisals.",
            descriptionSecondary: isId
              ? "Group bisa dinonaktifkan tanpa menghapus data historis penilaian."
              : "Groups can be deactivated without deleting historical appraisal data.",
            screenshotKey: "evaluationGroup",
            placeholder: "multiOutlet",
          },
        ]}
        finalTitle={isId ? "Buka potensi maksimal setiap karyawan Anda" : "Unlock the maximum potential of every employee"}
        finalPoints={[
          {
            title: isId ? "Metrik yang adil" : "Fair metrics",
            description: isId
              ? "Pastikan setiap karyawan dinilai secara objektif."
              : "Ensure every employee is judged objectively.",
          },
          {
            title: isId ? "Otomatisasi skor" : "Automated scoring",
            description: isId
              ? "Hemat waktu dengan perhitungan yang tidak pernah salah."
              : "Save time with calculations that never miss.",
          },
        ]}
        finalCtaLabel={isId ? "Mulai Uji Coba Gratis Anda" : "Start My Free Trial"}
      />
    </>
  );
}
