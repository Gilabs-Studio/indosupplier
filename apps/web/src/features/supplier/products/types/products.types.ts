export interface ProductPhoto {
  id?: string;
  file_url: string;
  caption?: string;
  sort_order: number;
}

export interface Category {
  id: string;
  parent_id: string | null;
  slug: string;
  name: string;
  description: string;
}

export interface Product {
  id: string;
  supplier_profile_id: string;
  category_id: string;
  category?: Category;
  name: string;
  description: string;
  moq: string;
  starting_price: number;
  currency: string;
  capacity_text: string;
  is_featured: boolean;
  sort_order: number;
  photos: ProductPhoto[];
  created_at: string;
  updated_at: string;
}

export interface CreateProductPayload {
  category_id: string;
  name: string;
  description: string;
  moq: string;
  starting_price: number;
  currency: string;
  capacity_text: string;
  is_featured: boolean;
  sort_order: number;
  photos: ProductPhoto[];
}

export type UpdateProductPayload = CreateProductPayload;
