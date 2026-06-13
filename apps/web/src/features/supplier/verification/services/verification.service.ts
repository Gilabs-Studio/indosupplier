import { apiClient } from "@/lib/api-client";
import type { VerificationData } from "../types/verification.types";

const INITIAL_VERIFICATION_DATA: VerificationData = {
  businessInfo: {
    legalName: "",
    nibNumber: "",
    establishedDate: "",
    industry: "",
    phone: "",
    description: "",
    address: "",
  },
  stakeholderInfo: {
    directorName: "",
    directorNik: "",
  },
  bankAccountInfo: {
    bankName: "",
    accountNumber: "",
    accountName: "",
  },
  status: "unverified",
};

export const verificationService = {
  async getVerificationData(): Promise<VerificationData> {
    try {
      const response = await apiClient.get<{ data: VerificationData }>("/supplier/verification");
      if (response.data?.data) {
        return response.data.data;
      }
    } catch {
      // Fallback
    }

    if (typeof window !== "undefined") {
      const stored = localStorage.getItem("supplier_verification_data");
      if (stored) {
        try {
          const parsed = JSON.parse(stored);
          const isVerified = localStorage.getItem("supplier_verified") === "true";
          return {
            ...parsed,
            status: isVerified ? "verified" : parsed.status,
          };
        } catch {
          // ignore
        }
      }
      localStorage.setItem("supplier_verification_data", JSON.stringify(INITIAL_VERIFICATION_DATA));
    }
    return INITIAL_VERIFICATION_DATA;
  },

  async updateVerificationData(data: Partial<VerificationData>): Promise<VerificationData> {
    const current = await this.getVerificationData();
    const updated = {
      ...current,
      ...data,
      businessInfo: { ...current.businessInfo, ...data.businessInfo },
      stakeholderInfo: { ...current.stakeholderInfo, ...data.stakeholderInfo },
      bankAccountInfo: { ...current.bankAccountInfo, ...data.bankAccountInfo },
    };

    try {
      await apiClient.put("/supplier/verification", updated);
    } catch {
      // Fallback
    }

    if (typeof window !== "undefined") {
      localStorage.setItem("supplier_verification_data", JSON.stringify(updated));
    }
    return updated;
  },

  async submitVerification(): Promise<VerificationData> {
    const current = await this.getVerificationData();
    const updated: VerificationData = {
      ...current,
      status: "verified",
    };

    try {
      await apiClient.post("/supplier/verification/submit");
    } catch {
      // Fallback
    }

    if (typeof window !== "undefined") {
      localStorage.setItem("supplier_verification_data", JSON.stringify(updated));
      localStorage.setItem("supplier_verified", "true");
    }
    return updated;
  },
};
