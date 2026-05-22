import apiClient from "@/lib/api-client";

export interface OnboardingSteps {
  company: boolean;
  outlet: boolean;
  floor_layout: boolean;
  products: boolean;
  warehouse: boolean;
  users: boolean;
  fiscal_year: boolean;
}

export interface OnboardingState {
  business_type: string;
  completed: boolean;
  steps?: OnboardingSteps;
}

export const onboardingService = {
  async getState(): Promise<OnboardingState> {
    const response = await apiClient.get<{ data: OnboardingState }>(
      "/general/onboarding",
    );
    return response.data.data;
  },

  async setBusinessType(businessType: string): Promise<OnboardingState> {
    const response = await apiClient.put<{ data: OnboardingState }>(
      "/general/onboarding/business-type",
      { business_type: businessType },
    );
    return response.data.data;
  },

  async markComplete(): Promise<OnboardingState> {
    const response = await apiClient.put<{ data: OnboardingState }>(
      "/general/onboarding/complete",
      {},
    );
    return response.data.data;
  },
};
