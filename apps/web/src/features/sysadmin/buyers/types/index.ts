export interface Buyer {
  id: string;
  name: string;
  email: string;
  companyName: string;
  country: string;
  status: "active" | "suspended" | "review";
  rfqCount: number;
  leadQualityScore: number;
  createdAt: string;
}
