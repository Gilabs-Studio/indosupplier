export interface TicketMessage {
  id: string;
  sender: "user" | "agent" | "system";
  senderName: string;
  content: string;
  createdAt: string;
  isInternalNote?: boolean;
}

export interface SupportTicket {
  id: string;
  title: string;
  userType: "buyer" | "supplier";
  userName: string;
  category: string;
  priority: "low" | "medium" | "high" | "urgent";
  status: "open" | "in_progress" | "resolved" | "closed";
  assignedAgent?: string;
  createdAt: string;
  updatedAt: string;
  slaDeadline: string;
  messages: TicketMessage[];
}
