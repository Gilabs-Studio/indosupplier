export interface SupportTicket {
  id: string;
  subject: string;
  date: string;
  status: "Open" | "Closed";
}

export interface SupportMessage {
  sender: "buyer" | "support";
  name: string;
  text: string;
  time: string;
}

export interface SupportTicketDetail extends SupportTicket {
  messages: SupportMessage[];
}

export interface CreateTicketPayload {
  subject: string;
  message: string;
}
