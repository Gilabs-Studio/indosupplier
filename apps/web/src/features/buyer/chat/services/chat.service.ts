import { apiClient } from "@/lib/api-client";
import type { ChatRoom, ChatMessage } from "../types/chat.types";

interface ApiResponse<T> {
  success: boolean;
  data: T;
}

export const chatService = {
  async getRooms(): Promise<ChatRoom[]> {
    const response = await apiClient.get<ApiResponse<ChatRoom[]>>("/buyer/chat/rooms");
    return response.data.data || [];
  },

  async getOrCreateRoom(supplierProfileId: string): Promise<ChatRoom> {
    const response = await apiClient.post<ApiResponse<ChatRoom>>("/buyer/chat/rooms", {
      supplier_profile_id: supplierProfileId,
    });
    return response.data.data;
  },

  async getMessages(roomId: string): Promise<ChatMessage[]> {
    const response = await apiClient.get<ApiResponse<ChatMessage[]>>(`/buyer/chat/rooms/${roomId}/messages`);
    return response.data.data || [];
  },

  async sendMessage(roomId: string, body: string): Promise<ChatMessage> {
    const response = await apiClient.post<ApiResponse<ChatMessage>>(`/buyer/chat/rooms/${roomId}/messages`, {
      body,
    });
    return response.data.data;
  },

  async markAsRead(roomId: string): Promise<void> {
    await apiClient.post(`/buyer/chat/rooms/${roomId}/read`);
  },
};
