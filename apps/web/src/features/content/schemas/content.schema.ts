import { z } from "zod";

export const contentArticleSchema = z.object({
  type: z.enum(["news", "feature", "tips", "editorial_review", "video"]),
  locale: z.enum(["id", "en", "both"]),
  title: z.string().min(3),
  slug: z.string().optional(),
  excerpt: z.string().optional(),
  body: z.string().optional(),
  authorName: z.string().optional(),
  imageUrl: z.string().url().optional().or(z.literal("")),
  videoUrl: z.string().url().optional().or(z.literal("")),
  duration: z.string().optional(),
  viewCount: z.number().int().min(0).optional(),
  supplierProfileId: z.string().uuid().optional().or(z.literal("")),
  supplierProductId: z.string().uuid().optional().or(z.literal("")),
  status: z.enum(["draft", "published", "archived"]),
  isFeatured: z.boolean(),
  sortOrder: z.number().int(),
});

export type ContentArticleFormValues = z.infer<typeof contentArticleSchema>;
