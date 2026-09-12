import { z } from "zod";

export const photoSchema = z.object({
  id: z.string().optional(),
  file_url: z.string().min(1, "Invalid image URL"),
  caption: z.string().max(255).optional(),
  sort_order: z.number().int(),
});

export const SUPPORTED_CURRENCIES = ["IDR", "USD", "SGD", "EUR", "CNY"] as const;
export type SupportedCurrency = (typeof SUPPORTED_CURRENCIES)[number];

export const CURRENCY_OPTIONS: { value: SupportedCurrency; label: string; symbol: string }[] = [
  { value: "IDR", label: "IDR (Rp) — Rupiah Indonesia", symbol: "Rp" },
  { value: "USD", label: "USD ($) — US Dollar", symbol: "$" },
  { value: "SGD", label: "SGD (S$) — Singapore Dollar", symbol: "S$" },
  { value: "EUR", label: "EUR (€) — Euro", symbol: "€" },
  { value: "CNY", label: "CNY (¥) — Chinese Yuan", symbol: "¥" },
];

export const productFormSchema = z.object({
  name: z
    .string()
    .min(3, "Product name must be at least 3 characters")
    .max(255, "Product name cannot exceed 255 characters"),
  category_id: z.string().uuid("Please select a category"),
  description: z.string().min(10, "Description must be at least 10 characters"),
  moq: z.string().min(1, "Minimum Order Quantity is required"),
  starting_price: z.number().min(0, "Price must be greater than or equal to 0"),
  currency: z.enum(SUPPORTED_CURRENCIES),
  capacity_text: z.string().min(1, "Supply capacity is required"),
  is_featured: z.boolean(),
  sort_order: z.number().int(),
  photos: z.array(photoSchema).min(1, "At least one product image is required"),
});

export type ProductFormValues = z.infer<typeof productFormSchema>;
