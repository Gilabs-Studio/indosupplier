import { useMutation } from "@tanstack/react-query";
import { onboardingService } from "../services/onboarding.service";
import type { OnboardingPayload } from "../types/onboarding.types";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import { useRouter } from "@/i18n/routing";
import { toast } from "sonner";

export function useBuyerOnboarding() {
  const { user, setUser } = useAuthStore();
  const router = useRouter();

  return useMutation({
    mutationFn: (data: OnboardingPayload) => onboardingService.submit(data),
    onSuccess: (data) => {
      if (user) {
        setUser({
          ...user,
          capabilities: {
            ...user.capabilities,
            buyer: true,
          },
          buyer_profile: {
            id: data.buyer_profile_id,
            status: data.status,
          },
        });
      }
      toast.success("Profil B2B Anda berhasil diperbarui! Selamat berbelanja.");
      router.push("/dashboard");
    },
    onError: (error) => {
      console.error("Onboarding failed:", error);
      toast.error("Gagal melengkapi profil B2B. Harap coba lagi.");
    },
  });
}
