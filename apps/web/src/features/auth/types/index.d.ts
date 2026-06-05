// Auth types aligned with new API

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
  company_name?: string;
  industry?: string;
}

export interface SupplierOnboardingRequest {
  company_name: string;
  primary_category: string;
  subcategory: string;
  province_id: string;
  city_id: string;
  phone: string;
  whatsapp: string;
  email: string;
  website?: string;
  company_type: string;
  tax_status: string;
  npwp?: string;
  nib?: string;
  address?: string;
  business_hours?: string;
  timezone?: string;
  description: string;
  first_product_name?: string;
  first_product_price?: string;
}

export interface LoginResponse {
  success: boolean;
  data: {
    user: User;
    access_token: string; // Empty in strict mode (HttpOnly cookies)
    refresh_token: string; // Empty in strict mode (HttpOnly cookies)
  };
}

export interface UserResponse {
  success: boolean;
  data: {
    user: User;
  };
}

export interface User {
  id: string;
  name: string;
  email: string;
  capabilities: AccountCapabilities;
  buyer_profile?: AccountProfileRef | null;
  supplier_profile?: AccountProfileRef | null;
}

export interface AccountCapabilities {
  buyer: boolean;
  supplier: boolean;
}

export interface AccountProfileRef {
  id: string;
  status?: string;
}

export interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
}
