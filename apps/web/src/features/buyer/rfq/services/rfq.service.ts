import { apiClient } from "@/lib/api-client";
import type { RFQItem, RFQBid, RFQDetail, CreateRfqPayload } from "../types/rfq.types";

interface ApiResponse<T> {
  success: boolean;
  data: T;
  meta?: {
    pagination?: {
      page: number;
      per_page: number;
      total: number;
      total_pages: number;
    };
  };
}

export const rfqService = {
  async listRfqs(params?: {
    page?: number;
    per_page?: number;
    status?: string;
  }): Promise<{ items: RFQItem[]; total: number }> {
    const response = await apiClient.get<ApiResponse<RFQItem[]>>("/buyer/rfqs", { params });
    return {
      items: response.data.data || [],
      total: response.data.meta?.pagination?.total || 0,
    };
  },

  async getRfqByID(id: string): Promise<RFQDetail> {
    const response = await apiClient.get<ApiResponse<RFQDetail>>(`/buyer/rfqs/${id}`);
    return response.data.data;
  },

  async createRfq(data: CreateRfqPayload): Promise<RFQItem> {
    const response = await apiClient.post<ApiResponse<RFQItem>>("/buyer/rfqs", data);
    return response.data.data;
  },

  async getRfqBids(id: string): Promise<RFQBid[]> {
    const response = await apiClient.get<ApiResponse<RFQBid[]>>(`/buyer/rfqs/${id}/bids`);
    return response.data.data || [];
  },

  async acceptBid(rfqId: string, bidId: string): Promise<void> {
    await apiClient.post(`/buyer/rfqs/${rfqId}/bids/${bidId}/accept`);
  },

  async uploadSpecFile(file: File): Promise<{ url: string; filename: string }> {
    const formData = new FormData();
    formData.append("file", file);
    const response = await apiClient.post<ApiResponse<{ url: string; filename: string }>>(
      "/upload/image?folder=rfqs",
      formData,
      {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      }
    );
    return response.data.data;
  },
};
