import React from "react";
import { useTranslations } from "next-intl";
import { getTranslations } from "next-intl/server";
import { Link } from "@/i18n/routing";
import WaitingListForm from "@/features/waiting-list/components/waiting-list-form";
import { ShieldCheck, BarChart3, Database, Globe, Landmark, Sparkles, Building2, Layers, CheckCircle } from "lucide-react";
import LanguageSwitcher from "@/components/navigation/language-switcher";

export default async function LandingPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  const t = await getTranslations("landing");
  const tWaitlist = await getTranslations("waitingList");

  return (
    <div className="min-h-screen bg-neutral-50 text-neutral-900 dark:bg-neutral-950 dark:text-neutral-50 transition-colors duration-300">
      {/* Background Decorative Grid */}
      <div className="absolute inset-0 bg-[linear-gradient(to_right,#8080800a_1px,transparent_1px),linear-gradient(to_bottom,#8080800a_1px,transparent_1px)] bg-[size:14px_24px] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_100%)] pointer-events-none" />

      {/* Header */}
      <header className="sticky top-0 z-50 border-b border-neutral-200/80 dark:border-neutral-800/80 bg-white/80 dark:bg-neutral-950/80 backdrop-blur-md">
        <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span className="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-tr from-blue-600 to-indigo-600 text-white font-bold text-lg shadow-md shadow-blue-500/20">
              IS
            </span>
            <span className="font-extrabold text-xl tracking-tight bg-clip-text text-transparent bg-gradient-to-r from-neutral-900 to-neutral-700 dark:from-white dark:to-neutral-300">
              IndoSupplier
            </span>
          </div>

          <div className="flex items-center gap-4">
            <LanguageSwitcher currentLocale={locale} />
            <Link
              href="/login"
              className="rounded-lg px-4 py-2 text-sm font-semibold text-neutral-700 dark:text-neutral-300 hover:bg-neutral-100 dark:hover:bg-neutral-800 transition-colors"
            >
              Sign In
            </Link>
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <section className="relative pt-20 pb-24 overflow-hidden">
        <div className="max-w-7xl mx-auto px-6 grid grid-cols-1 lg:grid-cols-12 gap-12 items-center">
          {/* Hero Copy */}
          <div className="lg:col-span-6 space-y-6">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-blue-200 dark:border-blue-900/50 bg-blue-50 dark:bg-blue-950/30 text-blue-600 dark:text-blue-400 text-xs font-semibold">
              <Sparkles className="h-3 w-3 animate-spin" />
              <span>⚡ Special Early Access</span>
            </div>

            <h1 className="text-4xl md:text-5xl lg:text-6xl font-extrabold tracking-tight text-neutral-950 dark:text-white leading-[1.1] bg-clip-text">
              {t("title")}
            </h1>

            <p className="text-lg text-neutral-600 dark:text-neutral-400 leading-relaxed max-w-xl">
              {t("subtitle")}
            </p>

            <div className="space-y-3.5 pt-4">
              <div className="flex items-center gap-2 text-neutral-700 dark:text-neutral-300">
                <CheckCircle className="h-5 w-5 text-emerald-500 flex-shrink-0" />
                <span>Multi-Warehouse & Stock Control</span>
              </div>
              <div className="flex items-center gap-2 text-neutral-700 dark:text-neutral-300">
                <CheckCircle className="h-5 w-5 text-emerald-500 flex-shrink-0" />
                <span>Indonesian e-Faktur & Tax Automation</span>
              </div>
              <div className="flex items-center gap-2 text-neutral-700 dark:text-neutral-300">
                <CheckCircle className="h-5 w-5 text-emerald-500 flex-shrink-0" />
                <span>Interactive CRM and Sales Pipeline</span>
              </div>
            </div>
          </div>

          {/* Waitlist Form */}
          <div className="lg:col-span-6">
            <WaitingListForm />
          </div>
        </div>
      </section>

      {/* Features Grid */}
      <section className="py-20 border-t border-neutral-200/60 dark:border-neutral-800/60 bg-neutral-100/50 dark:bg-neutral-900/30">
        <div className="max-w-7xl mx-auto px-6 space-y-12">
          <div className="text-center space-y-4 max-w-3xl mx-auto">
            <h2 className="text-3xl font-bold tracking-tight text-neutral-950 dark:text-white sm:text-4xl">
              {t("featuresTitle")}
            </h2>
            <p className="text-lg text-neutral-600 dark:text-neutral-400">
              {t("featuresSubtitle")}
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {/* ERP */}
            <div className="rounded-xl border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-900 p-6 shadow-sm hover:shadow-md transition-shadow">
              <span className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-100 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400 mb-4">
                <Database className="h-5 w-5" />
              </span>
              <h3 className="text-xl font-semibold mb-2 text-neutral-950 dark:text-white">
                {t("featureList.erp.title")}
              </h3>
              <p className="text-neutral-600 dark:text-neutral-400 leading-relaxed text-sm">
                {t("featureList.erp.desc")}
              </p>
            </div>

            {/* CRM */}
            <div className="rounded-xl border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-900 p-6 shadow-sm hover:shadow-md transition-shadow">
              <span className="flex h-10 w-10 items-center justify-center rounded-lg bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 mb-4">
                <BarChart3 className="h-5 w-5" />
              </span>
              <h3 className="text-xl font-semibold mb-2 text-neutral-950 dark:text-white">
                {t("featureList.crm.title")}
              </h3>
              <p className="text-neutral-600 dark:text-neutral-400 leading-relaxed text-sm">
                {t("featureList.crm.desc")}
              </p>
            </div>

            {/* Finance */}
            <div className="rounded-xl border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-900 p-6 shadow-sm hover:shadow-md transition-shadow">
              <span className="flex h-10 w-10 items-center justify-center rounded-lg bg-purple-100 dark:bg-purple-900/30 text-purple-600 dark:text-purple-400 mb-4">
                <Landmark className="h-5 w-5" />
              </span>
              <h3 className="text-xl font-semibold mb-2 text-neutral-950 dark:text-white">
                {t("featureList.finance.title")}
              </h3>
              <p className="text-neutral-600 dark:text-neutral-400 leading-relaxed text-sm">
                {t("featureList.finance.desc")}
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-950 py-12">
        <div className="max-w-7xl mx-auto px-6 flex flex-col md:flex-row items-center justify-between gap-6">
          <div className="flex items-center gap-3">
            <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-tr from-blue-600 to-indigo-600 text-white font-bold text-sm shadow-md">
              IS
            </span>
            <span className="font-extrabold text-lg tracking-tight bg-clip-text text-transparent bg-gradient-to-r from-neutral-900 to-neutral-700 dark:from-white dark:to-neutral-300">
              IndoSupplier
            </span>
          </div>
          <p className="text-sm text-neutral-500 dark:text-neutral-400">
            © {new Date().getFullYear()} IndoSupplier Platform. All rights reserved.
          </p>
        </div>
      </footer>
    </div>
  );
}
