import { apiClient } from "@/lib/api-client";
import type { TransactionItem, CreateTransactionPayload } from "../types/transaction.types";

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

const mockTransactions: TransactionItem[] = [
  {
    id: "tx-1",
    po_number: "PO-20260613-A01B2C",
    buyer_profile_id: "buyer-1",
    supplier_profile_id: "supplier-1",
    supplier_name: "PT Baja Sentosa",
    product_name: "Reinforced Steel Bar (Rebar) D10",
    quantity_value: 20,
    quantity_unit: "Ton",
    price_per_unit: 12500000,
    total_amount: 250000000,
    status: "processing",
    payment_status: "paid",
    delivery_address: "Kawasan Industri MM2100, Cibitung, Bekasi",
    notes: "Mohon diikat kuat per bundel 2 ton.",
    created_at: "2026-06-13T05:00:00Z",
    updated_at: "2026-06-13T05:30:00Z",
  },
  {
    id: "tx-2",
    po_number: "PO-20260612-C02D3E",
    buyer_profile_id: "buyer-1",
    supplier_profile_id: "supplier-2",
    supplier_name: "CV Tekstil Nusantara",
    product_name: "Raw Indigo Denim Fabric 12oz",
    quantity_value: 1000,
    quantity_unit: "Meters",
    price_per_unit: 35000,
    total_amount: 35000000,
    status: "completed",
    payment_status: "paid",
    delivery_address: "Jl. Industri Garmen No. 45, Bandung",
    notes: "Gulungan dilapis plastik tebal antiair.",
    created_at: "2026-06-12T03:00:00Z",
    updated_at: "2026-06-12T04:15:00Z",
  },
  {
    id: "tx-3",
    po_number: "PO-20260611-F03G4H",
    buyer_profile_id: "buyer-1",
    supplier_profile_id: "supplier-3",
    supplier_name: "PT Agro Indo Sejahtera",
    product_name: "Sumatra Gayo Arabica Coffee Beans Grade 1",
    quantity_value: 500,
    quantity_unit: "Kg",
    price_per_unit: 92000,
    total_amount: 46000000,
    status: "pending",
    payment_status: "unpaid",
    delivery_address: "Ruko Sentra Bisnis Blok B, Jakarta Barat",
    notes: "Sertakan sertifikat asal biji kopi gayo.",
    created_at: "2026-06-11T08:00:00Z",
    updated_at: "2026-06-11T08:00:00Z",
  },
  {
    id: "tx-4",
    po_number: "PO-20260610-H04I5J",
    buyer_profile_id: "buyer-1",
    supplier_profile_id: "supplier-1",
    supplier_name: "PT Baja Sentosa",
    product_name: "Hot Rolled Carbon Steel Plate 6mm",
    quantity_value: 5,
    quantity_unit: "Ton",
    price_per_unit: 14200000,
    total_amount: 71000000,
    status: "cancelled",
    payment_status: "unpaid",
    delivery_address: "Kawasan Industri MM2100, Cibitung, Bekasi",
    notes: "Pembatalan karena perubahan spek konstruksi.",
    created_at: "2026-06-10T02:00:00Z",
    updated_at: "2026-06-10T10:00:00Z",
  },
];

export const transactionService = {
  async listTransactions(params?: {
    page?: number;
    per_page?: number;
    status?: string;
  }): Promise<{ items: TransactionItem[]; total: number }> {
    try {
      const response = await apiClient.get<ApiResponse<TransactionItem[]>>("/buyer/transactions", { params });
      return {
        items: response.data.data || [],
        total: response.data.meta?.pagination?.total || 0,
      };
    } catch (error) {
      console.warn("Backend GET /buyer/transactions not implemented or error. Using local mock data.", error);
      
      const filtered = params?.status && params.status !== "all"
        ? mockTransactions.filter((t) => t.status === params.status)
        : mockTransactions;

      const page = params?.page || 1;
      const perPage = params?.per_page || 10;
      const startIndex = (page - 1) * perPage;
      const paginated = filtered.slice(startIndex, startIndex + perPage);

      return {
        items: paginated,
        total: filtered.length,
      };
    }
  },

  async getTransactionByID(id: string): Promise<TransactionItem> {
    try {
      const response = await apiClient.get<ApiResponse<TransactionItem>>(`/buyer/transactions/${id}`);
      return response.data.data;
    } catch (error) {
      console.warn(`Backend GET /buyer/transactions/${id} not implemented or error. Using local mock data.`, error);
      const tx = mockTransactions.find((t) => t.id === id || t.po_number === id);
      if (!tx) {
        throw new Error("Transaction not found");
      }
      return tx;
    }
  },

  async createTransaction(data: CreateTransactionPayload): Promise<TransactionItem> {
    try {
      const response = await apiClient.post<ApiResponse<TransactionItem>>("/buyer/transactions", data);
      return response.data.data;
    } catch (error) {
      console.warn("Backend POST /buyer/transactions not implemented or error. Simulating creation.", error);
      
      const newTx: TransactionItem = {
        id: "tx-" + Math.floor(100 + Math.random() * 900),
        po_number: "PO-" + new Date().toISOString().split("T")[0].replace(/-/g, "") + "-" + Math.random().toString(36).substring(2, 8).toUpperCase(),
        buyer_profile_id: "buyer-1",
        supplier_profile_id: data.supplier_profile_id,
        supplier_name: "PT Baja Sentosa",
        rfq_id: data.rfq_id,
        product_name: data.product_name,
        quantity_value: data.quantity_value,
        quantity_unit: data.quantity_unit,
        price_per_unit: data.price_per_unit,
        total_amount: data.quantity_value * data.price_per_unit,
        status: "pending",
        payment_status: "unpaid",
        delivery_address: data.delivery_address,
        notes: data.notes,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      
      mockTransactions.unshift(newTx);
      return newTx;
    }
  },
};
