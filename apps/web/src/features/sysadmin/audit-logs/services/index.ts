import type { AuditLog } from "../types";

const mockLogs: AuditLog[] = [
  {
    id: "LOG-5001",
    actorName: "John Doe",
    actorEmail: "john@indosupplier.com",
    action: "user.suspend",
    target: "Buyer: Ken Tan (Global Trading)",
    metadata: '{"reason": "Fraudulent Activity reported by CV Nusantara Textile", "duration_days": 30}',
    createdAt: "2026-06-04T09:12:00Z"
  },
  {
    id: "LOG-5002",
    actorName: "System",
    actorEmail: "system@indosupplier.local",
    action: "waitlist.auto_approve",
    target: "Supplier: PT Woodindo Pratama",
    metadata: '{"nib": "1283928192839", "company_type": "supplier"}',
    createdAt: "2026-06-03T18:45:00Z"
  },
  {
    id: "LOG-5003",
    actorName: "Jane Smith",
    actorEmail: "jane@indosupplier.com",
    action: "ads.approve",
    target: "Campaign: CAM-001 (Ramadan Boost)",
    metadata: '{"approved_by": "Jane Smith", "budget": 1500000}',
    createdAt: "2026-06-02T11:30:00Z"
  },
  {
    id: "LOG-5004",
    actorName: "John Doe",
    actorEmail: "john@indosupplier.com",
    action: "faq.create",
    target: "Article: Kebijakan Transaksi Aman",
    metadata: '{"slug": "kebijakan-transaksi-aman", "locale": "id"}',
    createdAt: "2026-06-01T15:20:00Z"
  }
];

export const auditLogService = {
  async list(params: { page: number; limit: number; action?: string; search?: string }): Promise<{ items: AuditLog[]; total: number }> {
    return new Promise((resolve) => {
      setTimeout(() => {
        let filtered = [...mockLogs];
        if (params.action) {
          filtered = filtered.filter(l => l.action === params.action);
        }
        if (params.search) {
          const s = params.search.toLowerCase();
          filtered = filtered.filter(l => 
            l.actorName.toLowerCase().includes(s) || 
            l.actorEmail.toLowerCase().includes(s) || 
            l.target.toLowerCase().includes(s)
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
  }
};
