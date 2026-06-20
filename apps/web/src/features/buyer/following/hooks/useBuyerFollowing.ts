import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { toast } from "sonner";

import { useAuthStore } from "@/features/auth/stores/use-auth-store";

import { followingService } from "../services/following.service";

export function useBuyerFollowing() {
  const queryClient = useQueryClient();
  const { isAuthenticated } = useAuthStore();
  const t = useTranslations("buyer.following");

  const followingQuery = useQuery({
    queryKey: ["buyer-following"],
    queryFn: () => followingService.getFollowingSuppliers(),
    enabled: isAuthenticated,
  });

  const followMutation = useMutation({
    mutationFn: (supplierProfileId: string) => followingService.followSupplier(supplierProfileId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["buyer-following"] });
      toast.success(t("toastFollowSuccess"));
    },
    onError: (error) => {
      console.error(error);
      toast.error(t("toastFollowError"));
    },
  });

  const unfollowMutation = useMutation({
    mutationFn: (supplierProfileId: string) => followingService.unfollowSupplier(supplierProfileId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["buyer-following"] });
      toast.success(t("toastUnfollowSuccess"));
    },
    onError: (error) => {
      console.error(error);
      toast.error(t("toastUnfollowError"));
    },
  });

  const following = followingQuery.data || [];

  return {
    following,
    isLoading: followingQuery.isLoading && isAuthenticated,
    isError: followingQuery.isError,
    followSupplier: followMutation.mutate,
    unfollowSupplier: unfollowMutation.mutate,
    isFollowingSupplier: (supplierProfileId: string) =>
      following.some((item) => item.supplierProfileId === supplierProfileId),
    isMutating: followMutation.isPending || unfollowMutation.isPending,
  };
}
