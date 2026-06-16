export interface EligibleTransaction {
  id: string;
  poNumber: string;
  supplierProfileId: string;
  supplierName: string;
  productName: string;
  quantityValue: number;
  quantityUnit: string;
  totalAmount: number;
  date: string;
}

export interface ReviewHistory {
  id: string;
  poNumber: string;
  supplierProfileId: string;
  supplierName: string;
  productName: string;
  quantityValue: number;
  quantityUnit: string;
  totalAmount: number;
  rating: number;
  reviewText: string;
  status: string;
  date: string;
}

export interface CreateReviewPayload {
  purchaseOrderId: string;
  rating: number;
  reviewText: string;
}
