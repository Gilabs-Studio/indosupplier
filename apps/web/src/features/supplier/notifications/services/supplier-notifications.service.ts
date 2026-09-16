import { apiClient } from "@/lib/api-client";
import type { SupplierNotification } from "../types/notifications.types";

interface ApiResponse<T> {
  success: boolean;
  data: T;
}

export const supplierNotificationsService = {
  async getNotifications(): Promise<SupplierNotification[]> {
    const response = await apiClient.get<ApiResponse<SupplierNotification[]>>("/supplier/notifications");
    return response.data.data || [];
  },

  async markAllRead(): Promise<void> {
    await apiClient.post("/supplier/notifications/mark-read");
  },
};
