package handler

import (
	"strings"

	"github.com/gilabs/gims/api/internal/ai/domain/dto"
	"github.com/gilabs/gims/api/internal/ai/domain/usecase"
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/infrastructure/cerebras"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gin-gonic/gin"
)

// ChatHandler handles AI chat HTTP requests
type ChatHandler struct {
	usecase        usecase.AIChatUsecase
	cerebrasClient *cerebras.Client
}

// NewChatHandler creates a new ChatHandler
func NewChatHandler(uc usecase.AIChatUsecase, client *cerebras.Client) *ChatHandler {
	return &ChatHandler{usecase: uc, cerebrasClient: client}
}

// SendMessage handles POST /ai/chat/send
func (h *ChatHandler) SendMessage(c *gin.Context) {
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

	// Extract user permissions from context (set by auth middleware)
	userPermissions := extractPermissions(c)
	isAdmin := c.GetString("user_role") == "admin"

	resp, err := h.usecase.SendMessage(c.Request.Context(), &req, userID, userPermissions, isAdmin)
	if err != nil {
		handleAIError(c, err)
		return
	}

	response.SuccessResponse(c, resp, nil)
}

// ConfirmAction handles POST /ai/chat/confirm
func (h *ChatHandler) ConfirmAction(c *gin.Context) {
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

	userPermissions := extractPermissions(c)
	isAdmin := c.GetString("user_role") == "admin"

	resp, err := h.usecase.ConfirmAction(c.Request.Context(), &req, userID, userPermissions, isAdmin)
	if err != nil {
		handleAIError(c, err)
		return
	}

	response.SuccessResponse(c, resp, nil)
}

// ListModels handles GET /ai/models
func (h *ChatHandler) ListModels(c *gin.Context) {
	models := h.cerebrasClient.AvailableModels()
	response.SuccessResponse(c, models, nil)
}

// extractPermissions retrieves the permission map from gin context
func extractPermissions(c *gin.Context) map[string]bool {
	if perms, exists := c.Get("user_permissions"); exists {
		if permMap, ok := perms.(map[string]bool); ok {
			return permMap
		}
	}
	return map[string]bool{}
}

// handleAIError maps usecase errors to appropriate HTTP responses
func handleAIError(c *gin.Context, err error) {
	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "AI_SERVICE_NOT_CONFIGURED"):
		errors.ErrorResponse(c, "AI_SERVICE_NOT_CONFIGURED", map[string]interface{}{"message": "AI service is not configured. Please contact administrator."}, nil)
	case strings.Contains(errMsg, "AI_SESSION_NOT_FOUND"):
		errors.ErrorResponse(c, "AI_SESSION_NOT_FOUND", map[string]interface{}{"message": "Chat session not found."}, nil)
	case strings.Contains(errMsg, "AI_ACTION_NOT_FOUND"):
		errors.ErrorResponse(c, "AI_ACTION_NOT_FOUND", map[string]interface{}{"message": "Action not found."}, nil)
	case strings.Contains(errMsg, "AI_ACTION_INVALID_STATE"):
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": errMsg}, nil)
	case strings.Contains(errMsg, "FORBIDDEN"):
		errors.ForbiddenResponse(c, "insufficient permissions for AI action", nil)
	case strings.Contains(errMsg, "AI_CHAT_FAILED"):
		errors.ErrorResponse(c, "AI_CHAT_FAILED", map[string]interface{}{"message": "An error occurred while processing your message."}, nil)
	default:
		errors.InternalServerErrorResponse(c, "an unexpected error occurred")
	}
}
