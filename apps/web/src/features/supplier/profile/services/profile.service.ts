import { apiClient } from "@/lib/api-client";
import type { SupplierProfileData, UpdateProfilePayload } from "../types/profile.types";

export const profileService = {
  async getProfile(): Promise<SupplierProfileData> {
    const response = await apiClient.get<{ data: SupplierProfileData }>("/supplier/profile");
    return response.data.data;
  },

  async updateProfile(payload: UpdateProfilePayload): Promise<SupplierProfileData> {
    const response = await apiClient.put<{ data: SupplierProfileData }>("/supplier/profile", payload);
    return response.data.data;
  },
};
