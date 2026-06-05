export interface AuctionSession {
  id: string;
  category: string;
  slots: number;
  minBid: number;
  bidsCount: number;
  highestBid: number;
  status: "open" | "closed" | "draft";
  startDate: string;
  endDate: string;
}
