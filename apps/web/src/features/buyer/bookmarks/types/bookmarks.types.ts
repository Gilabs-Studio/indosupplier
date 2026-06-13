export interface BookmarkSupplier {
  id: string;
  companyName: string;
  category: string;
  location: string;
  businessType: string;
  establishedYear: number;
  rating: number;
  reviewCount: number;
  isVerified: boolean;
  keyProducts: string[];
}
