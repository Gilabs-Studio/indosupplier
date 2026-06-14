export interface TransactionItem {
  id: string;
  po_number: string;
  buyer_profile_id: string;
  supplier_profile_id: string;
  supplier_name: string;
  rfq_id?: string;
  product_name: string;
  quantity_value: number;
  quantity_unit: string;
  price_per_unit: number;
  total_amount: number;
  status: "pending" | "processing" | "shipped" | "completed" | "cancelled";
  payment_status: "unpaid" | "paid" | "refunded";
  delivery_address: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateTransactionPayload {
  supplier_profile_id: string;
  rfq_id?: string;
  product_name: string;
  quantity_value: number;
  quantity_unit: string;
  price_per_unit: number;
  delivery_address: string;
  notes?: string;
}
