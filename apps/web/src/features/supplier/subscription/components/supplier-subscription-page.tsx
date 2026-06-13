"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import {
  ShieldCheck,
  History,
  TrendingUp,
  Package,
  Layers,
  Check,
  Loader2,
  Download,
  Gem,
  Award,
  Flame,
  CreditCard,
  Sliders,
} from "lucide-react";
import { useBillingOverview, useUpgradePlan } from "../hooks/useSubscription";

type TabType = "overview" | "plans" | "history";

export function SupplierSubscriptionPage() {
  const t = useTranslations("supplier.subscription");
  const { data: billing, isLoading } = useBillingOverview();
  const upgradeMutation = useUpgradePlan();
  const [activeTab, setActiveTab] = useState<TabType>("overview");

  if (isLoading || !billing) {
    return (
      <div className="flex flex-col items-center justify-center p-12 min-h-[400px]">
        <Loader2 className="h-8 w-8 animate-spin text-primary mb-4" />
        <span className="text-sm font-semibold text-muted-foreground">{t("loading")}</span>
      </div>
    );
  }

  // GIMS pricing plans catalog localized
  const plans = [
    {
      id: "free",
      name: t("plans.free.name"),
      price: t("plans.free.price"),
      period: t("plans.free.period"),
      icon: Award,
      features: [
        t("plans.free.features.0"),
        t("plans.free.features.1"),
        t("plans.free.features.2"),
        t("plans.free.features.3"),
      ],
    },
    {
      id: "bronze",
      name: t("plans.bronze.name"),
      price: t("plans.bronze.price"),
      period: t("plans.bronze.period"),
      icon: Flame,
      features: [
        t("plans.bronze.features.0"),
        t("plans.bronze.features.1"),
        t("plans.bronze.features.2"),
        t("plans.bronze.features.3"),
        t("plans.bronze.features.4"),
      ],
    },
    {
      id: "silver",
      name: t("plans.silver.name"),
      price: t("plans.silver.price"),
      period: t("plans.silver.period"),
      icon: Gem,
      features: [
        t("plans.silver.features.0"),
        t("plans.silver.features.1"),
        t("plans.silver.features.2"),
        t("plans.silver.features.3"),
        t("plans.silver.features.4"),
        t("plans.silver.features.5"),
      ],
    },
    {
      id: "gold",
      name: t("plans.gold.name"),
      price: t("plans.gold.price"),
      period: t("plans.gold.period"),
      icon: ShieldCheck,
      features: [
        t("plans.gold.features.0"),
        t("plans.gold.features.1"),
        t("plans.gold.features.2"),
        t("plans.gold.features.3"),
        t("plans.gold.features.4"),
        t("plans.gold.features.5"),
        t("plans.gold.features.6"),
      ],
    },
  ];

  const activePlanId = billing.subscriptions.find((s) => s.active)?.planId || "free";

  const handleUpgrade = (planId: string) => {
    if (planId === activePlanId) return;
    upgradeMutation.mutate(planId);
  };

  const tabs: { id: TabType; label: string; icon: React.ComponentType<{ className?: string }> }[] = [
    { id: "overview", label: t("tabOverview"), icon: CreditCard },
    { id: "plans", label: t("tabPlans"), icon: Sliders },
    { id: "history", label: t("tabHistory"), icon: History },
  ];

  const activePlanName = activePlanId === "gold"
    ? t("plans.gold.name")
    : activePlanId === "silver"
    ? t("plans.silver.name")
    : activePlanId === "bronze"
    ? t("plans.bronze.name")
    : t("plans.free.name");

  const activePlanRate = activePlanId === "gold"
    ? `${t("plans.gold.price")} ${t("plans.gold.period")}`
    : activePlanId === "silver"
    ? `${t("plans.silver.price")} ${t("plans.silver.period")}`
    : activePlanId === "bronze"
    ? `${t("plans.bronze.price")} ${t("plans.bronze.period")}`
    : `${t("plans.free.price")} ${t("plans.free.period")}`;

  const renderLimit = (used: number, limit: string | number) => {
    if (limit === "Unlimited") {
      return `${used} / ${t("unlimited")}`;
    }
    return `${used} / ${limit}`;
  };

  const getInvoiceDescription = (desc: string) => {
    if (desc.includes("Gold Enterprise - Annual plan")) {
      return t("invoiceDescGoldAnnual");
    }
    if (desc.includes("Silver Pro - Upgrade from Free")) {
      return t("invoiceDescSilverUpgrade");
    }
    if (desc.includes("Gold Enterprise - Subscription Upgrade")) {
      return t("invoiceDescGoldUpgrade");
    }
    if (desc.includes("Silver Pro - Subscription Upgrade")) {
      return t("invoiceDescSilverUpgradeSub");
    }
    if (desc.includes("Bronze Seller - Subscription Upgrade")) {
      return t("invoiceDescBronzeUpgrade");
    }
    return desc;
  };

  return (
    <div className="space-y-8 text-left max-w-5xl mx-auto py-4">
      {/* Title */}
      <div className="border-b border-border/60 pb-6">
        <h1 className="text-2xl font-extrabold tracking-tight text-foreground font-heading">
          {t("title")}
        </h1>
        <p className="text-sm text-muted-foreground mt-1 font-medium">
          {t("subtitle")}
        </p>
      </div>

      {/* Sidebar navigation + content layout split */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-8 items-start">
        {/* Left Side Tab Navigation */}
        <div className="md:col-span-1 flex flex-row md:flex-col gap-1 border-b md:border-b-0 md:border-r border-border/85 pb-4 md:pb-0 md:pr-4 overflow-x-auto select-none">
          {tabs.map((tabItem) => {
            const Icon = tabItem.icon;
            const isActive = activeTab === tabItem.id;
            return (
              <button
                key={tabItem.id}
                onClick={() => setActiveTab(tabItem.id)}
                className={`flex items-center gap-3 px-3 py-2.5 rounded-lg text-xs font-bold transition-all duration-200 shrink-0 text-left cursor-pointer ${
                  isActive
                    ? "bg-secondary text-primary font-extrabold border-l-2 border-primary md:translate-x-0.5"
                    : "text-muted-foreground hover:text-foreground hover:bg-muted/30"
                }`}
              >
                <Icon className={`h-4 w-4 ${isActive ? "text-primary" : "text-muted-foreground"}`} />
                <span>{tabItem.label}</span>
              </button>
            );
          })}
        </div>

        {/* Right Side Tab Content Area */}
        <div className="md:col-span-3 min-h-[400px] space-y-8 animate-fade-in pl-0 md:pl-4">
          
          {/* Tab 1: Overview */}
          {activeTab === "overview" && (
            <div className="space-y-10">
              {/* Active Plan Detail Section */}
              <div className="space-y-4">
                <div className="text-xs font-bold text-muted-foreground uppercase tracking-wider">
                  {t("activePlanLabel")}
                </div>
                <div className="space-y-2">
                  <div className="flex flex-wrap items-baseline gap-2.5">
                    <h2 className="text-3xl font-extrabold text-foreground tracking-tight leading-none">
                      {activePlanName}
                    </h2>
                    <Badge className="bg-success/15 text-success border border-success/30 font-extrabold text-[9px] uppercase px-1.5 py-0.5 rounded-sm">
                      {t("badgeActive")}
                    </Badge>
                  </div>
                  <p className="text-sm font-semibold text-muted-foreground">
                    {t("rateLabel")} <span className="text-foreground font-bold">{activePlanRate}</span>
                  </p>
                  <p className="text-xs text-muted-foreground font-medium">
                    {t("renewalLabel")} <span className="font-bold text-foreground">{billing.nextPaymentDate}</span>.
                  </p>
                </div>
              </div>

              <hr className="border-border/60" />

              {/* Metered usage indicators */}
              <div className="space-y-6">
                <div>
                  <h3 className="text-sm font-bold uppercase tracking-wider text-muted-foreground mb-1">
                    {t("usageTitle")}
                  </h3>
                  <p className="text-xs text-muted-foreground font-medium">
                    {t("usageSubtitle")}
                  </p>
                </div>

                <div className="space-y-5">
                  {/* Product listings limit */}
                  <div className="space-y-2 text-xs">
                    <div className="flex items-center justify-between font-bold text-foreground">
                      <span className="flex items-center gap-1.5 font-semibold text-muted-foreground">
                        <Package className="h-4 w-4 text-muted-foreground" />
                        {t("productUploadsLimitLabel")}
                      </span>
                      <span>
                        {renderLimit(billing.meteredUsage.productUploadsUsed, billing.meteredUsage.productUploadsLimit)}
                      </span>
                    </div>
                    <div className="w-full bg-muted h-1 rounded-full overflow-hidden">
                      <div
                        className="bg-primary h-1 rounded-full"
                        style={{
                          width: billing.meteredUsage.productUploadsLimit === "Unlimited" ? "100%" : "30%",
                        }}
                      />
                    </div>
                  </div>

                  {/* RFQ bids limit */}
                  <div className="space-y-2 text-xs">
                    <div className="flex items-center justify-between font-bold text-foreground">
                      <span className="flex items-center gap-1.5 font-semibold text-muted-foreground">
                        <TrendingUp className="h-4 w-4 text-muted-foreground" />
                        {t("rfqBidsLimitLabel")}
                      </span>
                      <span>
                        {renderLimit(billing.meteredUsage.rfqBidsUsed, billing.meteredUsage.rfqBidsLimit)}
                      </span>
                    </div>
                    <div className="w-full bg-muted h-1 rounded-full overflow-hidden">
                      <div
                        className="bg-primary h-1 rounded-full"
                        style={{
                          width: billing.meteredUsage.rfqBidsLimit === "Unlimited" ? "100%" : "45%",
                        }}
                      />
                    </div>
                  </div>

                  {/* Auction bids limit */}
                  <div className="space-y-2 text-xs">
                    <div className="flex items-center justify-between font-bold text-foreground">
                      <span className="flex items-center gap-1.5 font-semibold text-muted-foreground">
                        <Layers className="h-4 w-4 text-muted-foreground" />
                        {t("auctionSlotsLimitLabel")}
                      </span>
                      <span>
                        {renderLimit(billing.meteredUsage.auctionSlotsUsed, billing.meteredUsage.auctionSlotsLimit)}
                      </span>
                    </div>
                    <div className="w-full bg-muted h-1 rounded-full overflow-hidden">
                      <div
                        className="bg-primary h-1 rounded-full"
                        style={{
                          width: billing.meteredUsage.auctionSlotsLimit === "Unlimited" ? "100%" : "25%",
                        }}
                      />
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Tab 2: Pricing plans upgrade */}
          {activeTab === "plans" && (
            <div className="space-y-6">
              <div>
                <h3 className="text-base font-bold font-heading text-foreground">{t("plansTitle")}</h3>
                <p className="text-xs text-muted-foreground font-medium mt-1">
                  {t("plansSubtitle")}
                </p>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-6 items-stretch pt-2">
                {plans.map((p) => {
                  const Icon = p.icon;
                  const isCurrent = p.id === activePlanId;
                  return (
                    <div
                      key={p.id}
                      className={`border rounded-lg p-5 flex flex-col justify-between transition-all duration-300 ${
                        isCurrent
                          ? "border-primary ring-1 ring-primary bg-card"
                          : "border-border/80 bg-card hover:border-primary/40"
                      }`}
                    >
                      <div className="space-y-4">
                        <div className="flex items-center justify-between">
                          <div className="h-8 w-8 rounded-lg bg-primary/5 text-primary flex items-center justify-center shrink-0 border border-border/80">
                            <Icon className="h-4.5 w-4.5" />
                          </div>
                          {isCurrent && (
                            <Badge className="bg-primary text-primary-foreground font-bold text-[8px] uppercase px-1.5 rounded-sm">
                              {t("badgeCurrentPlan")}
                            </Badge>
                          )}
                        </div>
                        
                        <div className="space-y-1">
                          <h4 className="text-sm font-extrabold text-foreground font-heading">{p.name}</h4>
                          <div className="flex items-baseline gap-1">
                            <span className="text-lg font-extrabold text-foreground">{p.price}</span>
                            <span className="text-[10px] text-muted-foreground font-semibold">
                              {p.period}
                            </span>
                          </div>
                        </div>

                        <ul className="space-y-2 pt-2 border-t border-border/60">
                          {p.features.map((feat) => (
                            <li
                              key={feat}
                              className="flex items-start gap-2 text-xs text-muted-foreground font-medium"
                            >
                              <Check className="h-3.5 w-3.5 text-success shrink-0 mt-0.5" />
                              <span>{feat}</span>
                            </li>
                          ))}
                        </ul>
                      </div>

                      <div className="pt-5 mt-4">
                        <Button
                          onClick={() => handleUpgrade(p.id)}
                          disabled={isCurrent || upgradeMutation.isPending}
                          variant={isCurrent ? "outline" : "default"}
                          className="w-full text-xs font-bold py-3.5 h-8.5 cursor-pointer transition-all duration-200 hover:-translate-y-0.5 flex items-center justify-center gap-1.5"
                        >
                          {upgradeMutation.isPending && !isCurrent ? (
                            <>
                              <Loader2 className="h-3.5 w-3.5 animate-spin" /> {t("upgradingButton")}
                            </>
                          ) : isCurrent ? (
                            t("currentPlanButton")
                          ) : (
                            t("upgradeButton")
                          )}
                        </Button>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {/* Tab 3: Payment invoices history logs */}
          {activeTab === "history" && (
            <div className="space-y-6">
              <div>
                <h3 className="text-base font-bold font-heading text-foreground">{t("historyTitle")}</h3>
                <p className="text-xs text-muted-foreground font-medium mt-1">
                  {t("historySubtitle")}
                </p>
              </div>

              <div className="border border-border/80 rounded-lg overflow-hidden bg-card">
                <Table>
                  <TableHeader className="bg-muted/10">
                    <TableRow>
                      <TableHead className="font-bold text-xs">{t("invoiceTableDate")}</TableHead>
                      <TableHead className="font-bold text-xs">{t("invoiceTableDesc")}</TableHead>
                      <TableHead className="font-bold text-xs">{t("invoiceTableAmount")}</TableHead>
                      <TableHead className="font-bold text-xs">{t("invoiceTableStatus")}</TableHead>
                      <TableHead className="font-bold text-xs text-right">{t("invoiceTableDownload")}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {billing.invoices.map((inv) => (
                      <TableRow key={inv.id}>
                        <TableCell className="text-xs font-semibold">{inv.date}</TableCell>
                        <TableCell className="text-xs text-muted-foreground font-medium">
                          {getInvoiceDescription(inv.description)}
                        </TableCell>
                        <TableCell className="text-xs font-bold text-foreground">{inv.amount}</TableCell>
                        <TableCell>
                          <Badge className="bg-success/15 text-success border border-success/30 font-bold text-[8px] uppercase px-1.5 py-0.5 rounded-sm">
                            {inv.status === "paid" ? t("invoiceStatusPaid") : inv.status}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-right">
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-7 px-2 font-bold text-xs text-primary hover:text-primary cursor-pointer hover:bg-muted/50"
                          >
                            <Download className="h-3.5 w-3.5 mr-1" /> PDF
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
