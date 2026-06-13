import { apiClient } from "@/lib/api-client";
import type { BuyerDashboardData } from "../types/dashboard.types";

interface ApiResponse<T> {
  success: boolean;
  data: T;
}

export const dashboardService = {
  async getDashboardData(): Promise<BuyerDashboardData> {
    try {
      const response = await apiClient.get<ApiResponse<BuyerDashboardData>>("/buyer/dashboard");
      return response.data.data;
    } catch (error) {
      console.warn("Backend GET /buyer/dashboard not implemented. Using local mock data.", error);
      return {
        profile_completeness: 85,
        stats: {
          active_rfqs: 6,
          saved_suppliers: 24,
          notifications: 9,
        },
        recent_rfqs: [
          {
            id: "RFQ-2026-004",
            product_name: "Garnet Sand Mesh 80",
            created_at: "2026-06-01",
            status: "Waiting for Quotes",
            replies_count: 3,
          },
          {
            id: "RFQ-2026-003",
            product_name: "Bentonite Clay Powder",
            created_at: "2026-05-28",
            status: "Offers Received",
            replies_count: 8,
          },
          {
            id: "RFQ-2026-002",
            product_name: "Quartz Powder 325 Mesh",
            created_at: "2026-05-15",
            status: "Completed",
            replies_count: 5,
          },
        ],
        suggested_suppliers: [
          {
            id: "1",
            company_name: "PT Rempah Nusantara",
            category: "Agriculture",
            location: "Surabaya",
            rating: 4.8,
            is_verified: true,
          },
          {
            id: "2",
            company_name: "CV Nusantara Garment",
            category: "Textile & Apparel",
            location: "Bandung",
            rating: 4.6,
            is_verified: true,
          },
          {
            id: "3",
            company_name: "PT Logam Steel Jaya",
            category: "Manufacturing",
            location: "Jakarta",
            rating: 4.7,
            is_verified: false,
          },
        ],
      };
    }
  },
};
