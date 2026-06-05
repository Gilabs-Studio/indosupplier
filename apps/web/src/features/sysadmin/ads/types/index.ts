export interface AdCampaign {
  id: string;
  supplierName: string;
  placement: string;
  budget: number;
  clicks: number;
  impressions: number;
  status: "pending" | "approved" | "rejected" | "revised";
  createdAt: string;
}
