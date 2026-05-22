// Session status for AI chat sessions
export type AIChatSessionStatus = "ACTIVE" | "CLOSED" | "EXPIRED";

// Action status for AI actions
export type AIActionStatus =
  | "PENDING_CONFIRMATION"
  | "SUCCESS"
  | "FAILED"
  | "CANCELLED";

// Message role in conversation
export type MessageRole = "user" | "assistant" | "system";

// Chat message from the API
export interface AIChatMessage {
  id: string;
  session_id: string;
  role: MessageRole;
  content: string;
  intent?: string;
  token_usage?: string;
  model?: string;
  duration_ms?: number;
  created_at: string;
}

// Chat session list item (lightweight)
export interface AIChatSession {
  id: string;
  user_id: string;
  title: string;
  status: AIChatSessionStatus;
  last_activity: string;
  message_count: number;
  created_at: string;
}

// Chat session detail (with messages and actions)
export interface AIChatSessionDetail {
  id: string;
  user_id: string;
  title: string;
  status: AIChatSessionStatus;
  last_activity: string;
  message_count: number;
  metadata?: string;
  messages: AIChatMessage[];
  actions: AIActionLog[];
  pending_action?: AIActionPreview | null;
  created_at: string;
}

// Action log entry
export interface AIActionLog {
  id: string;
  session_id: string;
  message_id?: string;
  user_id: string;
  intent: string;
  entity_type?: string;
  entity_id?: string;
  action: string;
  request_payload?: string;
  response_payload?: string;
  status: AIActionStatus;
  error_message?: string;
  permission_used?: string;
  duration_ms?: number;
  created_at: string;
}

// Action preview for confirmation flow
export interface AIActionPreview {
  id: string;
  intent: string;
  status: string;
  entity_type?: string;
  entity_id?: string;
  payload_preview?: unknown;
  duration_ms?: number;
}

// Token usage statistics
export interface TokenUsage {
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
}

// Intent registry entry
export interface AIIntentRegistry {
  id: string;
  intent_code: string;
  display_name: string;
  description: string;
  module: string;
  action_type: string;
  required_permission: string;
  requires_confirmation: boolean;
  endpoint_path: string;
  parameter_schema?: string;
  is_active: boolean;
}

// --- API Request Payloads ---

export interface SendMessagePayload {
  message: string;
  session_id?: string;
  model?: string;
}

// Cerebras model info from the API
export interface AIModelInfo {
  id: string;
  display_name: string;
  description: string;
  is_default: boolean;
}

export interface AIModelsResponse {
  success: boolean;
  data: AIModelInfo[];
  timestamp: string;
  request_id: string;
}

export interface ConfirmActionPayload {
  session_id?: string;
  action_id: string;
  confirmed: boolean;
}

// --- API Response Types ---

export interface ChatResponse {
  session_id: string;
  message: AIChatMessage;
  action?: AIActionPreview;
  requires_confirmation: boolean;
  token_usage?: TokenUsage;
}

export interface SendMessageResponse {
  success: boolean;
  data: ChatResponse;
  timestamp: string;
  request_id: string;
}

export interface SessionsListResponse {
  success: boolean;
  data: AIChatSession[];
  meta: {
    pagination: {
      page: number;
      per_page: number;
      total: number;
      total_pages: number;
      has_next: boolean;
      has_prev: boolean;
    };
  };
  timestamp: string;
  request_id: string;
}

export interface SessionDetailResponse {
  success: boolean;
  data: AIChatSessionDetail;
  timestamp: string;
  request_id: string;
}

export interface ActionLogsResponse {
  success: boolean;
  data: AIActionLog[];
  meta: {
    pagination: {
      page: number;
      per_page: number;
      total: number;
      total_pages: number;
      has_next: boolean;
      has_prev: boolean;
    };
  };
  timestamp: string;
  request_id: string;
}

export interface IntentRegistryResponse {
  success: boolean;
  data: AIIntentRegistry[];
  timestamp: string;
  request_id: string;
}

// --- Streaming Event Types (mirrors backend tools.StreamEvent) ---

export type StreamEventType =
  | "message_start"
  | "content_delta"
  | "tool_call"
  | "tool_result"
  | "thinking"
  | "message_end"
  | "error";

export interface StreamEvent {
  type: StreamEventType;
  content?: string;
  data?: unknown;
}

export interface StreamMessageStartData {
  session_id: string;
}

export interface StreamToolCallData {
  name: string;
  parameters: Record<string, unknown>;
}

export interface StreamToolResultData {
  call: {
    id: string;
    name: string;
    parameters: Record<string, unknown>;
  };
  result: {
    success: boolean;
    message: string;
    data?: unknown;
    entity_type?: string;
    entity_id?: string;
    action: string;
    duration_ms: number;
    error_code?: string;
    error_message?: string;
  };
}

export interface StreamMessageEndData {
  duration_ms: number;
  turn_count: number;
}

// --- Filter/Query Params ---

export interface SessionFilters {
  page?: number;
  per_page?: number;
  status?: AIChatSessionStatus;
  search?: string;
}

export interface ActionLogFilters {
  page?: number;
  per_page?: number;
  user_id?: string;
  intent?: string;
  status?: AIActionStatus;
}
