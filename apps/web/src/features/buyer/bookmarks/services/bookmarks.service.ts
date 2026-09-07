import { apiClient } from "@/lib/api-client";
import type { BookmarkItem } from "../types/bookmarks.types";

interface ApiResponse<T> {
  success: boolean;
  data: T;
}

export interface ToggleBookmarkResult {
  bookmarked: boolean;
  action: "added" | "removed";
  bookmark?: BookmarkItem;
}

export const bookmarksService = {
  async getBookmarks(): Promise<BookmarkItem[]> {
    const response = await apiClient.get<ApiResponse<BookmarkItem[]>>("/buyer/bookmarks");
    return response.data.data || [];
  },

  async addBookmark(supplierProfileId: string, supplierProductId?: string): Promise<BookmarkItem> {
    const response = await apiClient.post<ApiResponse<BookmarkItem>>("/buyer/bookmarks", {
      supplierProfileId,
      supplierProductId,
    });
    return response.data.data;
  },

  async removeBookmark(id: string): Promise<void> {
    await apiClient.delete(`/buyer/bookmarks/${id}`);
  },

  async toggleBookmark(supplierProfileId: string, supplierProductId?: string): Promise<ToggleBookmarkResult> {
    const response = await apiClient.post<ApiResponse<ToggleBookmarkResult>>("/buyer/bookmarks/toggle", {
      supplierProfileId,
      supplierProductId,
    });
    return response.data.data;
  },
};
