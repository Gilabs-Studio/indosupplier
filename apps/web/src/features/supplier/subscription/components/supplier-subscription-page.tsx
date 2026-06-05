"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { toast } from "sonner";
import { Check, ShieldCheck, Award, Flame, Gem } from "lucide-react";

export function SupplierSubscriptionPage() {
  const t = useTranslations("supplier.subscription");
  const [activePlan, setActivePlan] = useState("gold");

  const plans = [
    {
      id: "free",
      name: "Free Basic",
      price: "Rp 0",
      period: t("periodMonth"),
      icon: Award,
      features: [
        "Basic Directory Listing",
        "Upload up to 3 products",
        "Submit 1 RFQ bid / month",
        "Standard support responses"
      ],
      tone: "default"
    },
    {
      id: "bronze",
      name: "Bronze Seller",
      price: "Rp 2.000.000",
      period: t("periodYear"),
      icon: Flame,
      features: [
        "Verified Bronze badge",
        "Upload up to 15 products",
        "Submit 5 RFQ bids / month",
        "Standard search placement",
        "Email support response SLA"
      ],
      tone: "info"
    },
    {
      id: "silver",
      name: "Silver Pro",
      price: "Rp 5.000.000",
      period: t("periodYear"),
      icon: Gem,
      features: [
        "Verified Silver badge",
        "Upload up to 50 products",
        "Submit 15 RFQ bids / month",
        "Medium search placement boost",
        "1 active auction bidding slot",
        "Chat support response SLA"
      ],
      tone: "premium"
    },
    {
      id: "gold",
      name: "Gold Enterprise",
      price: "Rp 12.000.000",
      period: t("periodYear"),
      icon: ShieldCheck,
      features: [
        "Verified Gold Badge",
        "Unlimited products upload",
        "Unlimited RFQ bids submissions",
        "Priority top-tier search placement",
        "Unlimited auction bidding slots",
        "24/7 Priority support hotline",
        "Analytical marketing reports"
      ],
      tone: "success"
    }
  ];

  const handleUpgrade = (planId: string) => {
    if (planId === activePlan) {
      toast.info("This is your current active subscription plan.");
      return;
    }
    toast.promise(
      new Promise((resolve) => setTimeout(resolve, 1000)),
      {
        loading: "Redirecting to payment gateway...",
        success: () => {
          setActivePlan(planId);
          return "Subscription upgraded successfully! Welcome to your new tier.";
        },
        error: "Upgrade failed",
      }
    );
  };

  const getCardStyle = (planId: string) => {
    return planId === activePlan
      ? "border-primary ring-2 ring-primary bg-card"
      : "border-border hover:border-primary/50 bg-card";
  };

  return (
    <div className="space-y-6 text-left">
      {/* Header */}
      <div className="border-b border-border/80 pb-6">
        <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
          {t("title")}
        </h1>
        <p className="text-sm text-muted-foreground">
          {t("subtitle")}
        </p>
      </div>

      {/* Grid Comparison */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 items-stretch">
        {plans.map((p) => {
          const Icon = p.icon;
          const isCurrent = p.id === activePlan;
          return (
            <Card key={p.id} className={`border rounded-2xl shadow-xs overflow-hidden flex flex-col justify-between transition-all duration-300 ${getCardStyle(p.id)}`}>
              <div>
                <CardHeader className="border-b border-border bg-muted/10 py-5">
                  <div className="flex items-center justify-between gap-2">
                    <div className="h-9 w-9 bg-primary/10 text-primary border border-border rounded-lg flex items-center justify-center shrink-0">
                      <Icon className="h-5 w-5" />
                    </div>
                    {isCurrent && (
                      <Badge className="bg-primary text-primary-foreground font-bold text-[9px] uppercase">
                        {t("current")}
                      </Badge>
                    )}
                  </div>
                  <CardTitle className="text-base font-bold font-heading mt-3">{p.name}</CardTitle>
                  <div className="flex items-baseline gap-1 mt-2">
                    <span className="text-xl font-extrabold text-foreground">{p.price}</span>
                    <span className="text-[10px] text-muted-foreground font-semibold">{p.period}</span>
                  </div>
                </CardHeader>

                <CardContent className="p-5">
                  <ul className="space-y-2.5">
                    {p.features.map((feat) => (
                      <li key={feat} className="flex items-start gap-2 text-xs text-muted-foreground font-semibold">
                        <Check className="h-4 w-4 text-success shrink-0 mt-0.5" />
                        <span>{feat}</span>
                      </li>
                    ))}
                  </ul>
                </CardContent>
              </div>

              <div className="p-5 border-t border-border bg-muted/5">
                <Button
                  onClick={() => handleUpgrade(p.id)}
                  variant={isCurrent ? "outline" : "default"}
                  className={`w-full text-xs font-semibold py-4 h-9 cursor-pointer transition-all duration-200 hover:-translate-y-0.5`}
                >
                  {isCurrent ? "Active Package" : t("upgrade")}
                </Button>
              </div>
            </Card>
          );
        })}
      </div>
    </div>
  );
}
