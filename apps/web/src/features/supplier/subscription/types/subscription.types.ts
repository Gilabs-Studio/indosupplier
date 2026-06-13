export interface BillingInvoice {
  id: string;
  date: string;
  description: string;
  amount: string;
  status: "paid" | "pending" | "failed";
  receiptUrl: string;
}

export interface SubscriptionDetail {
  planId: string;
  planName: string;
  price: string;
  period: string;
  active: boolean;
  renewalDate: string;
}

export interface MeteredUsageStats {
  productUploadsUsed: number;
  productUploadsLimit: string;
  rfqBidsUsed: number;
  rfqBidsLimit: string;
  auctionSlotsUsed: number;
  auctionSlotsLimit: string;
}

export interface BillingOverview {
  currentMeteredUsage: string;
  currentIncludedUsage: string;
  nextPaymentDue: string;
  nextPaymentDate: string;
  subscriptions: SubscriptionDetail[];
  meteredUsage: MeteredUsageStats;
  invoices: BillingInvoice[];
}
