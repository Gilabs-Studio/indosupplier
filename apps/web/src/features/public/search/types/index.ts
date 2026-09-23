export interface SupplierProductDto {
  id: string;
  name: string;
  description?: string;
  price?: number;
  currency?: string;
  minOrder?: string;
  capacityText?: string;
  categoryName?: string;
  photos?: string[];
}

export interface SupplierCertificationDto {
  id: string;
  name: string;
  institution?: string;
  year?: number;
}

export interface PublicSupplierDto {
  id: string;
  slug: string;
  companyName: string;
  businessType: string;
  establishedYear: number;
  employeeCount: string | number;
  location: string;
  province?: string;
  address: string;
  description: string;
  isVerified: boolean;
  verificationLevel?: number;
  isPremiumVerified?: boolean;
  taxStatus?: string;
  responseRate?: number;
  responseTime?: string;
  rating: number;
  reviewCount: number;
  keyProducts: string[];
  certifications: string[];
  products?: SupplierProductDto[];
  certificationList?: SupplierCertificationDto[];
  reviews?: PublicReviewDto[];
  phone?: string;
  whatsApp?: string;
  email?: string;
  website?: string;
  logo?: string;
}

export interface PublicCategoryDto {
  id: string;
  slug: string;
  name: string;
  description: string;
  icon?: string;
  iconUrl?: string;
  icon_url?: string;
  supplierCount?: number;
  supplier_count?: number;
  productCount?: number;
  product_count?: number;
}

export interface SupplierSearchParams {
  query?: string;
  category?: string;
  region?: string;
  verifiedOnly?: boolean;
}

export interface PublicProductDto {
  id: string;
  name: string;
  description: string;
  price: number;
  currency: string;
  minOrder: string;
  capacityText: string;
  categoryName: string;
  photos: string[];
  supplierId: string;
  supplierCompanyName: string;
  supplierSlug: string;
  supplierLocation: string;
  supplierVerified: boolean;
  supplierRating: number;
  supplierReviewCount: number;
  rating?: number;
  reviewCount?: number;
}

export interface PublicReviewDto {
  id: string;
  buyerName: string;
  rating: number;
  reviewText: string;
  supplierReply: string;
  supplierRepliedAt?: string;
  createdAt: string;
}

export interface ProductReviewSummary {
  averageRating: number;
  totalReviews: number;
  ratingBreakdown: Record<string, number>;
  positivePercent: number;
}

export interface ProductReviewItem {
  id: string;
  buyerName: string;
  buyerCompany: string;
  rating: number;
  reviewText: string;
  supplierReply?: string;
  supplierRepliedAt?: string;
  createdAt: string;
}

export interface ProductReviewsPagination {
  currentPage: number;
  perPage: number;
  totalItems: number;
  totalPages: number;
  hasMore: boolean;
}

export interface ProductReviewsResponse {
  summary: ProductReviewSummary;
  reviews: ProductReviewItem[];
  pagination: ProductReviewsPagination;
}

export interface PublicProductDetailDto {
  product: PublicProductDto;
  supplier: PublicSupplierDto;
  reviews: PublicReviewDto[];
  relatedProducts: PublicProductDto[];
}
