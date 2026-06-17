export type ContentType = "news" | "feature" | "tips" | "editorial_review" | "video";
export type ContentLocale = "id" | "en" | "both";
export type ContentStatus = "draft" | "published" | "archived";

export interface ContentArticle {
  id: string;
  type: ContentType;
  locale: ContentLocale;
  title: string;
  slug: string;
  excerpt: string;
  body: string;
  authorName: string;
  imageUrl: string;
  videoUrl: string;
  duration: string;
  viewCount: number;
  supplierProfileId?: string;
  supplierProductId?: string;
  status: ContentStatus;
  isFeatured: boolean;
  sortOrder: number;
  publishedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface ContentListParams {
  type?: ContentType;
  locale?: "id" | "en";
  status?: ContentStatus | "all";
  search?: string;
  page?: number;
  perPage?: number;
}

export interface ContentArticlePayload {
  type: ContentType;
  locale: ContentLocale;
  title: string;
  slug?: string;
  excerpt?: string;
  body?: string;
  authorName?: string;
  imageUrl?: string;
  videoUrl?: string;
  duration?: string;
  viewCount?: number;
  supplierProfileId?: string;
  supplierProductId?: string;
  status: ContentStatus;
  isFeatured: boolean;
  sortOrder: number;
}

export interface PaginationMeta {
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

export interface ContentListResult {
  items: ContentArticle[];
  pagination?: PaginationMeta;
}
