import type { AxiosError } from "axios";

export interface ApiErrorResponse {
  success: false;
  error: {
    code: string;
    message: string;
  };
  timestamp: string;
  request_id: string;
}

export type AuthError = AxiosError<ApiErrorResponse>;
