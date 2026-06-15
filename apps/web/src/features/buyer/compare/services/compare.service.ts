import { apiClient } from "@/lib/api-client";
import type { ComparedSupplier, ComparedProduct } from "../types/compare.types";

interface ApiResponse<T> {
  success: boolean;
  data: T;
}

export const compareService = {
  // Supplier comparison
  async getComparedSuppliers(): Promise<ComparedSupplier[]> {
    const response = await apiClient.get<ApiResponse<ComparedSupplier[]>>("/buyer/compare");
    return response.data.data || [];
  },

  async addComparedSupplier(supplierProfileId: string): Promise<ComparedSupplier[]> {
    const response = await apiClient.post<ApiResponse<ComparedSupplier[]>>("/buyer/compare", {
      supplierProfileId,
    });
    return response.data.data || [];
  },

  async removeComparedSupplier(id: string): Promise<ComparedSupplier[]> {
    const response = await apiClient.delete<ApiResponse<ComparedSupplier[]>>(`/buyer/compare/${id}`);
    return response.data.data || [];
  },

  // Product comparison
  async getComparedProducts(): Promise<ComparedProduct[]> {
    const response = await apiClient.get<ApiResponse<ComparedProduct[]>>("/buyer/compare/products");
    return response.data.data || [];
  },

  async addComparedProduct(supplierProductId: string): Promise<ComparedProduct[]> {
    const response = await apiClient.post<ApiResponse<ComparedProduct[]>>("/buyer/compare/products", {
      supplierProductId,
    });
    return response.data.data || [];
  },

  async removeComparedProduct(id: string): Promise<ComparedProduct[]> {
    const response = await apiClient.delete<ApiResponse<ComparedProduct[]>>(`/buyer/compare/products/${id}`);
    return response.data.data || [];
  },
};
