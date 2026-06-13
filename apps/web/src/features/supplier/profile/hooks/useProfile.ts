import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { profileService } from "../services/profile.service";
import type { UpdateProfilePayload } from "../types/profile.types";
import { toast } from "sonner";

export function useSupplierProfile() {
  return useQuery({
    queryKey: ["supplier-profile"],
    queryFn: () => profileService.getProfile(),
  });
}

export function useUpdateSupplierProfile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateProfilePayload) => profileService.updateProfile(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["supplier-profile"] });
      toast.success("Profile updated successfully!");
    },
    onError: (error: unknown) => {
      const axiosError = error as { response?: { data?: { message?: string } } };
      const msg = axiosError.response?.data?.message || "Failed to update profile";
      toast.error(msg);
    },
  });
}
