package dto

// SendMessageRequest represents a chat message request from the user
type SendMessageRequest struct {
	SessionID *string `json:"session_id" binding:"omitempty,uuid"`
	Message   string  `json:"message" binding:"required,min=1,max=2000"`
	Model     string  `json:"model" binding:"omitempty,max=100"`
}

// ConfirmActionRequest represents a confirmation or cancellation of a pending action
type ConfirmActionRequest struct {
	SessionID *string `json:"session_id" binding:"omitempty,uuid"`
	ActionID  string  `json:"action_id" binding:"required,uuid"`
	Confirmed bool    `json:"confirmed"`
}

// ListSessionsRequest represents query parameters for listing sessions
type ListSessionsRequest struct {
	Page    int    `form:"page,default=1" binding:"omitempty,min=1"`
	PerPage int    `form:"per_page,default=20" binding:"omitempty,min=1,max=100"`
	Status  string `form:"status" binding:"omitempty,oneof=ACTIVE CLOSED EXPIRED"`
	Search  string `form:"search" binding:"omitempty,max=200"`
}

// ListActionsRequest represents query parameters for listing action logs
type ListActionsRequest struct {
	Page    int    `form:"page,default=1" binding:"omitempty,min=1"`
	PerPage int    `form:"per_page,default=20" binding:"omitempty,min=1,max=100"`
	UserID  string `form:"user_id" binding:"omitempty,uuid"`
	Intent  string `form:"intent" binding:"omitempty,max=100"`
	Status  string `form:"status" binding:"omitempty,oneof=SUCCESS FAILED PENDING_CONFIRMATION CANCELLED"`
}

// ChatMessageResponse represents a single message in the response (legacy format)
type ChatMessageResponse struct {
	ID        string      `json:"id"`
	Role      string      `json:"role"`
	Content   string      `json:"content"`
	Intent    interface{} `json:"intent,omitempty"`
	CreatedAt string      `json:"created_at"`
}

// MessageResponse is the enhanced message format used by the conversation engine.
type MessageResponse struct {
	ID         string      `json:"id"`
	Role       string      `json:"role"`
	Content    string      `json:"content"`
	Intent     *string     `json:"intent,omitempty"`
	Model      string      `json:"model,omitempty"`
	DurationMs int         `json:"duration_ms,omitempty"`
	ToolCalls  interface{} `json:"tool_calls,omitempty"`
	CreatedAt  string      `json:"created_at"`
}

// ActionPreview represents a preview of an action to be confirmed
type ActionPreview struct {
	ID             string      `json:"id"`
	Intent         string      `json:"intent"`
	Status         string      `json:"status"`
	EntityType     string      `json:"entity_type,omitempty"`
	EntityID       string      `json:"entity_id,omitempty"`
	PayloadPreview interface{} `json:"payload_preview,omitempty"`
	DurationMs     int         `json:"duration_ms,omitempty"`
}

// TokenUsageResponse represents LLM token usage
type TokenUsageResponse struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatResponse represents the full response from the chat endpoint
type ChatResponse struct {
	SessionID            string              `json:"session_id"`
	Message              MessageResponse     `json:"message"`
	Action               *ActionPreview      `json:"action,omitempty"`
	RequiresConfirmation bool                `json:"requires_confirmation"`
	TokenUsage           *TokenUsageResponse `json:"token_usage,omitempty"`
}

// SessionListResponse represents a session in the list view
type SessionListResponse struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	MessageCount int    `json:"message_count"`
	LastActivity string `json:"last_activity"`
	CreatedAt    string `json:"created_at"`
}

// SessionDetailResponse represents a session with messages and actions
type SessionDetailResponse struct {
	ID            string              `json:"id"`
	Title         string              `json:"title"`
	Status        string              `json:"status"`
	LastActivity  string              `json:"last_activity,omitempty"`
	MessageCount  int                 `json:"message_count"`
	Messages      []MessageResponse   `json:"messages"`
	Actions       []ActionLogResponse `json:"actions,omitempty"`
	PendingAction *ActionPreview      `json:"pending_action,omitempty"`
	CreatedAt     string              `json:"created_at"`
}

// ActionLogResponse represents an action log entry
type ActionLogResponse struct {
	ID              string      `json:"id"`
	SessionID       string      `json:"session_id,omitempty"`
	UserID          string      `json:"user_id,omitempty"`
	Intent          string      `json:"intent"`
	Action          string      `json:"action"`
	EntityType      string      `json:"entity_type,omitempty"`
	EntityID        string      `json:"entity_id,omitempty"`
	Status          string      `json:"status"`
	RequestPayload  interface{} `json:"request_payload,omitempty"`
	ResponsePayload interface{} `json:"response_payload,omitempty"`
	ErrorMessage    string      `json:"error_message,omitempty"`
	PermissionUsed  string      `json:"permission_used,omitempty"`
	DurationMs      int         `json:"duration_ms,omitempty"`
	CreatedAt       string      `json:"created_at"`
}

// IntentRegistryResponse represents an intent from the registry
type IntentRegistryResponse struct {
	ID                   string `json:"id"`
	IntentCode           string `json:"intent_code"`
	DisplayName          string `json:"display_name"`
	Description          string `json:"description,omitempty"`
	Module               string `json:"module"`
	ActionType           string `json:"action_type"`
	RequiredPermission   string `json:"required_permission"`
	RequiresConfirmation bool   `json:"requires_confirmation"`
	IsActive             bool   `json:"is_active"`
}

// AIStatsResponse represents AI usage statistics
type AIStatsResponse struct {
	TotalSessions     int64 `json:"total_sessions"`
	TotalMessages     int64 `json:"total_messages"`
	TotalActions      int64 `json:"total_actions"`
	SuccessfulActions int64 `json:"successful_actions"`
	FailedActions     int64 `json:"failed_actions"`
}
