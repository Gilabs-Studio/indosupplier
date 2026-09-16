export interface SupplierNotification {
  id: string;
  title: string;
  desc: string;
  date: string;
  unread: boolean;
  type: "quote" | "message" | "alert" | "system";
  related_type?: string;
  related_id?: string;
}
