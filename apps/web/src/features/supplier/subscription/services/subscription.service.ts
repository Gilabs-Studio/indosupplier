import { apiClient } from "@/lib/api-client";
import type { BillingOverview } from "../types/subscription.types";

export const subscriptionService = {
  async getBillingOverview(): Promise<BillingOverview> {
    const response = await apiClient.get<{ data: BillingOverview }>("/supplier/billing-overview");
    return response.data.data;
  },

  async upgradePlan(planId: string): Promise<BillingOverview> {
    const response = await apiClient.post<{ data: BillingOverview }>("/supplier/subscription/upgrade", { planId });
    return response.data.data;
  },
};
