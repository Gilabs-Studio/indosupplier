import { apiClient } from "@/lib/api-client";
import type { SupportTicket, SupportTicketDetail, CreateTicketPayload, SupportMessage } from "../types/support.types";

interface ApiResponse<T> {
  success: boolean;
  data: T;
}

export const supportService = {
  async getTickets(): Promise<SupportTicket[]> {
    try {
      const response = await apiClient.get<ApiResponse<SupportTicket[]>>("/buyer/support/tickets");
      return response.data.data || [];
    } catch (error) {
      console.warn("Backend GET /buyer/support/tickets not implemented. Using local mock data.", error);
      return [
        { id: "TK-2026-001", subject: "Pertanyaan Pengajuan Limit Kredit Sourcing", date: "2026-06-02", status: "Open" },
        { id: "TK-2026-002", subject: "Verifikasi Berkas SIUP Lambat", date: "2026-05-20", status: "Closed" },
      ];
    }
  },

  async getTicketByID(id: string): Promise<SupportTicketDetail> {
    try {
      const response = await apiClient.get<ApiResponse<SupportTicketDetail>>(`/buyer/support/tickets/${id}`);
      return response.data.data;
    } catch (error) {
      console.warn(`Backend GET /buyer/support/tickets/${id} not implemented. Using local mock data.`, error);
      return {
        id: id || "TK-2026-001",
        subject: "Pertanyaan Pengajuan Limit Kredit Sourcing",
        status: "Open",
        date: "2026-06-02",
        messages: [
          {
            sender: "buyer",
            name: "Yohanes",
            text: "Selamat siang admin, saya ingin menanyakan mengenai verifikasi berkas untuk limit kredit perusahaan kami. Kami sudah upload SIUP dan NIB kemarin siang. Kapan kiranya limit kredit kami aktif?",
            time: "2026-06-02 14:00",
          },
          {
            sender: "support",
            name: "IndoSupplier Customer Support",
            text: "Selamat siang Bapak Yohanes, terima kasih telah menghubungi kami. Berkas legalitas Anda sedang dalam antrean verifikasi tim finance kami. Estimasi verifikasi memakan waktu 1-2 hari kerja. Kami akan memberikan notifikasi email segera setelah status limit Anda aktif.",
            time: "2026-06-02 15:30",
          },
        ],
      };
    }
  },

  async createTicket(data: CreateTicketPayload): Promise<SupportTicket> {
    try {
      const response = await apiClient.post<ApiResponse<SupportTicket>>("/buyer/support/tickets", data);
      return response.data.data;
    } catch (error) {
      console.warn("Backend POST /buyer/support/tickets not implemented. Simulating creation.", error);
      return {
        id: "TK-2026-" + Math.floor(100 + Math.random() * 900),
        subject: data.subject,
        date: new Date().toISOString().split("T")[0],
        status: "Open",
      };
    }
  },

  async replyTicket(id: string, text: string): Promise<SupportMessage> {
    try {
      const response = await apiClient.post<ApiResponse<SupportMessage>>(`/buyer/support/tickets/${id}/messages`, { text });
      return response.data.data;
    } catch (error) {
      console.warn(`Backend POST /buyer/support/tickets/${id}/messages not implemented. Simulating reply.`, error);
      return {
        sender: "buyer",
        name: "Yohanes",
        text,
        time: "Baru saja",
      };
    }
  },
};
