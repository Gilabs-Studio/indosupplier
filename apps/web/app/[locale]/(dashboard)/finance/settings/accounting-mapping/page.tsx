import { AccountingMappingForm } from "@/features/finance/settings/components/accounting-mapping-form";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { getTranslations } from "next-intl/server";

export async function generateMetadata({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: "financeSettings" });
  return {
    title: t("title"),
  };
}

export default async function AccountingMappingPage() {
  const t = await getTranslations("financeSettings");

  return (
    <PermissionGuard requiredPermission="account_mappings.read">
      <div className="container mx-auto py-10 w-full max-w-5xl">
        <div className="mb-6">
          <h1 className="text-3xl font-bold tracking-tight">{t("title")}</h1>
          <p className="text-muted-foreground mt-2">
            {t("description")}
          </p>
        </div>
        <div className="mt-8">
          <AccountingMappingForm />
        </div>
      </div>
    </PermissionGuard>
  );
}
