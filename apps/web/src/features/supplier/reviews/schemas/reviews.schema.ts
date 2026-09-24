import { z } from "zod";

export const replyReviewSchema = z.object({
  reply: z
    .string()
    .trim()
    .min(3, { message: "Balasan minimal 3 karakter." })
    .max(1000, { message: "Balasan maksimal 1000 karakter." }),
});

export type ReplyReviewFormValues = z.infer<typeof replyReviewSchema>;
