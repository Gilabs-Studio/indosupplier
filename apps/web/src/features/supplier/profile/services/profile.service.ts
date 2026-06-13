import { apiClient } from "@/lib/api-client";
import type { SupplierProfileData, UpdateProfilePayload } from "../types/profile.types";

const DEFAULT_PROFILE: SupplierProfileData = {
  id: "prof-123",
  companyName: "PT Nusantara Supplier Utama",
  businessType: "Manufacturer & Distributor",
  established: "2018",
  employees: "150 Employees",
  email: "info@nusantarasupplier.com",
  phone: "+62 811 2345 6789",
  website: "https://nusantarasupplier.com",
  taxId: "01.234.567.8-999.000 (NPWP)",
  nib: "9120001234567 (NIB)",
  overview: "We are the leading raw materials supplier in Indonesia, focusing on high-grade minerals, industrial grade garnet sand, quartz powder, and agricultural bulk products.",
  location: "Kawasan Industri Jababeka, Cikarang, Jawa Barat, Indonesia",
  status: "active",
};

export const profileService = {
  async getProfile(): Promise<SupplierProfileData> {
    try {
      // Attempt to hit the backend if an endpoint is ever available
      const response = await apiClient.get<{ data: SupplierProfileData }>("/supplier/profile");
      if (response.data?.data) {
        return response.data.data;
      }
    } catch {
      // Silence backend error and fall back to local storage
    }

    if (typeof window !== "undefined") {
      const stored = localStorage.getItem("supplier_profile");
      if (stored) {
        try {
          return JSON.parse(stored);
        } catch {
          // ignore parsing errors
        }
      }
      localStorage.setItem("supplier_profile", JSON.stringify(DEFAULT_PROFILE));
    }
    return DEFAULT_PROFILE;
  },

  async updateProfile(payload: UpdateProfilePayload): Promise<SupplierProfileData> {
    try {
      const response = await apiClient.put<{ data: SupplierProfileData }>("/supplier/profile", payload);
      if (response.data?.data) {
        return response.data.data;
      }
    } catch {
      // Silence backend error and fall back to local storage
    }

    const updatedProfile: SupplierProfileData = {
      id: "prof-123",
      status: "active",
      ...payload,
    };

    if (typeof window !== "undefined") {
      localStorage.setItem("supplier_profile", JSON.stringify(updatedProfile));
    }

    return updatedProfile;
  },
};
