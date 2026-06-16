import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { reviewsService } from "../services/reviews.service";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import { toast } from "sonner";
import type { CreateReviewPayload } from "../types/reviews.types";
import { useTranslations } from "next-intl";

export function useBuyerReviews() {
  const queryClient = useQueryClient();
  const { isAuthenticated } = useAuthStore();
  const t = useTranslations("buyerReviews");

  const eligibleQuery = useQuery({
    queryKey: ["buyer-reviews-eligible"],
    queryFn: () => reviewsService.getEligibleTransactions(),
    enabled: isAuthenticated,
  });

  const historyQuery = useQuery({
    queryKey: ["buyer-reviews-history"],
    queryFn: () => reviewsService.getReviewHistory(),
    enabled: isAuthenticated,
  });

  const submitReviewMutation = useMutation({
    mutationFn: (payload: CreateReviewPayload) => reviewsService.createReview(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["buyer-reviews-eligible"] });
      queryClient.invalidateQueries({ queryKey: ["buyer-reviews-history"] });
      toast.success(t("success"));
    },
    onError: (error: unknown) => {
      console.error(error);
      const axiosError = error as { response?: { data?: { message?: string } } };
      const errorMessage = axiosError.response?.data?.message || "Gagal mengirimkan ulasan.";
      toast.error(errorMessage);
    },
  });

  return {
    eligibleTransactions: eligibleQuery.data || [],
    isEligibleLoading: eligibleQuery.isLoading && isAuthenticated,
    isEligibleError: eligibleQuery.isError,
    reviewHistory: historyQuery.data || [],
    isHistoryLoading: historyQuery.isLoading && isAuthenticated,
    isHistoryError: historyQuery.isError,
    submitReview: submitReviewMutation.mutate,
    isSubmitting: submitReviewMutation.isPending,
    isSubmitSuccess: submitReviewMutation.isSuccess,
  };
}
