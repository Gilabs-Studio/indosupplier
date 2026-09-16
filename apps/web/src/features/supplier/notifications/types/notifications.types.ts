export interface SupplierNotification {
  id: string;
  title: string;
  desc: string;
  date: string;
  unread: boolean;
  type: "quote" | "message" | "alert" | "system";
}
