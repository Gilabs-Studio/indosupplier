import { getRequestConfig } from "next-intl/server";

import type { Locale } from "@/types/locale";

import { routing } from "./routing";

// Feature-level English translations
import authEn from "@/features/auth/i18n/en.json";
import waitingListEn from "@/features/sysadmin/waiting-list/i18n/en.json";
import publicSearchEn from "@/features/public/search/i18n/en.json";
import publicCategoryEn from "@/features/public/category/i18n/en.json";
import publicSupplierEn from "@/features/public/supplier-profile/i18n/en.json";
import publicRegisterEn from "@/features/public/registration/i18n/en.json";
import publicHelpEn from "@/features/public/help/i18n/en.json";
import publicFaqEn from "@/features/public/faq/i18n/en.json";
import publicDemoEn from "@/features/public/demo/i18n/en.json";
import buyerEn from "@/features/buyer/i18n/en.json";
import supplierLayoutEn from "@/features/supplier/layout/i18n/en.json";
import supplierDashboardEn from "@/features/supplier/dashboard/i18n/en.json";
import supplierRfqEn from "@/features/supplier/rfq/i18n/en.json";
import supplierAdsEn from "@/features/supplier/ads/i18n/en.json";
import supplierAuctionEn from "@/features/supplier/auction/i18n/en.json";
import supplierProfileEn from "@/features/supplier/profile/i18n/en.json";
import supplierBillingEn from "@/features/supplier/billing/i18n/en.json";
import supplierSubscriptionEn from "@/features/supplier/subscription/i18n/en.json";
import supplierVerificationEn from "@/features/supplier/verification/i18n/en.json";
import supplierSupportEn from "@/features/supplier/support/i18n/en.json";
import supplierReviewsEn from "@/features/supplier/reviews/i18n/en.json";
import supplierNotificationsEn from "@/features/supplier/notifications/i18n/en.json";
import supplierOnboardingEn from "@/features/supplier/onboarding/i18n/en.json";

// Sysadmin English translations
import faqEn from "@/features/sysadmin/faq/i18n/en.json";
import reviewsEn from "@/features/sysadmin/reviews/i18n/en.json";
import subscriptionPlansEn from "@/features/sysadmin/subscription-plans/i18n/en.json";
import suppliersEn from "@/features/sysadmin/suppliers/i18n/en.json";
import supportEn from "@/features/sysadmin/support/i18n/en.json";
import dashboardEn from "@/features/sysadmin/dashboard/i18n/en.json";
import waitingListAdminEn from "@/features/sysadmin/waiting-list/i18n/admin-en.json";
import buyersEn from "@/features/sysadmin/buyers/i18n/en.json";
import adsEn from "@/features/sysadmin/ads/i18n/en.json";
import categoriesEn from "@/features/sysadmin/categories/i18n/en.json";
import abuseReportsEn from "@/features/sysadmin/abuse-reports/i18n/en.json";
import auditLogsEn from "@/features/sysadmin/audit-logs/i18n/en.json";
import auctionsEn from "@/features/sysadmin/auctions/i18n/en.json";

// Feature-level Indonesian translations
import authId from "@/features/auth/i18n/id.json";
import waitingListId from "@/features/sysadmin/waiting-list/i18n/id.json";
import publicSearchId from "@/features/public/search/i18n/id.json";
import publicCategoryId from "@/features/public/category/i18n/id.json";
import publicSupplierId from "@/features/public/supplier-profile/i18n/id.json";
import publicRegisterId from "@/features/public/registration/i18n/id.json";
import publicHelpId from "@/features/public/help/i18n/id.json";
import publicFaqId from "@/features/public/faq/i18n/id.json";
import publicDemoId from "@/features/public/demo/i18n/id.json";
import buyerId from "@/features/buyer/i18n/id.json";
import supplierLayoutId from "@/features/supplier/layout/i18n/id.json";
import supplierDashboardId from "@/features/supplier/dashboard/i18n/id.json";
import supplierRfqId from "@/features/supplier/rfq/i18n/id.json";
import supplierAdsId from "@/features/supplier/ads/i18n/id.json";
import supplierAuctionId from "@/features/supplier/auction/i18n/id.json";
import supplierProfileId from "@/features/supplier/profile/i18n/id.json";
import supplierBillingId from "@/features/supplier/billing/i18n/id.json";
import supplierSubscriptionId from "@/features/supplier/subscription/i18n/id.json";
import supplierVerificationId from "@/features/supplier/verification/i18n/id.json";
import supplierSupportId from "@/features/supplier/support/i18n/id.json";
import supplierReviewsId from "@/features/supplier/reviews/i18n/id.json";
import supplierNotificationsId from "@/features/supplier/notifications/i18n/id.json";
import supplierOnboardingId from "@/features/supplier/onboarding/i18n/id.json";

// Sysadmin Indonesian translations
import faqId from "@/features/sysadmin/faq/i18n/id.json";
import reviewsId from "@/features/sysadmin/reviews/i18n/id.json";
import subscriptionPlansId from "@/features/sysadmin/subscription-plans/i18n/id.json";
import suppliersId from "@/features/sysadmin/suppliers/i18n/id.json";
import supportId from "@/features/sysadmin/support/i18n/id.json";
import dashboardId from "@/features/sysadmin/dashboard/i18n/id.json";
import waitingListAdminId from "@/features/sysadmin/waiting-list/i18n/admin-id.json";
import buyersId from "@/features/sysadmin/buyers/i18n/id.json";
import adsId from "@/features/sysadmin/ads/i18n/id.json";
import categoriesId from "@/features/sysadmin/categories/i18n/id.json";
import abuseReportsId from "@/features/sysadmin/abuse-reports/i18n/id.json";
import auditLogsId from "@/features/sysadmin/audit-logs/i18n/id.json";
import auctionsId from "@/features/sysadmin/auctions/i18n/id.json";

const messages = {
  en: {
    ...authEn,
    ...waitingListEn,
    ...buyerEn,
    sysadminFaq: faqEn,
    sysadminReviews: reviewsEn,
    sysadminSubscriptionPlans: subscriptionPlansEn,
    sysadminSuppliers: suppliersEn,
    sysadminSupport: supportEn,
    sysadminDashboard: dashboardEn,
    sysadminWaitingList: waitingListAdminEn,
    sysadminBuyers: buyersEn,
    sysadminAds: adsEn,
    sysadminCategories: categoriesEn,
    sysadminAbuseReports: abuseReportsEn,
    sysadminAuditLogs: auditLogsEn,
    sysadminAuctions: auctionsEn,
    supplier: {
      ...supplierLayoutEn,
      ...supplierDashboardEn,
      ...supplierRfqEn,
      ...supplierAdsEn,
      ...supplierAuctionEn,
      ...supplierProfileEn,
      ...supplierBillingEn,
      ...supplierSubscriptionEn,
      ...supplierVerificationEn,
      ...supplierSupportEn,
      ...supplierReviewsEn,
      ...supplierNotificationsEn,
      ...supplierOnboardingEn
    },
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
    ...waitingListId,
    ...buyerId,
    sysadminFaq: faqId,
    sysadminReviews: reviewsId,
    sysadminSubscriptionPlans: subscriptionPlansId,
    sysadminSuppliers: suppliersId,
    sysadminSupport: supportId,
    sysadminDashboard: dashboardId,
    sysadminWaitingList: waitingListAdminId,
    sysadminBuyers: buyersId,
    sysadminAds: adsId,
    sysadminCategories: categoriesId,
    sysadminAbuseReports: abuseReportsId,
    sysadminAuditLogs: auditLogsId,
    sysadminAuctions: auctionsId,
    supplier: {
      ...supplierLayoutId,
      ...supplierDashboardId,
      ...supplierRfqId,
      ...supplierAdsId,
      ...supplierAuctionId,
      ...supplierProfileId,
      ...supplierBillingId,
      ...supplierSubscriptionId,
      ...supplierVerificationId,
      ...supplierSupportId,
      ...supplierReviewsId,
      ...supplierNotificationsId,
      ...supplierOnboardingId
    },
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
