export interface Review {
  id: string;
  buyerName: string;
  supplierName: string;
  productName?: string;
  rating: number;
  content: string;
  reply?: string;
  status: "approved" | "pending" | "flagged" | "spam";
  createdAt: string;
}
