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
  async getProfile(): Promise<BuyerProfile & { phone?: string; website?: string; address?: string }> {
    try {
      const response = await apiClient.get<ApiResponse<BuyerProfile & { phone?: string; website?: string; address?: string }>>(
        "/buyer/profile"
      );
      return response.data.data;
    } catch (error) {
      console.warn("Backend GET /buyer/profile not implemented, using local mock data", error);
      return {
        id: "mock-buyer-profile-id",
        user_id: "mock-user-id",
        full_name: "Yohanes",
        company_name: "PT Global Sourcing Mandiri",
        country_code: "ID",
        industry: "Logistics & Supply Chain",
        purchase_frequency: "Bulanan",
        profile_completeness: 85,
        phone: "+62 812-3456-7890",
        website: "https://www.globalsourcing.co.id",
        address: "Sudirman Central Business District (SCBD), Jakarta Selatan",
      };
    }
  },

  async updatePersonal(data: ProfilePersonalPayload): Promise<void> {
    try {
      await apiClient.put("/buyer/profile/personal", data);
    } catch (error) {
      console.warn("Backend PUT /buyer/profile/personal not implemented", error);
    }
  },

  async updateCompany(data: ProfileCompanyPayload): Promise<void> {
    try {
      await apiClient.put("/buyer/profile/company", data);
    } catch (error) {
      console.warn("Backend PUT /buyer/profile/company not implemented", error);
    }
  },

  async getDocuments(): Promise<BuyerDocument[]> {
    try {
      const response = await apiClient.get<ApiResponse<BuyerDocument[]>>("/buyer/profile/documents");
      return response.data.data || [];
    } catch (error) {
      console.warn("Backend GET /buyer/profile/documents not implemented, using local mock data", error);
      return [
        {
          id: "doc-1",
          buyer_profile_id: "mock-buyer-profile-id",
          document_type: "Nomor Induk Berusaha (NIB)",
          document_number: "NIB-912048123",
          file_url: "/uploads/nib.pdf",
          status: "verified",
          created_at: "2026-01-10T12:00:00Z",
        },
        {
          id: "doc-2",
          buyer_profile_id: "mock-buyer-profile-id",
          document_type: "Surat Izin Usaha Perdagangan (SIUP)",
          document_number: "SIUP-412/10-24/2022",
          file_url: "/uploads/siup.pdf",
          status: "verified",
          created_at: "2026-01-10T12:00:00Z",
        },
        {
          id: "doc-3",
          buyer_profile_id: "mock-buyer-profile-id",
          document_type: "NPWP Perusahaan",
          document_number: "NPWP-01.234.567.8-012.000",
          file_url: "/uploads/npwp.pdf",
          status: "verified",
          created_at: "2026-01-10T12:00:00Z",
        },
      ];
    }
  },

  async uploadDocument(data: UploadDocumentPayload): Promise<BuyerDocument> {
    try {
      const response = await apiClient.post<ApiResponse<BuyerDocument>>("/buyer/profile/documents", data);
      return response.data.data;
    } catch (error) {
      console.warn("Backend POST /buyer/profile/documents not implemented, simulating upload", error);
      return {
        id: "doc-" + Math.random().toString(36).substr(2, 9),
        buyer_profile_id: "mock-buyer-profile-id",
        document_type: data.document_type,
        document_number: data.document_number,
        file_url: data.file_url,
        status: "pending",
        created_at: new Date().toISOString(),
      };
    }
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
