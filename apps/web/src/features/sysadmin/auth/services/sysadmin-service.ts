import { apiClient } from "@/lib/api-client";
import type { SysadminLoginPayload, SysadminLoginResponse, SysadminMeResponse, SystemAdmin } from "../types";
import type { LoginRequest } from "@/features/auth/types";

export const sysadminService = {
  async login(credentials: LoginRequest): Promise<SysadminLoginPayload> {
    const response = await apiClient.post<SysadminLoginResponse>(
      "/sysadmin/auth/login",
      credentials
    );
    return response.data.data;
  },

  async getMe(): Promise<SystemAdmin> {
    const response = await apiClient.get<SysadminMeResponse>("/sysadmin/auth/me");
    return response.data.data;
  },

  async logout(): Promise<void> {
    await apiClient.post("/sysadmin/auth/logout");
  },
};
