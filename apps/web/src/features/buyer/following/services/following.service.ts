import { apiClient } from "@/lib/api-client";
import type { FollowingSupplierItem } from "../types/following.types";

interface ApiResponse<T> {
  success: boolean;
  data: T;
}

export const followingService = {
  async getFollowingSuppliers(): Promise<FollowingSupplierItem[]> {
    const response = await apiClient.get<ApiResponse<FollowingSupplierItem[]>>("/buyer/following");
    return response.data.data || [];
  },

  async followSupplier(supplierProfileId: string): Promise<FollowingSupplierItem> {
    const response = await apiClient.post<ApiResponse<FollowingSupplierItem>>("/buyer/following", {
      supplierProfileId,
    });
    return response.data.data;
  },

  async unfollowSupplier(supplierProfileId: string): Promise<void> {
    await apiClient.delete(`/buyer/following/${supplierProfileId}`);
  },
};
