import { apiClient } from "@/lib/api-client";
import type { BillingOverview } from "../types/subscription.types";

const MOCK_BILLING_OVERVIEW: BillingOverview = {
  currentMeteredUsage: "Rp 0",
  currentIncludedUsage: "Gold Tier Limits Included",
  nextPaymentDue: "Rp 12.000.000",
  nextPaymentDate: "June 30, 2027",
  subscriptions: [
    {
      planId: "gold",
      planName: "GIMS Gold Enterprise",
      price: "Rp 12.000.000",
      period: "year",
      active: true,
      renewalDate: "June 30, 2027",
    },
    {
      planId: "analytics",
      planName: "Supplier Analytics Add-on",
      price: "Rp 0",
      period: "month",
      active: true,
      renewalDate: "June 30, 2027",
    },
  ],
  meteredUsage: {
    productUploadsUsed: 142,
    productUploadsLimit: "Unlimited",
    rfqBidsUsed: 85,
    rfqBidsLimit: "Unlimited",
    auctionSlotsUsed: 1,
    auctionSlotsLimit: "Unlimited",
  },
  invoices: [
    {
      id: "INV-2026-001",
      date: "June 30, 2026",
      description: "GIMS Gold Enterprise - Annual plan",
      amount: "Rp 12.000.000",
      status: "paid",
      receiptUrl: "#",
    },
    {
      id: "INV-2025-001",
      date: "June 30, 2025",
      description: "GIMS Gold Enterprise - Annual plan",
      amount: "Rp 12.000.000",
      status: "paid",
      receiptUrl: "#",
    },
    {
      id: "INV-2024-001",
      date: "June 30, 2024",
      description: "GIMS Silver Pro - Upgrade from Free",
      amount: "Rp 5.000.000",
      status: "paid",
      receiptUrl: "#",
    },
  ],
};

export const subscriptionService = {
  async getBillingOverview(): Promise<BillingOverview> {
    try {
      const response = await apiClient.get<{ data: BillingOverview }>("/supplier/billing-overview");
      if (response.data?.data) {
        return response.data.data;
      }
    } catch {
      // Fallback
    }

    if (typeof window !== "undefined") {
      const stored = localStorage.getItem("supplier_billing_overview");
      if (stored) {
        try {
          return JSON.parse(stored);
        } catch {
          // ignore
        }
      }
      localStorage.setItem("supplier_billing_overview", JSON.stringify(MOCK_BILLING_OVERVIEW));
    }
    return MOCK_BILLING_OVERVIEW;
  },

  async upgradePlan(planId: string): Promise<BillingOverview> {
    const current = await this.getBillingOverview();

    // Map plan detail
    const planName = planId === "gold" ? "GIMS Gold Enterprise" : planId === "silver" ? "GIMS Silver Pro" : "GIMS Bronze Seller";
    const price = planId === "gold" ? "Rp 12.000.000" : planId === "silver" ? "Rp 5.000.000" : "Rp 2.000.000";

    const nextRenewal = "June 30, 2027";

    const newInvoice = {
      id: `INV-2026-00${current.invoices.length + 1}`,
      date: new Date().toLocaleDateString("en-US", { year: "numeric", month: "long", day: "numeric" }),
      description: `${planName} - Subscription Upgrade`,
      amount: price,
      status: "paid" as const,
      receiptUrl: "#",
    };

    const updated: BillingOverview = {
      ...current,
      nextPaymentDue: price,
      nextPaymentDate: nextRenewal,
      subscriptions: [
        {
          planId,
          planName,
          price,
          period: "year",
          active: true,
          renewalDate: nextRenewal,
        },
        current.subscriptions[1], // keep analytics add-on
      ],
      invoices: [newInvoice, ...current.invoices],
    };

    try {
      await apiClient.post("/supplier/subscription/upgrade", { planId });
    } catch {
      // Fallback
    }

    if (typeof window !== "undefined") {
      localStorage.setItem("supplier_billing_overview", JSON.stringify(updated));
    }
    return updated;
  },
};
