import apiClient from "@/lib/api-client";
import type {
  ForgotPasswordRequest,
  ForgotPasswordResponse,
  ResetPasswordRequest,
  ResetPasswordResponse,
  ValidateTokenResponse,
} from "../types";

export const passwordResetService = {
  async forgotPassword(data: ForgotPasswordRequest): Promise<ForgotPasswordResponse> {
    const response = await apiClient.post<{ data: ForgotPasswordResponse }>(
      "/password-reset/forgot-password",
      data
    );
    return response.data.data;
  },

  async resetPassword(data: ResetPasswordRequest): Promise<ResetPasswordResponse> {
    const response = await apiClient.post<{ data: ResetPasswordResponse }>(
      "/password-reset/reset-password",
      data
    );
    return response.data.data;
  },

  async validateToken(token: string): Promise<ValidateTokenResponse> {
    const response = await apiClient.get<{ data: ValidateTokenResponse }>(
      `/password-reset/validate-token?token=${encodeURIComponent(token)}`
    );
    return response.data.data;
  },
};
