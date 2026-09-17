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
  imageUrl?: string;
  description?: string;
  shippingTerm?: string;
  targetDelivery?: string;
  submittedOffer?: {
    price: string;
    moq: string;
    deliveryTime: string;
    notes: string;
    submittedAt?: string;
    respondedAt?: string;
  };
  buyer: SupplierRfqBuyer;
}

export interface SubmitSupplierRfqProposalPayload {
  price: string;
  moq: string;
  deliveryTime: string;
  notes: string;
}

export interface RFQMessageItem {
  id: string;
  rfqId: string;
  supplierProfileId: string;
  senderType: "buyer" | "supplier" | "system";
  senderId: string;
  senderName: string;
  senderAvatar?: string;
  senderRole?: string;
  senderRating?: number;
  messageType: "message" | "offer" | "bid_accepted" | "system";
  body: string;
  price?: number;
  priceFormatted?: string;
  moq?: string;
  deliveryTime?: string;
  createdAt: string;
  createdAtFormatted: string;
  isMine: boolean;
}

export interface SendSupplierRFQMessagePayload {
  body: string;
  price?: string;
  moq?: string;
  deliveryTime?: string;
}
