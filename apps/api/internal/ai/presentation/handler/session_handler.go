package handler

import (
	"github.com/gilabs/gims/api/internal/ai/domain/dto"
	"github.com/gilabs/gims/api/internal/ai/domain/usecase"
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gin-gonic/gin"
)

// SessionHandler handles AI chat session HTTP requests
type SessionHandler struct {
	usecase usecase.AIChatUsecase
}

// NewSessionHandler creates a new SessionHandler
func NewSessionHandler(uc usecase.AIChatUsecase) *SessionHandler {
	return &SessionHandler{usecase: uc}
}

// ListSessions handles GET /ai/sessions
func (h *SessionHandler) ListSessions(c *gin.Context) {
	var req dto.ListSessionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		errors.HandleValidationError(c, err)
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		errors.UnauthorizedResponse(c, "authentication required")
		return
	}

	sessions, pagination, err := h.usecase.ListSessions(c.Request.Context(), &req, userID)
	if err != nil {
		handleAIError(c, err)
		return
	}

	meta := &response.Meta{
		Pagination: response.NewPaginationMeta(pagination.Page, pagination.PerPage, pagination.Total),
	}
	response.SuccessResponse(c, sessions, meta)
}

// GetSessionDetail handles GET /ai/sessions/:id
func (h *SessionHandler) GetSessionDetail(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": "Session ID is required"}, nil)
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		errors.UnauthorizedResponse(c, "authentication required")
		return
	}

	session, err := h.usecase.GetSessionDetail(c.Request.Context(), sessionID, userID)
	if err != nil {
		handleAIError(c, err)
		return
	}

	response.SuccessResponse(c, session, nil)
}

// DeleteSession handles DELETE /ai/sessions/:id
func (h *SessionHandler) DeleteSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": "Session ID is required"}, nil)
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		errors.UnauthorizedResponse(c, "authentication required")
		return
	}

	if err := h.usecase.DeleteSession(c.Request.Context(), sessionID, userID); err != nil {
		handleAIError(c, err)
		return
	}

	response.SuccessResponseDeleted(c, "session", sessionID, nil)
}
