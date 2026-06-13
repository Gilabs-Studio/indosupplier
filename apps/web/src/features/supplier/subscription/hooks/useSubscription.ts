import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { subscriptionService } from "../services/subscription.service";
import { toast } from "sonner";

export function useBillingOverview() {
  return useQuery({
    queryKey: ["supplier-billing-overview"],
    queryFn: () => subscriptionService.getBillingOverview(),
  });
}

export function useUpgradePlan() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (planId: string) => subscriptionService.upgradePlan(planId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["supplier-billing-overview"] });
      toast.success("Subscription upgraded successfully!");
    },
    onError: (error: unknown) => {
      const axiosError = error as { response?: { data?: { message?: string } } };
      const msg = axiosError.response?.data?.message || "Failed to upgrade subscription";
      toast.error(msg);
    },
  });
}
