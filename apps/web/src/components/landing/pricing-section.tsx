"use client";

import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useLocale, useTranslations } from "next-intl";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Check, XCircle, Minus, Plus, Phone } from "lucide-react";
import { authService, type SubscriptionPlanConfig } from "@/features/auth/services/auth-service";
import { useRouter } from "@/i18n/routing";
import { getSubscriptionPlanCopy } from "@/lib/subscription-plan-copy";
import { cn } from "@/lib/utils";

type BillingPeriod = "monthly" | "yearly";
type PlanTab = "bundle" | "modular";

interface BundlePlan {
  id: string;
  name: string;
  tagline: string;
  perUserMonthly: number;
  cta?: "select" | "contact";
}

interface ModularPlan {
  id: string;
  name: string;
  category: string;
  perUserMonthly: number;
  features: string[];
}

type ComparisonCellValue = string | "check" | "minus";

interface BundleComparisonRow {
  label: string;
  growthSuite: ComparisonCellValue;
  ultimateSuite: ComparisonCellValue;
  enterprise: ComparisonCellValue;
}

const CONTACT_SALES_WHATSAPP_URL = "https://wa.me/6289607700028?text=Hi%20SalesView%2C%20I%20want%20to%20consult%20about%20the%20Enterprise%20plan.";

const BUNDLE_COMPARISON_ROWS: BundleComparisonRow[] = [
  { label: "POS", growthSuite: "check", ultimateSuite: "check", enterprise: "check" },
  { label: "ERP", growthSuite: "check", ultimateSuite: "check", enterprise: "check" },
  { label: "FINANCE", growthSuite: "check", ultimateSuite: "check", enterprise: "check" },
  { label: "CRM", growthSuite: "check", ultimateSuite: "check", enterprise: "check" },
  { label: "HR", growthSuite: "minus", ultimateSuite: "check", enterprise: "check" },
  { label: "Advance REPORT", growthSuite: "minus", ultimateSuite: "check", enterprise: "check" },
  { label: "AI", growthSuite: "minus", ultimateSuite: "check", enterprise: "check" },
  { label: "TRAVEL PLANNER", growthSuite: "minus", ultimateSuite: "check", enterprise: "check" },
  { label: "usersLabel", growthSuite: "Per seat", ultimateSuite: "Per seat", enterprise: "Unlimited users" },
  { label: "storage", growthSuite: "100 GB", ultimateSuite: "500 GB", enterprise: "Unlimited storage" },
  { label: "support", growthSuite: "Business hours", ultimateSuite: "Priority support", enterprise: "24/7 dedicated support" },
  { label: "integrations", growthSuite: "Standard integrations", ultimateSuite: "Advanced integrations", enterprise: "Custom integrations" },
];

const FALLBACK_PLANS: SubscriptionPlanConfig[] = [
  {
    id: "pos_growth",
    slug: "pos_growth",
    name: "POS Modular",
    category: "pos",
    description: "",
    billing_type: "per_user",
    price_monthly_idr: 79_000,
    price_yearly_idr: 853_200,
    min_users: 1,
    max_users: 500,
    is_highlighted: false,
    sort_order: 10,
    features: [],
    module_slugs: [],
  },
  {
    id: "erp_pro",
    slug: "erp_pro",
    name: "ERP Modular",
    category: "erp",
    description: "",
    billing_type: "per_user",
    price_monthly_idr: 109_000,
    price_yearly_idr: 1_177_200,
    min_users: 1,
    max_users: 500,
    is_highlighted: false,
    sort_order: 20,
    features: [],
    module_slugs: [],
  },
  {
    id: "crm_growth",
    slug: "crm_growth",
    name: "CRM Modular",
    category: "crm",
    description: "",
    billing_type: "per_user",
    price_monthly_idr: 89_000,
    price_yearly_idr: 961_200,
    min_users: 1,
    max_users: 500,
    is_highlighted: false,
    sort_order: 30,
    features: [],
    module_slugs: [],
  },
  {
    id: "hr_growth",
    slug: "hr_growth",
    name: "HR Modular",
    category: "hr",
    description: "",
    billing_type: "per_user",
    price_monthly_idr: 89_000,
    price_yearly_idr: 961_200,
    min_users: 1,
    max_users: 500,
    is_highlighted: false,
    sort_order: 40,
    features: [],
    module_slugs: [],
  },
  {
    id: "growth_suite",
    slug: "growth_suite",
    name: "Growth Suite",
    category: "bundle",
    description: "",
    billing_type: "per_user",
    price_monthly_idr: 125_000,
    price_yearly_idr: 1_350_000,
    min_users: 1,
    max_users: 500,
    is_highlighted: false,
    sort_order: 1,
    features: [],
    module_slugs: ["pos_growth", "erp_pro", "crm_growth"],
  },
  {
    id: "ultimate_suite",
    slug: "ultimate_suite",
    name: "Ultimate Suite",
    category: "bundle",
    description: "",
    billing_type: "per_user",
    price_monthly_idr: 175_000,
    price_yearly_idr: 1_890_000,
    min_users: 1,
    max_users: 500,
    is_highlighted: true,
    sort_order: 2,
    features: [],
    module_slugs: ["pos_growth", "erp_pro", "crm_growth", "hr_growth"],
  },
];

function formatIDR(amount: number): string {
  if (amount === 0) return "Custom";
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
  }).format(amount);
}

function effectiveMonthlyPrice(perUser: number, users: number, billing: BillingPeriod): number {
  const base = perUser * users;
  return billing === "yearly" ? Math.round(base * 0.9) : base;
}

function totalInvoice(perUser: number, users: number, billing: BillingPeriod): number {
  const monthly = effectiveMonthlyPrice(perUser, users, billing);
  return billing === "yearly" ? monthly * 12 : monthly;
}

function buildEnterpriseCTA(locale: string): BundlePlan {
  const copy = getSubscriptionPlanCopy("enterprise", locale, {
    name: "Enterprise",
    description: "Custom for your business",
    features: [
      "Everything in Ultimate",
      "Unlimited users",
      "Dedicated SLA",
      "Custom integrations",
      "On-premise option",
    ],
  });

  return {
    id: "enterprise",
    name: copy.name,
    tagline: copy.description,
    perUserMonthly: 0,
    cta: "contact",
  };
}

function getCoreDiscountBadge(planId: string): { original: number; final: number; percent: number } | null {
  const pair = planId === "growth_suite"
    ? { original: 277_000, final: 125_000 }
    : planId === "ultimate_suite"
      ? { original: 366_000, final: 175_000 }
      : null;

  if (!pair) return null;
  return {
    ...pair,
    percent: Math.round(((pair.original - pair.final) / pair.original) * 100),
  };
}

export function PricingSection() {
  const t = useTranslations("landing");
  const tRegister = useTranslations("auth.register");
  const locale = useLocale();
  const router = useRouter();

  const [pricingTab, setPricingTab] = useState<PlanTab>("bundle");
  const [pricingBilling, setPricingBilling] = useState<BillingPeriod>("monthly");
  const [pricingUsers, setPricingUsers] = useState(1);
  const [selectedBundle] = useState<string>("ultimate_suite");
  const [selectedModular] = useState<string>("pos_growth");

  const { data: apiPlans } = useQuery({
    queryKey: ["landing-subscription-plans"],
    queryFn: () => authService.getSubscriptionPlans(),
    staleTime: 5 * 60 * 1000,
    retry: false,
  });

  const plans = apiPlans && apiPlans.length > 0 ? apiPlans : FALLBACK_PLANS;

  const dynamicBundles = useMemo<BundlePlan[]>(() => {
    const bundlePlans = plans
      .filter((p) => p.category === "bundle")
      .sort((a, b) => a.sort_order - b.sort_order)
      .map<BundlePlan>((p) => {
        const planCopy = getSubscriptionPlanCopy(p.slug, locale, {
          name: p.name,
          description: p.description,
          features: p.features ?? p.module_slugs ?? [],
        });

        return {
          id: p.slug,
          name: planCopy.name,
          tagline: planCopy.description,
          perUserMonthly: p.price_monthly_idr,
          cta: "select",
        };
      });

    return [...bundlePlans, buildEnterpriseCTA(locale)];
  }, [locale, plans]);

  const dynamicModular = useMemo<ModularPlan[]>(() => {
    return plans
      .filter((p) => p.category !== "bundle")
      .sort((a, b) => a.sort_order - b.sort_order)
      .map<ModularPlan>((p) => {
        const planCopy = getSubscriptionPlanCopy(p.slug, locale, {
          name: p.name,
          description: p.description,
          features: p.features ?? [],
        });

        return {
          id: p.slug,
          name: planCopy.name,
          category: p.category.toUpperCase(),
          perUserMonthly: p.price_monthly_idr,
          features: planCopy.features,
        };
      });
  }, [locale, plans]);

  const openContactSalesWhatsApp = () => {
    if (typeof window !== "undefined") {
      window.open(CONTACT_SALES_WHATSAPP_URL, "_blank", "noopener,noreferrer");
    }
  };

  const renderComparisonCell = (value: ComparisonCellValue) => {
    if (value === "check") {
      return <Check className="mx-auto h-4 w-4 text-success" aria-label="Included" />;
    }
    if (value === "minus") {
      return <XCircle className="mx-auto h-4 w-4 text-muted-foreground/30" aria-label="Not included" />;
    }
    return (
      <p className="w-full whitespace-normal text-center text-[10px] leading-normal text-muted-foreground sm:text-xs">
        {value}
      </p>
    );
  };

  return (
    <section id="pricing" className="flex min-h-screen flex-col items-center justify-center py-24">
      <div className="relative z-10 mx-auto w-full max-w-6xl px-6 lg:px-12">
        <div className="mx-auto mb-10 max-w-3xl text-center">
          <p className="mb-5 text-xs font-semibold uppercase tracking-[0.22em] text-muted-foreground">
            {t("pricing.eyebrow")}
          </p>
          <h2 className="text-4xl font-light tracking-tight text-foreground sm:text-5xl">
            {t("pricing.title1")}
            <br />
            {t("pricing.title2")}
          </h2>
          <p className="mt-5 text-base leading-relaxed text-muted-foreground">{t("pricing.subtitle")}</p>
        </div>

        <div className="space-y-5">
          <div className="flex flex-col gap-3 rounded-2xl border border-border/70 bg-card/50 p-4 backdrop-blur-sm sm:flex-row sm:items-center sm:justify-between">
            <div className="inline-flex rounded-lg border border-border/70 p-1 text-xs">
              <button
                type="button"
                onClick={() => setPricingTab("bundle")}
                className={cn(
                  "rounded-lg px-3 py-1.5 font-medium transition-colors",
                  pricingTab === "bundle" ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:text-foreground",
                )}
              >
                {t("pricing.tabs.bundle")}
              </button>
              <button
                type="button"
                onClick={() => setPricingTab("modular")}
                className={cn(
                  "rounded-lg px-3 py-1.5 font-medium transition-colors",
                  pricingTab === "modular" ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:text-foreground",
                )}
              >
                {t("pricing.tabs.modular")}
              </button>
            </div>

            <div className="flex items-center gap-2">
              <div className="inline-flex rounded-lg border border-border/70 p-1 text-xs">
                <button
                  type="button"
                  onClick={() => setPricingBilling("monthly")}
                  className={cn(
                    "rounded-lg px-3 py-1.5 font-medium transition-colors",
                    pricingBilling === "monthly" ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:text-foreground",
                  )}
                >
                  {t("pricing.monthly")}
                </button>
                <button
                  type="button"
                  onClick={() => setPricingBilling("yearly")}
                  className={cn(
                    "rounded-lg px-3 py-1.5 font-medium transition-colors",
                    pricingBilling === "yearly" ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:text-foreground",
                  )}
                >
                  {t("pricing.yearly")}
                </button>
              </div>

              <div className="flex items-center rounded-lg border border-border/70">
                <button
                  type="button"
                  onClick={() => setPricingUsers((count) => Math.max(1, count - 1))}
                  className="flex h-9 w-9 items-center justify-center rounded-l-lg text-muted-foreground transition-colors hover:bg-muted"
                  aria-label={tRegister("decreaseUsers")}
                >
                  <Minus className="h-3 w-3" />
                </button>
                <span className="min-w-16 border-x px-3 text-center text-sm font-medium">{pricingUsers}</span>
                <button
                  type="button"
                  onClick={() => setPricingUsers((count) => Math.min(500, count + 1))}
                  className="flex h-9 w-9 items-center justify-center rounded-r-lg text-muted-foreground transition-colors hover:bg-muted"
                  aria-label={tRegister("increaseUsers")}
                >
                  <Plus className="h-3 w-3" />
                </button>
              </div>
              <span className="text-xs font-medium text-muted-foreground">{t("pricing.users")}</span>
            </div>
          </div>

          {pricingTab === "bundle" ? (
            <div className="overflow-x-auto rounded-2xl border border-border/60 bg-card/50 backdrop-blur-sm">
              <table className="w-full min-w-[760px] border-collapse">
                <thead>
                  <tr className="border-b border-border/60 bg-muted/30">
                    <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-foreground sm:px-6 sm:py-4 sm:text-sm">
                      {tRegister("features")}
                    </th>
                    {dynamicBundles.map((bundle) => {
                      const isSelected = selectedBundle === bundle.id;
                      return (
                        <th
                          key={bundle.id}
                          className={cn("px-4 py-3 text-center sm:px-6 sm:py-4", isSelected && bundle.cta !== "contact" && "bg-primary/5")}
                        >
                          <div className="flex flex-col items-center gap-1">
                            <p className="text-sm font-bold leading-tight text-foreground sm:text-base">{bundle.name.replace(" Suite", "")}</p>
                            <p className="max-w-[140px] text-[10px] font-normal leading-tight text-muted-foreground sm:text-xs">{bundle.tagline}</p>
                          </div>
                        </th>
                      );
                    })}
                  </tr>
                </thead>
                <tbody>
                  <tr className="border-b border-border/60">
                    <th scope="row" className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-tight text-foreground sm:px-6 sm:py-4 sm:text-sm">
                      {tRegister("price")}
                    </th>
                    {dynamicBundles.map((bundle) => {
                      const isSelected = selectedBundle === bundle.id;
                      if (bundle.cta === "contact") {
                        return (
                          <td key={bundle.id} className={cn("px-4 py-3 text-center align-top sm:px-6 sm:py-4", isSelected && "bg-primary/5")}>
                            <p className="text-sm font-bold text-foreground sm:text-lg">{tRegister("custom")}</p>
                            <p className="mt-0.5 text-[10px] leading-tight text-muted-foreground">{t("pricing.contactSalesHint")}</p>
                          </td>
                        );
                      }

                      const total = totalInvoice(bundle.perUserMonthly, pricingUsers, pricingBilling);
                      const monthly = pricingBilling === "yearly" ? Math.round(total / 12) : total;
                      const strikingBadge = getCoreDiscountBadge(bundle.id);

                      return (
                        <td key={bundle.id} className={cn("px-4 py-3 text-center align-top sm:px-6 sm:py-4", isSelected && "bg-primary/5")}>
                          <div className="flex flex-col items-center gap-2">
                            {strikingBadge && (
                              <Badge className="relative overflow-hidden rounded-full border-0 bg-linear-to-r from-destructive via-chart-3 to-chart-5 px-3 py-1 text-[10px] font-bold tracking-wide text-primary-foreground shadow-md">
                                <span className="absolute inset-0 bg-[radial-gradient(circle_at_top_right,rgba(255,255,255,0.38),transparent_45%)]" />
                                <span className="relative flex items-center gap-1 whitespace-nowrap">
                                  <span className="line-through decoration-primary-foreground/80 decoration-2">{formatIDR(strikingBadge.original)}</span>
                                  <span>-{strikingBadge.percent}%</span>
                                </span>
                              </Badge>
                            )}
                            <p className="text-sm font-bold tracking-tight text-foreground sm:text-lg">
                              {formatIDR(monthly)}
                              <span className="ml-1 text-[10px] font-normal text-muted-foreground sm:text-xs">{t("pricing.perMonthSeat")}</span>
                            </p>
                            <p className="text-[10px] font-medium uppercase tracking-tight text-muted-foreground">
                              {t("pricing.totalInvoice")}: {formatIDR(total)}
                            </p>
                          </div>
                        </td>
                      );
                    })}
                  </tr>

                  {BUNDLE_COMPARISON_ROWS.map((row) => (
                    <tr key={row.label} className="group border-b border-border/50 transition-colors hover:bg-muted/10">
                      <th
                        scope="row"
                        className="px-4 py-2.5 text-left text-[11px] font-medium text-muted-foreground/80 transition-colors group-hover:bg-muted/5 group-hover:text-foreground sm:px-6 sm:py-3.5 sm:text-sm"
                      >
                        {tRegister(row.label)}
                      </th>
                      <td className={cn("px-4 py-2.5 text-center sm:px-6 sm:py-3.5", selectedBundle === "growth_suite" && "bg-primary/5")}>{renderComparisonCell(row.growthSuite)}</td>
                      <td className={cn("px-4 py-2.5 text-center sm:px-6 sm:py-3.5", selectedBundle === "ultimate_suite" && "bg-primary/5")}>{renderComparisonCell(row.ultimateSuite)}</td>
                      <td className="px-4 py-2.5 text-center sm:px-6 sm:py-3.5">{renderComparisonCell(row.enterprise)}</td>
                    </tr>
                  ))}

                  <tr className="bg-muted/10">
                    <td className="px-4 py-4 sm:px-6 sm:py-6" />
                    {dynamicBundles.map((bundle) => {
                      const isSelected = selectedBundle === bundle.id;
                      return (
                        <td key={bundle.id} className={cn("px-4 py-4 sm:px-6 sm:py-6", isSelected && bundle.cta !== "contact" && "bg-primary/5")}>
                          {bundle.cta === "contact" ? (
                            <Button type="button" variant="outline" onClick={openContactSalesWhatsApp} className="h-9 w-full text-xs font-semibold sm:h-10 sm:text-sm">
                              <Phone className="mr-1 h-3 w-3 sm:h-4 sm:w-4" />
                              {tRegister("contactSales")}
                            </Button>
                          ) : (
                            <Button
                              type="button"
                              variant={isSelected ? "default" : "outline"}
                              onClick={() => router.push("/register")}
                              className={cn(
                                "h-9 w-full text-xs font-bold transition-all duration-300 sm:h-10 sm:text-sm",
                                isSelected ? "scale-[1.02] shadow-md ring-2 ring-primary/20" : "hover:scale-[1.01]",
                              )}
                            >
                              {isSelected ? tRegister("selected") : t("pricing.selectBundle", { name: bundle.name.replace(" Suite", "") })}
                            </Button>
                          )}
                        </td>
                      );
                    })}
                  </tr>
                </tbody>
              </table>
            </div>
          ) : (
            <div className="space-y-4">
              <p className="text-sm text-muted-foreground">{t("pricing.modularSubtitle")}</p>
              <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
                {dynamicModular.map((plan) => {
                  const isSelected = selectedModular === plan.id;
                  const monthlyTotal = effectiveMonthlyPrice(plan.perUserMonthly, pricingUsers, pricingBilling);
                  const baseMonthly = plan.perUserMonthly * pricingUsers;
                  const total = totalInvoice(plan.perUserMonthly, pricingUsers, pricingBilling);
                  return (
                    <div
                      key={plan.id}
                      className={cn(
                        "relative flex min-h-64 flex-col rounded-xl border p-4 transition-all",
                        isSelected ? "border-primary bg-primary/5 shadow-sm" : "border-border/80 bg-card/60 hover:border-primary/40",
                      )}
                    >
                      {isSelected && (
                        <span className="absolute right-3 top-3 flex h-5 w-5 items-center justify-center rounded-lg bg-primary text-primary-foreground">
                          <Check className="h-2.5 w-2.5" />
                        </span>
                      )}
                      <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">{plan.category}</p>
                      <p className="mt-1 text-sm font-semibold text-foreground">{plan.name}</p>
                      <div className="mt-3">
                        <p className="text-lg font-semibold text-foreground">
                          {pricingBilling === "yearly" ? (
                            <span className="flex flex-col gap-0.5">
                              <span className="text-[10px] text-muted-foreground/70 line-through decoration-destructive/60 decoration-2 sm:text-xs">
                                {formatIDR(baseMonthly)}
                              </span>
                              <span>{formatIDR(monthlyTotal)}</span>
                            </span>
                          ) : (
                            formatIDR(monthlyTotal)
                          )}
                          <span className="ml-1 text-xs font-normal text-muted-foreground">{tRegister("perMonthShort")}</span>
                        </p>
                      </div>
                      <ul className="mt-3 space-y-1">
                        {plan.features.slice(0, 3).map((feature) => (
                          <li key={feature} className="flex items-center gap-1.5 text-xs text-muted-foreground">
                            <Check className="h-3 w-3 shrink-0 text-primary" />
                            {feature}
                          </li>
                        ))}
                      </ul>
                      <p className="mt-auto pt-3 text-xs text-muted-foreground">
                        {t("pricing.totalInvoice")}: {formatIDR(total)}
                      </p>
                      <Button
                        type="button"
                        variant={isSelected ? "default" : "outline"}
                        onClick={() => router.push("/register")}
                        className="mt-3 h-9 text-xs font-semibold"
                      >
                        {isSelected ? tRegister("selected") : t("pricing.selectModular")}
                      </Button>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          <p className="text-center text-xs text-muted-foreground">{t("pricing.guarantee")}</p>
        </div>
      </div>
    </section>
  );
}
