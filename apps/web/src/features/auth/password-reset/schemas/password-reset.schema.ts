import { z } from "zod";
import { useTranslations } from "next-intl";

export const forgotPasswordSchema = z.object({
  email: z.string().email("Invalid email address"),
});

export const resetPasswordSchema = z.object({
  new_password: z.string().min(6, "Password must be at least 6 characters"),
  confirm_password: z.string().min(6, "Password must be at least 6 characters"),
}).refine((data) => data.new_password === data.confirm_password, {
  message: "Passwords do not match",
  path: ["confirm_password"],
});

export type ForgotPasswordFormData = z.infer<typeof forgotPasswordSchema>;
export type ResetPasswordFormData = z.infer<typeof resetPasswordSchema>;

// Helper function to get validated schema with i18n messages
export const getForgotPasswordSchema = (t: ReturnType<typeof useTranslations>) => {
  return z.object({
    email: z.string()
      .min(1, t("emailRequired"))
      .email(t("invalidEmail")),
  });
};

export const getResetPasswordSchema = (t: ReturnType<typeof useTranslations>) => {
  return z.object({
    new_password: z.string()
      .min(6, t("passwordMinLength")),
    confirm_password: z.string()
      .min(6, t("passwordMinLength")),
  }).refine((data) => data.new_password === data.confirm_password, {
    message: t("passwordsDoNotMatch"),
    path: ["confirm_password"],
  });
};
