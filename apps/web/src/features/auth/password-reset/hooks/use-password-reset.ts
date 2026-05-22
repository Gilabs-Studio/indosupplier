import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import { passwordResetService } from "../services/password-reset-service";
import type { ForgotPasswordFormData, ResetPasswordFormData } from "../schemas/password-reset.schema";

type ApiLikeError = {
  response?: {
    data?: {
      error?: {
        message?: string;
      };
    };
  };
};

function getErrorMessage(error: unknown): string | undefined {
  return (error as ApiLikeError)?.response?.data?.error?.message;
}

export function useForgotPassword() {
  const t = useTranslations("passwordReset");

  return useMutation({
    mutationFn: async (data: ForgotPasswordFormData) => passwordResetService.forgotPassword(data),
    onSuccess: () => {
      toast.success(t("forgotPasswordSuccess") || "Password reset link sent to your email");
    },
    onError: (error: unknown) => {
      const errorMessage = getErrorMessage(error) || t("forgotPasswordError") || "Failed to process forgot password request";
      toast.error(errorMessage);
    },
  });
}

export function useResetPassword() {
  const t = useTranslations("passwordReset");

  return useMutation({
    mutationFn: async (data: ResetPasswordFormData & { token: string }) => {
      return passwordResetService.resetPassword({
        token: data.token,
        new_password: data.new_password,
        confirm_password: data.confirm_password,
      });
    },
    onSuccess: () => {
      toast.success(t("resetPasswordSuccess") || "Password reset successfully");
    },
    onError: (error: unknown) => {
      const errorMessage = getErrorMessage(error) || t("resetPasswordError") || "Failed to reset password";
      toast.error(errorMessage);
    },
  });
}

export function useValidateResetToken() {
  return useMutation({
    mutationFn: async (token: string) => passwordResetService.validateToken(token),
    onError: () => {
      // Token is invalid or expired, UI will handle this
    },
  });
}
