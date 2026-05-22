import type {
  ApiEnvelope,
  ListNotificationsParams,
  ListNotificationsResponse,
  NotificationResponse,
  UnreadCountResponse,
} from "../types";
import { apiClient } from "@/lib/api-client";

const BASE_PATH = "/notifications";

export const notificationService = {
  async list(params?: ListNotificationsParams): Promise<ListNotificationsResponse> {
    const response = await apiClient.get<ListNotificationsResponse>(BASE_PATH, { params });
    return response.data;
  },

  async getUnreadCount(): Promise<number> {
    const response = await apiClient.get<UnreadCountResponse>(`${BASE_PATH}/unread-count`);
    return response.data.data.unread_count;
  },

  async markAsRead(id: string): Promise<NotificationResponse> {
    const response = await apiClient.post<NotificationResponse>(`${BASE_PATH}/${id}/read`);
    return response.data;
  },

  async markAllAsRead(): Promise<ApiEnvelope<{ marked: number }>> {
    const response = await apiClient.post<ApiEnvelope<{ marked: number }>>(`${BASE_PATH}/read-all`);
    return response.data;
  },
};

