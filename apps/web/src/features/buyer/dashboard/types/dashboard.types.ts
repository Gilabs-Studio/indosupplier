export interface DashboardStats {
  active_rfqs: number;
  saved_suppliers: number;
  notifications: number;
}

export interface RecentRfq {
  id: string;
  product_name: string;
  created_at: string;
  status: string;
  replies_count: number;
}

export interface SuggestedSupplier {
  id: string;
  company_name: string;
  category: string;
  location: string;
  rating: number;
  is_verified: boolean;
}

export interface BuyerDashboardData {
  profile_completeness: number;
  stats: DashboardStats;
  recent_rfqs: RecentRfq[];
  suggested_suppliers: SuggestedSupplier[];
}
