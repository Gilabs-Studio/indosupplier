// Auth types aligned with new API

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  success: boolean;
  data: {
    user: User;
    access_token: string; // Empty in strict mode (HttpOnly cookies)
    refresh_token: string; // Empty in strict mode (HttpOnly cookies)
  };
}

export interface User {
  id: string;
  name: string;
  email: string;
  avatar_url: string;
  employee_id?: string;
  role: Role;
  permissions: Record<string, string>; // code -> scope (OWN|DIVISION|AREA|ALL)
  tenant_id?: string;
  tenant_name?: string;
  subscription_plan?: string;
  subscription_access?: SubscriptionAccess;
}

export interface SubscriptionAccess {
  state: "active" | "grace_period" | "suspended";
  enforcement: "full_access" | "hard_lock";
  days_overdue: number;
  grace_period_days: number;
  force_billing_redirect: boolean;
  allow_read: boolean;
  allow_write: boolean;
  message?: string;
  billing_path?: string;
}

export interface Role {
  code: string;
  name: string;
  data_scope: "ALL" | "DIVISION" | "AREA" | "OUTLET" | "OWN";
  /** True when this is the unique, protected tenant-owner role created at registration.
   *  Use this flag to gate owner-only UI (Billing, Payment) instead of string-matching on code. */
  is_owner?: boolean;
}

export interface Menu {
  id: number;
  name: string;
  icon: string;
  url: string;
  children?: Menu[];
  actions?: MenuAction[];
}

export interface MenuAction {
  id: number;
  code: string;
  name: string;
  action?: string;
  access: boolean;
}

export interface MenusResponse {
  success: boolean;
  data: {
    menus: Menu[];
  };
}

export interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
}
