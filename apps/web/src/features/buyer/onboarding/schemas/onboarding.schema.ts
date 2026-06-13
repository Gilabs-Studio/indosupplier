import { z } from "zod";

export const onboardingSchema = z.object({
  company_name: z
    .string()
    .min(3, "Nama perusahaan minimal 3 karakter")
    .max(100, "Nama perusahaan maksimal 100 karakter"),
  industry: z.string().min(1, "Harap pilih bidang industri"),
  phone: z
    .string()
    .min(8, "Nomor telepon minimal 8 digit")
    .max(20, "Nomor telepon maksimal 20 digit")
    .regex(/^\+?[0-9\s\-]+$/, "Format nomor telepon tidak valid"),
});

export type OnboardingFormData = z.infer<typeof onboardingSchema>;
