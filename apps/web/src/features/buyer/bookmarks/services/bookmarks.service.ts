import { apiClient } from "@/lib/api-client";
import type { BookmarkSupplier } from "../types/bookmarks.types";

interface ApiResponse<T> {
  success: boolean;
  data: T;
}

export const bookmarksService = {
  async getBookmarks(): Promise<BookmarkSupplier[]> {
    try {
      const response = await apiClient.get<ApiResponse<BookmarkSupplier[]>>("/buyer/bookmarks");
      return response.data.data || [];
    } catch (error) {
      console.warn("Backend GET /buyer/bookmarks not implemented. Using local mock data.", error);
      return [
        {
          id: "1",
          companyName: "PT Rempah Nusantara",
          category: "Agriculture",
          location: "Surabaya, Jawa Timur",
          businessType: "Manufacturer",
          establishedYear: 2012,
          rating: 4.8,
          reviewCount: 128,
          isVerified: true,
          keyProducts: ["Coffee Beans", "Coconut Sugar", "Spice Mixes"],
        },
        {
          id: "2",
          companyName: "CV Nusantara Garment",
          category: "Textile & Apparel",
          location: "Bandung, Jawa Barat",
          businessType: "Manufacturer & Exporter",
          establishedYear: 2015,
          rating: 4.6,
          reviewCount: 94,
          isVerified: true,
          keyProducts: ["Cotton Shirts", "Denim Jackets", "Uniforms"],
        },
        {
          id: "3",
          companyName: "PT Logam Steel Jaya",
          category: "Manufacturing",
          location: "Bekasi, Jawa Barat",
          businessType: "Manufacturer",
          establishedYear: 2008,
          rating: 4.7,
          reviewCount: 56,
          isVerified: false,
          keyProducts: ["Steel Pipes", "Wire Mesh", "Metal Sheets"],
        },
      ];
    }
  },

  async removeBookmark(id: string): Promise<void> {
    try {
      await apiClient.delete(`/buyer/bookmarks/${id}`);
    } catch (error) {
      console.warn(`Backend DELETE /buyer/bookmarks/${id} not implemented.`, error);
    }
  },
};
