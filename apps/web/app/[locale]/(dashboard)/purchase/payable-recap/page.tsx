import { Suspense } from "react";
import dynamic from "next/dynamic";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { PageMotion } from "@/components/motion";
import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import PayableRecapLoading from "./loading";

const PayableRecapList = dynamic(
  () =>
    import("@/features/purchase/payable-recap/components").then((mod) => ({
      default: mod.PayableRecapList,
    })),
  { loading: () => null },
);

type GenerateMetadataProps = {
  params: { locale: string } | Promise<{ locale: string }>;
};

export async function generateMetadata({ params }: GenerateMetadataProps): Promise<Metadata> {
  const { locale } = await Promise.resolve(params);
  const t = await getTranslations({ locale, namespace: "payableRecap" });
  return {
    title: t("title") + " | SalesView",
    description: t("description"),
  };
}

export default function PurchasePayableRecapPage() {
  return (
    <PermissionGuard requiredPermission="purchase_payment.read">
      <PageMotion>
        <Suspense fallback={<PayableRecapLoading />}>
          <PayableRecapList />
        </Suspense>
      </PageMotion>
    </PermissionGuard>
  );
}
