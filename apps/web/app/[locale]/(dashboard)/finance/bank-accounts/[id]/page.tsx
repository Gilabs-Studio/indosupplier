import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";

import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { BankAccountDetailPage } from "@/features/finance/bank-accounts/components/bank-account-detail-page";

type PageProps = {
  params: { locale: string; id: string } | Promise<{ locale: string; id: string }>;
};

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { locale } = await Promise.resolve(params);
  const t = await getTranslations({ locale, namespace: "financeBankAccounts" });

  return {
    title: `${t("detail.title")} | SalesView`,
    description: t("description"),
  };
}

export default async function FinanceBankAccountDetailRoute({ params }: PageProps) {
  const { id } = await Promise.resolve(params);

  return (
    <PermissionGuard requiredPermission="bank_account.read">
      <BankAccountDetailPage id={id} />
    </PermissionGuard>
  );
}
