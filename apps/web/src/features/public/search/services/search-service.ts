import { apiClient } from "@/lib/api-client";
import type { PublicSupplierDto, PublicCategoryDto, SupplierSearchParams } from "../types";

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
};
