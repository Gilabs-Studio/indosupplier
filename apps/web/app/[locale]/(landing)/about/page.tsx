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
    path: "/about",
    title: isId ? "Tentang SalesView | Platform Bisnis Terintegrasi" : "About SalesView | Integrated Business Platform",
    description: isId
      ? "Kenali SalesView, platform all-in-one untuk ERP, CRM, HRIS, POS, dan Finance yang membantu bisnis bertumbuh lebih cepat."
      : "Learn about SalesView, an all-in-one platform for ERP, CRM, HRIS, POS, and Finance that helps businesses scale faster.",
    keywords: isId
      ? ["tentang salesview", "profil salesview", "platform bisnis indonesia", "software bisnis terintegrasi"]
      : ["about salesview", "salesview company", "integrated business software", "erp crm hris pos"],
    imageAlt: "About SalesView",
  });
}

export default async function AboutPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);

  const isId = locale === "id";

  return (
    <PageMotion className="min-h-screen bg-background pt-20">
      <main className="mx-auto max-w-5xl px-6 py-16">
        <p className="text-xs font-semibold uppercase tracking-[0.22em] text-primary">
          SalesView
        </p>
        <h1 className="mt-4 text-4xl font-semibold tracking-tight text-foreground md:text-6xl">
          {isId ? "Kisah di Balik Mesin Bisnis Anda" : "The Story Behind Your Business Engine"}
        </h1>
        <p className="mt-6 max-w-3xl text-lg leading-relaxed text-muted-foreground">
          {isId
            ? "Kami tidak hanya membangun perangkat lunak; kami membangun fondasi agar bisnis Anda dapat berkembang tanpa batas. Kami percaya bahwa pertumbuhan tidak seharusnya dihambat oleh sistem yang rumit."
            : "We don't just build software; we build the foundation for your business to scale limitlessly. We believe that growth shouldn't be bottlenecked by complex systems."}
        </p>

        <section className="mt-12 grid gap-6 md:grid-cols-3">
          <article className="rounded-2xl border border-border bg-card p-6">
            <h2 className="text-lg font-semibold text-foreground">{isId ? "Misi" : "Mission"}</h2>
            <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
              {isId
                ? "Menghapus kebingungan operasional dan memberikan Anda kendali penuh untuk mendorong keuntungan, bukan sekadar mengelola administrasi."
                : "To eliminate operational chaos and give you absolute control to drive profits, rather than just managing administration."}
            </p>
          </article>
          <article className="rounded-2xl border border-border bg-card p-6">
            <h2 className="text-lg font-semibold text-foreground">{isId ? "Fokus" : "Focus"}</h2>
            <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
              {isId
                ? "Menciptakan satu ekosistem tanpa gesekan di mana setiap tim — penjualan, keuangan, SDM, dan operasional — bekerja dari sumber kebenaran yang sama."
                : "Creating a frictionless ecosystem where every team — sales, finance, HR, and operations — works from a single source of truth."}
            </p>
          </article>
          <article className="rounded-2xl border border-border bg-card p-6">
            <h2 className="text-lg font-semibold text-foreground">{isId ? "Nilai" : "Value"}</h2>
            <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
              {isId
                ? "Kejernihan tanpa kompromi, inovasi yang memprioritaskan pengguna, dan komitmen mutlak untuk mengubah tantangan bisnis Anda menjadi peluang pertumbuhan."
                : "Uncompromising clarity, user-first innovation, and an absolute commitment to turning your business challenges into growth opportunities."}
            </p>
          </article>
        </section>

        <div className="mt-12">
          <Link
            href="/pricing"
            className="inline-flex rounded-full bg-primary px-6 py-3 text-sm font-medium text-primary-foreground hover:bg-primary/90"
          >
            {isId ? "Mulai Transformasi Bisnis Anda" : "Start Your Business Transformation"}
          </Link>
        </div>
      </main>
    </PageMotion>
  );
}
