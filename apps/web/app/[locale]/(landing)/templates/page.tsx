import type { Metadata } from "next";
import { setRequestLocale } from "next-intl/server";
import { Link } from "@/i18n/routing";
import { PageMotion } from "@/components/motion/page-motion";
import { buildLandingMetadata } from "@/lib/seo";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string }>;
}): Promise<Metadata> {
  const { locale } = await params;
  const isId = locale === "id";

  return buildLandingMetadata({
    locale,
    path: "/templates",
    title: isId ? "Template Bisnis Sales, Finance, HR & Inventory | SalesView" : "Business Templates for Sales, Finance, HR & Inventory | SalesView",
    description: isId
      ? "Akses template bisnis siap pakai untuk sales, finance, HR, inventory, dan operasional agar tim Anda bisa implementasi proses lebih cepat dan rapi."
      : "Access ready-to-use templates for sales, finance, HR, inventory, and operations so your team can implement processes faster and with better consistency.",
    keywords: isId
      ? ["template bisnis", "template sales", "template invoice", "template laporan keuangan"]
      : ["business templates", "sales templates", "invoice template", "financial report template"],
    imageAlt: "SalesView business templates",
  });
}

const TEMPLATE_ITEMS = [
  { id: "sales", badge: "Sales", color: "bg-primary/10 text-primary" },
  { id: "finance", badge: "Finance", color: "bg-secondary text-secondary-foreground" },
  { id: "hr", badge: "HR", color: "bg-muted text-muted-foreground" },
  { id: "inventory", badge: "Inventory", color: "bg-accent text-accent-foreground" },
] as const;

export default async function TemplatesPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);

  const isId = locale === "id";

  return (
    <PageMotion className="min-h-screen bg-background pt-20">
      <main className="mx-auto max-w-6xl px-6 py-16">
        <p className="text-xs font-semibold uppercase tracking-[0.22em] text-primary">SalesView Resources</p>
        <h1 className="mt-4 text-4xl font-semibold tracking-tight text-foreground md:text-6xl">
          {isId ? "Mulai Cepat dengan Template Premium" : "Jumpstart Your Work with Premium Templates"}
        </h1>
        <p className="mt-6 max-w-3xl text-lg leading-relaxed text-muted-foreground">
          {isId
            ? "Akses perpustakaan format standar industri kami. Hemat ratusan jam dalam penyiapan operasional, pelaporan, dan standar proses."
            : "Access our library of industry-standard formats. Save hundreds of hours in operational setup, reporting, and process standardization."}
        </p>

        <section className="mt-12 grid gap-6 md:grid-cols-2">
          {TEMPLATE_ITEMS.map((item) => (
            <article key={item.id} className="rounded-2xl border border-border bg-card p-6">
              <span className={`inline-flex rounded-full px-3 py-1 text-xs font-semibold ${item.color}`}>
                {item.badge}
              </span>
              <h2 className="mt-4 text-xl font-semibold text-foreground">
                {isId
                  ? `Template Kelas Enterprise untuk ${item.badge}`
                  : `Enterprise-Grade Templates for ${item.badge}`}
              </h2>
              <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
                {isId
                  ? "Standarisasi operasional Anda seketika dengan format teruji yang dirancang untuk skalabilitas."
                  : "Instantly standardize your operations with battle-tested formats designed for scalability."}
              </p>
            </article>
          ))}
        </section>

        <div className="mt-12">
          <Link
            href="/register"
            className="inline-flex rounded-full bg-primary px-6 py-3 text-sm font-medium text-primary-foreground hover:bg-primary/90"
          >
            {isId ? "Akses Semua Template Sekarang" : "Access All Templates Now"}
          </Link>
        </div>
      </main>
    </PageMotion>
  );
}
