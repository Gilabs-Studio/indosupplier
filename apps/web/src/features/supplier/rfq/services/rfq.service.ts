import { apiClient } from "@/lib/api-client";
import type {
  SupplierRfqItem,
  SubmitSupplierRfqProposalPayload,
  RFQMessageItem,
  SendSupplierRFQMessagePayload,
} from "../types/rfq.types";

interface ApiResponse<T> {
  success: boolean;
  data: T;
  meta?: {
    pagination?: {
      total: number;
    };
  };
}

export const supplierRfqService = {
  async list(params?: { page?: number; per_page?: number }): Promise<{ items: SupplierRfqItem[]; total: number }> {
    const response = await apiClient.get<ApiResponse<SupplierRfqItem[]>>("/supplier/rfqs", { params });
    return {
      items: response.data.data || [],
      total: response.data.meta?.pagination?.total || 0,
    };
  },

  async getByID(id: string): Promise<SupplierRfqItem> {
    const response = await apiClient.get<ApiResponse<SupplierRfqItem>>(`/supplier/rfqs/${id}`);
    return response.data.data;
  },

  async submitProposal(id: string, payload: SubmitSupplierRfqProposalPayload): Promise<void> {
    await apiClient.post(`/supplier/rfqs/${id}/proposals`, payload);
  },

  async getThread(id: string): Promise<RFQMessageItem[]> {
    const response = await apiClient.get<ApiResponse<RFQMessageItem[]>>(`/supplier/rfqs/${id}/thread`);
    return response.data.data || [];
  },

  async sendMessage(id: string, payload: SendSupplierRFQMessagePayload): Promise<RFQMessageItem> {
    const response = await apiClient.post<ApiResponse<RFQMessageItem>>(`/supplier/rfqs/${id}/thread/messages`, payload);
    return response.data.data;
  },
};
