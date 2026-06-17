import { apiClient } from "@/lib/api-client";
import type {
  ContentArticle,
  ContentArticlePayload,
  ContentListParams,
  ContentListResult,
  PaginationMeta,
} from "../types/content.types";

interface ApiResponse<T> {
  data: T;
  meta?: {
    pagination?: PaginationMeta;
  };
}

function normalizeParams(params: ContentListParams = {}) {
  return {
    type: params.type,
    locale: params.locale,
    status: params.status === "all" ? undefined : params.status,
    search: params.search,
    page: params.page,
    per_page: params.perPage,
  };
}

export const contentService = {
  async listPublic(params: ContentListParams = {}): Promise<ContentListResult> {
    const response = await apiClient.get<ApiResponse<ContentArticle[]>>("/content/articles", {
      params: normalizeParams(params),
    });
    return {
      items: response.data.data || [],
      pagination: response.data.meta?.pagination,
    };
  },

  async getPublicBySlug(slug: string, locale: "id" | "en"): Promise<ContentArticle | null> {
    try {
      const response = await apiClient.get<ApiResponse<ContentArticle>>(`/content/articles/${slug}`, {
        params: { locale },
      });
      return response.data.data || null;
    } catch {
      return null;
    }
  },

  async listAdmin(params: ContentListParams = {}): Promise<ContentListResult> {
    const response = await apiClient.get<ApiResponse<ContentArticle[]>>("/sysadmin/content/articles", {
      params: normalizeParams(params),
    });
    return {
      items: response.data.data || [],
      pagination: response.data.meta?.pagination,
    };
  },

  async create(payload: ContentArticlePayload): Promise<ContentArticle> {
    const response = await apiClient.post<ApiResponse<ContentArticle>>("/sysadmin/content/articles", payload);
    return response.data.data;
  },

  async update(id: string, payload: Partial<ContentArticlePayload>): Promise<ContentArticle> {
    const response = await apiClient.put<ApiResponse<ContentArticle>>(`/sysadmin/content/articles/${id}`, payload);
    return response.data.data;
  },

  async delete(id: string): Promise<void> {
    await apiClient.delete(`/sysadmin/content/articles/${id}`);
  },
};
