import { apiClient } from "@/lib/api-client";
import type {
  PublicSupplierDto,
  PublicCategoryDto,
  SupplierSearchParams,
  PublicProductDto,
  PublicProductDetailDto,
} from "../types";

export const searchService = {
  async search(params: SupplierSearchParams): Promise<PublicSupplierDto[]> {
    try {
      const response = await apiClient.get<{ data: PublicSupplierDto[] }>("/suppliers", {
        params: {
          q: params.query,
          category: params.category,
          region: params.region,
          verified: params.verifiedOnly ? "true" : undefined,
        },
      });
      return response.data?.data || [];
    } catch (error) {
      console.warn("API error fetching suppliers, returning empty state:", error);
      // Backend does not have endpoint, return empty list per API-First rules
      return [];
    }
  },

  async getCategories(): Promise<PublicCategoryDto[]> {
    try {
      const response = await apiClient.get<{ data: PublicCategoryDto[] }>("/categories");
      return response.data?.data || [];
    } catch (error) {
      console.warn("API error fetching categories, returning empty state:", error);
      return [];
    }
  },

  async getSupplierBySlug(slug: string): Promise<PublicSupplierDto | null> {
    try {
      const response = await apiClient.get<{ data: PublicSupplierDto }>(`/suppliers/${slug}`);
      return response.data?.data || null;
    } catch (error) {
      console.warn(`API error fetching supplier ${slug}, returning null:`, error);
      return null;
    }
  },

  async searchProducts(q: string): Promise<PublicProductDto[]> {
    try {
      const response = await apiClient.get<{ data: PublicProductDto[] }>("/products", {
        params: { q },
      });
      return response.data?.data || [];
    } catch (error) {
      console.warn("API error fetching products, returning empty state:", error);
      return [];
    }
  },

  async getProductById(id: string): Promise<PublicProductDetailDto | null> {
    try {
      const response = await apiClient.get<{ data: PublicProductDetailDto }>(`/products/${id}`);
      return response.data?.data || null;
    } catch (error) {
      console.warn(`API error fetching product ${id}, returning null:`, error);
      return null;
    }
  },

  async lookupSuppliers(q: string, page: number): Promise<PublicSupplierDto[]> {
    try {
      const response = await apiClient.get<{ data: PublicSupplierDto[] }>("/suppliers/lookup", {
        params: { q, page, limit: 5 },
      });
      return response.data?.data || [];
    } catch (error) {
      console.warn("API error in lookupSuppliers, returning empty:", error);
      return [];
    }
  },

  async lookupProducts(q: string, page: number): Promise<PublicProductDto[]> {
    try {
      const response = await apiClient.get<{ data: PublicProductDto[] }>("/products/lookup", {
        params: { q, page, limit: 5 },
      });
      return response.data?.data || [];
    } catch (error) {
      console.warn("API error in lookupProducts, returning empty:", error);
      return [];
    }
  },
};
