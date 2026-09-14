import { apiClient } from "@/lib/api-client";
import type {
  PublicSupplierDto,
  PublicCategoryDto,
  SupplierSearchParams,
  PublicProductDto,
  PublicProductDetailDto,
  ProductReviewsResponse,
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

  async searchProducts(q: string, supplierId?: string, page?: number, limit?: number): Promise<PublicProductDto[]> {
    try {
      const response = await apiClient.get<{ data: PublicProductDto[] }>("/products", {
        params: { q, supplier_id: supplierId, page, limit },
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

  async getProductReviews(
    productId: string,
    params?: { page?: number; limit?: number; rating?: number }
  ): Promise<ProductReviewsResponse> {
    try {
      const response = await apiClient.get<{
        data: {
          summary: {
            average_rating: number;
            total_reviews: number;
            rating_breakdown: Record<string, number>;
            positive_percent: number;
          };
          reviews: Array<{
            id: string;
            buyer_name: string;
            buyer_company: string;
            rating: number;
            review_text: string;
            supplier_reply: string;
            supplier_replied_at?: string;
            created_at: string;
          }>;
          pagination: {
            current_page: number;
            per_page: number;
            total_items: number;
            total_pages: number;
            has_more: boolean;
          };
        };
      }>(`/products/${productId}/reviews`, { params });

      const raw = response.data?.data;
      if (!raw) {
        return {
          summary: { averageRating: 0, totalReviews: 0, ratingBreakdown: { "1": 0, "2": 0, "3": 0, "4": 0, "5": 0 }, positivePercent: 0 },
          reviews: [],
          pagination: { currentPage: 1, perPage: 5, totalItems: 0, totalPages: 0, hasMore: false },
        };
      }

      return {
        summary: {
          averageRating: raw.summary.average_rating || 0,
          totalReviews: raw.summary.total_reviews || 0,
          ratingBreakdown: raw.summary.rating_breakdown || { "1": 0, "2": 0, "3": 0, "4": 0, "5": 0 },
          positivePercent: raw.summary.positive_percent || 0,
        },
        reviews: (raw.reviews || []).map((r) => ({
          id: r.id,
          buyerName: r.buyer_name || "Pembeli Terverifikasi",
          buyerCompany: r.buyer_company || "",
          rating: r.rating || 5,
          reviewText: r.review_text || "",
          supplierReply: r.supplier_reply || "",
          supplierRepliedAt: r.supplier_replied_at,
          createdAt: r.created_at,
        })),
        pagination: {
          currentPage: raw.pagination.current_page || 1,
          perPage: raw.pagination.per_page || 5,
          totalItems: raw.pagination.total_items || 0,
          totalPages: raw.pagination.total_pages || 0,
          hasMore: !!raw.pagination.has_more,
        },
      };
    } catch (error) {
      console.warn(`API error in getProductReviews for ${productId}:`, error);
      return {
        summary: { averageRating: 0, totalReviews: 0, ratingBreakdown: { "1": 0, "2": 0, "3": 0, "4": 0, "5": 0 }, positivePercent: 0 },
        reviews: [],
        pagination: { currentPage: 1, perPage: 5, totalItems: 0, totalPages: 0, hasMore: false },
      };
    }
  },
};
