import { Metadata } from "next";
import { PageMotion } from "@/components/motion/page-motion";
import { setRequestLocale } from "next-intl/server";
import { Link } from "@/i18n/routing";
import { ThemeAwareImage } from "@/components/landing/theme-aware-image";
import {
  LANDING_THEME_IMAGES_BY_FEATURE,
  type LandingScreenshotKey,
} from "@/lib/landing-theme-images";
import { buildLandingMetadata } from "@/lib/seo";

export async function generateMetadata({ params }: { params: Promise<{ locale: string }> }): Promise<Metadata> {
  const { locale } = await params;
  
  const title = locale === 'id' 
    ? `Software POS Kasir untuk Retail & F&B | SalesView` 
    : `POS Software for Retail & F&B | SalesView`;
    
  const desc = locale === 'id'
    ? `Jalankan kasir POS dengan sinkronisasi stok real-time, laporan penjualan, dan dukungan multi-cabang untuk bisnis retail serta F&B.`
    : `Run POS cashier operations with real-time stock sync, sales reporting, and multi-branch support for retail and F&B businesses.`;

  return buildLandingMetadata({
    locale,
    path: "/pos",
    title,
    description: desc,
    keywords: locale === "id"
      ? ["software pos", "aplikasi kasir", "pos retail", "kasir f&b", "point of sale indonesia", "salesview pos"]
      : ["pos software", "point of sale", "retail pos", "restaurant pos", "multi-branch pos", "salesview pos"],
    imageAlt: "SalesView point-of-sale overview",
  });
}

export default async function PosLandingPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  setRequestLocale(locale);

  const schemaOrg = {
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    "@id": `https://salesview.id/${locale}/pos#product`,
    "name": locale === "id" ? "SalesView Sistem Point of Sale (POS) & Kasir" : "SalesView Point of Sale (POS) System",
    "applicationCategory": "BusinessApplication",
    "operatingSystem": "Web",
    "description": locale === "id"
      ? "Point of sale cerdas dengan stok real-time, loyalty pelanggan, dan dukungan multi-cabang untuk retail dan F&B."
      : "Smart point of sale with real-time stock, loyalty, and multi-branch support for retail and F&B.",
    "url": `https://salesview.id/${locale}/pos`,
    "offers": {
      "@type": "AggregateOffer",
      "priceCurrency": "IDR",
      "lowPrice": "0",
      "offerCount": "3"
    },
    "aggregateRating": {
      "@type": "AggregateRating",
      "ratingValue": "4.9",
      "reviewCount": "87",
      "bestRating": "5",
      "worstRating": "1"
    },
    "review": [
      {
        "@type": "Review",
        "author": { "@type": "Person", "name": "Rini Wulandari" },
        "datePublished": "2025-03-20",
        "reviewBody": "Aplikasi kasir POS terbaik yang pernah saya coba. Cepat, mudah digunakan, dan stok langsung tersinkron otomatis.",
        "reviewRating": {
          "@type": "Rating",
          "ratingValue": "5",
          "bestRating": "5",
          "worstRating": "1"
        }
      },
      {
        "@type": "Review",
        "author": { "@type": "Person", "name": "Hendra Kusuma" },
        "datePublished": "2025-04-05",
        "reviewBody": "Sistem kasir yang sangat andal untuk bisnis cafe multi-cabang kami. Laporan penjualan real-time sangat membantu manajemen.",
        "reviewRating": {
          "@type": "Rating",
          "ratingValue": "5",
          "bestRating": "5",
          "worstRating": "1"
        }
      }
    ]
  };

  const isId = locale === 'id';
  const featureBlocks: Array<{
    title: string;
    description: string;
    descriptionSecondary?: string;
    screenshotKey?: LandingScreenshotKey;
    placeholder: "terminalFlow" | "multiOutlet";
  }> = [
    {
      title: isId ? "Transaksi cepat, alur kasir tetap rapi" : "Fast checkout with a cleaner cashier flow",
      description: isId
        ? "Gabungkan item, diskon, dan pembayaran dalam satu alur yang ringkas supaya antrean bergerak lebih cepat di jam sibuk."
        : "Combine items, discounts, and payments in one focused flow so queues keep moving during peak hours.",
      descriptionSecondary: isId
        ? "Kasir tetap bisa melayani dengan konsisten karena tampilan terminal dibuat sederhana dan minim langkah berulang."
        : "Cashiers stay consistent because the terminal is built with fewer repetitive steps.",
      screenshotKey: "posTerminalFlow",
      placeholder: "terminalFlow",
    },
    {
      title: isId ? "Live table bantu servis tetap terkontrol" : "Live table keeps service under control",
      description: isId
        ? "Pantau status meja, durasi layanan, dan antrean pesanan secara real-time agar koordinasi tim outlet lebih cepat."
        : "Track table status, service duration, and order queues in real time so outlet teams coordinate faster.",
      screenshotKey: "posLiveTable",
      placeholder: "terminalFlow",
    },
    {
      title: isId ? "Atur layout meja sesuai ritme outlet" : "Design floor layout for each outlet flow",
      description: isId
        ? "Sesuaikan posisi meja dan area layanan supaya alur dine-in lebih efisien untuk jam ramai maupun jam normal."
        : "Arrange table positions and service areas to keep dine-in flow efficient during peak and normal hours.",
      screenshotKey: "posFloorLayout",
      placeholder: "multiOutlet",
    },
  ];

  return (
    <>
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(schemaOrg) }} />
      <PageMotion className="flex flex-col min-h-screen bg-background pt-16">
        <section className="pt-24 pb-10 px-6 max-w-5xl mx-auto text-center">
          <p className="text-xs font-semibold tracking-widest uppercase text-muted-foreground mb-6">
            SalesView Point of Sale
          </p>
          <h1 className="text-5xl md:text-7xl font-semibold tracking-tight text-foreground leading-[1.1] mb-6">
            {isId ? `Ubah Antrean Panjang Menjadi` : `Turn Long Queues Into`}
            <br />
            <span className="text-primary font-accent font-medium italic text-[4.5rem] md:text-[6.5rem] leading-none inline-block mt-2">
              {isId ? `Layanan Kilat` : `Lightning Service`}
            </span>
          </h1>
          <p className="text-xl md:text-2xl font-light text-muted-foreground/80 max-w-3xl mx-auto leading-relaxed mb-10">
            {isId 
              ? `Beri pelanggan Anda pengalaman transaksi bebas hambatan. POS cerdas kami menjaga kelancaran toko Anda bahkan tanpa koneksi internet.`
              : `Give your customers a frictionless checkout experience. Our smart POS keeps your store running smoothly even when the internet drops.`
            }
          </p>
          
          <div className="flex flex-wrap items-center justify-center gap-4">
            <Link 
              href="/register"
              className="rounded-full bg-primary px-8 py-4 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-transform hover:scale-105 shadow-sm"
            >
              {isId ? "Coba POS Gratis" : "Start My Free Trial"}
            </Link>
            <Link
              href="/login"
              className="rounded-full bg-secondary text-secondary-foreground px-8 py-4 text-sm font-medium hover:bg-secondary/80 transition-colors"
            >
              {isId ? "Lihat Cara Kerjanya" : "See How It Works"}
            </Link>
          </div>
          <p className="mt-6 text-sm text-muted-foreground">
            {isId ? "Gratis selamanya, tanpa batasan jumlah user." : "Free, forever, with absolutely no user limits."}
          </p>
        </section>

        <section className="px-6 max-w-7xl mx-auto w-full pb-24">
          <div className="relative mx-auto w-full max-w-6xl overflow-hidden rounded-2xl border border-border/40 bg-card shadow-[0_20px_50px_rgba(0,0,0,0.12)]">
            <div className="relative aspect-video w-full">
              <ThemeAwareImage
                lightSrc={LANDING_THEME_IMAGES_BY_FEATURE.posHero.light}
                darkSrc={LANDING_THEME_IMAGES_BY_FEATURE.posHero.dark}
                alt={isId ? "Tampilan POS SalesView" : "SalesView POS interface"}
                fill
                className="object-cover object-top"
                sizes="(max-width: 1280px) 100vw, 1280px"
                priority
              />
            </div>
          </div>
        </section>

        <section className="bg-secondary/20 py-24 md:py-32">
          <div className="max-w-5xl mx-auto px-6">
            <div className="text-center mb-14 md:mb-20 max-w-3xl mx-auto">
              <h2 className="text-4xl md:text-5xl font-semibold tracking-tight text-foreground mb-5">
                {isId ? "Berikan pengalaman checkout terbaik untuk pelanggan Anda" : "Deliver the best checkout experience for your customers"}
              </h2>
              <p className="text-lg md:text-xl text-muted-foreground font-light leading-relaxed">
                {isId
                  ? "Tingkatkan efisiensi kasir dan biarkan tim Anda fokus pada hal terpenting: kepuasan pelanggan."
                  : "Increase cashier efficiency and let your team focus on what matters most: customer satisfaction."}
              </p>
            </div>

            <div className="space-y-20 md:space-y-24">
              {featureBlocks.map((feature, index) => {
                const useScreenshot = Boolean(feature.screenshotKey);
                const reverse = index % 2 === 0;

                return (
                  <div
                    key={feature.title}
                    className={`flex flex-col gap-12 lg:gap-20 items-center ${reverse ? "lg:flex-row-reverse" : "lg:flex-row"}`}
                  >
                    <div className="lg:w-1/2 w-full">
                      <div className="aspect-video rounded-3xl overflow-hidden bg-background border border-border shadow-sm p-6 md:p-8 flex items-center justify-center relative">
                        {useScreenshot ? (
                          <ThemeAwareImage
                            lightSrc={LANDING_THEME_IMAGES_BY_FEATURE[feature.screenshotKey!].light}
                            darkSrc={LANDING_THEME_IMAGES_BY_FEATURE[feature.screenshotKey!].dark}
                            alt={feature.title}
                            fill
                            className="object-cover object-top"
                            sizes="(max-width: 1024px) 100vw, 50vw"
                          />
                        ) : feature.placeholder === "multiOutlet" ? (
                          <div className="w-full h-full border border-border/50 rounded-xl p-4 md:p-5 grid grid-cols-5 gap-4">
                            <div className="col-span-3 space-y-3">
                              <div className="h-10 rounded-md bg-secondary/70 border border-border/50" />
                              <div className="grid grid-cols-3 gap-2">
                                {[1, 2, 3].map((item) => (
                                  <div key={item} className="h-16 rounded-md bg-secondary/60 border border-border/40 p-2">
                                    <div className="h-2 w-2/3 bg-muted rounded mb-2" />
                                    <div className="h-5 w-1/2 bg-foreground/15 rounded" />
                                  </div>
                                ))}
                              </div>
                              <div className="h-28 rounded-md bg-card border border-border/50 p-3 flex items-end gap-2">
                                {[30, 45, 35, 60, 50, 70].map((height, i) => (
                                  <div key={i} className="flex-1 rounded-t-sm bg-primary/40" style={{ height: `${height}%` }} />
                                ))}
                              </div>
                            </div>
                            <div className="col-span-2 border-l border-border/50 pl-4 space-y-3">
                              {[1, 2, 3, 4].map((item) => (
                                <div key={item} className="rounded-md bg-secondary/60 border border-border/40 p-2">
                                  <div className="h-2 w-2/3 bg-muted rounded mb-2" />
                                  <div className="h-2 w-full bg-muted rounded" />
                                </div>
                              ))}
                              <div className="h-9 rounded-md bg-primary/80" />
                            </div>
                          </div>
                        ) : (
                          <div className="w-full h-full border border-border/50 rounded-xl p-4 grid grid-cols-3 gap-4">
                            <div className="col-span-2 space-y-2">
                              {[1, 2, 3].map((item) => (
                                <div key={item} className="h-12 bg-secondary rounded flex items-center px-4 justify-between">
                                  <div className="w-1/2 h-3 bg-muted rounded" />
                                  <div className="w-8 h-3 bg-muted rounded" />
                                </div>
                              ))}
                            </div>
                            <div className="col-span-1 border-l border-border/50 pl-4 flex flex-col justify-end">
                              <div className="h-20 bg-primary/20 rounded mb-4 flex items-center justify-center">
                                <div className="w-1/2 h-4 bg-primary/40 rounded" />
                              </div>
                              <div className="h-10 bg-primary rounded" />
                            </div>
                          </div>
                        )}
                      </div>
                    </div>

                    <div className="lg:w-1/2 space-y-6 text-center lg:text-left">
                      <h3 className="text-3xl md:text-4xl font-semibold tracking-tight text-foreground">
                        {feature.title}
                      </h3>
                      <p className="text-lg text-muted-foreground font-light leading-relaxed">
                        {feature.description}
                      </p>
                      {feature.descriptionSecondary ? (
                        <p className="text-lg text-muted-foreground font-light leading-relaxed">
                          {feature.descriptionSecondary}
                        </p>
                      ) : null}
                    </div>
                  </div>
                );
              })}
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
                ? "Hubungkan operasional POS dengan inventory dan purchasing untuk kontrol stok yang lebih presisi."
                : "Connect POS operations with inventory and purchasing for tighter stock control."}
            </p>
            <div className="mt-8 flex flex-wrap justify-center gap-3">
              <Link href="/stock" className="rounded-full border border-border bg-background px-4 py-2 text-sm font-medium hover:bg-secondary">
                {isId ? "Inventory" : "Inventory"}
              </Link>
              <Link href="/goods-receipt" className="rounded-full border border-border bg-background px-4 py-2 text-sm font-medium hover:bg-secondary">
                {isId ? "Penerimaan Barang" : "Goods Receipt"}
              </Link>
              <Link href="/movements" className="rounded-full border border-border bg-background px-4 py-2 text-sm font-medium hover:bg-secondary">
                {isId ? "Pergerakan Stok" : "Stock Movements"}
              </Link>
              <Link href="/purchase" className="rounded-full border border-border bg-background px-4 py-2 text-sm font-medium hover:bg-secondary">
                {isId ? "Purchase" : "Purchase"}
              </Link>
            </div>
          </div>
        </section>

        <section className="py-24 md:py-32 px-6 max-w-4xl mx-auto text-center space-y-8">
          <h2 className="text-4xl md:text-5xl font-semibold tracking-tight text-foreground">
            {isId ? "Siap melayani pelanggan Anda lebih cepat?" : "Ready to serve your customers faster?"}
          </h2>
          <div className="grid md:grid-cols-2 gap-8 text-left mt-16 max-w-3xl mx-auto">
            <div>
              <h4 className="font-semibold text-lg mb-2">{isId ? "Sesi kasir lebih terkontrol" : "Better cashier session control"}</h4>
              <p className="text-muted-foreground">
                {isId
                  ? "Buka dan tutup sesi kasir dengan alur yang jelas agar proses serah terima shift lebih aman."
                  : "Open and close cashier sessions with a clear flow for safer shift handover."}
              </p>
            </div>
            <div>
              <h4 className="font-semibold text-lg mb-2">{isId ? "Pembayaran tetap fleksibel" : "Flexible payment handling"}</h4>
              <p className="text-muted-foreground">
                {isId
                  ? "Kelola metode pembayaran tunai dan non-tunai dengan pencatatan yang tetap rapi di POS."
                  : "Handle cash and non-cash payments while keeping POS records clean and consistent."}
              </p>
            </div>
          </div>
          <div className="pt-12">
            <Link
              href="/register"
              className="rounded-full bg-foreground px-8 py-4 text-sm font-medium text-background hover:bg-foreground/90 transition-transform hover:scale-105 inline-flex shadow-xl"
            >
              {isId ? "Lihat semua fitur POS" : "See all POS features"}
            </Link>
          </div>
        </section>
      </PageMotion>
    </>
  );
}
