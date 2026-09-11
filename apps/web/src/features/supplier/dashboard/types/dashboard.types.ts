export interface TotalSalesMetric {
  current_amount: number;
  formatted_amount: string;
  previous_amount: number;
  growth_percentage: string;
  paid_order_count: number;
}

export interface ActiveProductsMetric {
  total_count: number;
  active_count: number;
  draft_count: number;
  featured_count: number;
}

export interface MatchingRFQsMetric {
  total_open: number;
  expiring_soon_count: number;
  new_this_week_count: number;
}

export interface MonthlySalesPerformance {
  month: string;
  month_name: string;
  month_number: number;
  year: number;
  amount: number;
  formatted_amount: string;
  order_count: number;
}

export interface SellerPerformanceMetric {
  verification_level: number;
  verification_badge: string;
  trust_description: string;
  star_rating: number;
  review_count: number;
  chat_response_rate: string;
  response_time_text: string;
  monthly_visitors: number;
}

export interface RecentRFQItem {
  id: string;
  rfq_number: string;
  product: string;
  category: string;
  quantity: string;
  target_port: string;
  date: string;
  replies: number;
  status: string;
  matching_category: boolean;
}

export interface SupplierDashboardData {
  total_sales: TotalSalesMetric;
  active_products: ActiveProductsMetric;
  matching_rfqs: MatchingRFQsMetric;
  sales_performance: MonthlySalesPerformance[];
  seller_performance: SellerPerformanceMetric;
  recent_rfqs: RecentRFQItem[];
}
