export interface DemoBannerHighlight {
  id: string;
  iconName: "dollar" | "box" | "receipt";
  titleId: string;
  titleEn: string;
}

export interface DemoHeroBanner {
  id: string;
  badgeId: string;
  badgeEn: string;
  titlePrimaryId: string;
  titlePrimaryEn: string;
  titleSecondaryId: string;
  titleSecondaryEn: string;
  highlights: DemoBannerHighlight[];
  supervisorImage: string;
}

export interface DemoQuickAction {
  id: string;
  badge?: string;
  titleId: string;
  titleEn: string;
  subtitleId: string;
  subtitleEn: string;
  iconType: "rfq" | "cart" | "shield" | "calendar";
  href: string;
  colorScheme: "emerald" | "cyan" | "green" | "teal";
}

export interface DemoCategoryItem {
  id: string;
  slug: string;
  nameId: string;
  nameEn: string;
  image: string;
  href: string;
}

export interface DemoProductFilterState {
  searchQuery: string;
  location: string;
  minPrice: number | undefined;
  maxPrice: number | undefined;
  minOrder: string;
  isPowerSupplier: boolean;
  isVerifiedSupplier: boolean;
  isReadyStock: boolean;
  sort: "terlaris" | "price_asc" | "price_desc" | "rating" | "newest";
  page: number;
}

export interface DemoProductItem {
  id: string;
  name: string;
  description?: string;
  price: number;
  currency: string;
  unit: string;
  minOrder: string;
  capacityText?: string;
  categoryName?: string;
  categorySlug?: string;
  photos: string[];
  supplierId: string;
  supplierCompanyName: string;
  supplierSlug: string;
  supplierLocation: string;
  supplierVerified: boolean;
  isPowerSupplier: boolean;
  supplierRating: number;
  supplierReviewCount: number;
  tags?: string[];
}
