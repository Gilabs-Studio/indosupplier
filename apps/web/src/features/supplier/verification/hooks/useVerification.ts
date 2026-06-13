import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { verificationService } from "../services/verification.service";
import type { VerificationData } from "../types/verification.types";
import { toast } from "sonner";

export function useVerificationData() {
  return useQuery({
    queryKey: ["supplier-verification"],
    queryFn: () => verificationService.getVerificationData(),
  });
}

export function useUpdateVerificationData() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: Partial<VerificationData>) => verificationService.updateVerificationData(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["supplier-verification"] });
    },
    onError: (error: unknown) => {
      const axiosError = error as { response?: { data?: { message?: string } } };
      const msg = axiosError.response?.data?.message || "Failed to update verification details";
      toast.error(msg);
    },
  });
}

export function useSubmitVerification() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => verificationService.submitVerification(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["supplier-verification"] });
      toast.success("Business verification submitted successfully!");
      // Reload page to update the layout warning banner immediately
      setTimeout(() => {
        window.location.reload();
      }, 800);
    },
    onError: (error: unknown) => {
      const axiosError = error as { response?: { data?: { message?: string } } };
      const msg = axiosError.response?.data?.message || "Failed to submit verification";
      toast.error(msg);
    },
  });
}
