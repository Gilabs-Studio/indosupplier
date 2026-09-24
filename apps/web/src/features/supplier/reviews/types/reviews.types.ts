export interface SupplierReviewItem {
  id: string;
  buyer_profile_id: string;
  buyer_company_name: string;
  buyer_user_name: string;
  buyer_avatar_url: string;
  product_id?: string;
  product_name: string;
  purchase_order_id?: string;
  po_number: string;
  rating: number;
  review_text: string;
  supplier_reply: string;
  supplier_replied_at: string | null;
  status: string;
  created_at: string;
}

export interface SupplierReviewStats {
  total_reviews: number;
  average_rating: number;
  replied_count: number;
  unreplied_count: number;
  rating_breakdown: Record<number, number>;
}

export interface SupplierReviewsPagination {
  current_page: number;
  total_pages: number;
  total_items: number;
  limit: number;
}

export interface SupplierReviewsResponse {
  reviews: SupplierReviewItem[];
  stats: SupplierReviewStats;
  pagination: SupplierReviewsPagination;
}

export interface ReplyReviewPayload {
  reply: string;
}

export interface SupplierReviewsFilter {
  page: number;
  limit: number;
  rating?: number;
  status?: "all" | "unreplied" | "replied";
  search?: string;
}
