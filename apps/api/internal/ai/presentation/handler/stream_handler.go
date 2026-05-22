package handler

import (
	"encoding/json"
	"io"
	"log"
	"strings"

	"github.com/gilabs/gims/api/internal/ai/domain/dto"
	"github.com/gilabs/gims/api/internal/ai/domain/usecase"
	"github.com/gilabs/gims/api/internal/ai/domain/usecase/tools"
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/infrastructure/cerebras"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gin-gonic/gin"
)

// StreamHandler handles the SSE streaming chat endpoint and v2 engine endpoints.
// This mirrors Claude Code's streaming response architecture where events
// are emitted in real-time: message_start, content_delta, tool_call,
// tool_result, message_end.
type StreamHandler struct {
	engineUsecase  *usecase.ChatEngineUsecase
	cerebrasClient *cerebras.Client
}

// NewStreamHandler creates a new streaming handler.
func NewStreamHandler(engineUsecase *usecase.ChatEngineUsecase, client *cerebras.Client) *StreamHandler {
	return &StreamHandler{
		engineUsecase:  engineUsecase,
		cerebrasClient: client,
	}
}

// StreamMessage handles POST /ai/chat/stream — SSE streaming chat endpoint.
// Events follow the Claude Code streaming protocol:
//   - message_start:  { session_id }
//   - content_delta:  { content: "chunk" }
//   - tool_call:      { name, parameters }
//   - tool_result:    { call, result }
//   - thinking:       { content: "reasoning" }
//   - message_end:    { duration_ms, turn_count }
//   - error:          { content: "error message" }
func (h *StreamHandler) StreamMessage(c *gin.Context) {
	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleValidationError(c, err)
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		errors.UnauthorizedResponse(c, "authentication required")
		return
	}

	permissions := extractStreamPermissions(c)
	isAdmin := c.GetString("user_role") == "admin"
	requestID := c.GetString("request_id")
	sessionID := ""
	if req.SessionID != nil {
		sessionID = *req.SessionID
	}
	modelName := req.Model
	if strings.TrimSpace(modelName) == "" {
		modelName = "(default)"
	}

	log.Printf(
		"[AI_STREAM] start request_id=%s user_id=%s session_id=%s model=%s message_len=%d message_preview=%q",
		requestID,
		userID,
		sessionID,
		modelName,
		len(req.Message),
		previewMessage(req.Message),
	)

	reqCtx := c.Request.Context()

	// Set SSE headers before streaming
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering

	eventChan := make(chan tools.StreamEvent, 64)
	doneChan := make(chan error, 1)

	go func() {
		defer close(eventChan)
		err := h.engineUsecase.SendMessageStream(
			reqCtx,
			&req,
			userID,
			permissions,
			isAdmin,
			eventChan,
		)
		if err != nil {
			log.Printf(
				"[AI_STREAM] usecase_error request_id=%s user_id=%s session_id=%s err=%v context_err=%v",
				requestID,
				userID,
				sessionID,
				err,
				reqCtx.Err(),
			)
		}

		if err != nil && reqCtx.Err() == nil {
			if !strings.Contains(err.Error(), "LLM streaming failed") {
				event := tools.StreamEvent{
					Type:    tools.EventError,
					Content: buildStreamErrorMessage(err),
				}
				select {
				case eventChan <- event:
				default:
				}
			}
		}
		doneChan <- err
	}()

	// Stream events to client as SSE
	c.Stream(func(w io.Writer) bool {
		select {
		case event, ok := <-eventChan:
			if !ok {
				return false
			}
			if event.Type == tools.EventError {
				log.Printf(
					"[AI_STREAM] event_error request_id=%s user_id=%s session_id=%s content=%q",
					requestID,
					userID,
					sessionID,
					event.Content,
				)
			}
			writeSSEEvent(c, event)
			return true

		case <-reqCtx.Done():
			log.Printf(
				"[AI_STREAM] client_disconnected request_id=%s user_id=%s session_id=%s reason=%v",
				requestID,
				userID,
				sessionID,
				reqCtx.Err(),
			)
			return false
		}
	})

	// Drain any remaining processing error
	if err := <-doneChan; err != nil {
		log.Printf(
			"[AI_STREAM] completed_with_error request_id=%s user_id=%s session_id=%s err=%v",
			requestID,
			userID,
			sessionID,
			err,
		)
		return
	}

	log.Printf(
		"[AI_STREAM] completed request_id=%s user_id=%s session_id=%s",
		requestID,
		userID,
		sessionID,
	)
}

// writeSSEEvent marshals a StreamEvent and writes it in SSE format.
func writeSSEEvent(c *gin.Context, event tools.StreamEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	c.SSEvent(string(event.Type), string(data))
	c.Writer.Flush()
}

// extractStreamPermissions reads user_permissions from gin context.
// Separate function to avoid name clash with the legacy handler's extractPermissions.
func extractStreamPermissions(c *gin.Context) map[string]bool {
	if perms, exists := c.Get("user_permissions"); exists {
		if permMap, ok := perms.(map[string]bool); ok {
			return permMap
		}
	}
	return map[string]bool{}
}

// SendMessageV2 handles POST /ai/chat/v2/send — non-streaming engine endpoint.
func (h *StreamHandler) SendMessageV2(c *gin.Context) {
	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleValidationError(c, err)
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		errors.UnauthorizedResponse(c, "authentication required")
		return
	}

	permissions := extractStreamPermissions(c)
	isAdmin := c.GetString("user_role") == "admin"

	result, err := h.engineUsecase.SendMessage(
		c.Request.Context(),
		&req,
		userID,
		permissions,
		isAdmin,
	)
	if err != nil {
		handleEngineError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// ConfirmActionV2 handles POST /ai/chat/v2/confirm — engine-based confirmation.
func (h *StreamHandler) ConfirmActionV2(c *gin.Context) {
	var req dto.ConfirmActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleValidationError(c, err)
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		errors.UnauthorizedResponse(c, "authentication required")
		return
	}

	permissions := extractStreamPermissions(c)
	isAdmin := c.GetString("user_role") == "admin"

	result, err := h.engineUsecase.ConfirmAction(
		c.Request.Context(),
		&req,
		userID,
		permissions,
		isAdmin,
	)
	if err != nil {
		handleEngineError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// ListModelsV2 handles GET /ai/v2/models
func (h *StreamHandler) ListModelsV2(c *gin.Context) {
	models := h.cerebrasClient.AvailableModels()
	response.SuccessResponse(c, models, nil)
}

// handleEngineError maps engine usecase errors to HTTP responses.
func handleEngineError(c *gin.Context, err error) {
	errMsg := err.Error()
	switch {
	case strings.Contains(errMsg, "AI_SERVICE_NOT_CONFIGURED"):
		errors.ErrorResponse(c, "AI_SERVICE_NOT_CONFIGURED",
			map[string]interface{}{"message": "AI service is not configured. Please contact administrator."}, nil)
	case strings.Contains(errMsg, "FORBIDDEN"):
		errors.ForbiddenResponse(c, "insufficient permissions for AI action", nil)
	case strings.Contains(errMsg, "AI_SESSION_NOT_FOUND"):
		errors.ErrorResponse(c, "AI_SESSION_NOT_FOUND",
			map[string]interface{}{"message": "Chat session not found."}, nil)
	case strings.Contains(errMsg, "model_not_found") || strings.Contains(errMsg, "does not exist or you do not have access to it"):
		errors.ErrorResponse(c, "AI_MODEL_NOT_AVAILABLE",
			map[string]interface{}{"message": "Selected AI model is not available. Please choose another model."}, nil)
	case strings.Contains(errMsg, "context_length_exceeded") || strings.Contains(errMsg, "Please reduce the length of the messages"):
		errors.ErrorResponse(c, "AI_CONTEXT_LENGTH_EXCEEDED",
			map[string]interface{}{"message": "Conversation context is too long. Please start a new chat or shorten the latest message."}, nil)
	case strings.Contains(errMsg, "AI_ACTION_NOT_FOUND"):
		errors.ErrorResponse(c, "AI_ACTION_NOT_FOUND",
			map[string]interface{}{"message": "Action not found."}, nil)
	case strings.Contains(errMsg, "AI_ACTION_INVALID_STATE"):
		errors.ErrorResponse(c, "VALIDATION_ERROR",
			map[string]interface{}{"message": errMsg}, nil)
	default:
		errors.InternalServerErrorResponse(c, "an unexpected error occurred")
	}
}

// buildStreamErrorMessage normalizes internal engine errors into user-facing text.
func buildStreamErrorMessage(err error) string {
	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "AI_SERVICE_NOT_CONFIGURED"):
		return "AI service is not configured. Please contact administrator."
	case strings.Contains(errMsg, "FORBIDDEN"):
		return "You do not have permission to perform this action."
	case strings.Contains(errMsg, "AI_SESSION_NOT_FOUND"):
		return "Chat session not found."
	case strings.Contains(errMsg, "model_not_found") || strings.Contains(errMsg, "does not exist or you do not have access to it"):
		return "Selected AI model is not available. Please choose another model."
	case strings.Contains(errMsg, "context_length_exceeded") || strings.Contains(errMsg, "Please reduce the length of the messages"):
		return "Conversation context is too long. Please start a new chat or shorten the latest message."
	case strings.Contains(errMsg, "AI_ACTION_NOT_FOUND"):
		return "Action not found."
	default:
		return "An error occurred while processing your message."
	}
}

func previewMessage(input string) string {
	trimmed := strings.TrimSpace(input)
	if len(trimmed) <= 160 {
		return trimmed
	}
	return trimmed[:160] + "..."
}
