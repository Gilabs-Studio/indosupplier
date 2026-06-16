export interface ChatMessage {
  id: string;
  chat_room_id: string;
  sender_id: string;
  sender_type: "buyer" | "supplier";
  body: string;
  is_read: boolean;
  created_at: string;
}

export interface ChatRoom {
  id: string;
  buyer_profile_id: string;
  supplier_profile_id: string;
  company_name: string;
  is_verified: boolean;
  last_message?: ChatMessage;
  unread_count: number;
  created_at: string;
  updated_at: string;
}

export interface WebSocketEvent<T = unknown> {
  type: string;
  data: T;
}
