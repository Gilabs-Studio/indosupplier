import { apiClient } from "@/lib/api-client";
import type {
  BuyerProfile,
  ProfilePersonalPayload,
  ProfileCompanyPayload,
  BuyerDocument,
  UploadDocumentPayload,
} from "../types/profile.types";

interface ApiResponse<T> {
  success: boolean;
  data: T;
}

export const profileService = {
  async getProfile(): Promise<BuyerProfile> {
    const response = await apiClient.get<ApiResponse<BuyerProfile>>(
      "/buyer/profile"
    );
    return response.data.data;
  },

  async updatePersonal(data: ProfilePersonalPayload): Promise<BuyerProfile> {
    const response = await apiClient.put<ApiResponse<BuyerProfile>>("/buyer/profile/personal", data);
    return response.data.data;
  },

  async updateCompany(data: ProfileCompanyPayload): Promise<BuyerProfile> {
    const response = await apiClient.put<ApiResponse<BuyerProfile>>("/buyer/profile/company", data);
    return response.data.data;
  },

  async getDocuments(): Promise<BuyerDocument[]> {
    const response = await apiClient.get<ApiResponse<BuyerDocument[]>>("/buyer/profile/documents");
    return response.data.data || [];
  },

  async uploadDocument(data: UploadDocumentPayload): Promise<BuyerDocument> {
    const response = await apiClient.post<ApiResponse<BuyerDocument>>("/buyer/profile/documents", data);
    return response.data.data;
  },

  async uploadRawFile(file: File): Promise<{ url: string; filename: string }> {
    const formData = new FormData();
    formData.append("file", file);
    const response = await apiClient.post<ApiResponse<{ url: string; filename: string }>>(
      "/upload/image?folder=documents",
      formData,
      {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      }
    );
    return response.data.data;
  },
};
