import { Suspense } from "react";
import { BuyerComparePage } from "@/features/buyer/compare/components/buyer-compare-page";
import { PublicLayout } from "@/features/public/components/public-layout";
import { CenteredLoading } from "@/components/loading";

interface ComparePageProps {
  params: Promise<{ locale: string }>;
}

export default async function ComparePage({ params }: ComparePageProps) {
  const { locale } = await params;

  return (
    <Suspense
      fallback={
        <PublicLayout locale={locale}>
          <div className="w-full min-h-[60vh] flex items-center justify-center">
            <CenteredLoading />
          </div>
        </PublicLayout>
      }
    >
      <BuyerComparePage />
    </Suspense>
  );
}
