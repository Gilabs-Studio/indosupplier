import { apiClient } from "@/lib/api-client";
import type { BuyerNotification } from "../types/notifications.types";

interface ApiResponse<T> {
  success: boolean;
  data: T;
}

export const notificationsService = {
  async getNotifications(): Promise<BuyerNotification[]> {
    try {
      const response = await apiClient.get<ApiResponse<BuyerNotification[]>>("/buyer/notifications");
      return response.data.data || [];
    } catch (error) {
      console.warn("Backend GET /buyer/notifications not implemented. Using local mock data.", error);
      return [
        {
          id: "1",
          title: "Penawaran Masuk Baru",
          desc: "PT Rempah Nusantara mengirimkan penawaran harga sebesar Rp 12.500 / Kg untuk RFQ Bentonite Clay Powder.",
          date: "2 Jam yang lalu",
          unread: true,
          type: "quote",
        },
        {
          id: "2",
          title: "Pesan Baru dari Supplier",
          desc: "CV Nusantara Garment membalas chat Anda tentang ukuran sampel uniform.",
          date: "5 Jam yang lalu",
          unread: true,
          type: "message",
        },
        {
          id: "3",
          title: "RFQ Berhasil Disiarkan",
          desc: "RFQ-2026-004 (Garnet Sand) Anda telah disetujui oleh admin dan disiarkan ke 12 supplier terdaftar.",
          date: "1 Hari yang lalu",
          unread: false,
          type: "system",
        },
        {
          id: "4",
          title: "Peringatan Dokumen Kedaluwarsa",
          desc: "Masa berlaku dokumen SIUP Anda akan berakhir dalam 30 hari. Segera perbarui untuk menghindari pemblokiran akun.",
          date: "3 Hari yang lalu",
          unread: false,
          type: "alert",
        },
      ];
    }
  },

  async markAllRead(): Promise<void> {
    try {
      await apiClient.post("/buyer/notifications/mark-read");
    } catch (error) {
      console.warn("Backend POST /buyer/notifications/mark-read not implemented.", error);
    }
  },
};
