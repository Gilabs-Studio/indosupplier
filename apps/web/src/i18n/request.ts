import { getRequestConfig } from "next-intl/server";

import type { Locale } from "@/types/locale";

import { routing } from "./routing";

// Feature-level English translations
import authEn from "@/features/auth/i18n/en.json";
import waitingListEn from "@/features/waiting-list/i18n/en.json";

// Feature-level Indonesian translations
import authId from "@/features/auth/i18n/id.json";
import waitingListId from "@/features/waiting-list/i18n/id.json";

const messages = {
  en: {
    ...authEn,
    ...waitingListEn,
  },
  id: {
    ...authId,
    ...waitingListId,
  },
} as const;

export default getRequestConfig(async ({ requestLocale }) => {
  let locale = await requestLocale;

  if (!locale || !routing.locales.includes(locale as Locale)) {
    locale = routing.defaultLocale;
  }

  return {
    locale,
    messages: messages[locale as keyof typeof messages],
  };
});
