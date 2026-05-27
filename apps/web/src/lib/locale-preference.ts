import { routing } from "@/i18n/routing";
import type { Locale } from "@/types/locale";

export const LOCALE_PREFERENCE_COOKIE = "indosupplier_locale_pref";
export const LOCALE_PREFERENCE_STORAGE_KEY = "locale";

export function normalizeLocale(value: string | null | undefined): Locale | null {
  if (value === "id" || value === "en") {
    return value;
  }

  return null;
}

function readCookieValue(cookieString: string | null | undefined, name: string): string | null {
  if (!cookieString) {
    return null;
  }

  const parts = cookieString.split("; ");

  for (const part of parts) {
    const separatorIndex = part.indexOf("=");

    if (separatorIndex === -1) {
      continue;
    }

    const cookieName = part.slice(0, separatorIndex);

    if (cookieName === name) {
      return part.slice(separatorIndex + 1) || null;
    }
  }

  return null;
}

export function getLocalePreferenceFromCookieString(cookieString: string | null | undefined): Locale | null {
  return normalizeLocale(readCookieValue(cookieString, LOCALE_PREFERENCE_COOKIE));
}

export function getLocalePreferenceFromCookieValue(value: string | null | undefined): Locale | null {
  return normalizeLocale(value);
}

export function getLocalePreferenceFromBrowser(): Locale | null {
  if (typeof window === "undefined") {
    return null;
  }

  const cookiePreference = getLocalePreferenceFromCookieString(globalThis.document.cookie);
  if (cookiePreference) {
    return cookiePreference;
  }

  const savedLocale = globalThis.localStorage.getItem(LOCALE_PREFERENCE_STORAGE_KEY);
  return normalizeLocale(savedLocale);
}

export function getLocaleFromCountryCode(countryCode: string | null | undefined): Locale {
  const normalizedCountryCode = countryCode?.trim().toUpperCase();

  if (normalizedCountryCode === "ID") {
    return "id";
  }

  return routing.defaultLocale;
}

export function getLocalePreferenceFromBrowserPath(pathname: string): Locale | null {
  const localeMatch = pathname.match(/^\/(en|id)(\/|$)/);

  if (!localeMatch) {
    return null;
  }

  return normalizeLocale(localeMatch[1]);
}
