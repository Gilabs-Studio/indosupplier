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
}

export interface PublicCategoryDto {
  id: string;
  slug: string;
  name: string;
  description: string;
  icon: string;
  supplierCount: number;
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

export interface PublicProductDetailDto {
  product: PublicProductDto;
  supplier: PublicSupplierDto;
  reviews: PublicReviewDto[];
  relatedProducts: PublicProductDto[];
}
