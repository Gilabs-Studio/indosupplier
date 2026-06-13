import { apiClient } from "@/lib/api-client";
import type { ComparedSupplier } from "../types/compare.types";

interface ApiResponse<T> {
  success: boolean;
  data: T;
}

export const compareService = {
  async getComparedSuppliers(): Promise<ComparedSupplier[]> {
    try {
      const response = await apiClient.get<ApiResponse<ComparedSupplier[]>>("/buyer/compare");
      return response.data.data || [];
    } catch (error) {
      console.warn("Backend GET /buyer/compare not implemented. Using local mock data.", error);
      return [
        {
          id: "1",
          companyName: "PT Rempah Nusantara",
          location: "Surabaya, Jawa Timur",
          businessType: "Manufacturer",
          establishedYear: 2012,
          rating: 4.8,
          reviewCount: 128,
          verified: true,
          moq: "500 Kg",
          responseTime: "2 Jam",
          capacity: "20 Ton / Bulan",
          certifications: ["BPOM", "Halal", "HACCP"],
        },
        {
          id: "2",
          companyName: "CV Nusantara Garment",
          location: "Bandung, Jawa Barat",
          businessType: "Manufacturer",
          establishedYear: 2015,
          rating: 4.6,
          reviewCount: 94,
          verified: true,
          moq: "100 Pcs",
          responseTime: "4 Jam",
          capacity: "10,000 Pcs / Bulan",
          certifications: ["SNI", "OEKO-TEX"],
        },
      ];
    }
  },

  async removeComparedSupplier(id: string): Promise<void> {
    try {
      await apiClient.delete(`/buyer/compare/${id}`);
    } catch (error) {
      console.warn(`Backend DELETE /buyer/compare/${id} not implemented.`, error);
    }
  },
};
