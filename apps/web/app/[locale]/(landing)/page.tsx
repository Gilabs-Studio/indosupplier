import React from "react";
import { getTranslations } from "next-intl/server";
import Image from "next/image";
import WaitingListForm from "@/features/sysadmin/waiting-list/components/waiting-list-form";
import { Header } from "@/components/navigation/header";
import { ScrollTextReveal } from "@/components/motion";
import { Button } from "@/components/ui/button";

const heroTitleFormats = {
  br: () => <br />,
  brHidden: () => <br className="hidden sm:inline" />,
  spanClass: (chunks: React.ReactNode) => (
    <span className="font-serif italic border-b border-[#E27D18]/50 pb-1">
      {chunks}
    </span>
  ),
};

export default async function LandingPage({
  params,
}: Readonly<{
  params: Promise<{ locale: string }>;
}>) {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: "landing" });

  return (
    <div className="min-h-screen bg-[#1a1a1a] text-[#ffffff] font-sans antialiased">
      {/* ── Navigation Header ── */}
      <Header locale={locale} />

      {/* ── SECTION 1: HERO ── */}
      <section
        className="relative flex min-h-svh items-center justify-start overflow-visible bg-[#1a1a1a] px-6 md:px-16 lg:px-24 pt-36 pb-36 md:pt-48 md:pb-48"
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
          <div className="absolute inset-0 bg-linear-to-r from-[#1a1a1a]/95 via-[#1a1a1a]/80 to-transparent md:from-[#1a1a1a]/90 md:via-[#1a1a1a]/55 md:to-transparent" />
        </div>

        <div className="relative z-10 w-full max-w-[1400px] mx-auto flex flex-col justify-center items-start">
          <div className="max-w-3xl text-left relative isolate">
            {/* Title with distinct typography highlighting */}
            <h1 className="mb-8 font-serif text-[56px] sm:text-[64px] md:text-[72px] lg:text-[80px] font-bold leading-[1.1] tracking-tight text-[#E8E6E3] animate-fade-in">
              {t.rich("hero.title", heroTitleFormats)}
            </h1>

            {/* Subheadline copy */}
            <p className="mb-12 max-w-2xl text-[16px] md:text-[18px] font-normal leading-relaxed text-neutral-400 animate-slide-up">
              {t("hero.subheadline")}
            </p>

            {/* CTA Button container */}
            <div className="relative z-20 flex justify-start gap-8 animate-slide-up delay-100">
              <Button
                asChild
                size="lg"
                className="px-8 py-6 bg-white hover:bg-neutral-200 text-neutral-900 border border-neutral-200 rounded-none text-[14px] font-semibold tracking-wider transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 cursor-pointer"
              >
                <a href="#join" className="cursor-pointer">
                  {t("hero.cta")}
                </a>
              </Button>
            </div>
          </div>

          {/* Stats container in editorial text format */}
          <div className="mt-20 grid grid-cols-1 md:grid-cols-3 gap-10 w-full max-w-4xl text-left animate-slide-up delay-200 border-t border-neutral-800/60 pt-10">
            {/* Item 1 */}
            <div>
              <h4 className="text-[12px] font-medium text-neutral-400 uppercase tracking-widest">
                {t("hero.stats.verifiedTitle")}
              </h4>
              <p className="text-[14px] text-neutral-300 font-normal leading-relaxed mt-2.5">
                {t("hero.stats.verifiedDesc")}
              </p>
            </div>

            {/* Item 2 */}
            <div>
              <h4 className="text-[12px] font-medium text-neutral-400 uppercase tracking-widest">
                {t("hero.stats.secureTitle")}
              </h4>
              <p className="text-[14px] text-neutral-300 font-normal leading-relaxed mt-2.5">
                {t("hero.stats.secureDesc")}
              </p>
            </div>

            {/* Item 3 */}
            <div>
              <h4 className="text-[12px] font-medium text-neutral-400 uppercase tracking-widest">
                {t("hero.stats.supportTitle")}
              </h4>
              <p className="text-[14px] text-neutral-300 font-normal leading-relaxed mt-2.5">
                {t("hero.stats.supportDesc")}
              </p>
            </div>
          </div>
        </div>

      </section>

      {/* ── SECTION 2: FEATURES / CAPABILITIES ── */}
      <section id="features" className="px-6 md:px-16 lg:px-24 pt-44 pb-36 md:pt-56 md:pb-48 bg-[#262626]/8 border-t border-[#333333]/30">
        <div className="max-w-[1400px] w-full mx-auto">
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
            <div className="lg:col-span-7">
              <span className="text-[12px] tracking-widest font-sans font-medium inline-block text-[#E27D18] uppercase mb-4">
                {t("features.badge")}
              </span>
              <h2 className="font-serif font-normal text-[36px] md:text-[48px] leading-[1.15] tracking-tight max-w-[720px] text-[#ffffff]">
                {t("features.headline")}
              </h2>
            </div>

            <div className="lg:col-span-5 lg:pl-10">
              <p className="text-[16px] md:text-[18px] font-normal leading-relaxed text-[#a6a6a6]/80 mt-4">
                {t("features.summary")}
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* ── SECTION 3: ABOUT / PHILOSOPHY ── */}
      <section id="about" className="min-h-screen flex items-center px-6 md:px-16 lg:px-24 py-36 md:py-48 bg-[#ffffff] text-[#1a1a1a] relative overflow-hidden">
        <div className="max-w-[1400px] w-full mx-auto grid grid-cols-1 lg:grid-cols-12 gap-12">
          {/* Left Column: Huge typography statement */}
          <div className="lg:col-span-8 flex flex-col justify-between">
            <div>
              <span className="text-[12px] tracking-widest font-sans font-medium inline-block text-[#E27D18] uppercase mb-8">
                {t("philosophy.title")}
              </span>
              <h2 className="font-serif font-normal text-[36px] md:text-[48px] leading-[1.15] tracking-tight max-w-[720px] text-[#1a1a1a]">
                <ScrollTextReveal text={t("philosophy.quote")} />
              </h2>
            </div>
          </div>

          {/* Right Column: Detailed narrative */}
          <div className="lg:col-span-4 lg:pl-8 flex flex-col justify-end">
            <p className="text-[16px] md:text-[18px] font-normal leading-relaxed text-[#1a1a1a]/80 mb-8">
              {t("philosophy.description")}
            </p>
            <div className="h-px bg-[#1a1a1a]/20 w-full mb-8" />
            <div className="flex justify-between items-center text-[12px] tracking-widest font-medium text-[#1a1a1a]/75 uppercase">
              <span>Optimized for Indonesia</span>
              <span>EST. 2026</span>
            </div>
          </div>
        </div>
      </section>

      {/* ── SECTION 4: WAITING LIST / CONVERSION ── */}
      <section
        id="join"
        aria-label={t("waitlist.headline")}
        className="relative flex flex-col min-h-screen px-6 md:px-16 lg:px-24 py-36 md:py-48 bg-[#1a1a1a] overflow-hidden border-t border-[#333333]/30"
      >
        {/* Background image layer */}
        <div aria-hidden className="absolute inset-0 pointer-events-none">
          <div
            className="absolute inset-0 bg-cover bg-no-repeat opacity-25 transition-opacity duration-500 bg-position-[65%_center] md:bg-right-center"
            style={{
              backgroundImage: "url('/waitlist_bg.png')",
              maskImage: "linear-gradient(to bottom, transparent 0%, black 15%, black 85%, transparent 100%)",
              WebkitMaskImage: "linear-gradient(to bottom, transparent 0%, black 15%, black 85%, transparent 100%)"
            }}
          />
          {/* Deep gradient tint for readability */}
          <div className="absolute inset-0 bg-linear-to-r from-[#1a1a1a]/98 via-[#1a1a1a]/90 to-[#1a1a1a]/60 md:from-[#1a1a1a]/98 md:via-[#1a1a1a]/85 md:to-[#1a1a1a]/40" />
          {/* Decorative ambient glow */}
          <div className="absolute top-1/4 right-0 w-[400px] h-[400px] bg-[#FFB300]/4 rounded-full blur-[100px] pointer-events-none" />
        </div>

        <div className="relative z-10 max-w-[1400px] w-full mx-auto flex flex-col flex-1">
          {/* ── Main Grid: Left copy + Right form ── */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-12 lg:gap-16 xl:gap-24 flex-1 items-center">
            {/* Left Column */}
            <div className="lg:col-span-7 flex flex-col justify-center">
              <h2 className="font-serif font-normal text-[36px] md:text-[48px] leading-[1.15] tracking-tight text-[#ffffff] mb-6">
                {t("waitlist.headline")}
              </h2>
              <p className="text-[16px] md:text-[18px] font-normal leading-relaxed text-[#a6a6a6] mb-10 max-w-[520px]">
                {t("waitlist.subheadline")}
              </p>

              {/* Benefits list */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-5 text-[14px] font-normal text-[#a6a6a6] border-t border-[#333333]/80 pt-8 mb-12">
                {[
                  t("waitlist.benefits.discount"),
                  t("waitlist.benefits.onboarding"),
                  t("waitlist.benefits.support"),
                  t("waitlist.benefits.noCard"),
                ].map((benefit) => (
                  <div key={benefit} className="flex gap-3 items-start">
                    <svg
                      className="h-5 w-5 text-[#FFB300] shrink-0 mt-0.5"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      strokeWidth={2}
                      aria-hidden
                    >
                      <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                    </svg>
                    <p className="leading-relaxed text-[#a6a6a6]">{benefit}</p>
                  </div>
                ))}
              </div>

              {/* Platform Stats Grid */}
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-0 border border-[#333333]/60 rounded-lg overflow-hidden">
                {[
                  { value: t("waitlist.stats.suppliersValue"), label: t("waitlist.stats.suppliersLabel") },
                  { value: t("waitlist.stats.categoriesValue"), label: t("waitlist.stats.categoriesLabel") },
                  { value: t("waitlist.stats.buyersValue"), label: t("waitlist.stats.buyersLabel") },
                  { value: t("waitlist.stats.savingsValue"), label: t("waitlist.stats.savingsLabel") },
                ].map((stat, idx) => (
                  <div
                    key={stat.label}
                    className={`flex flex-col justify-center p-4 sm:p-5 bg-[#1f1f1f]/60 backdrop-blur-sm ${
                      idx < 3 ? "border-r border-[#333333]/60" : ""
                    } ${idx >= 2 ? "border-t border-[#333333]/60 sm:border-t-0" : ""}`}
                  >
                    <span className="font-serif font-bold text-[#E27D18] leading-none mb-1 text-[20px] sm:text-[24px]">
                      {stat.value}
                    </span>
                    <span className="text-[11px] text-[#a6a6a6] font-normal leading-snug">{stat.label}</span>
                  </div>
                ))}
              </div>
            </div>

            {/* Right Column: Form */}
            <div className="lg:col-span-5 flex flex-col justify-center">
              {/* Form header above card */}

              <WaitingListForm />
            </div>
          </div>
        </div>
      </section>

      {/* ── Footer ── */}
      <footer className="py-8 px-6 md:px-16 lg:px-24 border-t border-[#333333] bg-[#1a1a1a]">
        <div className="max-w-[1400px] w-full mx-auto flex flex-col md:flex-row items-center justify-between gap-6 text-[13px] font-normal text-[#a6a6a6]">
          <div className="flex items-center gap-3">
            <Image
              src="/logo.png"
              alt="IndoSupplier Logo"
              width={120}
              height={24}
              className="h-5 w-auto object-contain"
            />
            <span className="font-normal tracking-widest text-[#ffffff] text-[13px]">
              IndoSupplier
            </span>
          </div>
          <span>
            {t("footer.copy", { year: new Date().getFullYear() })}
          </span>
          <span className="tracking-wider text-[#a6a6a6]/80">
            {t("footer.tagline")}
          </span>
        </div>
      </footer>
    </div>
  );
}
