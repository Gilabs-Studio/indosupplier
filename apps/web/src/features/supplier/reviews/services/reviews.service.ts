import { apiClient } from "@/lib/api-client";
import type {
  SupplierReviewsResponse,
  SupplierReviewItem,
  ReplyReviewPayload,
  SupplierReviewsFilter,
} from "../types/reviews.types";

interface ApiResponse<T> {
  success: boolean;
  data: T;
  meta?: unknown;
}

export const supplierReviewsService = {
  async getReviews(filter: SupplierReviewsFilter): Promise<SupplierReviewsResponse> {
    const params = new URLSearchParams();
    params.set("page", filter.page.toString());
    params.set("limit", filter.limit.toString());
    if (filter.rating) {
      params.set("rating", filter.rating.toString());
    }
    if (filter.status && filter.status !== "all") {
      params.set("status", filter.status);
    }
    if (filter.search && filter.search.trim()) {
      params.set("search", filter.search.trim());
    }

    const response = await apiClient.get<ApiResponse<SupplierReviewsResponse>>(
      `/supplier/reviews?${params.toString()}`
    );
    return response.data.data;
  },

  async replyReview(
    reviewId: string,
    payload: ReplyReviewPayload
  ): Promise<SupplierReviewItem> {
    const response = await apiClient.post<ApiResponse<SupplierReviewItem>>(
      `/supplier/reviews/${reviewId}/reply`,
      payload
    );
    return response.data.data;
  },
};
