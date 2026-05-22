package handler

import (
	"errors"
	"net/http"

	coreErrors "github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
	"github.com/gilabs/gims/api/internal/crm/domain/usecase"
	"github.com/gin-gonic/gin"
)

// LeadAutomationHandler handles server-side n8n automation calls for lead generation.
type LeadAutomationHandler struct {
	uc usecase.LeadAutomationUsecase
}

// NewLeadAutomationHandler creates a LeadAutomationHandler.
func NewLeadAutomationHandler(uc usecase.LeadAutomationUsecase) *LeadAutomationHandler {
	return &LeadAutomationHandler{uc: uc}
}

// TestConnection checks whether configured n8n webhook endpoint is reachable.
func (h *LeadAutomationHandler) TestConnection(c *gin.Context) {
	result, err := h.uc.TestConnection(c.Request.Context())
	if err != nil {
		h.handleAutomationError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// Trigger executes n8n lead generation webhook via backend-controlled endpoint.
func (h *LeadAutomationHandler) Trigger(c *gin.Context) {
	var req dto.LeadAutomationTriggerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	result, err := h.uc.Trigger(c.Request.Context(), req)
	if err != nil {
		h.handleAutomationError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

func (h *LeadAutomationHandler) handleAutomationError(c *gin.Context, err error) {
	if errors.Is(err, usecase.ErrN8NNotConfigured) {
		coreErrors.ErrorResponse(c, "SERVICE_UNAVAILABLE", map[string]interface{}{
			"message": "N8N_LEADS_WEBHOOK_URL is not configured on API server",
		}, nil)
		return
	}

	var upstreamErr *usecase.LeadAutomationUpstreamError
	if errors.As(err, &upstreamErr) {
		switch {
		case upstreamErr.Status == http.StatusNotFound:
			coreErrors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{
				"message":              "n8n webhook is not registered or workflow is inactive",
				"upstream_status":      upstreamErr.Status,
				"upstream_text":        upstreamErr.Body,
				"executed_webhook_url": upstreamErr.ExecutedWebhookURL,
			}, nil)
		case upstreamErr.Status >= http.StatusBadGateway:
			coreErrors.ErrorResponse(c, "SERVICE_UNAVAILABLE", map[string]interface{}{
				"message":              "n8n upstream service is unavailable",
				"upstream_status":      upstreamErr.Status,
				"upstream_text":        upstreamErr.Body,
				"executed_webhook_url": upstreamErr.ExecutedWebhookURL,
			}, nil)
		default:
			coreErrors.ErrorResponse(c, "INVALID_REQUEST_BODY", map[string]interface{}{
				"message":              "n8n rejected the request",
				"upstream_status":      upstreamErr.Status,
				"upstream_text":        upstreamErr.Body,
				"executed_webhook_url": upstreamErr.ExecutedWebhookURL,
			}, nil)
		}
		return
	}

	coreErrors.InternalServerErrorResponse(c, err.Error())
}
