export interface FAQArticle {
  id: string;
  title: string;
  slug: string;
  topic: string;
  language: "id" | "en" | "both";
  content: string;
  active: boolean;
  createdAt: string;
  updatedAt: string;
}
