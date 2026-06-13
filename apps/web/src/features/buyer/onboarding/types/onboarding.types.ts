export interface OnboardingPayload {
  company_name: string;
  industry: string;
  phone: string;
}

export interface OnboardingResponse {
  success: boolean;
  data: {
    buyer_profile_id: string;
    status: string;
  };
}
