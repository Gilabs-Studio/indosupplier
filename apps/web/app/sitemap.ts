import { MetadataRoute } from "next";

export default function sitemap(): MetadataRoute.Sitemap {
  const baseUrl = "https://indosuppliers.id";
  const locales = ["id", "en"] as const;
  const lastModified = new Date();

  return locales.map((locale) => ({
    url: `${baseUrl}/${locale}`,
    lastModified,
    changeFrequency: "weekly" as const,
    priority: 1.0,
    alternates: {
      languages: {
        id: `${baseUrl}/id`,
        en: `${baseUrl}/en`,
      },
    },
  }));
}
