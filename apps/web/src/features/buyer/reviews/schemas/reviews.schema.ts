import { z } from "zod";

export const reviewFormSchema = z.object({
  purchaseOrderId: z.string().uuid({ message: "ID transaksi tidak valid" }),
  rating: z.number().min(1, "Rating minimal 1").max(5, "Rating maksimal 5"),
  reviewText: z.string().min(3, "Ulasan minimal 3 karakter").max(1000, "Ulasan terlalu panjang"),
});

export type ReviewFormInput = z.infer<typeof reviewFormSchema>;
