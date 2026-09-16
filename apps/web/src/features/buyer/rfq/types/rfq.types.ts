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
