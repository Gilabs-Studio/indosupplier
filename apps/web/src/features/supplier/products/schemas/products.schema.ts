import { z } from "zod";

export const photoSchema = z.object({
  id: z.string().optional(),
  file_url: z.string().url("Invalid image URL"),
  caption: z.string().max(255).optional(),
  sort_order: z.number().int().default(0),
});

export const productFormSchema = z.object({
  name: z.string().min(3, "Product name must be at least 3 characters").max(255, "Product name cannot exceed 255 characters"),
  category_id: z.string().uuid("Please select a category"),
  description: z.string().min(10, "Description must be at least 10 characters"),
  moq: z.string().min(1, "Minimum Order Quantity is required"),
  starting_price: z.preprocess(
    (val) => (val === "" || val === undefined ? 0 : Number(val)),
    z.number().min(0, "Price must be greater than or equal to 0")
  ),
  currency: z.string().default("IDR"),
  capacity_text: z.string().min(1, "Supply capacity is required"),
  is_featured: z.boolean().default(false),
  sort_order: z.number().int().default(0),
  photos: z.array(photoSchema).min(1, "At least one product image is required"),
});

export type ProductFormValues = z.infer<typeof productFormSchema>;
