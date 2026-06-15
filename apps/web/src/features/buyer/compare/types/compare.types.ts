export interface ComparedReview {
  id: string;
  buyerName: string;
  rating: number;
  reviewText: string;
  createdAt: string;
}

export interface ComparedSupplier {
  id: string;
  companyName: string;
  slug: string;
  location: string;
  businessType: string;
  establishedYear: number;
  rating: number;
  reviewCount: number;
  verified: boolean;
  moq: string;
  responseTime: string;
  capacity: string;
  certifications: string[];
  reviews: ComparedReview[];
}

export interface ComparedProduct {
  id: string;
  name: string;
  imageUrl: string;
  price: number;
  moq: string;
  capacity: string;
  categoryName: string;
  description: string;
  supplierId: string;
  supplierCompanyName: string;
  supplierSlug: string;
  supplierRating: number;
  supplierReviewCount: number;
  supplierVerified: boolean;
  supplierLocation: string;
  supplierResponseTime: string;
  reviews: ComparedReview[];
}

