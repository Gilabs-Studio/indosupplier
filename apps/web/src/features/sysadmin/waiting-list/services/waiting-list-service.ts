import { apiClient, publicApiClient } from "@/lib/api-client";
import type { JoinWaitingListRequest, WaitingListEntry, WaitingListListResponse, SingleWaitingListResponse } from "../types";

export const waitingListService = {
  async join(data: JoinWaitingListRequest): Promise<WaitingListEntry> {
    const response = await publicApiClient.post<SingleWaitingListResponse>(
      "/waiting-list/join",
      data
    );
    return response.data.data;
  },

  async list(params: {
    page: number;
    limit: number;
    status?: string;
  }): Promise<{ items: WaitingListEntry[]; total: number }> {
    const response = await apiClient.get<WaitingListListResponse>(
      "/sysadmin/waiting-list",
      { params }
    );
    return {
      items: response.data.data || [],
      total: response.data.meta?.pagination?.total || 0,
    };
  },

  async updateStatus(id: string, status: string): Promise<WaitingListEntry> {
    const response = await apiClient.put<SingleWaitingListResponse>(
      `/sysadmin/waiting-list/${id}`,
      { status }
    );
    return response.data.data;
  },

  async delete(id: string): Promise<void> {
    await apiClient.delete(`/sysadmin/waiting-list/${id}`);
  },
};
