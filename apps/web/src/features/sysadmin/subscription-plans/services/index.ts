import type { SubscriptionPlan } from "../types";

let mockPlans: SubscriptionPlan[] = [
  {
    id: "SUB-001",
    name: "Free Onboarding Tier",
    price: 0,
    billingCycle: "monthly",
    features: ["Listing Produk Maks. 5", "RFQ Broadcast Publik", "Lencana Standard"],
    active: true,
    tier: "free"
  },
  {
    id: "SUB-002",
    name: "Supplier Basic",
    price: 499000,
    billingCycle: "monthly",
    features: ["Listing Produk Maks. 50", "Akses RFQ Eksklusif", "Lencana Verified Level 1", "Dukungan Email CS"],
    active: true,
    tier: "basic"
  },
  {
    id: "SUB-003",
    name: "Supplier Premium",
    price: 1499000,
    billingCycle: "monthly",
    features: ["Listing Produk Tanpa Batas", "Pemberitahuan Instan RFQ", "Lencana Verified Level 2", "Akses Sesi Lelang Iklan", "Prioritas Pencarian", "Dukungan WA CS Dedikasi"],
    active: true,
    tier: "premium"
  },
  {
    id: "SUB-004",
    name: "Supplier Premium (Tahunan)",
    price: 14990000,
    billingCycle: "annually",
    features: ["Listing Produk Tanpa Batas", "Pemberitahuan Instan RFQ", "Lencana Verified Level 2 & 3", "Akses Sesi Lelang Iklan", "Prioritas Pencarian Utama", "Diskon 2 Bulan Berlangganan", "Dukungan Akun Manajer VIP"],
    active: true,
    tier: "premium"
  }
];

export const subscriptionPlanService = {
  async list(): Promise<SubscriptionPlan[]> {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve([...mockPlans]);
      }, 100);
    });
  },

  async update(id: string, updates: Partial<SubscriptionPlan>): Promise<SubscriptionPlan> {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        const plan = mockPlans.find(p => p.id === id);
        if (!plan) {
          reject(new Error("Plan not found"));
          return;
        }
        Object.assign(plan, updates);
        resolve({ ...plan });
      }, 100);
    });
  },

  async create(data: Omit<SubscriptionPlan, "id">): Promise<SubscriptionPlan> {
    return new Promise((resolve) => {
      setTimeout(() => {
        const newPlan: SubscriptionPlan = {
          ...data,
          id: `SUB-00${mockPlans.length + 1}`
        };
        mockPlans.push(newPlan);
        resolve(newPlan);
      }, 100);
    });
  },

  async delete(id: string): Promise<void> {
    return new Promise((resolve) => {
      setTimeout(() => {
        mockPlans = mockPlans.filter(p => p.id !== id);
        resolve();
      }, 100);
    });
  }
};
