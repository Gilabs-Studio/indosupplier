import { routing } from "@/i18n/routing";
import type { Locale } from "@/types/locale";
import {
  getLocalePreferenceFromBrowser,
  getLocalePreferenceFromBrowserPath,
} from "@/lib/locale-preference";

/**
 * Extract locale from pathname or return default locale
 * Handles paths with and without locale prefix
 * 
 * @param pathname - The pathname to extract locale from (e.g., "/en/dashboard" or "/dashboard")
 * @returns The locale string ("en" or "id")
 */
export function getLocaleFromPathname(pathname: string): Locale {
  const pathLocale = getLocalePreferenceFromBrowserPath(pathname);
  if (pathLocale && routing.locales.includes(pathLocale)) {
    return pathLocale;
  }

  const savedLocale = getLocalePreferenceFromBrowser();
  if (savedLocale) {
    return savedLocale;
  }

  // Default to configured default locale
  return routing.defaultLocale;
}

