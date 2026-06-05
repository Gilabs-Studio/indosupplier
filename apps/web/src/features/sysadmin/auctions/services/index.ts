import type { AuctionSession } from "../types";

let mockAuctions: AuctionSession[] = [
  {
    id: "AUC-001",
    category: "Food & Ingredients",
    slots: 2,
    minBid: 2000000,
    bidsCount: 8,
    highestBid: 4200000,
    status: "open",
    startDate: "2026-06-01T00:00:00Z",
    endDate: "2026-06-15T23:59:59Z"
  },
  {
    id: "AUC-002",
    category: "Textile & Garment",
    slots: 3,
    minBid: 3500000,
    bidsCount: 12,
    highestBid: 6800000,
    status: "open",
    startDate: "2026-06-01T00:00:00Z",
    endDate: "2026-06-14T23:59:59Z"
  },
  {
    id: "AUC-003",
    category: "Wooden Furniture",
    slots: 2,
    minBid: 1750000,
    bidsCount: 3,
    highestBid: 2100000,
    status: "closed",
    startDate: "2026-05-15T00:00:00Z",
    endDate: "2026-05-30T23:59:59Z"
  },
  {
    id: "AUC-004",
    category: "Agriculture & Copra",
    slots: 2,
    minBid: 1500000,
    bidsCount: 0,
    highestBid: 0,
    status: "draft",
    startDate: "2026-06-15T00:00:00Z",
    endDate: "2026-06-30T23:59:59Z"
  }
];

export const auctionSessionService = {
  async list(params: { page: number; limit: number; status?: string }): Promise<{ items: AuctionSession[]; total: number }> {
    return new Promise((resolve) => {
      setTimeout(() => {
        let filtered = [...mockAuctions];
        if (params.status) {
          filtered = filtered.filter(a => a.status === params.status);
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

  async updateStatus(id: string, status: AuctionSession["status"]): Promise<AuctionSession> {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        const auction = mockAuctions.find(a => a.id === id);
        if (!auction) {
          reject(new Error("Auction not found"));
          return;
        }
        auction.status = status;
        resolve({ ...auction });
      }, 100);
    });
  },

  async create(data: Omit<AuctionSession, "id" | "bidsCount" | "highestBid">): Promise<AuctionSession> {
    return new Promise((resolve) => {
      setTimeout(() => {
        const newAuction: AuctionSession = {
          ...data,
          id: `AUC-00${mockAuctions.length + 1}`,
          bidsCount: 0,
          highestBid: 0
        };
        mockAuctions = [newAuction, ...mockAuctions];
        resolve(newAuction);
      }, 100);
    });
  },

  async delete(id: string): Promise<void> {
    return new Promise((resolve) => {
      setTimeout(() => {
        mockAuctions = mockAuctions.filter(a => a.id !== id);
        resolve();
      }, 100);
    });
  }
};
