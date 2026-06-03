"use client";

import React from "react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";
import { PublicLayout } from "@/features/public/components/public-layout";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Factory,
  Layers,
  Building,
  ShieldCheck,
  Zap,
  CheckCircle,
  TrendingUp,
} from "lucide-react";

interface DemoHomePageProps {
  locale: string;
}

export function DemoHomePage({ locale }: DemoHomePageProps) {
  const t = useTranslations("public.demoHome");
  const tCat = useTranslations("public.categories");

  const categories = [
    { id: "manufacturing", name: tCat("manufacturing"), icon: Factory, desc: "Industrial equipment, heavy machinery, and tooling." },
    { id: "agriculture", name: tCat("agriculture"), icon: Layers, desc: "Organic crops, processed foodstuffs, spices, and grains." },
    { id: "textile", name: tCat("textile"), icon: Layers, desc: "Garments, traditional fabrics (Batik), yarns, and industrial textiles." },
    { id: "chemical", name: tCat("chemical"), icon: Layers, desc: "Plastics, polymers, rubber products, and specialty chemicals." },
    { id: "furniture", name: tCat("furniture"), icon: Building, desc: "Jepara teak furniture, rattan crafts, and home decors." },
    { id: "construction", name: tCat("construction"), icon: Building, desc: "Building materials, steel structures, cement, and piping." },
  ];

  return (
    <PublicLayout locale={locale}>
      {/* Hero Banner */}
      <section className="bg-neutral-950 text-white py-24 relative overflow-hidden">
        <div className="absolute inset-0 opacity-20 bg-[radial-gradient(circle_at_30%_30%,#3b82f6,transparent_50%)]" />
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 relative z-10">
          <div className="max-w-3xl space-y-6">
            <Badge className="bg-blue-500/10 text-blue-400 border border-blue-500/20 px-3 py-1 text-xs font-semibold rounded-full uppercase tracking-wider">
              {t("badge")}
            </Badge>
            <h1 className="text-4xl font-extrabold tracking-tight sm:text-5xl lg:text-6xl font-heading leading-tight">
              {t("title")}
            </h1>
            <p className="text-neutral-400 text-base sm:text-lg max-w-xl font-light">
              {t("subtitle")}
            </p>

            <div className="flex flex-wrap gap-4 pt-4">
              <Button asChild size="lg" className="bg-white text-neutral-900 hover:bg-neutral-100 font-semibold cursor-pointer">
                <Link href="/demo/search">{t("btnSearch")}</Link>
              </Button>
              <Button asChild size="lg" variant="outline" className="border-white/20 text-white hover:bg-white/10 font-semibold cursor-pointer">
                <Link href="/demo/register">{t("btnRegister")}</Link>
              </Button>
            </div>
          </div>
        </div>
      </section>

      {/* Key Value Proposition */}
      <section className="bg-background py-16 border-b border-border">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            <div className="flex gap-4 items-start p-4">
              <div className="p-3 bg-emerald-500/10 text-emerald-600 rounded-xl shrink-0">
                <ShieldCheck className="h-6 w-6" />
              </div>
              <div className="space-y-1">
                <h3 className="font-semibold text-foreground text-sm">{t("verifiedTitle")}</h3>
                <p className="text-xs text-muted-foreground leading-relaxed">
                  {t("verifiedDesc")}
                </p>
              </div>
            </div>

            <div className="flex gap-4 items-start p-4">
              <div className="p-3 bg-blue-500/10 text-blue-600 rounded-xl shrink-0">
                <Zap className="h-6 w-6" />
              </div>
              <div className="space-y-1">
                <h3 className="font-semibold text-foreground text-sm">{t("rfqTitle")}</h3>
                <p className="text-xs text-muted-foreground leading-relaxed">
                  {t("rfqDesc")}
                </p>
              </div>
            </div>

            <div className="flex gap-4 items-start p-4">
              <div className="p-3 bg-purple-500/10 text-purple-600 rounded-xl shrink-0">
                <TrendingUp className="h-6 w-6" />
              </div>
              <div className="space-y-1">
                <h3 className="font-semibold text-foreground text-sm">{t("directTitle")}</h3>
                <p className="text-xs text-muted-foreground leading-relaxed">
                  {t("directDesc")}
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Explore Industries (OTO inspired category view) */}
      <section className="bg-muted py-20">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 space-y-10">
          <div className="text-center max-w-2xl mx-auto space-y-3">
            <h2 className="text-2xl font-bold tracking-tight text-foreground font-heading sm:text-3xl">
              {tCat("title")}
            </h2>
            <p className="text-sm text-muted-foreground leading-relaxed">
              {tCat("subtitle")}
            </p>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {categories.map((cat) => {
              const IconComponent = cat.icon;
              return (
                <Card
                  key={cat.id}
                  className="bg-card border border-border shadow-xs hover:shadow-md hover:-translate-y-0.5 transition-all duration-300 rounded-xl overflow-hidden cursor-pointer"
                >
                  <CardContent className="p-6 space-y-4">
                    <div className="flex items-center gap-3">
                      <div className="p-3 bg-primary text-primary-foreground rounded-lg">
                        <IconComponent className="h-5 w-5" />
                      </div>
                      <h3 className="font-semibold text-foreground text-sm">{cat.name}</h3>
                    </div>
                    <p className="text-xs text-muted-foreground leading-relaxed h-12 line-clamp-3">
                      {cat.desc}
                    </p>
                    <Button asChild variant="link" className="p-0 text-xs font-semibold text-foreground hover:text-primary cursor-pointer">
                      <Link href={`/demo/search?category=${cat.id}`} className="flex items-center gap-1">
                        {tCat("viewSuppliers")} →
                      </Link>
                    </Button>
                  </CardContent>
                </Card>
              );
            })}
          </div>
        </div>
      </section>

      {/* Trust & Safety Banner */}
      <section className="bg-background py-16 border-t border-border">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 text-center space-y-8">
          <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
            {t("trustTitle")}
          </h2>
          <div className="flex flex-wrap justify-center gap-8 text-foreground font-medium text-sm">
            <span className="flex items-center gap-2">
              <CheckCircle className="h-4 w-4 text-emerald-500" /> {t("nibVerified")}
            </span>
            <span className="flex items-center gap-2">
              <CheckCircle className="h-4 w-4 text-emerald-500" /> {t("factoryInspection")}
            </span>
            <span className="flex items-center gap-2">
              <CheckCircle className="h-4 w-4 text-emerald-500" /> {t("exportCompliant")}
            </span>
            <span className="flex items-center gap-2">
              <CheckCircle className="h-4 w-4 text-emerald-500" /> {t("secureChat")}
            </span>
          </div>
        </div>
      </section>
    </PublicLayout>
  );
}
