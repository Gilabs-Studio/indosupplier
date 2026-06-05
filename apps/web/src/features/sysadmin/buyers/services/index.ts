import type { Buyer } from "../types";

let mockBuyers: Buyer[] = [
  {
    id: "BYR-001",
    name: "Sarah Lee",
    email: "sarah.lee@sourcing.au",
    companyName: "Aust Sourcing Ltd",
    country: "Australia",
    status: "active",
    rfqCount: 8,
    leadQualityScore: 92,
    createdAt: "2026-05-15T09:00:00Z"
  },
  {
    id: "BYR-002",
    name: "Budi Hartono",
    email: "budi@retailjaya.id",
    companyName: "PT Retail Jaya",
    country: "Indonesia",
    status: "active",
    rfqCount: 12,
    leadQualityScore: 84,
    createdAt: "2026-05-20T11:30:00Z"
  },
  {
    id: "BYR-003",
    name: "Ken Tan",
    email: "ken@globalfoods.sg",
    companyName: "Global Foods Pte Ltd",
    country: "Singapore",
    status: "review",
    rfqCount: 3,
    leadQualityScore: 56,
    createdAt: "2026-06-01T14:00:00Z"
  },
  {
    id: "BYR-004",
    name: "Michel Dupont",
    email: "m.dupont@import.fr",
    companyName: "Dupont Importation",
    country: "France",
    status: "suspended",
    rfqCount: 1,
    leadQualityScore: 32,
    createdAt: "2026-05-10T10:15:00Z"
  }
];

export const buyerService = {
  async list(params: { page: number; limit: number; status?: string; search?: string }): Promise<{ items: Buyer[]; total: number }> {
    return new Promise((resolve) => {
      setTimeout(() => {
        let filtered = [...mockBuyers];
        if (params.status) {
          filtered = filtered.filter(b => b.status === params.status);
        }
        if (params.search) {
          const s = params.search.toLowerCase();
          filtered = filtered.filter(b => 
            b.name.toLowerCase().includes(s) || 
            b.email.toLowerCase().includes(s) || 
            b.companyName.toLowerCase().includes(s)
          );
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

  async updateStatus(id: string, status: Buyer["status"]): Promise<Buyer> {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        const buyer = mockBuyers.find(b => b.id === id);
        if (!buyer) {
          reject(new Error("Buyer not found"));
          return;
        }
        buyer.status = status;
        resolve({ ...buyer });
      }, 100);
    });
  },

  async delete(id: string): Promise<void> {
    return new Promise((resolve) => {
      setTimeout(() => {
        mockBuyers = mockBuyers.filter(b => b.id !== id);
        resolve();
      }, 100);
    });
  }
};
