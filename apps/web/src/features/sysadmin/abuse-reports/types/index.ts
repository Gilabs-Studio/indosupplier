export interface AbuseReport {
  id: string;
  reporterName: string;
  reporterEmail: string;
  reportedName: string;
  reportedType: "buyer" | "supplier";
  reason: string;
  description: string;
  status: "pending" | "dismissed" | "warned" | "suspended";
  createdAt: string;
}
