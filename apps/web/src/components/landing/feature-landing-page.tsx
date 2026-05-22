import { PageMotion } from "@/components/motion/page-motion";
import { Link } from "@/i18n/routing";
import { ThemeAwareImage } from "@/components/landing/theme-aware-image";
import {
  LANDING_THEME_IMAGES_BY_FEATURE,
  type LandingScreenshotKey,
} from "@/lib/landing-theme-images";

type PlaceholderVariant =
  | "terminalFlow"
  | "multiOutlet"
  | "timelineFlow"
  | "invoiceFlow"
  | "scorecardFlow"
  | "warehouseFlow";

type FeatureBlock = {
  title: string;
  description: string;
  descriptionSecondary?: string;
  screenshotKey?: LandingScreenshotKey;
  placeholder: PlaceholderVariant;
};

type FeaturePoint = {
  title: string;
  description: string;
};

type InternalLink = {
  href: string;
  label: {
    id: string;
    en: string;
  };
};

const RELATED_LINK_GROUPS: Record<"sales" | "finance" | "inventory" | "hr", InternalLink[]> = {
  sales: [
    { href: "/crm", label: { id: "CRM", en: "CRM" } },
    { href: "/quotations", label: { id: "Quotations", en: "Quotations" } },
    { href: "/sales", label: { id: "Sales Orders", en: "Sales Orders" } },
    { href: "/invoicing", label: { id: "Invoicing", en: "Invoicing" } },
  ],
  finance: [
    { href: "/accounting", label: { id: "Accounting", en: "Accounting" } },
    { href: "/fixed-assets", label: { id: "Aset Tetap", en: "Fixed Assets" } },
    { href: "/financial-reports", label: { id: "Laporan Keuangan", en: "Financial Reports" } },
    { href: "/reconciliation", label: { id: "Rekonsiliasi", en: "Reconciliation" } },
  ],
  inventory: [
    { href: "/stock", label: { id: "Inventory", en: "Inventory" } },
    { href: "/goods-receipt", label: { id: "Penerimaan Barang", en: "Goods Receipt" } },
    { href: "/movements", label: { id: "Pergerakan Stok", en: "Stock Movements" } },
    { href: "/purchase", label: { id: "Purchase", en: "Purchase" } },
  ],
  hr: [
    { href: "/employees", label: { id: "Karyawan", en: "Employees" } },
    { href: "/attendance", label: { id: "Attendance", en: "Attendance" } },
    { href: "/recruitment", label: { id: "Recruitment", en: "Recruitment" } },
    { href: "/travel-planner", label: { id: "Travel Planner", en: "Travel Planner" } },
    { href: "/evaluation", label: { id: "Evaluasi", en: "Evaluation" } },
  ],
};

const FEATURE_PATH_BY_LABEL: Record<string, string> = {
  "SalesView Accounting": "/accounting",
  "SalesView Financial Reports": "/financial-reports",
  "SalesView Reconciliation": "/reconciliation",
  "SalesView Fixed Assets": "/fixed-assets",
  "SalesView Quotations": "/quotations",
  "SalesView Sales Orders": "/sales",
  "SalesView Invoicing": "/invoicing",
  "SalesView Purchase": "/purchase",
  "SalesView Goods Receipt": "/goods-receipt",
  "SalesView Movements": "/movements",
  "SalesView Inventory": "/stock",
  "SalesView Employees": "/employees",
  "SalesView Attendance": "/attendance",
  "SalesView Recruitment": "/recruitment",
  "SalesView Travel Planner": "/travel-planner",
  "SalesView Evaluation": "/evaluation",
};

function getRelatedLinks(label: string): InternalLink[] {
  const lowerLabel = label.toLowerCase();

  let group: keyof typeof RELATED_LINK_GROUPS = "sales";
  if (
    lowerLabel.includes("accounting") ||
    lowerLabel.includes("financial") ||
    lowerLabel.includes("reconciliation") ||
    lowerLabel.includes("fixed assets")
  ) {
    group = "finance";
  } else if (
    lowerLabel.includes("inventory") ||
    lowerLabel.includes("purchase") ||
    lowerLabel.includes("goods receipt") ||
    lowerLabel.includes("movements")
  ) {
    group = "inventory";
  } else if (
    lowerLabel.includes("employees") ||
    lowerLabel.includes("attendance") ||
    lowerLabel.includes("recruitment") ||
    lowerLabel.includes("travel planner") ||
    lowerLabel.includes("evaluation")
  ) {
    group = "hr";
  }

  const currentPath = FEATURE_PATH_BY_LABEL[label];
  return RELATED_LINK_GROUPS[group].filter((item) => item.href !== currentPath);
}

type FeatureLandingPageProps = {
  locale: string;
  label: string;
  heroTitle: string;
  heroAccent: string;
  heroDescription: string;
  heroImageAlt: string;
  heroImageKey: LandingScreenshotKey;
  introTitle: string;
  introDescription: string;
  featureBlocks: FeatureBlock[];
  finalTitle: string;
  finalPoints?: [FeaturePoint, FeaturePoint];
  finalCtaLabel: string;
  finalCtaHref?: string;
};

function PlaceholderVisual({ variant }: { variant: PlaceholderVariant }) {
  if (variant === "multiOutlet") {
    return (
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
            {[30, 45, 35, 60, 50, 70].map((height, idx) => (
              <div key={idx} className="flex-1 rounded-t-sm bg-primary/40" style={{ height: `${height}%` }} />
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
    );
  }

  if (variant === "timelineFlow") {
    return (
      <div className="w-full h-full border border-border/50 rounded-xl p-4 space-y-3">
        {[1, 2, 3, 4].map((item) => (
          <div key={item} className="flex items-center gap-3 rounded-md border border-border/40 p-3 bg-secondary/50">
            <div className="h-8 w-8 rounded-full bg-primary/25 shrink-0" />
            <div className="flex-1 space-y-2">
              <div className="h-2 w-1/3 bg-muted rounded" />
              <div className="h-2 w-2/3 bg-muted rounded" />
            </div>
          </div>
        ))}
      </div>
    );
  }

  if (variant === "invoiceFlow") {
    return (
      <div className="w-full h-full border border-border/50 rounded-xl p-4 space-y-4 bg-card">
        <div className="h-8 w-2/5 bg-secondary rounded" />
        <div className="grid grid-cols-2 gap-3">
          <div className="h-10 bg-secondary/60 rounded" />
          <div className="h-10 bg-secondary/60 rounded" />
        </div>
        {[1, 2, 3].map((item) => (
          <div key={item} className="h-8 w-full bg-secondary/50 rounded flex items-center justify-between px-3">
            <div className="h-2 w-1/3 bg-muted rounded" />
            <div className="h-2 w-10 bg-muted rounded" />
          </div>
        ))}
        <div className="flex justify-end pt-2">
          <div className="h-10 w-32 bg-primary rounded" />
        </div>
      </div>
    );
  }

  if (variant === "scorecardFlow") {
    return (
      <div className="w-full h-full border border-border/50 rounded-xl p-4 space-y-4">
        <div className="h-8 bg-secondary/70 rounded border border-border/40" />
        {[1, 2, 3].map((item) => (
          <div key={item} className="rounded-lg border border-border/40 p-3 bg-card">
            <div className="h-2 w-1/3 bg-muted rounded mb-3" />
            <div className="h-2 w-full bg-muted rounded mb-2" />
            <div className="h-2 w-2/3 bg-primary/40 rounded" />
          </div>
        ))}
      </div>
    );
  }

  if (variant === "warehouseFlow") {
    return (
      <div className="w-full h-full border border-border/50 rounded-xl p-4 space-y-4">
        <div className="h-8 rounded bg-secondary/70 border border-border/50" />
        {[1, 2, 3].map((item) => (
          <div key={item} className="rounded-lg border border-border/40 p-3 bg-card flex items-center justify-between">
            <div className="space-y-2 w-2/3">
              <div className="h-2 w-1/2 bg-muted rounded" />
              <div className="h-2 w-full bg-muted rounded" />
            </div>
            <div className="h-8 w-16 bg-primary/30 rounded" />
          </div>
        ))}
      </div>
    );
  }

  return (
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
  );
}

export function FeatureLandingPage({
  locale,
  label,
  heroTitle,
  heroAccent,
  heroDescription,
  heroImageAlt,
  heroImageKey,
  introTitle,
  introDescription,
  featureBlocks,
  finalTitle,
  finalPoints,
  finalCtaLabel,
  finalCtaHref = "/register",
}: FeatureLandingPageProps) {
  const isId = locale === "id";
  const relatedLinks = getRelatedLinks(label);

  return (
    <PageMotion className="flex flex-col min-h-screen bg-background pt-16">
      <section className="pt-24 pb-10 px-6 max-w-5xl mx-auto text-center">
        <p className="text-xs font-semibold tracking-widest uppercase text-muted-foreground mb-6">{label}</p>
        <h1 className="text-5xl md:text-7xl font-semibold tracking-tight text-foreground leading-[1.1] mb-6">
          {heroTitle}
          <br />
          <span className="text-primary font-accent font-medium italic text-[4.5rem] md:text-[6.5rem] leading-none inline-block mt-2">
            {heroAccent}
          </span>
        </h1>
        <p className="text-xl md:text-2xl font-light text-muted-foreground/80 max-w-3xl mx-auto leading-relaxed mb-10">
          {heroDescription}
        </p>

        <div className="flex flex-wrap items-center justify-center gap-4">
          <Link
            href="/register"
            className="cursor-pointer rounded-full bg-primary px-8 py-4 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-transform hover:scale-105 shadow-sm"
          >
            {isId ? "Mulai sekarang - Gratis" : "Start now - Free"}
          </Link>
          <Link
            href="/login"
            className="cursor-pointer rounded-full bg-secondary text-secondary-foreground px-8 py-4 text-sm font-medium hover:bg-secondary/80 transition-colors"
          >
            {isId ? "Ketahui lebih lanjut" : "Learn more"}
          </Link>
        </div>
      </section>

      <section className="px-6 max-w-7xl mx-auto w-full pb-24">
        <div className="relative mx-auto w-full max-w-6xl overflow-hidden rounded-2xl border border-border/40 bg-card shadow-sm">
          <div className="relative aspect-video w-full">
            <ThemeAwareImage
                lightSrc={LANDING_THEME_IMAGES_BY_FEATURE[heroImageKey].light}
                darkSrc={LANDING_THEME_IMAGES_BY_FEATURE[heroImageKey].dark}
                alt={heroImageAlt}
                fill
                className="object-cover object-top"
                sizes="(max-width: 1280px) 100vw, 1280px"
                priority
                fallback={<PlaceholderVisual variant={featureBlocks?.[0]?.placeholder ?? "multiOutlet"} />}
              />
          </div>
        </div>
      </section>

      <section className="bg-secondary/20 py-24 md:py-32">
        <div className="max-w-5xl mx-auto px-6">
          <div className="text-center mb-14 md:mb-20 max-w-3xl mx-auto">
            <h2 className="text-4xl md:text-5xl font-semibold tracking-tight text-foreground mb-5">{introTitle}</h2>
            <p className="text-lg md:text-xl text-muted-foreground font-light leading-relaxed">{introDescription}</p>
          </div>

          <div className="space-y-20 md:space-y-24">
            {featureBlocks.map((feature, index) => {
              const useScreenshot = Boolean(feature.screenshotKey);
              const reverse = index % 2 === 0;

              return (
                <div
                  key={feature.title}
                  className={`flex flex-col gap-12 lg:gap-20 items-center justify-center ${reverse ? "lg:flex-row-reverse" : "lg:flex-row"}`}
                >
                  <div className="lg:w-1/2 w-full max-w-lg mx-auto lg:mx-0">
                    <div className="aspect-video rounded-3xl overflow-hidden bg-background border border-border shadow-sm p-6 md:p-8 flex items-center justify-center relative">
                      {useScreenshot ? (
                        <ThemeAwareImage
                            lightSrc={LANDING_THEME_IMAGES_BY_FEATURE[feature.screenshotKey!].light}
                            darkSrc={LANDING_THEME_IMAGES_BY_FEATURE[feature.screenshotKey!].dark}
                            alt={feature.title}
                            fill
                            className="object-cover object-top"
                            sizes="(max-width: 1024px) 100vw, 50vw"
                            fallback={<PlaceholderVisual variant={feature.placeholder} />}
                          />
                      ) : (
                        <PlaceholderVisual variant={feature.placeholder} />
                      )}
                    </div>
                  </div>

                  <div className="lg:w-1/2 space-y-6 text-center lg:text-left">
                    <h3 className="text-3xl md:text-4xl font-semibold tracking-tight text-foreground">{feature.title}</h3>
                    <p className="text-lg text-muted-foreground font-light leading-relaxed">{feature.description}</p>
                    {feature.descriptionSecondary ? (
                      <p className="text-lg text-muted-foreground font-light leading-relaxed">{feature.descriptionSecondary}</p>
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
              ? "Hubungkan proses antar tim dengan modul yang saling melengkapi dalam satu alur kerja."
              : "Connect team workflows by exploring modules that complement each other in one platform."}
          </p>
          <div className="mt-8 flex flex-wrap justify-center gap-3">
            {relatedLinks.map((item) => (
              <Link
                key={item.href}
                href={item.href}
                className="rounded-full border border-border bg-background px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-secondary"
              >
                {isId ? item.label.id : item.label.en}
              </Link>
            ))}
          </div>
        </div>
      </section>

      <section className="py-24 md:py-32 px-6 max-w-4xl mx-auto text-center space-y-8">
        <h2 className="text-4xl md:text-5xl font-semibold tracking-tight text-foreground">{finalTitle}</h2>
        {finalPoints ? (
          <div className="grid md:grid-cols-2 gap-8 text-left mt-16 max-w-3xl mx-auto">
            {finalPoints.map((point) => (
              <div key={point.title}>
                <h4 className="font-semibold text-lg mb-2">{point.title}</h4>
                <p className="text-muted-foreground">{point.description}</p>
              </div>
            ))}
          </div>
        ) : null}
        <div className="pt-12">
          <Link
            href={finalCtaHref}
            className="cursor-pointer rounded-full bg-foreground px-8 py-4 text-sm font-medium text-background hover:bg-foreground/90 transition-transform hover:scale-105 inline-flex shadow-xl"
          >
            {finalCtaLabel}
          </Link>
        </div>
      </section>
    </PageMotion>
  );
}
