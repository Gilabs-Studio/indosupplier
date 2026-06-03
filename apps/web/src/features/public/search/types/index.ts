export interface SupplierProductDto {
  id: string;
  name: string;
  description?: string;
  price?: number;
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
  employeeCount: number;
  location: string;
  address: string;
  description: string;
  isVerified: boolean;
  rating: number;
  reviewCount: number;
  keyProducts: string[];
  certifications: string[];
  products?: SupplierProductDto[];
  certificationList?: SupplierCertificationDto[];
  phone?: string;
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
