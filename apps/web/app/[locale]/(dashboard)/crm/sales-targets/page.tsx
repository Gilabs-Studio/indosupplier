import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { PageMotion } from "@/components/motion";
import { SalesTargetsDualUI } from "@/features/crm/sales-targets/components/sales-targets-dual-ui";
import SalesTargetsLoading from "./loading";

export async function generateMetadata({
  params: { locale },
}: {
  params: { locale: string };
}) {
  const t = await getTranslations({ locale, namespace: "salesTargets" });

  return {
    title: `${t("title")} | CRM`,
    description: t("subtitle"),
  };
}

export default function SalesTargetsPage() {
  return (
    <PermissionGuard requiredPermission="sales_target.read">
      <PageMotion>
        <Suspense fallback={<SalesTargetsLoading />}>
          <SalesTargetsDualUI />
        </Suspense>
      </PageMotion>
    </PermissionGuard>
  );
}
