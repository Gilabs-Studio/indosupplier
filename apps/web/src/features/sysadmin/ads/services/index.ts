import type { AdCampaign } from "../types";

let mockCampaigns: AdCampaign[] = [
  {
    id: "CAM-001",
    supplierName: "PT Java Woodworking",
    placement: "Search Top Placement",
    budget: 1500000,
    clicks: 120,
    impressions: 4800,
    status: "approved",
    createdAt: "2026-06-02T08:00:00Z"
  },
  {
    id: "CAM-002",
    supplierName: "PT Bumi Rempah Indonesia",
    placement: "Category Page Banner",
    budget: 3500000,
    clicks: 340,
    impressions: 12500,
    status: "approved",
    createdAt: "2026-06-01T10:30:00Z"
  },
  {
    id: "CAM-003",
    supplierName: "CV Nusantara Garment",
    placement: "Featured Product Card",
    budget: 800000,
    clicks: 0,
    impressions: 0,
    status: "pending",
    createdAt: "2026-06-03T11:00:00Z"
  },
  {
    id: "CAM-004",
    supplierName: "UD Metal Perkasa",
    placement: "Search Top Placement",
    budget: 1200000,
    clicks: 45,
    impressions: 1100,
    status: "revised",
    createdAt: "2026-05-28T09:00:00Z"
  }
];

export const adCampaignService = {
  async list(params: { page: number; limit: number; status?: string }): Promise<{ items: AdCampaign[]; total: number }> {
    return new Promise((resolve) => {
      setTimeout(() => {
        let filtered = [...mockCampaigns];
        if (params.status) {
          filtered = filtered.filter(c => c.status === params.status);
        }
        const start = (params.page - 1) * params.limit;
        const items = filtered.slice(start, start + params.limit);
        resolve({
          items,
          total: filtered.length
        });
      }, 100);
    });
  },

  async updateStatus(id: string, status: AdCampaign["status"]): Promise<AdCampaign> {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        const campaign = mockCampaigns.find(c => c.id === id);
        if (!campaign) {
          reject(new Error("Campaign not found"));
          return;
        }
        campaign.status = status;
        resolve({ ...campaign });
      }, 100);
    });
  },

  async delete(id: string): Promise<void> {
    return new Promise((resolve) => {
      setTimeout(() => {
        mockCampaigns = mockCampaigns.filter(c => c.id !== id);
        resolve();
      }, 100);
    });
  }
};
