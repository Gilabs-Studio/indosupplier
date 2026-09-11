import { z } from "zod";

const currentYear = new Date().getFullYear();

export const supplierProfileSchema = z.object({
  companyName: z
    .string()
    .min(2, "Nama perusahaan minimal 2 karakter")
    .max(255, "Nama perusahaan maksimal 255 karakter"),
  businessType: z
    .string()
    .min(2, "Jenis usaha minimal 2 karakter")
    .max(80, "Jenis usaha maksimal 80 karakter"),
  established: z
    .string()
    .max(80, "Tahun berdiri maksimal 80 karakter")
    .refine(
      (val) => {
        if (!val || val.trim() === "") return true;
        const clean = val.trim();
        if (!/^\d{4}$/.test(clean)) return false;
        const year = parseInt(clean, 10);
        return year >= 1800 && year <= currentYear + 1;
      },
      {
        message: `Tahun berdiri harus berupa 4 digit angka tahun yang valid (1800 - ${currentYear + 1})`,
      }
    ),
  employees: z.string().max(80, "Jumlah karyawan maksimal 80 karakter"),
  email: z
    .string()
    .min(1, "Email wajib diisi")
    .email("Format email tidak valid")
    .max(255, "Email maksimal 255 karakter"),
  phone: z
    .string()
    .min(6, "Nomor telepon minimal 6 digit")
    .max(30, "Nomor telepon maksimal 30 karakter")
    .regex(
      /^(\+?[0-9\s\-()]+)$/,
      "Format nomor telepon harus berupa angka yang valid (contoh: +6281234567890)"
    ),
  website: z
    .string()
    .max(255, "URL situs web maksimal 255 karakter")
    .refine(
      (val) => !val || /^https?:\/\/.+/i.test(val),
      { message: "URL situs web harus diawali dengan http:// atau https://" }
    ),
  taxId: z
    .string()
    .max(80, "NPWP maksimal 80 karakter")
    .refine(
      (val) => !val || /^[\d.\-]+$/.test(val.trim()),
      { message: "NPWP harus berupa kombinasi angka dan pemisah (contoh: 01.234.567.8-901.000)" }
    ),
  nib: z
    .string()
    .max(80, "NIB maksimal 80 karakter")
    .refine(
      (val) => !val || /^\d{9,16}$/.test(val.replace(/[\s\-]/g, "")),
      { message: "NIB harus berupa kode angka (9-16 digit angka)" }
    ),
  overview: z.string().max(5000, "Ikhtisar maksimal 5000 karakter"),
  location: z
    .string()
    .min(3, "Lokasi/Alamat minimal 3 karakter")
    .max(500, "Lokasi/Alamat maksimal 500 karakter"),
  logo: z.string().max(1000).optional(),
});

export type SupplierProfileFormValues = z.infer<typeof supplierProfileSchema>;
