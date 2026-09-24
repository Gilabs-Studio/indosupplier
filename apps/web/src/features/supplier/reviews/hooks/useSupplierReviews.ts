import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import { supplierReviewsService } from "../services/reviews.service";
import { useSupplierReviewsStore } from "../stores/useSupplierReviewsStore";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import type { ReplyReviewPayload } from "../types/reviews.types";

export function useSupplierReviews() {
  const queryClient = useQueryClient();
  const t = useTranslations("supplier.reviews");
  const { isAuthenticated } = useAuthStore();
  const {
    filter,
    setFilter,
    resetFilter,
    activeReview,
    isReplyDialogOpen,
    openReplyDialog,
    closeReplyDialog,
  } = useSupplierReviewsStore();

  const reviewsQuery = useQuery({
    queryKey: ["supplier-reviews", filter],
    queryFn: () => supplierReviewsService.getReviews(filter),
    enabled: isAuthenticated,
  });

  const replyMutation = useMutation({
    mutationFn: ({ reviewId, payload }: { reviewId: string; payload: ReplyReviewPayload }) =>
      supplierReviewsService.replyReview(reviewId, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["supplier-reviews"] });
      queryClient.invalidateQueries({ queryKey: ["supplier-dashboard"] });
      toast.success(t("replySuccess"));
      closeReplyDialog();
    },
    onError: (error: unknown) => {
      const axiosError = error as {
        response?: { data?: { error?: { message?: string } } };
      };
      const errorMessage =
        axiosError.response?.data?.error?.message ||
        t("replyError") ||
        "Gagal mengirimkan balasan.";
      toast.error(errorMessage);
    },
  });

  const handleReplySubmit = (payload: ReplyReviewPayload) => {
    if (!activeReview) return;
    replyMutation.mutate({
      reviewId: activeReview.id,
      payload,
    });
  };

  return {
    reviews: reviewsQuery.data?.reviews || [],
    stats: reviewsQuery.data?.stats,
    pagination: reviewsQuery.data?.pagination,
    isLoading: reviewsQuery.isLoading,
    isFetching: reviewsQuery.isFetching,
    isError: reviewsQuery.isError,
    refetch: reviewsQuery.refetch,
    filter,
    setFilter,
    resetFilter,
    activeReview,
    isReplyDialogOpen,
    openReplyDialog,
    closeReplyDialog,
    handleReplySubmit,
    isReplying: replyMutation.isPending,
  };
}
