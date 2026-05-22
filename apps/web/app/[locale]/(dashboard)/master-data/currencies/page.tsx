import { getTranslations } from "next-intl/server";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { CurrencyContainer } from "@/features/master-data/currencies/components";

export async function generateMetadata() {
  const t = await getTranslations("currency");
  return {
    title: t("title"),
    description: t("description"),
  };
}

export default function CurrenciesPage() {
  return (
    <PermissionGuard requiredPermission="currency.read">
      <CurrencyContainer />
    </PermissionGuard>
  );
}