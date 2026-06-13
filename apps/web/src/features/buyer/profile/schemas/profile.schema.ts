import { z } from "zod";

export const personalProfileSchema = z.object({
  full_name: z
    .string()
    .min(3, "Nama lengkap minimal 3 karakter")
    .max(100, "Nama lengkap maksimal 100 karakter"),
  phone: z
    .string()
    .min(8, "Nomor telepon minimal 8 digit")
    .max(20, "Nomor telepon maksimal 20 digit")
    .regex(/^\+?[0-9\s\-]+$/, "Format nomor telepon tidak valid"),
});

export const companyProfileSchema = z.object({
  company_name: z
    .string()
    .min(3, "Nama perusahaan minimal 3 karakter")
    .max(100, "Nama perusahaan maksimal 100 karakter"),
  industry: z.string().min(1, "Harap pilih bidang industri"),
  website: z
    .string()
    .url("Format URL website tidak valid")
    .optional()
    .or(z.literal("")),
  address: z
    .string()
    .min(5, "Alamat minimal 5 karakter")
    .max(500, "Alamat maksimal 500 karakter"),
});

export type PersonalProfileFormData = z.infer<typeof personalProfileSchema>;
export type CompanyProfileFormData = z.infer<typeof companyProfileSchema>;
