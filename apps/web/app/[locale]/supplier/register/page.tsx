"use client";

import { useLocale } from "next-intl";
import { PublicLayout } from "@/features/public/components/public-layout";
import { SupplierOnboardingPage } from "@/features/supplier/onboarding/components/supplier-onboarding-page";

export default function SupplierRegisterPage() {
  const locale = useLocale();
  return (
    <PublicLayout locale={locale}>
      <SupplierOnboardingPage />
    </PublicLayout>
  );
}
