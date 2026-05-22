import { Suspense } from "react";
import dynamic from "next/dynamic";
import { Skeleton } from "@/components/ui/skeleton";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";

const PurchasePaymentsContainer = dynamic(
  () =>
    import("@/features/purchase/payments/components").then((mod) => ({
      default: mod.PurchasePaymentsContainer,
    })),
  { loading: () => null },
);
type GenerateMetadataProps = {
  params: { locale: string } | Promise<{ locale: string }>;
};

export async function generateMetadata({ params }: GenerateMetadataProps): Promise<Metadata> {
  const { locale } = await Promise.resolve(params);
  const t = await getTranslations({ locale, namespace: "purchasePayment" });
  return {
    title: t("title") + " | SalesView",
    description: t("description"),
  };
}

function PaymentsSkeleton() {
  return (
    <div className="space-y-4">
      <Skeleton className="h-10 w-64" />
      <Skeleton className="h-10 w-80" />
      <div className="rounded-md border">
        <div className="p-4 space-y-4">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-12 w-full" />
          ))}
        </div>
      </div>
    </div>
  );
}

export default function PurchasePaymentsPage() {
  return (
    <PermissionGuard requiredPermission="purchase_payment.read">
      <Suspense fallback={<PaymentsSkeleton />}>
        <PurchasePaymentsContainer />
      </Suspense>
    </PermissionGuard>
  );
}
