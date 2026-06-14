import { z } from "zod";

export const createTransactionSchema = z.object({
  supplier_profile_id: z.string().uuid("Invalid supplier ID"),
  rfq_id: z.string().uuid("Invalid RFQ ID").optional(),
  product_name: z.string().min(3, "Product name must be at least 3 characters"),
  quantity_value: z.number().gt(0, "Quantity must be greater than 0"),
  quantity_unit: z.string().min(1, "Unit is required"),
  price_per_unit: z.number().gt(0, "Price must be greater than 0"),
  delivery_address: z.string().min(10, "Delivery address must be at least 10 characters"),
  notes: z.string().optional(),
});

export type CreateTransactionInput = z.infer<typeof createTransactionSchema>;
