import type { ContentArticle, ContentType } from "../types/content.types";

export const contentTypeLabels: Record<ContentType, { id: string; en: string }> = {
  news: { id: "Berita", en: "News" },
  feature: { id: "Artikel Feature", en: "Feature" },
  tips: { id: "Tips", en: "Tips" },
  editorial_review: { id: "Review Redaksi", en: "Editorial Review" },
  video: { id: "Video", en: "Video" },
};

export function formatContentDate(article: ContentArticle, locale: string): string {
  const value = article.publishedAt || article.createdAt;
  const formatter = new Intl.DateTimeFormat(locale === "en" ? "en-US" : "id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  });
  return formatter.format(new Date(value));
}

export function formatViewCount(value: number, locale: string): string {
  const formatted = new Intl.NumberFormat(locale === "en" ? "en-US" : "id-ID", {
    notation: "compact",
    maximumFractionDigits: 1,
  }).format(value);
  return locale === "en" ? `${formatted} views` : `${formatted} tayangan`;
}
