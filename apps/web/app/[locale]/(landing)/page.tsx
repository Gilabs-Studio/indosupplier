import React from "react";
import { getTranslations } from "next-intl/server";
import Image from "next/image";
import WaitingListForm from "@/features/waiting-list/components/waiting-list-form";
import { Header } from "@/components/navigation/header";
import { ScrollTextReveal } from "@/components/motion";
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
        className="relative flex min-h-svh items-center justify-center overflow-hidden bg-background px-6 pt-28 pb-16 md:pt-32"
      >
        <div aria-hidden className="absolute inset-0">
          <div
            className="absolute inset-0 bg-center bg-cover bg-no-repeat opacity-50"
            style={{ backgroundImage: "url('/hero.png')" }}
          />
          <div className="absolute inset-0 bg-linear-to-b from-background/10 via-background/15 to-background/55" />
          <div className="absolute inset-x-0 top-0 h-32 bg-linear-to-b from-background/90 via-background/40 to-transparent" />
          <div className="absolute inset-x-0 bottom-0 h-56 bg-linear-to-b from-transparent via-background/30 to-background" />
        </div>

        <div className="relative z-10 mx-auto flex w-full max-w-5xl flex-col items-center justify-center text-center">
          <div className="relative isolate px-4 py-14 sm:px-8 sm:py-16">
            <div
              aria-hidden
              className="absolute -inset-x-56 top-1/2 -z-10 h-184 -translate-y-1/2 rounded-full bg-[radial-gradient(circle_at_center,rgba(255,255,255,1)_0%,rgba(255,255,250,1)_8%,rgba(255,253,242,0.98)_16%,rgba(255,247,228,0.9)_28%,rgba(255,238,196,0.64)_44%,rgba(255,232,170,0.28)_62%,transparent_86%)] blur-[150px] opacity-100 mix-blend-screen"
            />
            <div className="absolute -inset-x-[38%] -top-12 -z-10 h-56 rounded-full bg-[radial-gradient(circle_at_center,rgba(255,255,255,1)_0%,rgba(255,255,255,0.98)_14%,rgba(255,255,255,0.62)_32%,rgba(255,255,255,0.18)_58%,transparent_84%)] blur-[110px] opacity-90" />
            <div className="absolute -inset-x-[78%] -bottom-28 -z-10 h-96 rounded-full bg-[radial-gradient(circle_at_center,rgba(255,244,219,1)_0%,rgba(255,237,199,0.96)_16%,rgba(255,223,165,0.6)_36%,rgba(255,223,165,0.2)_60%,transparent_84%)] blur-[210px] opacity-100 mix-blend-screen" />
            <div className="absolute -inset-x-[34%] top-[10%] -z-10 h-52 rounded-full bg-[radial-gradient(circle_at_center,rgba(255,255,255,1)_0%,rgba(255,255,255,0.9)_18%,rgba(255,250,238,0.54)_40%,rgba(255,248,235,0.18)_64%,transparent_84%)] blur-[90px] opacity-100" />

            <h1
              className="mb-6 max-w-4xl font-bold leading-[1.08] tracking-[-0.04em] text-foreground drop-shadow-[0_8px_72px_rgba(255,248,235,1)] animate-fade-in"
              style={{ fontSize: "clamp(2.4rem, 5.7vw, 4.4rem)" }}
            >
              {t("hero.headline")}
            </h1>

            <p className="mx-auto mb-10 max-w-2xl text-[16px] font-normal leading-relaxed text-muted-foreground/85 animate-slide-up md:text-[18px]">
              {t("hero.subheadline")}
            </p>

            <div className="relative z-20 flex flex-wrap items-center justify-center gap-8 animate-slide-up delay-100">
              <RainbowButton asChild size="lg" className="text-[13px] font-semibold tracking-widest uppercase transition-all duration-300 hover:scale-[1.02] active:scale-[0.98]">
                <a href="#join">
                  {t("hero.cta")}
                </a>
              </RainbowButton>
            </div>

            <div className="mt-10 flex flex-wrap items-center justify-center gap-x-5 gap-y-3 text-xs font-medium text-muted-foreground/85 animate-slide-up delay-200 md:text-sm">
              <div className="flex items-center gap-2">
                <ShieldCheck className="h-4 w-4 shrink-0 text-[#d4af37]" />
                <span>{t("hero.badges.verified")}</span>
              </div>
              <div className="hidden h-3.5 w-px shrink-0 bg-border/70 md:block" />
              <div className="flex items-center gap-2">
                <Handshake className="h-4.5 w-4.5 shrink-0 text-[#d4af37]" />
                <span>{t("hero.badges.secure")}</span>
              </div>
              <div className="hidden h-3.5 w-px shrink-0 bg-border/70 md:block" />
              <div className="flex items-center gap-2">
                <Globe className="h-4 w-4 shrink-0 text-[#d4af37]" />
                <span>{t("hero.badges.direct")}</span>
              </div>
              <div className="hidden h-3.5 w-px shrink-0 bg-border/70 md:block" />
              <div className="flex items-center gap-2">
                <Award className="h-4 w-4 shrink-0 text-[#d4af37]" />
                <span>{t("hero.badges.certifications")}</span>
              </div>
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
