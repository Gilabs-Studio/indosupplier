import { getRequestConfig } from "next-intl/server";

import type { Locale } from "@/types/locale";

import { routing } from "./routing";

// Feature-level English translations
import authEn from "@/features/auth/i18n/en.json";
import indosupplierEn from "@/features/indosupplier/i18n/en.json";
import waitingListEn from "@/features/waiting-list/i18n/en.json";
import publicSearchEn from "@/features/public/search/i18n/en.json";
import publicCategoryEn from "@/features/public/category/i18n/en.json";
import publicSupplierEn from "@/features/public/supplier-profile/i18n/en.json";
import publicRegisterEn from "@/features/public/registration/i18n/en.json";
import publicHelpEn from "@/features/public/help/i18n/en.json";
import publicFaqEn from "@/features/public/faq/i18n/en.json";
import publicDemoEn from "@/features/public/demo/i18n/en.json";
import buyerEn from "@/features/buyer/i18n/en.json";

// Feature-level Indonesian translations
import authId from "@/features/auth/i18n/id.json";
import indosupplierId from "@/features/indosupplier/i18n/id.json";
import waitingListId from "@/features/waiting-list/i18n/id.json";
import publicSearchId from "@/features/public/search/i18n/id.json";
import publicCategoryId from "@/features/public/category/i18n/id.json";
import publicSupplierId from "@/features/public/supplier-profile/i18n/id.json";
import publicRegisterId from "@/features/public/registration/i18n/id.json";
import publicHelpId from "@/features/public/help/i18n/id.json";
import publicFaqId from "@/features/public/faq/i18n/id.json";
import publicDemoId from "@/features/public/demo/i18n/id.json";
import buyerId from "@/features/buyer/i18n/id.json";

const messages = {
  en: {
    ...authEn,
    ...indosupplierEn,
    ...waitingListEn,
    ...buyerEn,
    public: {
      ...publicSearchEn.public,
      ...publicCategoryEn.public,
      ...publicSupplierEn.public,
      ...publicRegisterEn.public,
      ...publicHelpEn.public,
      ...publicFaqEn.public,
      ...publicDemoEn.public,
    },
  },
  id: {
    ...authId,
    ...indosupplierId,
    ...waitingListId,
    ...buyerId,
    public: {
      ...publicSearchId.public,
      ...publicCategoryId.public,
      ...publicSupplierId.public,
      ...publicRegisterId.public,
      ...publicHelpId.public,
      ...publicFaqId.public,
      ...publicDemoId.public,
    },
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
