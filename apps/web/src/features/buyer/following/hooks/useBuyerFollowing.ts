import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { useAuthStore } from "@/features/auth/stores/use-auth-store";

import { followingService } from "../services/following.service";

export function useBuyerFollowing() {
  const queryClient = useQueryClient();
  const { isAuthenticated } = useAuthStore();

  const followingQuery = useQuery({
    queryKey: ["buyer-following"],
    queryFn: () => followingService.getFollowingSuppliers(),
    enabled: isAuthenticated,
  });

  const followMutation = useMutation({
    mutationFn: (supplierProfileId: string) => followingService.followSupplier(supplierProfileId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["buyer-following"] });
      toast.success("Supplier berhasil diikuti.");
    },
    onError: (error) => {
      console.error(error);
      toast.error("Gagal mengikuti supplier.");
    },
  });

  const unfollowMutation = useMutation({
    mutationFn: (supplierProfileId: string) => followingService.unfollowSupplier(supplierProfileId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["buyer-following"] });
      toast.success("Supplier tidak lagi diikuti.");
    },
    onError: (error) => {
      console.error(error);
      toast.error("Gagal memperbarui following supplier.");
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
