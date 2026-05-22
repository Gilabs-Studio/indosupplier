export type NotificationType = "approval_request" | "info" | "warning";

export interface Notification {
  id: string;
  user_id: string;
  title: string;
  message: string;
  type: NotificationType;
  entity_type: string;
  entity_id: string;
  entity_link?: string;
  is_read: boolean;
  read_at: string | null;
  created_at: string;
}

export interface ListNotificationsParams {
  page?: number;
  per_page?: number;
  type?: NotificationType;
  entity?: string;
  is_read?: boolean;
}

export interface ListNotificationsResponse {
  success: boolean;
  data: Notification[];
  meta: {
    pagination: {
      page: number;
      per_page: number;
      total: number;
      total_pages: number;
      has_next: boolean;
      has_prev: boolean;
    };
    filters?: Record<string, unknown>;
  };
  timestamp: string;
  request_id: string;
}

export interface NotificationResponse {
  success: boolean;
  data: Notification;
  meta?: {
    filters?: Record<string, unknown>;
  };
  timestamp: string;
  request_id: string;
}

export interface UnreadCountResponse {
  success: boolean;
  data: {
    unread_count: number;
  };
  timestamp: string;
  request_id: string;
}

export interface ApiEnvelope<T> {
  success: boolean;
  data: T;
  meta?: {
    pagination?: {
      page: number;
      per_page: number;
      total: number;
      total_pages: number;
      has_next: boolean;
      has_prev: boolean;
    };
    filters?: Record<string, unknown>;
  };
  timestamp: string;
  request_id: string;
}

export interface WebSocketMessage {
  event?: "notification.created" | "notification.updated" | "notification.deleted";
  type: "notification.created" | "notification.updated" | "notification.deleted";
  data: Notification | { user_id: string; notification_id: string };
}

