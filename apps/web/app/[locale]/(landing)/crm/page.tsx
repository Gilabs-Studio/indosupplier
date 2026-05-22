import { Metadata } from "next";
import { PageMotion } from "@/components/motion/page-motion";
import { setRequestLocale } from "next-intl/server";
// ArrowRight not used — removed
import { Link } from "@/i18n/routing";
import { ThemeAwareImage } from "@/components/landing/theme-aware-image";
import { LANDING_THEME_IMAGES_BY_FEATURE } from "@/lib/landing-theme-images";
import { buildLandingMetadata } from "@/lib/seo";

export async function generateMetadata({ params }: { params: Promise<{ locale: string }> }): Promise<Metadata> {
  const { locale } = await params;
  
  const title = locale === 'id' 
    ? `Software CRM untuk Pipeline & Follow-up Sales | SalesView` 
    : `CRM Software for Pipeline & Sales Follow-ups | SalesView`;
    
  const desc = locale === 'id'
    ? `Kelola pipeline, aktivitas tim sales, dan follow-up prospek dalam satu CRM yang terhubung dengan penawaran dan sales order.`
    : `Manage pipeline, sales activities, and lead follow-ups in one CRM connected to quotations and sales orders.`;

  return buildLandingMetadata({
    locale,
    path: "/crm",
    title,
    description: desc,
    keywords: locale === "id"
      ? [
          "software crm",
          "crm pipeline sales",
          "follow up prospek",
          "crm b2b indonesia",
          "salesview crm",
        ]
      : [
          "crm software",
          "sales pipeline crm",
          "lead follow up",
          "b2b crm platform",
          "salesview crm",
        ],
    imageAlt: "SalesView CRM overview",
  });
}

export default async function CrmLandingPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  setRequestLocale(locale);

  // LocalBusiness & SoftwareApplication tailored SEO Schema
  const schemaOrg = {
    "@context": "https://schema.org",
    "@type": ["LocalBusiness", "SoftwareApplication"],
    "@id": `https://salesview.id/${locale}/crm#business`,
    "name": locale === 'id' ? "Sistem CRM Terbaik Terintegrasi" : "Best Integrated CRM System",
    "applicationCategory": "BusinessApplication",
    "operatingSystem": "Web",
    "description": locale === 'id' 
      ? `Aplikasi CRM enterprise terbaik. Bebas, tanpa batasan pengguna, untuk penjualan yang lebih modern.`
      : `Best enterprise CRM application. Seamless, without limits, designed for modern sales.`,
    "address": {
      "@type": "PostalAddress",
      "addressLocality": "Semarang",
      "addressRegion": "Jawa Tengah",
      "addressCountry": "ID"
    },
    "geo": {
      "@type": "GeoCoordinates",
      "latitude": -7.005145,
      "longitude": 110.438125
    },
    "url": `https://salesview.id/${locale}/crm`,
    "telephone": "+628111222333"
  };

  const isId = locale === 'id';

  return (
    <>
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(schemaOrg) }} />
      <PageMotion className="flex flex-col min-h-screen bg-background pt-16">
        
        {/* APPLE-LIKE HERO (Super Minimalist) */}
        <section className="pt-24 pb-16 px-6 max-w-5xl mx-auto text-center">
          <p className="text-xs font-semibold tracking-widest uppercase text-muted-foreground mb-6">
            SalesView CRM
          </p>
          <h1 className="text-5xl md:text-7xl font-semibold tracking-tight text-foreground leading-[1.1] mb-6">
            {isId ? `Ubah Lead Dingin Menjadi` : `Turn Cold Leads Into`}
            <br />
            <span className="text-primary font-accent font-medium italic text-[4.5rem] md:text-[6.5rem] leading-none inline-block mt-2">
              {isId ? `Penjualan Nyata` : `Real Sales`}
            </span>
          </h1>
          <p className="text-xl md:text-2xl font-light text-muted-foreground/80 max-w-3xl mx-auto leading-relaxed mb-10">
            {isId 
              ? `Berdayakan tim sales Anda dengan CRM berbasis AI untuk melacak setiap interaksi, memprediksi pendapatan, dan menutup transaksi lebih cepat.`
              : `Empower your sales team with an AI-native CRM to track every interaction, forecast revenue accurately, and close deals faster.`
            }
          </p>
          
          <div className="flex flex-wrap items-center justify-center gap-4">
            <Link 
              href="/register"
              className="rounded-full bg-primary px-8 py-4 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-transform hover:scale-105 shadow-sm"
            >
              {isId ? "Coba CRM Gratis" : "Start My Free Trial"}
            </Link>
            <Link 
              href="/login"
              className="rounded-full bg-secondary text-secondary-foreground px-8 py-4 text-sm font-medium hover:bg-secondary/80 transition-colors"
            >
              {isId ? "Jadwalkan Demo" : "Book a Demo"}
            </Link>
          </div>
          <p className="mt-6 text-sm text-muted-foreground">
            {isId ? "Bebas, selamanya, tanpa batasan pengguna." : "Free, forever, absolutely no user limits."}
          </p>
        </section>
        
        {/* HERO IMAGE SHOWCASE (Clean no heavy borders) */}
        <section className="px-6 max-w-7xl mx-auto w-full pb-24">
          <div className="relative aspect-video w-full rounded-2xl overflow-hidden shadow-[0_20px_50px_rgba(0,0,0,0.1)] border border-border/40 bg-card">
            {/* Using abstract Apple-like minimalist representation or screenshot if available */}
            <div className="absolute inset-0 bg-linear-to-tr from-background to-secondary/30" />
            <ThemeAwareImage
              lightSrc={LANDING_THEME_IMAGES_BY_FEATURE.crmHero.light}
              darkSrc={LANDING_THEME_IMAGES_BY_FEATURE.crmHero.dark}
              alt="CRM Interface" 
              fill 
              className="object-cover object-top"
              priority
            />
          </div>
        </section>

        {/* FEATURE INTRODUCTION GRID */}
        <section className="bg-secondary/20 py-24 md:py-32">
          <div className="max-w-5xl mx-auto px-6">
            <div className="text-center mb-20 md:mb-32 max-w-3xl mx-auto">
              <h2 className="text-4xl md:text-5xl font-semibold tracking-tight text-foreground mb-6">
                {isId ? "Kuasai seluruh alur penjualan Anda tanpa repot" : "Master your entire sales pipeline effortlessly"}
              </h2>
              <p className="text-xl text-muted-foreground font-light">
                {isId ? "Jangan biarkan ada prospek yang lolos. Kelola semua peluang Anda dalam satu tampilan visual yang mudah digunakan." : "Never let a prospect slip away. Manage all your opportunities in one easy-to-use visual interface."}
              </p>
            </div>

            <div className="space-y-32">
              
              {/* Feature 1 */}
              <div className="flex flex-col gap-12 lg:gap-24 items-center lg:flex-row">
                <div className="lg:w-1/2 space-y-6 text-center lg:text-left">
                  <h3 className="text-3xl md:text-4xl font-semibold tracking-tight text-foreground">
                    {isId ? "Semua tertata rapi dan efisien" : "Neatly organized and efficient"}
                  </h3>
                  <p className="text-lg md:text-xl text-muted-foreground font-light leading-relaxed">
                    {isId 
                      ? `Tampilan kanban mengatur opportunity berdasarkan tahap mereka. Tarik dan lepas kartu di pipeline untuk menggerakkanya dari tahap ke tahap.`
                      : `The kanban view organizes opportunities by their stage. Drag and drop cards in your pipeline to easily move them from stage to stage.`
                    }
                  </p>
                </div>
                
                <div className="lg:w-1/2 w-full">
                  <div className="relative aspect-video rounded-3xl overflow-hidden bg-background border border-border shadow-sm flex items-center justify-center p-6 md:p-8">
                    <ThemeAwareImage
                      lightSrc={LANDING_THEME_IMAGES_BY_FEATURE.crmKanbanPipeline.light}
                      darkSrc={LANDING_THEME_IMAGES_BY_FEATURE.crmKanbanPipeline.dark}
                      alt={isId ? "CRM kanban pipeline" : "CRM kanban pipeline"}
                      fill
                      className="object-cover object-top"
                    />
                  </div>
                </div>
              </div>

               {/* Feature 2 */}
              <div className="flex flex-col gap-12 lg:gap-24 items-center lg:flex-row-reverse">
                <div className="lg:w-1/2 space-y-6 text-center lg:text-left">
                  <h3 className="text-3xl md:text-4xl font-semibold tracking-tight text-foreground">
                    {isId ? "Tidak pernah lagi lupa follow-up" : "Never forget a follow-up"}
                  </h3>
                  <p className="text-lg md:text-xl text-muted-foreground font-light leading-relaxed">
                    {isId 
                      ? `Jadwalkan panggilan, rapat, email, atau penawaran, dan SalesView AI secara otomatis merencanakan aktivitas Anda berikutnya.`
                      : `Schedule calls, meetings, mailings, or quotations, and SalesView AI automatically plans your next activity.`
                    }
                  </p>
                  <p className="text-lg md:text-xl text-muted-foreground font-light leading-relaxed">
                    {isId 
                      ? `Komunikasi adalah kunci untuk meningkatkan hubungan pelanggan. Semua masuk ke satu tempat agar mudah diakses.`
                      : `Communication is key to building customer relationships. All touchpoints merge in one place for easy access.`
                    }
                  </p>
                </div>
                
                <div className="lg:w-1/2 w-full">
                  <div className="relative aspect-video rounded-3xl overflow-hidden bg-background border border-border shadow-sm flex items-center justify-center p-6 md:p-8">
                    <ThemeAwareImage
                      lightSrc={LANDING_THEME_IMAGES_BY_FEATURE.crmActivityFollowup.light}
                      darkSrc={LANDING_THEME_IMAGES_BY_FEATURE.crmActivityFollowup.dark}
                      alt={isId ? "CRM aktivitas dan follow-up" : "CRM activities and follow-up"}
                      fill
                      className="object-cover object-top"
                    />
                  </div>
                </div>
              </div>

               {/* Feature 3 */}
              <div className="flex flex-col gap-12 lg:gap-24 items-center lg:flex-row">
                <div className="lg:w-1/2 space-y-6 text-center lg:text-left">
                  <h3 className="text-3xl md:text-4xl font-semibold tracking-tight text-foreground">
                    {isId ? "Pemetaan area untuk prioritas lapangan" : "Area mapping for field priority"}
                  </h3>
                  <p className="text-lg md:text-xl text-muted-foreground font-light leading-relaxed">
                    {isId 
                      ? `Visualkan lead, deal, dan aktivitas di peta area agar tim sales tahu wilayah mana yang paling aktif dan mana yang perlu ditingkatkan.`
                      : `Visualize leads, deals, and activities on a location map so sales teams can prioritize active regions and improve underperforming areas.`
                    }
                  </p>
                </div>
                
                <div className="lg:w-1/2 w-full">
                  <div className="relative aspect-video rounded-3xl overflow-hidden bg-background border border-border shadow-sm flex items-center justify-center p-6 md:p-8">
                    <ThemeAwareImage
                      lightSrc={LANDING_THEME_IMAGES_BY_FEATURE.crmAreaMapping.light}
                      darkSrc={LANDING_THEME_IMAGES_BY_FEATURE.crmAreaMapping.dark}
                      alt={isId ? "CRM area mapping" : "CRM area mapping"}
                      fill
                      className="object-cover object-top"
                    />
                  </div>
                </div>
              </div>

            </div>
          </div>
        </section>

        <section className="px-6 py-16 md:py-20">
          <div className="mx-auto max-w-4xl rounded-3xl border border-border/50 bg-card p-8 md:p-10 text-center">
            <p className="text-xs font-semibold tracking-widest uppercase text-muted-foreground mb-4">SalesView Apps</p>
            <h3 className="text-2xl md:text-3xl font-semibold tracking-tight text-foreground">
              {isId ? "Jelajahi Modul Terkait" : "Explore Related Modules"}
            </h3>
            <p className="mt-3 text-muted-foreground max-w-2xl mx-auto">
              {isId
                ? "Perkuat alur revenue dengan modul sales lain yang terintegrasi langsung dengan CRM."
                : "Strengthen your revenue workflow with sales modules directly integrated with CRM."}
            </p>
            <div className="mt-8 flex flex-wrap justify-center gap-3">
              <Link href="/quotations" className="rounded-full border border-border bg-background px-4 py-2 text-sm font-medium hover:bg-secondary">
                {isId ? "Quotations" : "Quotations"}
              </Link>
              <Link href="/sales" className="rounded-full border border-border bg-background px-4 py-2 text-sm font-medium hover:bg-secondary">
                {isId ? "Sales Orders" : "Sales Orders"}
              </Link>
              <Link href="/invoicing" className="rounded-full border border-border bg-background px-4 py-2 text-sm font-medium hover:bg-secondary">
                {isId ? "Invoicing" : "Invoicing"}
              </Link>
            </div>
          </div>
        </section>

        {/* BOTTOM CTA */}
        <section className="py-24 md:py-32 px-6 max-w-4xl mx-auto text-center space-y-8">
          <h2 className="text-4xl md:text-5xl font-semibold tracking-tight text-foreground">
            {isId ? "Siap untuk melipatgandakan tingkat konversi Anda?" : "Ready to double your conversion rate?"}
          </h2>
          <div className="grid md:grid-cols-2 gap-8 text-left mt-16 max-w-3xl mx-auto">
             <div>
                <h4 className="font-semibold text-lg mb-2">{isId ? "Aktifkan otomatisasi" : "Enable automation"}</h4>
                <p className="text-muted-foreground">{isId ? "Otomatiskan pembuatan lead, penugasan tim dan jadwal kegiatan agar lebih fokus pada yang penting." : "Automate lead creation, team assignment, and activities to focus on what matters."}</p>
             </div>
             <div>
                <h4 className="font-semibold text-lg mb-2">{isId ? "Lead scoring bertenaga AI" : "AI-Powered Lead Scoring"}</h4>
                <p className="text-muted-foreground">{isId ? "Terlalu banyak lead? AI kami menilai lead agar Anda selalu bekerja pada prioritas yang tepat." : "Too many leads? Our AI scores leads so you always work on the right priorities."}</p>
             </div>
          </div>
          <div className="pt-12">
            <Link 
              href="/register"
              className="rounded-full bg-foreground px-8 py-4 text-sm font-medium text-background hover:bg-foreground/90 transition-transform hover:scale-105 inline-flex shadow-xl"
            >
              {isId ? "Lihat semua fitur CRM" : "See all CRM features"}
            </Link>
          </div>
        </section>

      </PageMotion>
    </>
  );
}
