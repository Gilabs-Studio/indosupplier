type PlanCopyLocale = "en" | "id";

export interface PlanCopy {
  name: string;
  description: string;
  features: string[];
  badge?: string;
}

interface PlanCopyFallback {
  name?: string;
  description?: string;
  features?: string[];
  badge?: string;
}

const planCopyByLocale: Record<PlanCopyLocale, Record<string, PlanCopy>> = {
  en: {
    pos_growth: {
      name: "POS Modular",
      description:
        "Avoid losing transactions during rush hours. Run multi-outlet cashier, loyalty, and live sales visibility from one screen",
      features: [
        "Keep queues moving without missing orders",
        "Bring customers back without manual follow-up",
        "See outlet performance in real time before problems grow",
      ],
    },
    erp_pro: {
      name: "ERP Modular",
      description:
        "No more scattered invoices or delayed stock updates. Purchase, sales, inventory, and finance stay synchronized end to end",
      features: [
        "Prevent stock and invoice mismatches before month-end",
        "Close purchasing and sales workflows without manual reconciliation",
        "Keep finance entries consistent from the same operational flow",
      ],
    },
    crm_growth: {
      name: "CRM Modular",
      description:
        "Stop pipeline leaks. Track leads, activities, and follow-ups so opportunities do not disappear between handovers",
      features: [
        "Move deals forward with clear next actions",
        "Track team activities and targets in one place",
        "Turn more leads into revenue with predictable follow-up",
      ],
    },
    hr_growth: {
      name: "HR Modular",
      description:
        "Protect payroll and attendance accuracy. HR operations stay structured so requests and approvals do not get lost",
      features: [
        "Reduce payroll disputes with cleaner attendance records",
        "Handle leave and overtime requests with clear approval trails",
        "Keep HR evaluations and schedules consistent across teams",
      ],
    },
    growth_suite: {
      name: "Growth Suite",
      description: "One package for POS, ERP, and CRM",
      features: [
        "Connect cashier, operations, and sales pipeline in one flow",
        "Avoid cross-team blind spots from disconnected tools",
        "Scale faster with one source of truth across core units",
      ],
      badge: "coret",
    },
    ultimate_suite: {
      name: "Ultimate Suite",
      description: "All modules with priority setup & support",
      features: [
        "Run POS, ERP, CRM, and HR in one integrated stack",
        "Get priority guidance for setup, migration, and onboarding",
        "Use full reporting and API capabilities without fragmented add-ons",
      ],
      badge: "coret",
    },
    enterprise: {
      name: "Enterprise",
      description: "Custom for your business with dedicated support",
      features: [
        "Everything in Ultimate",
        "Unlimited users",
        "Dedicated SLA",
        "Custom integrations",
        "On-premise option",
      ],
    },
  },
  id: {
    pos_growth: {
      name: "POS Modular",
      description:
        "Jangan kehilangan transaksi saat outlet ramai. Kasir multi-outlet, loyalty, dan visibilitas penjualan berjalan dari satu layar",
      features: [
        "Antrean tetap lancar tanpa order terlewat",
        "Pelanggan kembali tanpa follow-up manual",
        "Kinerja outlet terlihat realtime sebelum masalah membesar",
      ],
    },
    erp_pro: {
      name: "ERP Modular",
      description:
        "Tidak ada lagi invoice tercecer atau update stok terlambat. Purchase, sales, inventory, dan finance tersinkron ujung ke ujung",
      features: [
        "Cegah selisih stok dan invoice sebelum tutup bulan",
        "Alur pembelian dan penjualan selesai tanpa rekap manual",
        "Pencatatan finance konsisten dari alur operasional yang sama",
      ],
    },
    crm_growth: {
      name: "CRM Modular",
      description:
        "Hentikan kebocoran pipeline. Lead, aktivitas, dan follow-up tercatat sehingga peluang tidak hilang saat handover",
      features: [
        "Deal bergerak dengan next action yang jelas",
        "Aktivitas tim dan target termonitor dalam satu tempat",
        "Lebih banyak lead jadi revenue dengan follow-up terukur",
      ],
    },
    hr_growth: {
      name: "HR Modular",
      description:
        "Amankan akurasi payroll dan absensi. Operasional HR tetap rapi agar request dan approval tidak terlewat",
      features: [
        "Kurangi sengketa payroll lewat data absensi yang rapi",
        "Request cuti dan lembur terlacak dengan approval yang jelas",
        "Evaluasi dan jadwal HR konsisten lintas tim",
      ],
    },
    growth_suite: {
      name: "Growth Suite",
      description: "Satu paket untuk POS, ERP, dan CRM",
      features: [
        "Kasir, operasional, dan pipeline sales terhubung dalam satu alur",
        "Hindari blind spot antar tim akibat tools yang terpisah",
        "Scale lebih cepat dengan single source of truth",
      ],
      badge: "coret",
    },
    ultimate_suite: {
      name: "Ultimate Suite",
      description: "Semua modul dengan dukungan & setup prioritas",
      features: [
        "POS, ERP, CRM, dan HR berjalan dalam satu stack terintegrasi",
        "Dapat pendampingan prioritas untuk setup, migrasi, dan onboarding",
        "Fitur laporan penuh dan API tanpa add-on yang terpecah",
      ],
      badge: "coret",
    },
    enterprise: {
      name: "Enterprise",
      description: "Kustom sesuai bisnis Anda dengan dukungan implementasi dedicated",
      features: [
        "Semua fitur Ultimate",
        "Pengguna tanpa batas",
        "SLA dedicated",
        "Integrasi kustom",
        "Opsi on-premise",
      ],
    },
  },
};

function normalizeLocale(locale?: string): PlanCopyLocale {
  const normalized = (locale ?? "en").toLowerCase();
  return normalized === "id" ? "id" : "en";
}

function normalizeFeatures(values?: string[]): string[] {
  if (!Array.isArray(values)) {
    return [];
  }
  return values.filter((value) => value.trim() !== "");
}

export function getSubscriptionPlanCopy(
  slug: string,
  locale?: string,
  fallback?: PlanCopyFallback,
): PlanCopy {
  const normalizedSlug = slug.trim().toLowerCase();
  const localeKey = normalizeLocale(locale);
  const localized = planCopyByLocale[localeKey][normalizedSlug];

  return {
    name: localized?.name ?? fallback?.name ?? slug,
    description: localized?.description ?? fallback?.description ?? fallback?.name ?? slug,
    features: localized?.features ?? normalizeFeatures(fallback?.features),
    badge: localized?.badge ?? fallback?.badge,
  };
}