import type { Metadata } from "next";

export const SEO_BASE_URL = "https://indosuppliers.id";
export const SUPPORTED_SEO_LOCALES = ["id", "en"] as const;

type SupportedLocale = (typeof SUPPORTED_SEO_LOCALES)[number];

type LandingMetadataInput = {
  locale: string;
  path: string;
  title: string;
  description: string;
  keywords: string[];
  imageAlt: string;
  imageUrl?: string;
};

function getDefaultImageByPath(path: string): string {
  const normalized = withLeadingSlash(path).toLowerCase();

  if (normalized.includes("/crm")) return "/screenshot/pipeline.webp";
  if (normalized.includes("/sales") || normalized.includes("/quotations") || normalized.includes("/invoicing")) {
    return "/screenshot/sales-order.webp";
  }
  if (
    normalized.includes("/stock") ||
    normalized.includes("/goods-receipt") ||
    normalized.includes("/movements") ||
    normalized.includes("/purchase")
  ) {
    return "/screenshot/stock-inventory.webp";
  }
  if (
    normalized.includes("/accounting") ||
    normalized.includes("/financial-reports") ||
    normalized.includes("/reconciliation") ||
    normalized.includes("/fixed-assets") ||
    normalized.includes("/pricing")
  ) {
    return "/screenshot/profit-loss.webp";
  }
  if (
    normalized.includes("/employees") ||
    normalized.includes("/attendance") ||
    normalized.includes("/recruitment") ||
    normalized.includes("/travel-planner") ||
    normalized.includes("/evaluation")
  ) {
    return "/screenshot/salary.webp";
  }

  return "/screenshot/dashboard.webp";
}

function normalizeLocale(locale: string): SupportedLocale {
  return locale === "id" ? "id" : "en";
}

function withLeadingSlash(path: string): string {
  return path.startsWith("/") ? path : `/${path}`;
}

export function getLocalizedPath(path: string, locale: string): string {
  const normalizedLocale = normalizeLocale(locale);
  const normalizedPath = withLeadingSlash(path);

  if (normalizedPath === "/") {
    return `/${normalizedLocale}`;
  }

  return `/${normalizedLocale}${normalizedPath}`;
}

export function getLanguageAlternates(path: string): Record<string, string> {
  return {
    id: getLocalizedPath(path, "id"),
    en: getLocalizedPath(path, "en"),
    "x-default": getLocalizedPath(path, "en"),
  };
}

export function buildLandingMetadata({
  locale,
  path,
  title,
  description,
  keywords,
  imageAlt,
  imageUrl,
}: LandingMetadataInput): Metadata {
  const canonicalPath = getLocalizedPath(path, locale);
  const selectedImage = imageUrl ?? getDefaultImageByPath(path);

  return {
    title,
    description,
    keywords,
    alternates: {
      canonical: canonicalPath,
      languages: getLanguageAlternates(path),
    },
    openGraph: {
      type: "website",
      locale: locale === "id" ? "id_ID" : "en_US",
      url: `${SEO_BASE_URL}${canonicalPath}`,
      title,
      description,
      siteName: "Indosupplier",
      images: [
        {
          url: selectedImage,
          width: 1920,
          height: 1080,
          alt: imageAlt,
        },
      ],
    },
    twitter: {
      card: "summary_large_image",
      title,
      description,
      images: [selectedImage],
    },
  };
}
