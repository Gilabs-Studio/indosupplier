"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Link } from "@/i18n/routing";
import { motion, AnimatePresence, type Variants } from "framer-motion";
import dynamic from "next/dynamic";
import { useTranslations } from "next-intl";
import { ThemeAwareImage } from "@/components/landing/theme-aware-image";
import {
  LANDING_THEME_IMAGES_BY_FEATURE,
  type LandingScreenshotKey,
} from "@/lib/landing-theme-images";
import {
  LayoutDashboard,
  Database,
  Package,
  CreditCard,
  Users,
  BarChart3,
  ArrowUpRight,
  Truck,
  PhoneCall,
  ChevronDown,
  LineChart,
  Sparkles,
  UserCheck,
} from "lucide-react";
import { LazyMount } from "@/features/general/dashboard/components/lazy-mount";

// WebGL background — single instance, fixed behind entire page
import type { LightRaysProps } from "@/components/landing/light-rays";
const LightRays = dynamic(
  () => import("@/components/landing/light-rays").then((m) => m.default as React.ComponentType<LightRaysProps>),
  { ssr: false }
);

// Leaflet map — requires browser environment
const MarketingMapDemo = dynamic(
  () =>
    import("@/components/landing/marketing-map-demo").then(
      (m) => m.MarketingMapDemo
    ),
  { ssr: false }
);

const PricingSection = dynamic(
  () => import("@/components/landing/pricing-section").then((m) => m.PricingSection),
  {
    ssr: false,
    loading: () => <div className="min-h-[600px] w-full" aria-hidden="true" />,
  }
);

const PRIMARY_HEX = "#FFFFFF";

const fadeUp: Variants = {
  hidden: { opacity: 0, y: 32 },
  visible: (delay = 0) => ({
    opacity: 1,
    y: 0,
    transition: { duration: 0.8, ease: "easeOut", delay },
  }),
};

// Common content width constraint
const W = "relative z-10 mx-auto w-full max-w-6xl px-6 lg:px-12";

const MODULE_ICONS = [
  LayoutDashboard,
  Database,
  BarChart3,
  Truck,
  Package,
  CreditCard,
  Users,
  PhoneCall,
  LineChart,
  Sparkles,
  UserCheck,
];

const BENEFIT_ICONS = [LayoutDashboard, Package, CreditCard, Users, LineChart, Sparkles];

function FaqItem({
  q,
  a,
  index,
}: {
  q: string;
  a: string;
  index: number;
}) {
  const [open, setOpen] = useState(false);
  return (
    <motion.div
      variants={fadeUp}
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true, margin: "-20px" }}
      custom={index * 0.07}
      className="border-b border-border last:border-0"
    >
      <button
        onClick={() => setOpen((v) => !v)}
        className="flex w-full cursor-pointer items-start justify-between gap-4 py-6 text-left"
        aria-expanded={open}
      >
        <span className="text-sm font-medium text-foreground sm:text-base">
          {q}
        </span>
        <ChevronDown
          className={`mt-0.5 h-4 w-4 shrink-0 text-muted-foreground transition-transform duration-300 ${
            open ? "rotate-180" : ""
          }`}
        />
      </button>
      <AnimatePresence initial={false}>
        {open && (
          <motion.div
            key="body"
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.35, ease: "easeOut" }}
            className="overflow-hidden"
          >
            <p className="pb-6 text-sm leading-relaxed text-muted-foreground">
              {a}
            </p>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  );
}

export default function LandingPage() {
  const [activeScreenshot, setActiveScreenshot] = useState(0);
  const [showBackgroundFx, setShowBackgroundFx] = useState(false);
  const [followBeamPointer, setFollowBeamPointer] = useState(false);

  const t = useTranslations("landing");

  useEffect(() => {
    // Defer WebGL background rendering to allow hero section to paint first
    // This improves LCP by avoiding GPU initialization blocking main thread
    const timerId = window.setTimeout(() => {
      setShowBackgroundFx(true);
    }, 1200);

    return () => window.clearTimeout(timerId);
  }, []);

  useEffect(() => {
    const mediaQuery = window.matchMedia("(pointer: fine) and (min-width: 1024px)");

    const updatePointerBehavior = () => {
      setFollowBeamPointer(mediaQuery.matches);
    };

    updatePointerBehavior();
    mediaQuery.addEventListener("change", updatePointerBehavior);

    return () => mediaQuery.removeEventListener("change", updatePointerBehavior);
  }, []);

  const MODULE_KEYS = [
    "dashboard",
    "masterData",
    "sales",
    "purchase",
    "stock",
    "finance",
    "hrd",
    "hr",
    "crm",
    "reports",
    "aiAssistant",
  ] as const;

  const BENEFIT_KEYS = [
    "singleSource",
    "operationalControl",
    "financeAudit",
    "hrdDiscipline",
    "insightFast",
    "automation",
  ] as const;

  const SCREENSHOT_KEYS: readonly LandingScreenshotKey[] = [
    "pipeline",
    "dashboard",
    "salesOrder",
    "stockInventory",
    "profitLoss",
    "customer",
    "geoPerformance",
    "salary",
  ] as const;

  const FAQ_KEYS = ["q1", "q2", "q3", "q4", "q5", "q6"] as const;

  return (
    <>
      {/* ── Single global WebGL background (fixed, below everything) ── */}
      <div className="fixed inset-0 z-0 pointer-events-none" aria-hidden="true">
        {showBackgroundFx ? (
          <LightRays
            raysOrigin="top-center"
            raysColor={PRIMARY_HEX}
            raysSpeed={0.35}
            lightSpread={2.0}
            rayLength={2.8}
            fadeDistance={0.92}
            saturation={0.5}
            pulsating
            followMouse={followBeamPointer}
            mouseInfluence={0.07}
          />
        ) : null}
      </div>

      <main className="relative z-10 overflow-x-hidden" id="main-content">

        {/* ════════════════════════════════════════════════════════
            1. HERO — with dashboard screenshot below
        ════════════════════════════════════════════════════════ */}
        <section className="flex flex-col items-center pt-32 pb-0 text-center">
          <div className={W}>
            <motion.p
              variants={fadeUp}
              initial={false}
              animate="visible"
              custom={0}
              className="mb-6 text-xs font-semibold uppercase tracking-[0.22em] text-primary"
            >
              {t("hero.badge")}
            </motion.p>

            <motion.h1
              variants={fadeUp}
              initial={false}
              animate="visible"
              custom={0.1}
              style={{ opacity: 1, transform: "none" }}
              className="mx-auto max-w-5xl text-center text-balance text-5xl font-light leading-[1.06] tracking-tight text-foreground sm:text-6xl lg:text-[5.5rem]"
            >
              {t("hero.title1")}
              <br />
              <span className="inline-block text-[3.5rem] font-medium leading-[1.1] tracking-normal text-primary italic font-accent sm:text-[4.5rem] lg:text-[5.5rem]">
                {t("hero.title2")}
              </span>
            </motion.h1>

            <motion.p
              variants={fadeUp}
              initial={false}
              animate="visible"
              custom={0.2}
              className="mx-auto mt-8 max-w-xl text-base leading-relaxed text-muted-foreground sm:text-lg"
            >
              {t("hero.subtitle")}
            </motion.p>

            {/* Minimal hero — no tracking input for a clean Apple-like look */}
            <motion.div
              variants={fadeUp}
              initial="hidden"
              animate="visible"
              custom={0.3}
              className="mx-auto mt-8 w-full max-w-2xl"
            />

            {/* Clean screenshot — simplified frame for minimalist aesthetic */}
            <motion.div
              variants={fadeUp}
              initial={false}
              animate="visible"
              custom={0.45}
              className="relative mt-12 w-full overflow-hidden rounded-2xl bg-card/60 shadow-lg"
            >
              <div className="relative aspect-video w-full">
                <ThemeAwareImage
                  lightSrc={LANDING_THEME_IMAGES_BY_FEATURE.dashboard.light}
                  darkSrc={LANDING_THEME_IMAGES_BY_FEATURE.dashboard.dark}
                  alt="SalesView Dashboard"
                  fill
                  className="object-cover object-top"
                  sizes="(max-width: 1200px) 100vw, 1200px"
                  priority
                  fetchPriority="high"
                />
              </div>
            </motion.div>
          </div>
        </section>

        {/* ════════════════════════════════════════════════════════
            2. BENEFITS GRID
        ════════════════════════════════════════════════════════ */}
        <section id="features" className="flex min-h-screen flex-col items-center justify-center py-24">
          <div className={W}>
            <motion.div
              variants={fadeUp}
              initial="hidden"
              whileInView="visible"
              viewport={{ once: true, margin: "-80px" }}
              className="mx-auto max-w-3xl text-center"
            >
              <p className="mb-5 text-xs font-semibold uppercase tracking-[0.22em] text-muted-foreground">
                {t("benefits.eyebrow")}
              </p>
              <h2 className="text-4xl font-light leading-tight tracking-tight text-foreground sm:text-5xl">
                {t("benefits.title")}
              </h2>
              <p className="mt-6 text-base leading-relaxed text-muted-foreground sm:text-lg">
                {t("benefits.subtitle")}
              </p>
            </motion.div>

            <div className="mt-20 grid gap-8 sm:grid-cols-2 lg:grid-cols-3">
              {BENEFIT_KEYS.map((key, index) => {
                const Icon = BENEFIT_ICONS[index];
                return (
                  <motion.div
                    key={key}
                    variants={fadeUp}
                    initial="hidden"
                    whileInView="visible"
                    viewport={{ once: true, margin: "-40px" }}
                    custom={index * 0.1}
                    className="overflow-hidden rounded-3xl border border-border bg-card/40 p-8 shadow-sm backdrop-blur-sm"
                  >
                    <div className="mb-6 inline-flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary">
                      <Icon className="h-6 w-6" />
                    </div>
                    <h3 className="mb-3 text-lg font-medium text-foreground">
                      {t(`benefits.items.${key}.title`)}
                    </h3>
                    <p className="text-sm leading-relaxed text-muted-foreground">
                      {t(`benefits.items.${key}.description`)}
                    </p>
                  </motion.div>
                );
              })}
            </div>
          </div>
        </section>

        {/* ════════════════════════════════════════════════════════
            2.5. SCREENSHOT SHOWCASE TABS
        ════════════════════════════════════════════════════════ */}
        <section className="flex flex-col items-center justify-center py-24 relative">
          <div className="absolute inset-0 bg-muted/10 -z-10" />
          <div className={`${W} max-w-7xl`}>
            <motion.div
              variants={fadeUp}
              initial="hidden"
              whileInView="visible"
              viewport={{ once: true, margin: "-80px" }}
              className="mb-12 text-center"
            >
              <h2 className="text-3xl font-light tracking-tight text-foreground sm:text-4xl">
                {t("screenshots.title")}
              </h2>
            </motion.div>

            {/* TAB BUTTONS */}
            <div className="mb-8 flex overflow-x-auto pb-4 scrollbar-hide justify-start lg:justify-center">
              <div className="flex gap-2 px-4 lg:px-0">
                {SCREENSHOT_KEYS.map((key, index) => (
                  <button
                    key={key}
                    onClick={() => setActiveScreenshot(index)}
                    className={`whitespace-nowrap rounded-full px-5 py-2.5 text-sm cursor-pointer ${
                      activeScreenshot === index
                        ? "bg-primary text-primary-foreground font-medium"
                        : "bg-muted text-muted-foreground hover:bg-muted/80"
                    }`}
                  >
                    {t(`screenshots.items.${key}.title`)}
                  </button>
                ))}
              </div>
            </div>

            {/* TAB CONTENT */}
            <motion.div
              variants={fadeUp}
              initial="hidden"
              whileInView="visible"
              viewport={{ once: true, margin: "-40px" }}
              className="relative overflow-hidden rounded-2xl border border-border bg-card shadow-sm mx-auto"
            >
              <div className="p-4 sm:p-6 border-b border-border/50 bg-card/80 backdrop-blur-sm">
                <h3 className="text-lg font-semibold text-foreground">
                  {t(`screenshots.items.${SCREENSHOT_KEYS[activeScreenshot]}.title`)}
                </h3>
                <p className="text-sm text-muted-foreground">
                  {t(`screenshots.items.${SCREENSHOT_KEYS[activeScreenshot]}.caption`)}
                </p>
              </div>
              <div className="relative aspect-video w-full overflow-hidden bg-muted/20">
                <ThemeAwareImage
                  lightSrc={LANDING_THEME_IMAGES_BY_FEATURE[SCREENSHOT_KEYS[activeScreenshot]].light}
                  darkSrc={LANDING_THEME_IMAGES_BY_FEATURE[SCREENSHOT_KEYS[activeScreenshot]].dark}
                  alt={t(`screenshots.items.${SCREENSHOT_KEYS[activeScreenshot]}.title`)}
                  fill
                  className="object-contain p-2 sm:p-4"
                  sizes="(max-width: 1200px) 100vw, 1200px"
                  loading="lazy"
                />
              </div>
            </motion.div>
          </div>
        </section>

        {/* ════════════════════════════════════════════════════════
            3. TURNING POINT
        ════════════════════════════════════════════════════════ */}
        <section className="flex min-h-screen flex-col items-center justify-center py-24 text-center">
          <div className={W}>
            <motion.div
              variants={fadeUp} initial="hidden"
              whileInView="visible" viewport={{ once: true, margin: "-120px" }}
              className="mx-auto max-w-4xl"
            >
              <h2 className="text-4xl font-light leading-[1.1] tracking-tight text-foreground sm:text-5xl lg:text-6xl">
                {t("turningPoint.title1")}
                <br />
                {t("turningPoint.title2")}
                <br />
                <span className="text-primary font-accent font-medium italic text-4xl sm:text-5xl lg:text-6xl inline-block mt-1 tracking-normal">{t("turningPoint.title3")}</span>
              </h2>
              <p className="mt-8 text-base leading-relaxed text-muted-foreground sm:text-lg">
                {t("turningPoint.subtitle")}
              </p>
            </motion.div>
          </div>
        </section>

        {/* ════════════════════════════════════════════════════════
            4. FEATURE: MAP (real Leaflet)
        ════════════════════════════════════════════════════════ */}
        <section id="features" className="flex min-h-screen flex-col items-center justify-center py-24">
          <div className={W}>
            <div className="grid gap-14 lg:grid-cols-2 lg:gap-20 items-center">
              <motion.div
                variants={fadeUp} initial="hidden"
                whileInView="visible" viewport={{ once: true, margin: "-80px" }}
              >
                <h2 className="mb-5 text-3xl font-light leading-tight tracking-tight text-foreground sm:text-4xl lg:text-5xl">
                  {t("map.title1")}
                  <br />
                  {t("map.title2")}
                </h2>
                <p className="mb-5 text-base leading-relaxed text-muted-foreground">
                  {t("map.subtitle")}
                </p>
                <p className="text-sm leading-relaxed text-muted-foreground">
                  {t("map.detail")}
                </p>
              </motion.div>

              <motion.div
                variants={fadeUp} initial="hidden"
                whileInView="visible" viewport={{ once: true, margin: "-60px" }}
                custom={0.1}
                className="relative h-80 lg:h-[420px] overflow-hidden rounded-2xl border border-border bg-card/60 backdrop-blur-sm shadow-sm"
              >
                {/* useRemote=false: never call protected APIs from the public landing page */}
                <LazyMount rootMargin="280px">
                  <MarketingMapDemo useRemote={false} />
                </LazyMount>

                {/* Legend overlay */}
                <div className="pointer-events-none absolute top-3 right-3 z-500 flex flex-col gap-1.5 rounded-xl border border-border bg-card/90 px-3 py-2.5 backdrop-blur-sm">
                  <div className="flex items-center gap-2">
                    <div className="h-2.5 w-2.5 rounded-full bg-primary opacity-90" />
                    <span className="text-xs text-muted-foreground">{t("map.legend.mainWarehouse")}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="h-2 w-2 rounded-full bg-primary opacity-50" />
                    <span className="text-xs text-muted-foreground">{t("map.legend.distributionPoint")}</span>
                  </div>
                </div>

                <div className="pointer-events-none absolute bottom-3 left-3 z-500">
                  <span className="rounded-lg border border-primary/25 bg-card/90 px-3 py-1.5 text-xs font-medium text-primary backdrop-blur-sm">
                    {t("map.badge")}
                  </span>
                </div>
              </motion.div>
            </div>
          </div>
        </section>

        {/* ════════════════════════════════════════════════════════
            5. FEATURE: AI
        ════════════════════════════════════════════════════════ */}
        <section className="flex min-h-screen flex-col items-center justify-center py-24">
          <div className={W}>
            <div className="grid gap-14 lg:grid-cols-2 lg:gap-20 items-center">
              <motion.div
                variants={fadeUp} initial="hidden"
                whileInView="visible" viewport={{ once: true, margin: "-60px" }}
                className="order-2 lg:order-1 overflow-hidden rounded-2xl border border-border bg-card/60 backdrop-blur-sm shadow-sm"
              >
                <div className="flex items-center gap-2 border-b border-border bg-muted/40 px-5 py-3.5">
                  <div className="h-3 w-3 rounded-full bg-destructive/50" />
                  <div className="h-3 w-3 rounded-full bg-chart-3/50" />
                  <div className="h-3 w-3 rounded-full bg-chart-2/50" />
                  <span className="ml-3 font-mono text-xs text-muted-foreground">SalesView AI Assistant</span>
                </div>
                <div className="space-y-5 p-6 font-mono text-sm">
                  <div className="space-y-1.5">
                    <p className="text-xs font-semibold text-muted-foreground">{t("ai.chat.labelUser")}</p>
                    <p className="rounded-lg bg-primary/8 px-4 py-2.5 text-sm text-foreground">
                      {t("ai.chat.user1")}
                    </p>
                  </div>
                  <div className="space-y-1.5">
                    <p className="text-xs font-semibold text-primary">{t("ai.chat.labelAi")}</p>
                    <p className="rounded-lg border border-border bg-card/50 px-4 py-2.5 text-sm leading-relaxed text-muted-foreground">
                      {t("ai.chat.ai1", { count: 342 })}
                    </p>
                  </div>
                  <div className="space-y-1.5">
                    <p className="text-xs font-semibold text-muted-foreground">{t("ai.chat.labelUser")}</p>
                    <p className="rounded-lg bg-primary/8 px-4 py-2.5 text-sm text-foreground">
                      {t("ai.chat.user2")}
                    </p>
                  </div>
                  <div className="space-y-1.5">
                    <p className="text-xs font-semibold text-primary">{t("ai.chat.labelAi")}</p>
                    <p className="rounded-lg border border-border bg-card/50 px-4 py-2.5 text-sm leading-relaxed text-muted-foreground">
                      {t("ai.chat.ai2", { po: "PO-2026-0847" })}
                    </p>
                  </div>
                </div>
              </motion.div>

              <motion.div
                variants={fadeUp} initial="hidden"
                whileInView="visible" viewport={{ once: true, margin: "-80px" }}
                custom={0.1}
                className="order-1 lg:order-2"
              >
                <h2 className="mb-5 text-3xl font-light leading-tight tracking-tight text-foreground sm:text-4xl lg:text-5xl">
                  {t("ai.title1")}
                  <br />
                  {t("ai.title2")}
                </h2>
                <p className="mb-5 text-base leading-relaxed text-muted-foreground">
                  {t("ai.subtitle")}
                </p>
                <p className="text-sm leading-relaxed text-muted-foreground">
                  {t("ai.detail")}
                </p>
              </motion.div>
            </div>
          </div>
        </section>

        <LazyMount rootMargin="320px">
          <PricingSection />
        </LazyMount>

        {/* ════════════════════════════════════════════════════════
            7. MODULES
        ════════════════════════════════════════════════════════ */}
        <section id="modules" className="flex min-h-screen flex-col items-center justify-center py-24">
          <div className={W}>
            <motion.div
              variants={fadeUp} initial="hidden"
              whileInView="visible" viewport={{ once: true, margin: "-80px" }}
              className="mx-auto mb-14 max-w-2xl text-center"
            >
              <p className="mb-5 text-xs font-semibold uppercase tracking-[0.22em] text-muted-foreground">
                {t("modules.eyebrow")}
              </p>
              <h2 className="text-4xl font-light tracking-tight text-foreground sm:text-5xl">
                {t("modules.title1")}
                <br />
                {t("modules.title2")}
              </h2>
              <p className="mt-5 text-base leading-relaxed text-muted-foreground">
                {t("modules.subtitle")}
              </p>
            </motion.div>

            <div className="grid gap-px bg-border/50 rounded-2xl overflow-hidden sm:grid-cols-2 lg:grid-cols-3">
              {MODULE_KEYS.map((key, i) => {
                const Icon = MODULE_ICONS[i];
                return (
                  <motion.div
                    key={key}
                    variants={fadeUp} initial="hidden"
                    whileInView="visible" viewport={{ once: true, margin: "-30px" }}
                    custom={i * 0.07}
                    className="group flex flex-col gap-4 bg-card/60 p-8 backdrop-blur-sm transition-colors hover:bg-card/80"
                  >
                    <div className="flex h-9 w-9 items-center justify-center rounded-lg border border-border bg-background/60 transition-colors group-hover:border-primary/30 group-hover:bg-primary/5">
                      <Icon className="h-[18px] w-[18px] text-muted-foreground transition-colors group-hover:text-primary" />
                    </div>
                    <div>
                      <h3 className="mb-1 text-sm font-semibold text-foreground">
                        {t(`modules.items.${key}.title`)}
                      </h3>
                      <p className="text-sm leading-relaxed text-muted-foreground">
                        {t(`modules.items.${key}.description`)}
                      </p>
                    </div>
                  </motion.div>
                );
              })}
            </div>
          </div>
        </section>

        {/* ════════════════════════════════════════════════════════
          8. FAQ
        ════════════════════════════════════════════════════════ */}
        <section id="faq" className="flex min-h-screen flex-col items-center justify-center py-24">
          <div className={W}>
            <div className="grid gap-16 lg:grid-cols-[1fr_1.5fr] lg:gap-24 items-start">
              <motion.div
                variants={fadeUp} initial="hidden"
                whileInView="visible" viewport={{ once: true, margin: "-80px" }}
                className="lg:sticky lg:top-24"
              >
                <p className="mb-5 text-xs font-semibold uppercase tracking-[0.22em] text-muted-foreground">
                  {t("faq.eyebrow")}
                </p>
                <h2 className="text-4xl font-light leading-tight tracking-tight text-foreground sm:text-5xl">
                  {t("faq.title1")}
                  <br />
                  {t("faq.title2")}
                </h2>
                <p className="mt-5 text-base leading-relaxed text-muted-foreground">
                  {t("faq.subtitle")}
                </p>
                <div className="mt-8">
                  <Link href="/login">
                    <Button variant="outline" className="cursor-pointer bg-card/60 backdrop-blur-sm">
                      {t("faq.contactCta")}
                      <ArrowUpRight className="ml-1.5 h-3.5 w-3.5" />
                    </Button>
                  </Link>
                </div>
              </motion.div>

              <div className="divide-y divide-border rounded-2xl border border-border bg-card/50 px-6 backdrop-blur-sm">
                {FAQ_KEYS.map((key, i) => (
                  <FaqItem
                    key={key}
                    q={t(`faq.items.${key}`)}
                    a={t(`faq.items.${key.replace("q", "a")}`)}
                    index={i}
                  />
                ))}
              </div>
            </div>
          </div>
        </section>

        {/* ════════════════════════════════════════════════════════
          9. CTA
        ════════════════════════════════════════════════════════ */}
        <section className="flex min-h-screen flex-col items-center justify-center py-24 text-center">
          <div className={W}>
            <motion.div
              variants={fadeUp} initial="hidden"
              whileInView="visible" viewport={{ once: true, margin: "-80px" }}
              className="mx-auto max-w-2xl"
            >
              <h2 className="mb-5 text-4xl font-light leading-tight tracking-tight text-foreground sm:text-5xl lg:text-6xl">
                {t("cta.title1")}
                <br />
                {t("cta.title2")}
              </h2>
              <p className="mb-10 text-base leading-relaxed text-muted-foreground">
                {t("cta.subtitle")}
              </p>
              <div className="flex flex-col items-center gap-4 sm:flex-row sm:justify-center">
                <Link href="#features">
                  <Button size="lg" className="cursor-pointer px-10 text-base shadow-md">
                    {t("cta.seeFeatures")}
                  </Button>
                </Link>
                <Link href="/login">
                  <Button
                    variant="outline"
                    size="lg"
                    className="cursor-pointer px-8 text-base bg-card/60 backdrop-blur-sm"
                  >
                    {t("cta.signIn")}
                    <ArrowUpRight className="ml-1.5 h-4 w-4" />
                  </Button>
                </Link>
              </div>
            </motion.div>
          </div>
        </section>

      </main>
    </>
  );
}
