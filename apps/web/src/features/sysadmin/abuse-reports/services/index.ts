import type { AbuseReport } from "../types";

let mockReports: AbuseReport[] = [
  {
    id: "REP-101",
    reporterName: "Sarah Connor",
    reporterEmail: "sarah@connor.me",
    reportedName: "PT Java Woodworking",
    reportedType: "supplier",
    reason: "RFQ Spam",
    description: "Sends repeated automated messages offering unrelated plastic products.",
    status: "pending",
    createdAt: "2026-06-03T10:00:00Z"
  },
  {
    id: "REP-102",
    reporterName: "Budi Santoso",
    reporterEmail: "budi@santoso.co.id",
    reportedName: "Ken Tan (Global Trading)",
    reportedType: "buyer",
    reason: "Fraudulent Activity",
    description: "Requested samples but did not pay freight costs as agreed, then stopped replying.",
    status: "pending",
    createdAt: "2026-06-02T14:30:00Z"
  },
  {
    id: "REP-103",
    reporterName: "Alice Miller",
    reporterEmail: "alice@miller.com",
    reportedName: "PT Rempah Tropis",
    reportedType: "supplier",
    reason: "Inappropriate Content",
    description: "Product descriptions contain copyright-infringed text and stock watermarked photos.",
    status: "warned",
    createdAt: "2026-05-28T09:15:00Z"
  },
  {
    id: "REP-104",
    reporterName: "Joko Widodo",
    reporterEmail: "jokowi@indotrade.net",
    reportedName: "Sourcing Ltd",
    reportedType: "buyer",
    reason: "Spamming",
    description: "Flooding the categories with identical RFQs every hour.",
    status: "dismissed",
    createdAt: "2026-05-25T16:45:00Z"
  }
];

export const abuseReportService = {
  async list(params: { page: number; limit: number; status?: string }): Promise<{ items: AbuseReport[]; total: number }> {
    return new Promise((resolve) => {
      setTimeout(() => {
        let filtered = [...mockReports];
        if (params.status) {
          filtered = filtered.filter(r => r.status === params.status);
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

  async updateStatus(id: string, status: AbuseReport["status"]): Promise<AbuseReport> {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        const report = mockReports.find(r => r.id === id);
        if (!report) {
          reject(new Error("Report not found"));
          return;
        }
        report.status = status;
        resolve({ ...report });
      }, 100);
    });
  },

  async delete(id: string): Promise<void> {
    return new Promise((resolve) => {
      setTimeout(() => {
        mockReports = mockReports.filter(r => r.id !== id);
        resolve();
      }, 100);
    });
  }
};
