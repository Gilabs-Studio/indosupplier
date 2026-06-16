import { apiClient } from "@/lib/api-client";
import type { EligibleTransaction, ReviewHistory, CreateReviewPayload } from "../types/reviews.types";

interface ApiResponse<T> {
  success: boolean;
  data: T;
}

export const reviewsService = {
  async getEligibleTransactions(): Promise<EligibleTransaction[]> {
    const response = await apiClient.get<ApiResponse<EligibleTransaction[]>>("/buyer/reviews/eligible");
    return response.data.data || [];
  },

  async getReviewHistory(): Promise<ReviewHistory[]> {
    const response = await apiClient.get<ApiResponse<ReviewHistory[]>>("/buyer/reviews/history");
    return response.data.data || [];
  },

  async createReview(payload: CreateReviewPayload): Promise<unknown> {
    const response = await apiClient.post<ApiResponse<unknown>>("/buyer/reviews", payload);
    return response.data.data;
  },
};
