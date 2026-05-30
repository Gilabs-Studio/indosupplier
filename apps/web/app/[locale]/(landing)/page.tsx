import React from "react";
import { getTranslations } from "next-intl/server";
import { Link } from "@/i18n/routing";
import WaitingListForm from "@/features/waiting-list/components/waiting-list-form";
import LanguageSwitcher from "@/components/navigation/language-switcher";
import { Button } from "@/components/ui/button";
import { ScrollTextReveal } from "@/components/motion";


export default async function LandingPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: "landing" });

  return (
    <div className="min-h-screen bg-background text-foreground font-jost antialiased selection:bg-primary selection:text-primary-foreground">
      {/* ── Navigation Header ── */}
      <header className="w-full py-6 px-6 md:px-16 lg:px-24 z-50">
        <div className="max-w-[1400px] mx-auto flex items-center justify-between">
          <div className="flex items-center gap-3">
            <img src="/logo.png" alt="IndoSupplier Logo" className="h-6 w-auto object-contain" />
            <span className="font-normal text-[15px] tracking-widest uppercase text-foreground">
              IndoSupplier
            </span>
          </div>

          {/* Center menu links */}
          <div className="hidden md:flex items-center gap-12 text-[12px] tracking-widest uppercase font-light text-muted-foreground">
            <a href="#features" className="hover:text-foreground transition-colors">
              {t("features.badge")}
            </a>
            <a href="#about" className="hover:text-foreground transition-colors">
              {t("nav.about")}
            </a>
            <a href="#join" className="hover:text-foreground transition-colors">
              {t("nav.waitlist")}
            </a>
          </div>

          {/* Right menu actions */}
          <div className="flex items-center gap-6 text-[12px] font-medium tracking-widest uppercase">
            <LanguageSwitcher currentLocale={locale} />
            <Link
              href="/login"
              className="hover:underline text-foreground"
            >
              {t("nav.signIn")}
            </Link>
          </div>
        </div>
      </header>

      {/* ── SECTION 1: HERO (Full Viewport Height) ── */}
      <section className="min-h-[calc(100vh-80px)] flex items-center px-6 md:px-16 lg:px-24 py-12 md:py-20 relative">
        <div className="max-w-[1400px] w-full mx-auto grid grid-cols-1 lg:grid-cols-12 gap-12 lg:gap-16 items-center">
          {/* Hero Content Left */}
          <div className="lg:col-span-7 flex flex-col justify-center">
            {/* Main Headline */}
            <h1
              className="font-light leading-[1.08] tracking-[-0.04em] text-foreground mb-8 max-w-[850px] animate-fade-in"
              style={{ fontSize: "clamp(2.8rem, 5.5vw, 4.8rem)" }}
            >
              {t("hero.headline")}
            </h1>

            {/* Subheadline */}
            <p className="text-[17px] md:text-[19px] font-light leading-relaxed text-muted-foreground max-w-[580px] mb-12 animate-slide-up">
              {t("hero.subheadline")}
            </p>

            {/* Call to action */}
            <div className="flex flex-wrap items-center gap-8 animate-slide-up">
              <Button variant="cta" asChild>
                <a href="#join">
                  {t("hero.cta")}
                </a>
              </Button>
              <span className="text-[13px] tracking-wider uppercase font-light text-muted-foreground">
                {t("hero.trustLabel")}
              </span>
            </div>
          </div>

          {/* Hero Visual Right */}
          <div className="lg:col-span-5 flex justify-center animate-fade-in">
            <img
              src="/hero2.png"
              alt="IndoSupplier Platform Visual"
              className="w-full h-auto object-contain rounded-md"
            />
          </div>
        </div>
      </section>

      {/* ── SECTION 2: FEATURES / CAPABILITIES ── */}
      <section id="features" className="px-6 md:px-16 lg:px-24 py-20 md:py-28 bg-secondary/35 border-t border-border/40">
        <div className="max-w-[1400px] w-full mx-auto">
          <div className="mb-16">
            <span className="text-[11px] uppercase tracking-widest font-medium text-muted-foreground block mb-3">
              {t("features.badge")}
            </span>
            <h2 className="text-[28px] md:text-[36px] font-light tracking-tight text-foreground">
              {t("features.headline")}
            </h2>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {[
              {
                key: "erp",
                label: t("features.items.erp.title"),
                desc: t("features.items.erp.desc"),
              },
              {
                key: "crm",
                label: t("features.items.crm.title"),
                desc: t("features.items.crm.desc"),
              },
              {
                key: "finance",
                label: t("features.items.finance.title"),
                desc: t("features.items.finance.desc"),
              },
            ].map((card, idx) => (
              <div
                key={card.key}
                className="bg-card/50 border border-border/50 p-8 rounded-lg hover:bg-card hover:border-border transition-all duration-300 flex flex-col justify-between min-h-[200px]"
              >
                <span className="text-[12px] uppercase tracking-wider font-light text-muted-foreground/60">
                  0{idx + 1}
                </span>
                <div className="mt-12">
                  <h3 className="text-[18px] font-normal text-foreground mb-3">
                    {card.label}
                  </h3>
                  <p className="text-[14px] font-light leading-relaxed text-muted-foreground">
                    {card.desc}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── SECTION 3: ABOUT / PHILOSOPHY ── */}
      <section id="about" className="min-h-screen flex items-center px-6 md:px-16 lg:px-24 py-20 bg-primary text-primary-foreground relative overflow-hidden">
        <div className="max-w-[1400px] w-full mx-auto grid grid-cols-1 lg:grid-cols-12 gap-12">
          {/* Left Column: Huge typography statement */}
          <div className="lg:col-span-8 flex flex-col justify-between">
            <div>
              <span className="text-[11px] uppercase tracking-widest font-medium text-primary-foreground/60 block mb-8">
                {t("philosophy.title")}
              </span>
              <h2
                className="font-light leading-[1.2] tracking-[-0.03em] max-w-[720px] text-primary-foreground"
                style={{ fontSize: "clamp(2rem, 4.5vw, 3.4rem)" }}
              >
                <ScrollTextReveal text={t("philosophy.quote")} />
              </h2>
            </div>
          </div>

          {/* Right Column: Detailed narrative */}
          <div className="lg:col-span-4 lg:pl-8 flex flex-col justify-end">
            <p className="text-[16px] font-light leading-relaxed text-primary-foreground/80 mb-8">
              {t("philosophy.description")}
            </p>
            <div className="h-px bg-primary-foreground/20 w-full mb-8" />
            <div className="flex justify-between items-center text-[13px] uppercase tracking-wider font-light text-primary-foreground/75">
              <span>Optimized for Indonesia</span>
              <span>EST. 2026</span>
            </div>
          </div>
        </div>
      </section>

      {/* ── SECTION 4: WAITING LIST / CONVERSION ── */}
      <section id="join" className="px-6 md:px-16 lg:px-24 py-20 md:py-28 bg-secondary">
        <div className="max-w-[1400px] w-full mx-auto grid grid-cols-1 lg:grid-cols-2 gap-16 lg:gap-24 items-center">
          {/* Left Column */}
          <div>
            <span className="text-[11px] uppercase tracking-widest font-medium text-muted-foreground block mb-6">
              {t("waitlist.badge")}
            </span>
            <h2
              className="font-light leading-[1.08] tracking-[-0.04em] text-foreground mb-6"
              style={{ fontSize: "clamp(2.4rem, 5vw, 4rem)" }}
            >
              {t("waitlist.headline")}
            </h2>
            <p className="text-[17px] font-light leading-relaxed text-muted-foreground mb-12 max-w-[500px]">
              {t("waitlist.subheadline")}
            </p>

            {/* Structured list in grid */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-8 text-[14px] font-light text-muted-foreground border-t border-border/80 pt-8">
              {[
                t("waitlist.benefits.discount"),
                t("waitlist.benefits.onboarding"),
                t("waitlist.benefits.support"),
                t("waitlist.benefits.noCard"),
              ].map((benefit, idx) => (
                <div key={idx} className="flex gap-3 items-start">
                  <svg
                    className="h-5 w-5 text-foreground shrink-0 mt-0.5"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    strokeWidth={2}
                  >
                    <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                  <p className="leading-relaxed text-muted-foreground">{benefit}</p>
                </div>
              ))}
            </div>
          </div>

          {/* Right Column (Form container) */}
          <div className="relative">
            <WaitingListForm />
          </div>
        </div>
      </section>

      {/* ── Footer ── */}
      <footer className="py-12 px-6 md:px-16 lg:px-24 border-t border-border bg-background">
        <div className="max-w-[1400px] w-full mx-auto flex flex-col md:flex-row items-center justify-between gap-6 text-[13px] font-light text-muted-foreground">
          <div className="flex items-center gap-3">
            <img src="/logo.png" alt="IndoSupplier Logo" className="h-5 w-auto object-contain" />
            <span className="font-normal tracking-widest uppercase text-foreground text-[13px]">
              IndoSupplier
            </span>
          </div>
          <span>
            {t("footer.copy", { year: new Date().getFullYear() })}
          </span>
          <span className="uppercase tracking-wider text-muted-foreground/80">
            {t("footer.tagline")}
          </span>
        </div>
      </footer>
    </div>
  );
}
