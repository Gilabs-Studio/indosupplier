import { apiClient } from "@/lib/api-client";
import type { SupportTicket, SupportTicketDetail, CreateTicketPayload, SupportMessage } from "../types/support.types";

interface ApiResponse<T> {
  success: boolean;
  data: T;
}

export const supportService = {
  async getTickets(): Promise<SupportTicket[]> {
    const response = await apiClient.get<ApiResponse<SupportTicket[]>>("/buyer/support/tickets");
    return response.data.data || [];
  },

  async getTicketByID(id: string): Promise<SupportTicketDetail> {
    const response = await apiClient.get<ApiResponse<SupportTicketDetail>>(`/buyer/support/tickets/${id}`);
    return response.data.data;
  },

  async createTicket(data: CreateTicketPayload): Promise<SupportTicket> {
    const response = await apiClient.post<ApiResponse<SupportTicket>>("/buyer/support/tickets", data);
    return response.data.data;
  },

  async replyTicket(id: string, text: string): Promise<SupportMessage> {
    const response = await apiClient.post<ApiResponse<SupportMessage>>(`/buyer/support/tickets/${id}/messages`, { text });
    return response.data.data;
  },
};
