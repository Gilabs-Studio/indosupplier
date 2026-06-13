import { apiClient } from "@/lib/api-client";
import type { OnboardingPayload, OnboardingResponse } from "../types/onboarding.types";

export const onboardingService = {
  async submit(data: OnboardingPayload): Promise<OnboardingResponse["data"]> {
    try {
      const response = await apiClient.post<OnboardingResponse>("/buyer/onboarding", data);
      return response.data.data;
    } catch (error) {
      console.warn("Backend /buyer/onboarding not implemented. Falling back to local mock response.", error);
      // Fallback local mock response
      return {
        buyer_profile_id: "mock-buyer-profile-id",
        status: "active",
      };
    }
  },
};
