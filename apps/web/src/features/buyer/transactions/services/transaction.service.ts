import { apiClient } from "@/lib/api-client";
import type { TransactionItem, CreateTransactionPayload } from "../types/transaction.types";

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

export const transactionService = {
  async listTransactions(params?: {
    page?: number;
    per_page?: number;
    status?: string;
  }): Promise<{ items: TransactionItem[]; total: number }> {
    const response = await apiClient.get<ApiResponse<TransactionItem[]>>("/buyer/transactions", { params });
    return {
      items: response.data.data || [],
      total: response.data.meta?.pagination?.total || 0,
    };
  },

  async getTransactionByID(id: string): Promise<TransactionItem> {
    const response = await apiClient.get<ApiResponse<TransactionItem>>(`/buyer/transactions/${id}`);
    return response.data.data;
  },

  async createTransaction(data: CreateTransactionPayload): Promise<TransactionItem> {
    const response = await apiClient.post<ApiResponse<TransactionItem>>("/buyer/transactions", data);
    return response.data.data;
  },
};
