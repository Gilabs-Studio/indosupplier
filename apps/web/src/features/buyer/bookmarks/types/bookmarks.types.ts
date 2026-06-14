export interface BookmarkItem {
  id: string;
  supplierProfileId: string;
  supplierProductId?: string;
  type: "supplier" | "product";
  supplierSlug: string;
  companyName: string;
  category: string;
  location: string;
  businessType: string;
  establishedYear: number;
  rating: number;
  reviewCount: number;
  isVerified: boolean;
  keyProducts?: string[];

  // Product specific details
  productName?: string;
  productPrice?: number;
  productMinOrder?: string;
  productImage?: string;
}
