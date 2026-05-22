// Package engine implements the multi-turn conversation loop with tool execution,
// mirroring Claude Code's QueryEngine pattern: user message → model response →
// if tool_call, execute → feed result → repeat until end_turn.
package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	aiContext "github.com/gilabs/gims/api/internal/ai/domain/usecase/context"
	"github.com/gilabs/gims/api/internal/ai/domain/usecase/tools"
	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/cerebras"
)

// MaxToolTurns limits the number of tool loop iterations to prevent runaway conversations.
const MaxToolTurns = 5

// MaxContextMessages is the maximum number of historical messages included in the context window.
const MaxContextMessages = 8

// MaxCompletionTokens keeps headroom for small-context models (for example 8k context)
// so prompt + completion allocation does not overflow context limits.
const MaxCompletionTokens = 1024

const incompleteToolCallRecoveryMessage = "Maaf, respons aksi terpotong sebelum konfirmasi terbentuk. Mohon kirim ulang detail dalam satu pesan agar saya bisa menyiapkan konfirmasi dengan benar."

var legacyToolCallMarkerPattern = regexp.MustCompile(`<((?:create|update|delete|query|list|approve|reject|generate)_[a-z0-9_]+)>`)

// ConversationMessage represents a single message in the conversation history.
type ConversationMessage struct {
	Role    string `json:"role"` // "user", "assistant", "system"
	Content string `json:"content"`
}

// EngineResult holds the final output of a conversation engine run.
type EngineResult struct {
	Response             string                 `json:"response"`
	ToolCallResults      []tools.ToolCallResult `json:"tool_call_results,omitempty"`
	RequiresConfirmation bool                   `json:"requires_confirmation,omitempty"`
	PendingToolCall      *tools.ToolCall        `json:"pending_tool_call,omitempty"`
	TotalDurationMs      int64                  `json:"total_duration_ms"`
	TurnCount            int                    `json:"turn_count"`
	Events               []tools.StreamEvent    `json:"events,omitempty"`
}

// Engine is the multi-turn conversation engine with tool loop.
type Engine struct {
	cerebrasClient *cerebras.Client
	toolRegistry   *tools.Registry
	contextBuilder *aiContext.Builder
	normalizer     *MessageNormalizer
}

// NewEngine creates a new conversation engine.
func NewEngine(client *cerebras.Client, registry *tools.Registry, builder *aiContext.Builder) *Engine {
	return &Engine{
		cerebrasClient: client,
		toolRegistry:   registry,
		contextBuilder: builder,
		normalizer:     NewMessageNormalizer(),
	}
}

// ProcessMessage runs the synchronous multi-turn conversation loop.
// For each turn: send messages → parse response → if tool_call → execute → feed result → repeat.
func (e *Engine) ProcessMessage(
	ctx context.Context,
	userMessage string,
	history []ConversationMessage,
	userCtx *aiContext.UserContext,
	model string,
) (*EngineResult, error) {
	start := apptime.Now()

	// Filter tools by user permissions (Claude Code pattern: pre-filter before model sees them)
	var permissions map[string]bool
	var isAdmin bool
	if userCtx != nil {
		permissions = userCtx.Permissions
		isAdmin = userCtx.IsAdmin
	}
	availableTools := e.toolRegistry.FilterByPermissions(permissions, isAdmin)

	// Build system prompt
	systemPrompt := e.contextBuilder.BuildFlatSystemPrompt(availableTools, userCtx, e.toolRegistry)

	// Normalize history for context window
	normalizedHistory := e.normalizer.Normalize(history, MaxContextMessages)

	// Build the messages array for the LLM
	messages := buildLLMMessages(systemPrompt, normalizedHistory, userMessage)

	result := &EngineResult{}

	// Multi-turn tool loop
	for turn := 0; turn < MaxToolTurns; turn++ {
		result.TurnCount = turn + 1

		// Call the LLM
		chatResp, err := e.cerebrasClient.Chat(&cerebras.ChatRequest{
			Model:       model,
			Messages:    messages,
			Temperature: 0.3,
			MaxTokens:   MaxCompletionTokens,
		})
		if err != nil {
			return nil, fmt.Errorf("LLM call failed on turn %d: %w", turn+1, err)
		}

		llmResponse := chatResp.Message.Content

		// Parse for tool calls
		textBefore, toolCall, textAfter := ParseToolCall(llmResponse)

		if toolCall == nil {
			// No tool call — this is the final response.
			// Use textBefore to avoid persisting leaked/truncated tool artifacts.
			finalResponse := strings.TrimSpace(textBefore)
			if hasIncompleteToolCall(llmResponse) {
				finalResponse = incompleteToolCallRecoveryMessage
			}
			if finalResponse == "" {
				finalResponse = strings.TrimSpace(llmResponse)
			}
			result.Response = finalResponse
			result.TotalDurationMs = time.Since(start).Milliseconds()
			return result, nil
		}

		// Check if tool requires confirmation
		tool := e.toolRegistry.Get(toolCall.Name)
		if tool != nil && tool.NeedsConfirmation() {
			result.RequiresConfirmation = true
			result.PendingToolCall = toolCall
			result.Response = strings.TrimSpace(textBefore)
			if result.Response == "" {
				result.Response = fmt.Sprintf("I'd like to execute **%s**. Please confirm.", toolCall.Name)
			}
			result.TotalDurationMs = time.Since(start).Milliseconds()
			return result, nil
		}

		// Execute the tool
		toolResult, execErr := e.executeTool(ctx, toolCall, userCtx)

		tcResult := tools.ToolCallResult{
			Call:   toolCall,
			Result: toolResult,
		}
		if execErr != nil {
			tcResult.Error = execErr.Error()
		}
		result.ToolCallResults = append(result.ToolCallResults, tcResult)

		// Build tool result message to feed back
		toolResultContent := formatToolResult(toolCall, toolResult, execErr)

		// Add assistant response + tool result to messages for next turn
		assistantContent := textBefore
		if assistantContent == "" {
			assistantContent = fmt.Sprintf("[Calling tool: %s]", toolCall.Name)
		}
		messages = append(messages, cerebras.ChatMessage{Role: "assistant", Content: assistantContent})
		messages = append(messages, cerebras.ChatMessage{Role: "user", Content: toolResultContent})

		_ = textAfter // Discard post-tool text per protocol rules
	}

	// Exceeded max turns
	result.Response = "I've completed the maximum number of tool operations for this request. Please let me know if you need anything else."
	result.TotalDurationMs = time.Since(start).Milliseconds()
	return result, nil
}

// ProcessMessageStream runs the conversation loop with SSE streaming.
// Returns an EngineResult containing the accumulated response for DB persistence.
func (e *Engine) ProcessMessageStream(
	ctx context.Context,
	userMessage string,
	history []ConversationMessage,
	userCtx *aiContext.UserContext,
	model string,
	eventChan chan<- tools.StreamEvent,
) (*EngineResult, error) {
	start := apptime.Now()

	var permissions map[string]bool
	var isAdmin bool
	if userCtx != nil {
		permissions = userCtx.Permissions
		isAdmin = userCtx.IsAdmin
	}
	availableTools := e.toolRegistry.FilterByPermissions(permissions, isAdmin)
	systemPrompt := e.contextBuilder.BuildFlatSystemPrompt(availableTools, userCtx, e.toolRegistry)
	normalizedHistory := e.normalizer.Normalize(history, MaxContextMessages)
	messages := buildLLMMessages(systemPrompt, normalizedHistory, userMessage)

	result := &EngineResult{}

	for turn := 0; turn < MaxToolTurns; turn++ {
		result.TurnCount = turn + 1

		var fullResponse strings.Builder
		// emittedLen tracks how many bytes of fullResponse have been sent as content_delta.
		// This ensures we never emit <tool_call>...</tool_call> tokens to the frontend.
		emittedLen := 0
		toolCallStartIdx := -1

		_, streamErr := e.cerebrasClient.ChatStream(&cerebras.ChatRequest{
			Model:       model,
			Messages:    messages,
			Temperature: 0.3,
			MaxTokens:   MaxCompletionTokens,
		}, func(chunk string) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			fullResponse.WriteString(chunk)
			current := fullResponse.String()

			// Once we've found <tool_call>, stop emitting deltas (tool tags must not reach UI)
			if toolCallStartIdx >= 0 {
				return nil
			}

			idx := findToolCallMarkerIndex(current)
			if idx >= 0 {
				toolCallStartIdx = idx
				// Emit any text before the <tool_call> that hasn't been sent yet
				if idx > emittedLen {
					toEmit := current[emittedLen:idx]
					if toEmit != "" {
						eventChan <- tools.StreamEvent{
							Type:    tools.EventContentDelta,
							Content: toEmit,
						}
						emittedLen = idx
					}
				}
				return nil
			}

			// Safe to emit the new content since no tool_call detected yet
			if len(current) > emittedLen {
				toEmit := current[emittedLen:]
				eventChan <- tools.StreamEvent{
					Type:    tools.EventContentDelta,
					Content: toEmit,
				}
				emittedLen = len(current)
			}
			return nil
		})

		if streamErr != nil {
			eventChan <- tools.StreamEvent{
				Type:    tools.EventError,
				Content: formatStreamFailureMessage(streamErr),
			}
			return nil, fmt.Errorf("LLM streaming failed: %w", streamErr)
		}

		responseText := fullResponse.String()
		textBefore, toolCall, _ := ParseToolCall(responseText)

		if toolCall == nil {
			// Final response — use textBefore so that any truncated <tool_call>
			// fragment (stream cut off before </tool_call>) is silently dropped
			// and never persisted to the database.
			finalResponse := strings.TrimSpace(textBefore)
			incomplete := hasIncompleteToolCall(responseText)
			if incomplete {
				finalResponse = incompleteToolCallRecoveryMessage
				if emittedLen == 0 {
					eventChan <- tools.StreamEvent{Type: tools.EventContentDelta, Content: finalResponse}
				} else {
					eventChan <- tools.StreamEvent{Type: tools.EventContentDelta, Content: "\n\n" + finalResponse}
				}
			}
			if finalResponse == "" {
				finalResponse = strings.TrimSpace(responseText)
			}
			result.Response = finalResponse
			result.TotalDurationMs = time.Since(start).Milliseconds()
			eventChan <- tools.StreamEvent{
				Type: tools.EventMessageEnd,
				Data: map[string]interface{}{
					"duration_ms": result.TotalDurationMs,
					"turn_count":  result.TurnCount,
				},
			}
			return result, nil
		}

		// Emit tool_call event (visible in UI as a tool status card)
		eventChan <- tools.StreamEvent{
			Type: tools.EventToolCall,
			Data: map[string]interface{}{
				"name":       toolCall.Name,
				"parameters": toolCall.Parameters,
			},
		}

		// Check confirmation requirement
		tool := e.toolRegistry.Get(toolCall.Name)
		if tool != nil && tool.NeedsConfirmation() {
			result.RequiresConfirmation = true
			result.PendingToolCall = toolCall
			result.Response = strings.TrimSpace(textBefore)
			if result.Response == "" {
				result.Response = fmt.Sprintf("I'd like to execute **%s**. Please confirm.", toolCall.Name)
				// If the model emitted only a tool_call block (no text prefix), emit
				// a fallback confirmation text so the UI never appears blank.
				eventChan <- tools.StreamEvent{
					Type:    tools.EventContentDelta,
					Content: result.Response,
				}
			}
			result.TotalDurationMs = time.Since(start).Milliseconds()
			eventChan <- tools.StreamEvent{
				Type: tools.EventMessageEnd,
				Data: map[string]interface{}{
					"requires_confirmation": true,
					"pending_tool_call":     toolCall,
					"duration_ms":           result.TotalDurationMs,
					"turn_count":            result.TurnCount,
				},
			}
			return result, nil
		}

		// Execute tool and emit result
		toolResult, execErr := e.executeTool(ctx, toolCall, userCtx)

		tcResult := tools.ToolCallResult{Call: toolCall, Result: toolResult}
		if execErr != nil {
			tcResult.Error = execErr.Error()
		}
		result.ToolCallResults = append(result.ToolCallResults, tcResult)

		eventChan <- tools.StreamEvent{
			Type: tools.EventToolResult,
			Data: map[string]interface{}{
				"call":   toolCall,
				"result": toolResult,
				"error":  execErrString(execErr),
			},
		}

		// Feed result back to LLM for next turn
		toolResultContent := formatToolResult(toolCall, toolResult, execErr)
		assistantContent := textBefore
		if assistantContent == "" {
			assistantContent = fmt.Sprintf("[Calling tool: %s]", toolCall.Name)
		}
		messages = append(messages, cerebras.ChatMessage{Role: "assistant", Content: assistantContent})
		messages = append(messages, cerebras.ChatMessage{Role: "user", Content: toolResultContent})
	}

	result.Response = "I've completed the maximum number of tool operations for this request."
	result.TotalDurationMs = time.Since(start).Milliseconds()
	eventChan <- tools.StreamEvent{
		Type: tools.EventMessageEnd,
		Data: map[string]interface{}{
			"duration_ms": result.TotalDurationMs,
			"turn_count":  MaxToolTurns,
			"max_turns":   true,
		},
	}
	return result, nil
}

func formatStreamFailureMessage(err error) string {
	if err == nil {
		return "Respons terhenti karena gangguan saat memproses jawaban. Silakan kirim ulang pesan Anda."
	}

	errMsg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errMsg, "context_length_exceeded") || strings.Contains(errMsg, "please reduce the length of the messages"):
		ctxLenPattern := regexp.MustCompile(`current length is\s*(\d+)\s*while limit is\s*(\d+)`)
		matches := ctxLenPattern.FindStringSubmatch(errMsg)
		if len(matches) == 3 {
			return fmt.Sprintf("Respons terhenti karena konteks percakapan terlalu panjang (panjang saat ini %s, batas model %s). Silakan mulai chat baru atau ringkas pesan terakhir Anda.", matches[1], matches[2])
		}
		return "Respons terhenti karena konteks percakapan terlalu panjang untuk model saat ini. Silakan mulai chat baru atau ringkas pesan terakhir Anda."
	case strings.Contains(errMsg, "model_not_found") || strings.Contains(errMsg, "does not exist or you do not have access to it"):
		return "Model AI yang dipilih tidak tersedia. Silakan pilih model lain lalu kirim ulang permintaan Anda."
	default:
		return "Respons terhenti karena gangguan sementara pada layanan AI. Silakan coba lagi."
	}
}

// executeTool finds and runs a tool by name.
func (e *Engine) executeTool(ctx context.Context, call *tools.ToolCall, userCtx *aiContext.UserContext) (*tools.ToolResult, error) {
	tool := e.toolRegistry.Get(call.Name)
	if tool == nil {
		return &tools.ToolResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Unknown tool: %s", call.Name),
			Action:       "error",
		}, fmt.Errorf("tool '%s' not found in registry", call.Name)
	}

	execCtx := &tools.ExecutionContext{
		UserID: "",
	}
	if userCtx != nil {
		execCtx.UserID = userCtx.UserID
		execCtx.UserPermissions = userCtx.Permissions
		execCtx.IsAdmin = userCtx.IsAdmin
	}

	return tool.Execute(ctx, call.Parameters, execCtx)
}

// buildLLMMessages constructs the messages array for the Cerebras API.
func buildLLMMessages(systemPrompt string, history []ConversationMessage, userMessage string) []cerebras.ChatMessage {
	messages := make([]cerebras.ChatMessage, 0, len(history)+2)

	// System prompt
	messages = append(messages, cerebras.ChatMessage{Role: "system", Content: systemPrompt})

	// History
	for _, msg := range history {
		messages = append(messages, cerebras.ChatMessage{Role: msg.Role, Content: msg.Content})
	}

	// Current user message
	messages = append(messages, cerebras.ChatMessage{Role: "user", Content: userMessage})

	return messages
}

// formatToolResult formats the tool execution result as a message for the LLM.
func formatToolResult(call *tools.ToolCall, result *tools.ToolResult, execErr error) string {
	var sb strings.Builder
	sb.WriteString("<tool_result>\n")

	if execErr != nil {
		sb.WriteString(fmt.Sprintf(`{"tool": "%s", "success": false, "error": "%s"}`, call.Name, execErr.Error()))
	} else if result != nil {
		data, _ := json.Marshal(map[string]interface{}{
			"tool":    call.Name,
			"success": result.Success,
			"message": result.Message,
			"data":    result.Data,
		})
		sb.Write(data)
	} else {
		sb.WriteString(fmt.Sprintf(`{"tool": "%s", "success": false, "error": "no result returned"}`, call.Name))
	}

	sb.WriteString("\n</tool_result>")
	return sb.String()
}

func execErrString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

func findToolCallMarkerIndex(content string) int {
	markerIdx := strings.Index(content, "<tool_call>")
	legacyIdx := -1
	if idx := legacyToolCallMarkerPattern.FindStringIndex(content); len(idx) == 2 {
		legacyIdx = idx[0]
	}

	if markerIdx >= 0 && legacyIdx >= 0 {
		if markerIdx < legacyIdx {
			return markerIdx
		}
		return legacyIdx
	}
	if markerIdx >= 0 {
		return markerIdx
	}
	return legacyIdx
}

func hasIncompleteToolCall(content string) bool {
	if strings.Contains(content, "<tool_call>") && !strings.Contains(content, "</tool_call>") {
		return true
	}

	matches := legacyToolCallMarkerPattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return false
	}

	lowerContent := strings.ToLower(content)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		toolName := strings.ToLower(match[1])
		if toolName == "" {
			continue
		}
		closingTag := fmt.Sprintf("</%s>", toolName)
		if !strings.Contains(lowerContent, closingTag) {
			return true
		}
	}

	return false
}
