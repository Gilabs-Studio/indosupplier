export interface BuyerNotification {
  id: string;
  title: string;
  desc: string;
  date: string;
  unread: boolean;
  type: "quote" | "message" | "system" | "alert";
}
