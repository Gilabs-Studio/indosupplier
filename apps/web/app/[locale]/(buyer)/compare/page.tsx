import { Suspense } from "react";
import { BuyerComparePage } from "@/features/buyer/compare/components/buyer-compare-page";
import { BuyerLayout } from "@/features/buyer/components/buyer-layout";
import { CenteredLoading } from "@/components/loading";

export default function ComparePage() {
  return (
    <Suspense
      fallback={
        <BuyerLayout>
          <CenteredLoading />
        </BuyerLayout>
      }
    >
      <BuyerComparePage />
    </Suspense>
  );
}
