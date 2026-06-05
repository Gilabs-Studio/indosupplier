export interface SubscriptionPlan {
  id: string;
  name: string;
  price: number;
  billingCycle: "monthly" | "annually";
  features: string[];
  active: boolean;
  tier: "free" | "basic" | "premium" | "enterprise";
}
