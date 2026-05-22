package handler

import (
	"github.com/gilabs/gims/api/internal/ai/domain/dto"
	"github.com/gilabs/gims/api/internal/ai/domain/usecase"
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gin-gonic/gin"
)

// AdminHandler handles AI admin endpoints (action logs, intent registry)
type AdminHandler struct {
	usecase usecase.AIChatUsecase
}

// NewAdminHandler creates a new AdminHandler
func NewAdminHandler(uc usecase.AIChatUsecase) *AdminHandler {
	return &AdminHandler{usecase: uc}
}

// ListActions handles GET /ai/admin/actions
func (h *AdminHandler) ListActions(c *gin.Context) {
	var req dto.ListActionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		errors.HandleValidationError(c, err)
		return
	}

	actions, pagination, err := h.usecase.ListActions(c.Request.Context(), &req)
	if err != nil {
		handleAIError(c, err)
		return
	}

	meta := &response.Meta{
		Pagination: response.NewPaginationMeta(pagination.Page, pagination.PerPage, pagination.Total),
	}
	response.SuccessResponse(c, actions, meta)
}

// GetIntentRegistry handles GET /ai/admin/intents
func (h *AdminHandler) GetIntentRegistry(c *gin.Context) {
	intents, err := h.usecase.GetIntentRegistry(c.Request.Context())
	if err != nil {
		handleAIError(c, err)
		return
	}

	response.SuccessResponse(c, intents, nil)
}
