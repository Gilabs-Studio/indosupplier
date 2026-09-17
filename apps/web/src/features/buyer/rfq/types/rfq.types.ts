export interface RFQSupplierSummary {
  id: string;
  supplierProfileId: string;
  supplierName: string;
  status: string;
  price?: string;
  isAccepted: boolean;
}

export interface RFQItem {
  id: string;
  product: string;
  category: string;
  quantity: string;
  targetPort: string;
  date: string;
  status: string;
  replies: number;
  imageUrl?: string;
  suppliers?: RFQSupplierSummary[];
}

export interface RFQBid {
  id: string;
  supplierName: string;
  supplierProfileId?: string;
  price: string;
  moq: string;
  responseTime: string;
  verified: boolean;
}

export interface RFQDetail extends RFQItem {
  description: string;
  attachmentUrl?: string;
  attachmentName?: string;
  attachmentSize?: string;
}

export interface CreateRfqPayload {
  product_name: string;
  category: string;
  quantity: string;
  unit: string;
  target_port: string;
  description?: string;
  attachment_url?: string;
  attachment_name?: string;
  attachment_size?: number;
  target_supplier_id?: string;
  product_id?: string;
  image_url?: string;
  budget?: number;
  delivery_timeline?: string;
}

export interface RFQThreadSupplier {
  supplierProfileId: string;
  supplierName: string;
  supplierLogo?: string;
  verified: boolean;
  rating: number;
  city?: string;
  latestOffer?: string;
  latestOfferAt?: string;
  status: string;
  messageCount: number;
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

export interface SendRFQMessagePayload {
  body: string;
  price?: string;
  moq?: string;
  deliveryTime?: string;
}
