import type { MetadataRoute } from "next";

const BASE_URL = "https://salesview.id";

export default function robots(): MetadataRoute.Robots {
  return {
    rules: {
      userAgent: "*",
      allow: "/",
      disallow: ["/api/", "/id/dashboard", "/en/dashboard", "/id/system", "/en/system"],
    },
    sitemap: `${BASE_URL}/sitemap.xml`,
    host: BASE_URL,
  };
}
