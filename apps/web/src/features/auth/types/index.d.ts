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
}

export interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
}
