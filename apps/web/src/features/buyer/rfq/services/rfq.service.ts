import { apiClient } from "@/lib/api-client";
import type { RFQItem, RFQBid, RFQDetail, CreateRfqPayload } from "../types/rfq.types";

interface ApiResponse<T> {
  success: boolean;
  data: T;
  meta?: {
    pagination?: {
      page: number;
      per_page: number;
      total: number;
      total_pages: number;
    };
  };
}

export const rfqService = {
  async listRfqs(params?: {
    page?: number;
    per_page?: number;
    status?: string;
  }): Promise<{ items: RFQItem[]; total: number }> {
    try {
      const response = await apiClient.get<ApiResponse<RFQItem[]>>("/buyer/rfqs", { params });
      return {
        items: response.data.data || [],
        total: response.data.meta?.pagination?.total || 0,
      };
    } catch (error) {
      console.warn("Backend GET /buyer/rfqs not implemented. Using local mock data.", error);
      const mockRfqs: RFQItem[] = [
        {
          id: "RFQ-2026-004",
          product: "Garnet Sand Mesh 80",
          category: "Industrial Minerals",
          quantity: "50 Ton",
          targetPort: "Tanjung Priok, Jakarta",
          date: "2026-06-01",
          status: "Waiting for Quotes",
          replies: 3,
        },
        {
          id: "RFQ-2026-003",
          product: "Bentonite Clay Powder",
          category: "Chemicals",
          quantity: "20 Ton",
          targetPort: "Tanjung Perak, Surabaya",
          date: "2026-05-28",
          status: "Offers Received",
          replies: 8,
        },
        {
          id: "RFQ-2026-002",
          product: "Quartz Powder 325 Mesh",
          category: "Industrial Minerals",
          quantity: "100 Ton",
          targetPort: "Tanjung Priok, Jakarta",
          date: "2026-05-15",
          status: "Completed",
          replies: 5,
        },
        {
          id: "RFQ-2026-001",
          product: "Organic Coconut Sugar Organic Grade",
          category: "Agriculture & Food",
          quantity: "5 Ton",
          targetPort: "Port of Rotterdam (CIF)",
          date: "2026-05-01",
          status: "Completed",
          replies: 12,
        },
      ];
      
      const filtered = params?.status && params.status !== "all"
        ? mockRfqs.filter((r) => r.status.toLowerCase().includes(params.status!.toLowerCase()))
        : mockRfqs;

      return {
        items: filtered,
        total: filtered.length,
      };
    }
  },

  async getRfqByID(id: string): Promise<RFQDetail> {
    try {
      const response = await apiClient.get<ApiResponse<RFQDetail>>(`/buyer/rfqs/${id}`);
      return response.data.data;
    } catch (error) {
      console.warn(`Backend GET /buyer/rfqs/${id} not implemented. Using local mock data.`, error);
      return {
        id: id || "RFQ-2026-003",
        product: "Bentonite Clay Powder",
        category: "Chemicals",
        quantity: "20 Ton",
        targetPort: "Tanjung Perak, Surabaya",
        date: "2026-05-28",
        status: "Offers Received",
        replies: 3,
        description: "Membutuhkan Bentonite Clay Powder kualitas industri untuk konstruksi sipil. Kemasan bag 25kg, total kebutuhan 20 ton dikirim ke pelabuhan Surabaya. Sertakan Certificate of Analysis (COA) terbaru.",
        attachmentUrl: "/uploads/coa.pdf",
        attachmentName: "COA_Bentonite_Clay_2026.pdf",
        attachmentSize: "1.4 MB",
      };
    }
  },

  async createRfq(data: CreateRfqPayload): Promise<RFQItem> {
    try {
      const response = await apiClient.post<ApiResponse<RFQItem>>("/buyer/rfqs", data);
      return response.data.data;
    } catch (error) {
      console.warn("Backend POST /buyer/rfqs not implemented. Simulating creation.", error);
      return {
        id: "RFQ-2026-" + Math.floor(100 + Math.random() * 900),
        product: data.product_name,
        category: data.category,
        quantity: `${data.quantity} ${data.unit}`,
        targetPort: data.target_port,
        date: new Date().toISOString().split("T")[0],
        status: "Waiting for Quotes",
        replies: 0,
      };
    }
  },

  async getRfqBids(id: string): Promise<RFQBid[]> {
    try {
      const response = await apiClient.get<ApiResponse<RFQBid[]>>(`/buyer/rfqs/${id}/bids`);
      return response.data.data || [];
    } catch (error) {
      console.warn(`Backend GET /buyer/rfqs/${id}/bids not implemented. Using local mock data.`, error);
      return [
        {
          id: "bid-1",
          supplierName: "PT Rempah Nusantara",
          price: "Rp 12.500 / Kg",
          moq: "5 Ton",
          responseTime: "2 Jam",
          verified: true,
        },
        {
          id: "bid-2",
          supplierName: "CV Indo Mineral Utama",
          price: "Rp 11.800 / Kg",
          moq: "10 Ton",
          responseTime: "3 Jam",
          verified: true,
        },
        {
          id: "bid-3",
          supplierName: "PT Sumber Batuan Jaya",
          price: "Rp 13.000 / Kg",
          moq: "1 Ton",
          responseTime: "5 Jam",
          verified: false,
        },
      ];
    }
  },

  async acceptBid(rfqId: string, bidId: string): Promise<void> {
    try {
      await apiClient.post(`/buyer/rfqs/${rfqId}/bids/${bidId}/accept`);
    } catch (error) {
      console.warn(`Backend POST /buyer/rfqs/${rfqId}/bids/${bidId}/accept not implemented.`, error);
    }
  },

  async uploadSpecFile(file: File): Promise<{ url: string; filename: string }> {
    const formData = new FormData();
    formData.append("file", file);
    const response = await apiClient.post<ApiResponse<{ url: string; filename: string }>>(
      "/upload/image?folder=rfqs",
      formData,
      {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      }
    );
    return response.data.data;
  },
};
