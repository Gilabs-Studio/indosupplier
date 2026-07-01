export interface SupplierRfqBuyer {
  name: string;
  established: string;
  location: string;
  rating: string;
}

export interface SupplierRfqItem {
  id: string;
  product: string;
  category: string;
  quantity: string;
  port: string;
  date: string;
  budget: string;
  status: "open" | "new" | "responded" | "processing" | "accepted" | "closed";
  description?: string;
  shippingTerm?: string;
  targetDelivery?: string;
  buyer: SupplierRfqBuyer;
}

export interface SubmitSupplierRfqProposalPayload {
  price: string;
  moq: string;
  deliveryTime: string;
  notes: string;
}
