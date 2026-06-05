import { MetadataRoute } from "next";

export default function sitemap(): MetadataRoute.Sitemap {
  const baseUrl = "https://indosupplier.id";
  const locales = ["id", "en"] as const;
  const lastModified = new Date();

  const homepageEntries = locales.map((locale) => ({
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

  // Feature anchor-based URLs for important sections that can be
  // independently linked or shared (e.g., /id#features, /id#join)
  const sectionEntries = locales.flatMap((locale) => [
    {
      url: `${baseUrl}/${locale}#features`,
      lastModified,
      changeFrequency: "monthly" as const,
      priority: 0.8,
    },
    {
      url: `${baseUrl}/${locale}#about`,
      lastModified,
      changeFrequency: "monthly" as const,
      priority: 0.7,
    },
    {
      url: `${baseUrl}/${locale}#join`,
      lastModified,
      changeFrequency: "weekly" as const,
      priority: 0.9,
    },
  ]);

  return [...homepageEntries, ...sectionEntries];
}
