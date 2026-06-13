import { z } from "zod";

export const rfqSchema = z.object({
  product_name: z
    .string()
    .min(3, "Nama produk minimal 3 karakter")
    .max(100, "Nama produk maksimal 100 karakter"),
  category: z.string().min(1, "Harap pilih kategori"),
  quantity: z
    .string()
    .min(1, "Harap tentukan volume kebutuhan")
    .regex(/^\d+(\.\d+)?$/, "Volume kebutuhan harus berupa angka"),
  unit: z.string().min(1, "Harap pilih satuan"),
  target_port: z
    .string()
    .min(3, "Pelabuhan tujuan minimal 3 karakter")
    .max(100, "Pelabuhan tujuan maksimal 100 karakter"),
  description: z
    .string()
    .max(1000, "Deskripsi maksimal 1000 karakter")
    .optional()
    .or(z.literal("")),
  attachment_url: z.string().optional().or(z.literal("")),
});

export type RfqFormData = z.infer<typeof rfqSchema>;
