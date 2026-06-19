export interface FollowingProductItem {
  id: string;
  name: string;
  image: string;
}

export interface FollowingSupplierItem {
  id: string;
  supplierProfileId: string;
  supplierSlug: string;
  companyName: string;
  category: string;
  location: string;
  businessType: string;
  establishedYear: number;
  rating: number;
  reviewCount: number;
  isVerified: boolean;
  keyProducts: FollowingProductItem[];
  logo?: string;
  createdAt: string;
}
