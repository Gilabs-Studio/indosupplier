import React from "react";
import { getTranslations } from "next-intl/server";
import { Link } from "@/i18n/routing";
import LanguageSwitcher from "@/components/navigation/language-switcher";
import WaitingListForm from "@/features/waiting-list/components/waiting-list-form";

export default async function LandingPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: "landing" });

  return (
    <div className="min-h-screen bg-[#EAE6DF] text-neutral-800 font-jost antialiased selection:bg-neutral-900 selection:text-white">
      {/* ── Navigation Header ── */}
      <header className="w-full py-6 px-6 md:px-16 lg:px-24 z-50">
        <div className="max-w-[1400px] mx-auto flex items-center justify-between">
          <div className="flex items-center gap-3">
            <img src="/logo.png" alt="IndoSupplier Logo" className="h-6 w-auto object-contain" />
            <span className="font-normal text-[15px] tracking-widest uppercase text-neutral-900">
              IndoSupplier
            </span>
          </div>

          {/* Center menu links */}
          <div className="hidden md:flex items-center gap-12 text-[12px] tracking-widest uppercase font-light text-neutral-500">
            <a href="#features" className="hover:text-neutral-900 transition-colors">
              {t("features.badge")}
            </a>
            <a href="#about" className="hover:text-neutral-900 transition-colors">
              About
            </a>
            <a href="#join" className="hover:text-neutral-900 transition-colors">
              Waitlist
            </a>
          </div>

          {/* Right menu actions */}
          <div className="flex items-center gap-6 text-[12px] font-medium tracking-widest uppercase">
            <LanguageSwitcher currentLocale={locale} />
            <Link
              href="/login"
              className="hover:underline text-neutral-850"
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
              className="font-light leading-[1.08] tracking-[-0.04em] text-neutral-900 mb-8 max-w-[850px] animate-fade-in"
              style={{ fontSize: "clamp(2.8rem, 5.5vw, 4.8rem)" }}
            >
              {t("hero.headline")}
            </h1>

            {/* Subheadline */}
            <p className="text-[17px] md:text-[19px] font-light leading-relaxed text-neutral-600 max-w-[580px] mb-12 animate-slide-up">
              {t("hero.subheadline")}
            </p>

            {/* Call to action */}
            <div className="flex flex-wrap items-center gap-8 animate-slide-up">
              <a
                href="#join"
                className="bg-neutral-900 text-white text-[13px] tracking-widest uppercase font-medium px-8 py-4 hover:bg-neutral-800 transition-all duration-300"
              >
                {t("hero.cta")}
              </a>
              <span className="text-[13px] tracking-wider uppercase font-light text-neutral-500">
                {t("hero.trustLabel")}
              </span>
            </div>
          </div>

          {/* Hero Visual Right */}
          <div className="lg:col-span-5 flex justify-center animate-fade-in">
              <img
                src="/hero2.png"
                alt="IndoSupplier Platform Visual"
                className="w-full h-full object-cover rounded-md"
              />
          </div>
        </div>
      </section>

      {/* ── SECTION 2: FEATURES / CAPABILITIES ── */}
      <section id="features" className="px-6 md:px-16 lg:px-24 py-20 md:py-28 bg-[#F2F2F0]/50 border-t border-neutral-200/50">
        <div className="max-w-[1400px] w-full mx-auto">
          <div className="mb-16">
            <span className="text-[11px] uppercase tracking-widest font-medium text-neutral-500 block mb-3">
              {t("features.badge")}
            </span>
            <h2 className="text-[28px] md:text-[36px] font-light tracking-tight text-neutral-900">
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
                className="bg-white/40 border border-neutral-200/60 p-8 rounded-lg hover:bg-white hover:border-neutral-300 transition-all duration-300 flex flex-col justify-between min-h-[200px]"
              >
                <span className="text-[12px] uppercase tracking-wider font-light text-neutral-400">
                  0{idx + 1}
                </span>
                <div className="mt-12">
                  <h3 className="text-[18px] font-normal text-neutral-900 mb-3">
                    {card.label}
                  </h3>
                  <p className="text-[14px] font-light leading-relaxed text-neutral-600">
                    {card.desc}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── SECTION 3: ABOUT / PHILOSOPHY ── */}
      <section id="about" className="px-6 md:px-16 lg:px-24 py-20 md:py-28 bg-neutral-900 text-neutral-100">
        <div className="max-w-[1400px] w-full mx-auto grid grid-cols-1 lg:grid-cols-12 gap-12">
          {/* Left Column: Huge typography statement */}
          <div className="lg:col-span-8 flex flex-col justify-between">
            <div>
              <span className="text-[11px] uppercase tracking-widest font-medium text-neutral-400 block mb-8">
                {t("philosophy.title")}
              </span>
              <h2
                className="font-light leading-[1.2] tracking-[-0.03em] max-w-[720px] text-white"
                style={{ fontSize: "clamp(2rem, 4.5vw, 3.4rem)" }}
              >
                {t("philosophy.quote")}
              </h2>
            </div>
          </div>

          {/* Right Column: Detailed narrative */}
          <div className="lg:col-span-4 lg:pl-8 flex flex-col justify-end">
            <p className="text-[16px] font-light leading-relaxed text-neutral-400 mb-8">
              {t("philosophy.description")}
            </p>
            <div className="h-px bg-neutral-800 w-full mb-8" />
            <div className="flex justify-between items-center text-[13px] uppercase tracking-wider font-light text-neutral-400">
              <span>Optimized for Indonesia</span>
              <span>EST. 2026</span>
            </div>
          </div>
        </div>
      </section>

      {/* ── SECTION 4: WAITING LIST / CONVERSION ── */}
      <section id="join" className="px-6 md:px-16 lg:px-24 py-20 md:py-28 bg-[#F2F2F0]">
        <div className="max-w-[1400px] w-full mx-auto grid grid-cols-1 lg:grid-cols-2 gap-16 lg:gap-24 items-center">
          {/* Left Column */}
          <div>
            <span className="text-[11px] uppercase tracking-widest font-medium text-neutral-500 block mb-6">
              {t("waitlist.badge")}
            </span>
            <h2
              className="font-light leading-[1.08] tracking-[-0.04em] text-neutral-900 mb-6"
              style={{ fontSize: "clamp(2.4rem, 5vw, 4rem)" }}
            >
              {t("waitlist.headline")}
            </h2>
            <p className="text-[17px] font-light leading-relaxed text-neutral-600 mb-12 max-w-[500px]">
              {t("waitlist.subheadline")}
            </p>

            {/* Structured list in grid */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-8 text-[14px] font-light text-neutral-600 border-t border-neutral-300/80 pt-8">
              {[
                { label: "30% off first year", desc: "Special pricing for early members." },
                { label: "Onboarding session", desc: "Dedicated guidance from team." },
                { label: "No credit card", desc: "No risk, join list immediately." },
                { label: "Priority feedback", desc: "Request features directly." },
              ].map((item) => (
                <div key={item.label}>
                  <h4 className="font-normal text-neutral-900 mb-1 uppercase tracking-wider text-[12px]">
                    {item.label}
                  </h4>
                  <p className="leading-relaxed">{item.desc}</p>
                </div>
              ))}
            </div>
          </div>

          {/* Right Column (Form container) */}
          <div className="bg-white p-8 md:p-12 border border-neutral-200/80">
            <WaitingListForm />
          </div>
        </div>
      </section>

      {/* ── Footer ── */}
      <footer className="py-12 px-6 md:px-16 lg:px-24 border-t border-neutral-200/85 bg-[#EAE6DF]">
        <div className="max-w-[1400px] w-full mx-auto flex flex-col md:flex-row items-center justify-between gap-6 text-[13px] font-light text-neutral-500">
          <div className="flex items-center gap-3">
            <img src="/logo.png" alt="IndoSupplier Logo" className="h-5 w-auto object-contain" />
            <span className="font-normal tracking-widest uppercase text-neutral-900 text-[13px]">
              IndoSupplier
            </span>
          </div>
          <span>
            {t("footer.copy", { year: new Date().getFullYear() })}
          </span>
          <span className="uppercase tracking-wider">
            {t("footer.tagline")}
          </span>
        </div>
      </footer>
    </div>
  );
}
