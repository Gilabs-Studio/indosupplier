import React from "react";
import { getTranslations } from "next-intl/server";
import Image from "next/image";
import WaitingListForm from "@/features/waiting-list/components/waiting-list-form";
import { Header } from "@/components/navigation/header";
import { ScrollTextReveal } from "@/components/motion";
import { LightRays } from "@/components/ui/light-rays";
import { RainbowButton } from "@/components/ui/rainbow-button";
import { Handshake, Globe, Award, ShieldCheck } from "lucide-react";


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
      <Header locale={locale} />

      {/* ── SECTION 1: HERO (Full Viewport Height) ── */}
      <section
        className="min-h-[90vh] md:min-h-[95vh] flex items-center justify-center px-6 pt-32 pb-20 relative overflow-hidden bg-background"
      >
        {/* Faint Background Map / Sketch Illustration */}
        <div 
          className="absolute inset-0 z-0 bg-center bg-no-repeat bg-size-[150%_auto] md:bg-size-[105%_auto] opacity-10 pointer-events-none"
          style={{ backgroundImage: "url('/hero.png')" }}
        />

        <LightRays
          aria-hidden
          className="z-1"
          count={8}
          color="rgba(255, 255, 255, 0.9)"
          blur={44}
          speed={18}
          length="78vh"
        />

        <div className="max-w-4xl w-full mx-auto flex flex-col items-center justify-center text-center relative z-10">
          {/* Main Headline */}
          <h1
            className="font-bold leading-[1.15] tracking-[-0.03em] text-foreground mb-6 max-w-3xl animate-fade-in"
            style={{ fontSize: "clamp(2.3rem, 5.5vw, 4.2rem)" }}
          >
            {t("hero.headline")}
          </h1>

          {/* Subheadline */}
          <p className="text-[16px] md:text-[18px] font-normal leading-relaxed text-muted-foreground/80 max-w-2xl mb-10 animate-slide-up">
            {t("hero.subheadline")}
          </p>

          {/* Call to action */}
          <div className="flex flex-wrap items-center justify-center gap-8 animate-slide-up delay-100 relative z-20">
            <RainbowButton asChild size="lg" className="text-[13px] font-semibold tracking-widest uppercase hover:scale-[1.02] active:scale-[0.98] transition-all duration-300">
              <a href="#join">
                {t("hero.cta")}
              </a>
            </RainbowButton>
          </div>

          {/* Trust Badges */}
          <div className="flex flex-wrap items-center justify-center gap-x-5 gap-y-3 text-xs md:text-sm font-medium text-muted-foreground/80 mt-10 animate-slide-up delay-200 relative z-10">
            <div className="flex items-center gap-2">
              <ShieldCheck className="h-4 w-4 text-[#d4af37] shrink-0" />
              <span>{t("hero.badges.verified")}</span>
            </div>
            <div className="hidden md:block h-3.5 w-px bg-border/70 shrink-0" />
            <div className="flex items-center gap-2">
              <Handshake className="h-4.5 w-4.5 text-[#d4af37] shrink-0" />
              <span>{t("hero.badges.secure")}</span>
            </div>
            <div className="hidden md:block h-3.5 w-px bg-border/70 shrink-0" />
            <div className="flex items-center gap-2">
              <Globe className="h-4 w-4 text-[#d4af37] shrink-0" />
              <span>{t("hero.badges.direct")}</span>
            </div>
            <div className="hidden md:block h-3.5 w-px bg-border/70 shrink-0" />
            <div className="flex items-center gap-2">
              <Award className="h-4 w-4 text-[#d4af37] shrink-0" />
              <span>{t("hero.badges.certifications")}</span>
            </div>
          </div>
        </div>
      </section>

      {/* ── SECTION 2: FEATURES / CAPABILITIES ── */}
      <section id="features" className="px-6 md:px-16 lg:px-24 py-20 md:py-28 bg-secondary/8 border-t border-border/30">
        <div className="max-w-[1400px] w-full mx-auto">
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
            <div className="lg:col-span-7">
              <span className="text-[11px] uppercase tracking-widest font-medium text-muted-foreground block mb-4">
                {t("features.badge")}
              </span>
              <h2
                className="font-light leading-[1.2] tracking-[-0.03em] max-w-[720px] text-foreground"
                style={{ fontSize: "clamp(2rem, 4.5vw, 3.4rem)" }}
              >
                {t("features.headline")}
              </h2>
            </div>

            <div className="lg:col-span-5 lg:pl-10">
              <p className="text-[16px] font-light leading-relaxed text-muted-foreground/80 mt-4">
                {t("features.summary")}
              </p>
            </div>
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
      <footer className="py-3 px-6 md:px-16 lg:px-24 border-t border-border bg-background">
        <div className="max-w-[1400px] w-full mx-auto flex flex-col md:flex-row items-center justify-between gap-6 text-[13px] font-light text-muted-foreground">
          <div className="flex items-center gap-3">
            <Image
              src="/logo.png"
              alt="IndoSupplier Logo"
              width={120}
              height={24}
              className="h-5 w-auto object-contain"
            />
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
