// Package tools defines the Tool interface and core types for the AI tool system.
// Inspired by Claude Code's tool architecture: each tool is a self-contained unit
// with schema, permissions, and execution logic — replacing the monolithic switch-case executor.
package tools

import (
	"context"
)

// Tool represents a single executable capability exposed to the AI model.
// The model selects which tool to call based on the tool's name, description, and parameters.
type Tool interface {
	// Name returns the unique identifier (e.g. "query_stock", "create_holiday")
	Name() string
	// Description returns a human-readable description for the LLM system prompt
	Description() string
	// Module returns the ERP module this tool belongs to (e.g. "hrd", "sales", "finance")
	Module() string
	// Category returns the action category: "query", "create", "update", "delete", "approve"
	Category() string
	// Permission returns the required permission string (e.g. "sales.quotation.create")
	Permission() string
	// NeedsConfirmation returns whether this tool requires user confirmation before execution
	NeedsConfirmation() bool
	// IsConcurrencySafe returns true if the tool can run in parallel with other safe tools
	IsConcurrencySafe() bool
	// Parameters returns the parameter definitions for the LLM
	Parameters() []ToolParameter
	// Execute runs the tool with the given parameters and returns a result
	Execute(ctx context.Context, params map[string]interface{}, execCtx *ExecutionContext) (*ToolResult, error)
}

// ToolParameter describes a single parameter that the LLM should extract and provide.
type ToolParameter struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`        // "string", "number", "boolean", "date", "uuid", "array", "object"
	Description string   `json:"description"`
	Required    bool     `json:"required"`
	Enum        []string `json:"enum,omitempty"` // Allowed values for constrained parameters
	Default     string   `json:"default,omitempty"`
}

// ExecutionContext carries per-request state needed by tool execution.
type ExecutionContext struct {
	UserID          string
	UserPermissions map[string]bool
	IsAdmin         bool
	SessionID       string
	// ResolvedEntities maps parameter names to resolved database IDs
	// (e.g. "customer_name" → {ID: "uuid", Name: "PT ABC"})
	ResolvedEntities map[string]*ResolvedEntity
}

// ResolvedEntity holds a resolved entity reference from natural language input.
type ResolvedEntity struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"` // "customer", "employee", "product", etc.
}

// ToolResult holds the outcome of a tool execution.
type ToolResult struct {
	Success      bool        `json:"success"`
	Data         interface{} `json:"data,omitempty"`
	Message      string      `json:"message"`
	EntityType   string      `json:"entity_type,omitempty"`
	EntityID     string      `json:"entity_id,omitempty"`
	Action       string      `json:"action"`
	DurationMs   int64       `json:"duration_ms"`
	ErrorCode    string      `json:"error_code,omitempty"`
	ErrorMessage string      `json:"error_message,omitempty"`
}

// ToolCall represents a parsed tool invocation from the LLM response.
type ToolCall struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
}

// ToolCallResult pairs a tool call with its execution result.
type ToolCallResult struct {
	Call   *ToolCall   `json:"call"`
	Result *ToolResult `json:"result"`
	Error  string      `json:"error,omitempty"`
}

// StreamEvent represents an event emitted during streamed conversation.
type StreamEvent struct {
	Type    StreamEventType `json:"type"`
	Content string          `json:"content,omitempty"`
	Data    interface{}     `json:"data,omitempty"`
}

// StreamEventType categorizes streaming events (mirrors Claude Code's streaming model).
type StreamEventType string

const (
	EventMessageStart StreamEventType = "message_start"
	EventContentDelta StreamEventType = "content_delta"
	EventToolCall     StreamEventType = "tool_call"
	EventToolResult   StreamEventType = "tool_result"
	EventThinking     StreamEventType = "thinking"
	EventMessageEnd   StreamEventType = "message_end"
	EventError        StreamEventType = "error"
)
