import { apiClient } from "@/lib/api-client";
import type { BuyerNotification } from "../types/notifications.types";

interface ApiResponse<T> {
  success: boolean;
  data: T;
}

export const notificationsService = {
  async getNotifications(): Promise<BuyerNotification[]> {
    // TODO(notifications): wire the buyer notification endpoints after the
    // event-driven notification flow is finalized.
    const response = await apiClient.get<ApiResponse<BuyerNotification[]>>("/buyer/notifications");
    return response.data.data || [];
  },

  async markAllRead(): Promise<void> {
    await apiClient.post("/buyer/notifications/mark-read");
  },
};
