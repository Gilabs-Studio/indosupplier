import { getTranslations } from "next-intl/server";

import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { FinanceSettingsNav } from "@/features/finance/settings/components/settings-nav";

export async function generateMetadata({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: "financeSettings" });
  return {
    title: t("title"),
  };
}

export default async function FinanceSettingsPage() {
  return (
    <PermissionGuard requiredPermission="account_mappings.read">
      <div className="container mx-auto py-10 w-full max-w-6xl">
        <h1 className="text-3xl font-bold tracking-tight">Finance Settings</h1>
        <p className="text-muted-foreground mt-2">
          Foundation setup for fiscal year, tax, account mapping, and opening balance.
        </p>
        <div className="mt-6">
          <FinanceSettingsNav />
        </div>
      </div>
    </PermissionGuard>
  );
}
