import { getTranslations } from "next-intl/server";

import { PageMotion } from "@/components/motion";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { TaxConfigPanel } from "@/features/finance/settings/components/tax-config-panel";

export async function generateMetadata({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: "financeSettings" });
  return {
    title: `${t("title")} - Tax Configuration`,
  };
}

export default async function FinanceSettingsTaxConfigPage() {
  return (
    <PermissionGuard requiredPermission="tax_configuration.read">
      <PageMotion className="container mx-auto py-10 w-full max-w-6xl">
        <h1 className="text-3xl font-bold tracking-tight">Tax Configuration</h1>
        <p className="text-muted-foreground mt-2">
          Configure PPN and PPh mappings for automatic journal posting.
        </p>
        <div className="mt-6">
          <TaxConfigPanel />
        </div>
      </PageMotion>
    </PermissionGuard>
  );
}
