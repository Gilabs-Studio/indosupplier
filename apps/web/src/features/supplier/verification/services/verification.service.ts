import { apiClient } from "@/lib/api-client";
import type { VerificationData } from "../types/verification.types";

export const verificationService = {
  async getVerificationData(): Promise<VerificationData> {
    const response = await apiClient.get<{ data: VerificationData }>("/supplier/verification");
    return response.data.data;
  },

  async updateVerificationData(data: Partial<VerificationData>): Promise<VerificationData> {
    const response = await apiClient.put<{ data: VerificationData }>("/supplier/verification", data);
    return response.data.data;
  },

  async submitVerification(): Promise<VerificationData> {
    const response = await apiClient.post<{ data: VerificationData }>("/supplier/verification/submit");
    return response.data.data;
  },
};
