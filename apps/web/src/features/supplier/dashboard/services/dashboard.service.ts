import { apiClient } from "@/lib/api-client";
import { SupplierDashboardData } from "../types/dashboard.types";

interface ApiResponse<T> {
  success: boolean;
  data: T;
  timestamp?: string;
  request_id?: string;
}

export const supplierDashboardService = {
  async getDashboard(): Promise<SupplierDashboardData> {
    const response = await apiClient.get<ApiResponse<SupplierDashboardData>>("/supplier/dashboard");
    return response.data.data;
  },
};
