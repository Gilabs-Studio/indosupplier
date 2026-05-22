import { Metadata } from "next";
import { setRequestLocale } from "next-intl/server";
import { FeatureLandingPage } from "@/components/landing/feature-landing-page";
import { buildLandingMetadata } from "@/lib/seo";

export async function generateMetadata({ params }: { params: Promise<{ locale: string }> }): Promise<Metadata> {
  const { locale } = await params;

  return buildLandingMetadata({
    locale,
    path: "/attendance",
    title:
      locale === "id"
        ? "Sistem Absensi Karyawan: Clock-in GPS, WFH, Statistik Bulanan | SalesView"
        : "Employee Attendance System: GPS Clock-in, WFH, Monthly Stats | SalesView",
    description:
      locale === "id"
        ? "Clock-in/out dengan GPS dan foto (NORMAL/WFH/FIELD WORK), kalender kehadiran, statistik bulanan (present, late, absent, overtime), dan manajemen absensi manual oleh admin — semua terhubung ke data karyawan dan evaluation."
        : "Clock-in/out with GPS and photo (NORMAL/WFH/FIELD WORK), attendance calendar, monthly stats (present, late, absent, overtime), and manual attendance management by admin — all connected to employee data and evaluation.",
    keywords:
      locale === "id"
        ? [
            "absensi karyawan",
            "attendance GPS",
            "clock in clock out",
            "WFH attendance",
            "kalender kehadiran",
            "statistik absensi bulanan",
            "absensi manual",
            "salesview attendance",
          ]
        : [
            "employee attendance",
            "GPS attendance",
            "clock in clock out",
            "WFH attendance",
            "attendance calendar",
            "monthly attendance stats",
            "manual attendance",
            "salesview attendance",
          ],
    imageAlt: "SalesView employee attendance system overview",
  });
}

export default async function AttendanceLandingPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  setRequestLocale(locale);
  const isId = locale === "id";

  const schemaOrg = {
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    "@id": `https://salesview.id/${locale}/attendance#product`,
    name: "SalesView Attendance",
    applicationCategory: "BusinessApplication",
    operatingSystem: "Web",
    description: isId
      ? "Sistem absensi karyawan dengan clock-in GPS, WFH, kalender kehadiran, dan statistik bulanan."
      : "Employee attendance system with GPS clock-in, WFH, attendance calendar, and monthly stats.",
    url: `https://salesview.id/${locale}/attendance`,
  };

  return (
    <>
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(schemaOrg) }} />
      <FeatureLandingPage
        locale={locale}
        label="SalesView Attendance"
        heroTitle={isId ? "Ubah Absensi Kuno" : "Turn Outdated Attendance"}
        heroAccent={isId ? "Menjadi Kedisiplinan Presisi" : "Into Precise Discipline"}
        heroDescription={
          isId
            ? "Selamat tinggal kecurangan absen. Lacak kehadiran dengan GPS, verifikasi foto, dan otomatisasi statistik bulanan yang terhubung langsung ke evaluasi."
            : "Say goodbye to attendance fraud. Track clock-ins with GPS, photo verification, and automated monthly stats linked directly to performance evaluations."
        }
        heroImageAlt={isId ? "Tampilan attendance SalesView" : "SalesView attendance interface"}
        heroImageKey="attendanceHero"
        introTitle={isId ? "Pantau kehadiran tim Anda dari mana saja" : "Monitor your team's attendance from anywhere"}
        introDescription={
          isId
            ? "Baik di kantor, WFH, atau di lapangan, pastikan tim Anda tetap akuntabel dengan pelacakan lokasi dan waktu yang tak terbantahkan."
            : "Whether in the office, WFH, or in the field, keep your team accountable with undeniable time and location tracking."
        }
        featureBlocks={[
          {
            title: isId ? "Clock-in / clock-out dengan GPS dan foto" : "Clock-in / clock-out with GPS and photo",
            description: isId
              ? "Karyawan clock-in dan clock-out dengan tipe NORMAL, WFH, atau FIELD WORK — lengkap dengan koordinat GPS, alamat, dan foto selfie via kamera."
              : "Employees clock-in and clock-out with NORMAL, WFH, or FIELD WORK types — complete with GPS coordinates, address, and selfie photo via camera.",
            descriptionSecondary: isId
              ? "Jika terlambat, alasan keterlambatan bisa langsung diisi tanpa meninggalkan layar."
              : "If late, the late reason can be filled in directly without leaving the screen.",
            screenshotKey: "attendanceClockIn",
            placeholder: "terminalFlow",
          },
          {
            title: isId ? "Kalender kehadiran dan statistik bulanan" : "Attendance calendar and monthly stats",
            description: isId
              ? "Lihat kehadiran dalam tampilan kalender — PRESENT, LATE, ABSENT, WFH, LEAVE. Statistik bulanan mencakup present days, late days, overtime hours, dan attendance percentage."
              : "View attendance in a calendar view — PRESENT, LATE, ABSENT, WFH, LEAVE. Monthly stats include present days, late days, overtime hours, and attendance percentage.",
            screenshotKey: "attendanceCalendar",
            placeholder: "scorecardFlow",
          },
          {
            title: isId ? "Manajemen absensi oleh admin" : "Admin attendance management",
            description: isId
              ? "Admin dapat mencatat absensi manual dengan alasan, melihat daftar hadir seluruh karyawan, dan menyetujui entri manual — semua dengan audit trail."
              : "Admins can record manual attendance with reasons, view the full employee attendance list, and approve manual entries — all with an audit trail.",
            screenshotKey: "attendanceAdmin",
            placeholder: "multiOutlet",
          },
        ]}
        finalTitle={isId ? "Tingkatkan kedisiplinan dan produktivitas tim Anda" : "Elevate your team's discipline and productivity"}
        finalPoints={[
          {
            title: isId ? "Akuntabilitas tinggi" : "High accountability",
            description: isId
              ? "Hilangkan celah kecurangan dengan verifikasi ganda."
              : "Close loopholes with dual verification.",
          },
          {
            title: isId ? "Analisis instan" : "Instant analysis",
            description: isId
              ? "Pahami tren kehadiran untuk evaluasi akurat."
              : "Understand attendance trends for accurate reviews.",
          },
        ]}
        finalCtaLabel={isId ? "Mulai Uji Coba Gratis Anda" : "Start My Free Trial"}
      />
    </>
  );
}
