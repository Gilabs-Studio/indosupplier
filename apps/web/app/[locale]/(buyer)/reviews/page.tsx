import { Suspense } from "react";
import { BuyerReviewsPage } from "@/features/buyer/reviews/components/buyer-reviews-page";

export default function ReviewsPage() {
  return (
    <Suspense fallback={null}>
      <BuyerReviewsPage />
    </Suspense>
  );
}
