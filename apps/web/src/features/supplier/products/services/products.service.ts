import { apiClient, publicApiClient } from "@/lib/api-client";
import type { Product, Category, CreateProductPayload, UpdateProductPayload } from "../types/products.types";

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

export const productsService = {
  async list(params?: {
    search?: string;
    category_id?: string;
    page?: number;
    per_page?: number;
  }): Promise<{ items: Product[]; total: number }> {
    const response = await apiClient.get<ApiResponse<Product[]>>(
      "/supplier/products",
      { params }
    );
    return {
      items: response.data.data || [],
      total: response.data.meta?.pagination?.total || 0,
    };
  },

  async getByID(id: string): Promise<Product> {
    const response = await apiClient.get<ApiResponse<Product>>(
      `/supplier/products/${id}`
    );
    return response.data.data;
  },

  async create(data: CreateProductPayload): Promise<Product> {
    const response = await apiClient.post<ApiResponse<Product>>(
      "/supplier/products",
      data
    );
    return response.data.data;
  },

  async update(id: string, data: UpdateProductPayload): Promise<Product> {
    const response = await apiClient.put<ApiResponse<Product>>(
      `/supplier/products/${id}`,
      data
    );
    return response.data.data;
  },

  async delete(id: string): Promise<void> {
    await apiClient.delete(`/supplier/products/${id}`);
  },

  async listCategories(): Promise<Category[]> {
    const response = await publicApiClient.get<ApiResponse<Category[]>>(
      "/categories"
    );
    return response.data.data || [];
  },

  async uploadImage(file: File): Promise<{ url: string; filename: string }> {
    const formData = new FormData();
    formData.append("file", file);
    const response = await apiClient.post<ApiResponse<{ url: string; filename: string }>>(
      "/upload/image?folder=products",
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
