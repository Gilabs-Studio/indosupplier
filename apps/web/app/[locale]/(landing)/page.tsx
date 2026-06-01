import React from "react";
import { getTranslations } from "next-intl/server";
import Image from "next/image";
import WaitingListForm from "@/features/waiting-list/components/waiting-list-form";
import { Header } from "@/components/navigation/header";
import { ScrollTextReveal } from "@/components/motion";
import { RainbowButton } from "@/components/ui/rainbow-button";
import { ShieldCheck, Lock, Headset } from "lucide-react";


export default async function LandingPage({
  params,
}: Readonly<{
  params: Promise<{ locale: string }>;
}>) {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: "landing" });

  return (
    <div className="min-h-screen bg-background text-foreground font-jost antialiased">
      {/* ── Navigation Header ── */}
      <Header locale={locale} />

      {/* ── SECTION 1: HERO ── */}
      <section
        className="relative flex min-h-svh items-center justify-start overflow-visible bg-background px-6 md:px-16 lg:px-24 pt-28 pb-20 md:pt-32 md:pb-20"
      >
        <div aria-hidden className="absolute inset-0">
          {/* Progressive top/bottom fade of the background image */}
          <div
            className="absolute inset-0 bg-position-[35%_center] md:bg-right-center bg-cover bg-no-repeat opacity-95 transition-opacity duration-500"
            style={{
              backgroundImage: "url('/hero.png')",
              maskImage: "linear-gradient(to bottom, transparent 0%, black 12%, black 90%, transparent 100%)",
              WebkitMaskImage: "linear-gradient(to bottom, transparent 0%, black 12%, black 90%, transparent 100%)"
            }}
          />
          {/* Graduated background tint overlay for maximum text readability */}
          <div className="absolute inset-0 bg-linear-to-r from-background/95 via-background/80 to-transparent md:from-background/90 md:via-background/55 md:to-transparent" />
        </div>

        <div className="relative z-10 w-full max-w-[1400px] mx-auto flex flex-col justify-center items-start">
          <div className="max-w-2xl text-left relative isolate">
            {/* Title with distinct typography highlighting */}
            <h1 className="mb-5 font-sans text-[36px] sm:text-[48px] md:text-[56px] lg:text-[66px] font-bold leading-[1.1] tracking-[-0.03em] text-[#E2E8F0] animate-fade-in">
              {t.rich("hero.title", {
                br: () => <br />,
                brHidden: () => <br className="hidden sm:inline" />,
                spanClass: (chunks) => (
                  <span className="font-jawa-palsu inline-block bg-linear-to-r from-[#E27D18] to-[#FFB300] bg-clip-text text-transparent font-medium">
                    {chunks}
                  </span>
                ),
              })}
            </h1>

            {/* Subheadline copy */}
            <p className="mb-8 max-w-xl text-[15px] sm:text-[16px] md:text-[17px] font-normal leading-relaxed text-neutral-400 animate-slide-up">
              {t("hero.subheadline")}
            </p>

            {/* CTA Button container */}
            <div className="relative z-20 flex justify-start gap-8 animate-slide-up delay-100">
              <RainbowButton asChild size="lg" className="text-[13px] font-semibold tracking-widest transition-all duration-300 hover:scale-[1.02] active:scale-[0.98]">
                <a href="#join">
                  {t("hero.cta")}
                </a>
              </RainbowButton>
            </div>

            {/* Stats container using separator styling */}
            <div className="mt-10 p-4 sm:p-5 bg-neutral-900/40 backdrop-blur-md border border-neutral-800/60 rounded-[14px] flex flex-col md:flex-row items-stretch gap-6 md:gap-4 max-w-3xl text-left animate-slide-up delay-200">
              {/* Item 1 */}
              <div className="flex items-start gap-3 flex-1">
                <div className="p-2 bg-neutral-800 text-[#FFB300] shrink-0 mt-0.5">
                  <ShieldCheck className="h-5 w-5" />
                </div>
                <div>
                  <h4 className="text-xs sm:text-sm font-semibold text-neutral-200 leading-snug">
                    {t("hero.stats.verifiedTitle")}
                  </h4>
                  <p className="text-[11px] sm:text-xs text-neutral-400 font-normal leading-normal mt-0.5">
                    {t("hero.stats.verifiedDesc")}
                  </p>
                </div>
              </div>

              {/* Separator */}
              <div className="hidden md:block w-px bg-neutral-800 self-stretch my-1" />

              {/* Item 2 */}
              <div className="flex items-start gap-3 flex-1">
                <div className="p-2 bg-neutral-800 text-[#FFB300] shrink-0 mt-0.5">
                  <Lock className="h-5 w-5" />
                </div>
                <div>
                  <h4 className="text-xs sm:text-sm font-semibold text-neutral-200 leading-snug">
                    {t("hero.stats.secureTitle")}
                  </h4>
                  <p className="text-[11px] sm:text-xs text-neutral-400 font-normal leading-normal mt-0.5">
                    {t("hero.stats.secureDesc")}
                  </p>
                </div>
              </div>

              {/* Separator */}
              <div className="hidden md:block w-px bg-neutral-800 self-stretch my-1" />

              {/* Item 3 */}
              <div className="flex items-start gap-3 flex-1">
                <div className="p-2 bg-neutral-800 text-[#FFB300] shrink-0 mt-0.5">
                  <Headset className="h-5 w-5" />
                </div>
                <div>
                  <h4 className="text-xs sm:text-sm font-semibold text-neutral-200 leading-snug">
                    {t("hero.stats.supportTitle")}
                  </h4>
                  <p className="text-[11px] sm:text-xs text-neutral-400 font-normal leading-normal mt-0.5">
                    {t("hero.stats.supportDesc")}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>

      </section>

      {/* ── SECTION 2: FEATURES / CAPABILITIES ── */}
      <section id="features" className="px-6 md:px-16 lg:px-24 pt-32 pb-20 md:pt-40 md:pb-28 bg-secondary/8 border-t border-border/30">
        <div className="max-w-[1400px] w-full mx-auto">
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
            <div className="lg:col-span-7">
              <span className="text-[11px] tracking-widest font-jawa-palsu font-medium inline-block bg-linear-to-r from-[#E27D18] to-[#FFB300] bg-clip-text text-transparent mb-4">
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
              <span className="text-[11px] tracking-widest font-jawa-palsu font-medium inline-block bg-linear-to-r from-[#E27D18] to-[#E27D18] bg-clip-text text-transparent mb-8">
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
            <div className="flex justify-between items-center text-[13px] tracking-wider font-light text-primary-foreground/75">
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
            <span className="text-[11px] tracking-widest font-jawa-palsu font-medium inline-block bg-linear-to-r from-[#E27D18] to-[#FFB300] bg-clip-text text-transparent mb-6">
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
            <span className="font-normal tracking-widest text-foreground text-[13px]">
              IndoSupplier
            </span>
          </div>
          <span>
            {t("footer.copy", { year: new Date().getFullYear() })}
          </span>
          <span className="tracking-wider text-muted-foreground/80">
            {t("footer.tagline")}
          </span>
        </div>
      </footer>
    </div>
  );
}
