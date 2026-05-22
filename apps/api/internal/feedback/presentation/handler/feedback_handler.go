package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/feedback/domain/dto"
	"github.com/gilabs/gims/api/internal/feedback/domain/usecase"
	"github.com/gin-gonic/gin"
)

// FeedbackHandler handles all feedback-related HTTP requests.
type FeedbackHandler struct {
	uc usecase.FeedbackUsecase
}

// NewFeedbackHandler creates a new FeedbackHandler.
func NewFeedbackHandler(uc usecase.FeedbackUsecase) *FeedbackHandler {
	return &FeedbackHandler{uc: uc}
}

// ─── Form Management ──────────────────────────────────────────────────────────

// CreateForm creates a new feedback form for an outlet.
func (h *FeedbackHandler) CreateForm(c *gin.Context) {
	var req dto.CreateFeedbackFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErrorResponse(c, nil)
		return
	}

	userID, _ := c.Get("user_id")
	createdBy, _ := userID.(string)

	form, err := h.uc.CreateForm(c.Request.Context(), createdBy, &req)
	if err != nil {
		handleFeedbackError(c, err)
		return
	}
	response.SuccessResponseCreated(c, form, nil)
}

// UpdateForm updates an existing feedback form.
func (h *FeedbackHandler) UpdateForm(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateFeedbackFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErrorResponse(c, nil)
		return
	}

	userID, _ := c.Get("user_id")
	updatedBy, _ := userID.(string)

	form, err := h.uc.UpdateForm(c.Request.Context(), id, updatedBy, &req)
	if err != nil {
		handleFeedbackError(c, err)
		return
	}
	response.SuccessResponse(c, form, nil)
}

// DeleteForm soft-deletes a feedback form.
func (h *FeedbackHandler) DeleteForm(c *gin.Context) {
	id := c.Param("id")
	if err := h.uc.DeleteForm(c.Request.Context(), id); err != nil {
		handleFeedbackError(c, err)
		return
	}
	response.SuccessResponseNoContent(c)
}

// GetForm returns a single feedback form by ID.
func (h *FeedbackHandler) GetForm(c *gin.Context) {
	id := c.Param("id")
	form, err := h.uc.GetForm(c.Request.Context(), id)
	if err != nil {
		handleFeedbackError(c, err)
		return
	}
	response.SuccessResponse(c, form, nil)
}

// ListForms returns a paginated list of all feedback forms.
func (h *FeedbackHandler) ListForms(c *gin.Context) {
	page := 1
	perPage := 20
	outletID := c.Query("outlet_id")
	if p := c.Query("page"); p != "" {
		if v, err := parseInt(p); err == nil && v > 0 {
			page = v
		}
	}
	if pp := c.Query("per_page"); pp != "" {
		if v, err := parseInt(pp); err == nil && v > 0 && v <= 100 {
			perPage = v
		}
	}

	forms, pagination, err := h.uc.ListForms(c.Request.Context(), page, perPage, outletID)
	if err != nil {
		handleFeedbackError(c, err)
		return
	}

	meta := &response.Meta{
		Pagination: response.NewPaginationMeta(pagination.Page, pagination.PerPage, pagination.Total),
	}
	response.SuccessResponse(c, forms, meta)
}

// GetFormsByOutlet returns all forms for a specific outlet (query param outlet_id).
func (h *FeedbackHandler) GetFormsByOutlet(c *gin.Context) {
	outletID := c.Query("outlet_id")
	if outletID == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "outlet_id is required", nil, nil)
		return
	}
	forms, err := h.uc.GetFormsByOutlet(c.Request.Context(), outletID)
	if err != nil {
		handleFeedbackError(c, err)
		return
	}
	response.SuccessResponse(c, forms, nil)
}

// CopyForm clones an existing feedback form to one or more outlets.
func (h *FeedbackHandler) CopyForm(c *gin.Context) {
	formID := c.Param("id")
	var req dto.CopyFeedbackFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErrorResponse(c, nil)
		return
	}

	userID, _ := c.Get("user_id")
	createdBy, _ := userID.(string)

	copyResp, err := h.uc.CopyForm(c.Request.Context(), formID, createdBy, &req)
	if err != nil {
		handleFeedbackError(c, err)
		return
	}

	response.SuccessResponseCreated(c, copyResp, nil)
}

// ─── Token ────────────────────────────────────────────────────────────────────

// GenerateToken generates a one-time feedback token for a POS transaction.
// This endpoint is called by the POS receipt handler, not directly by end users.
func (h *FeedbackHandler) GenerateToken(c *gin.Context) {
	var req dto.GenerateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErrorResponse(c, nil)
		return
	}

	appBaseURL := strings.TrimRight(config.AppConfig.Server.FrontendBaseURL, "/") + "/en"
	tokenResp, err := h.uc.GenerateToken(c.Request.Context(), &req, appBaseURL)
	if err != nil {
		handleFeedbackError(c, err)
		return
	}
	response.SuccessResponseCreated(c, tokenResp, nil)
}

// ─── Public Endpoints (unauthenticated) ──────────────────────────────────────

// GetPublicForm returns the form data for a given token without requiring auth.
func (h *FeedbackHandler) GetPublicForm(c *gin.Context) {
	token := c.Param("token")
	form, err := h.uc.GetPublicForm(c.Request.Context(), token)
	if err != nil {
		handleFeedbackError(c, err)
		return
	}
	response.SuccessResponse(c, form, nil)
}

// SubmitFeedback accepts a customer's feedback submission.
func (h *FeedbackHandler) SubmitFeedback(c *gin.Context) {
	token := c.Param("token")
	var req dto.SubmitFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErrorResponse(c, nil)
		return
	}
	if err := h.uc.SubmitFeedback(c.Request.Context(), token, &req); err != nil {
		handleFeedbackError(c, err)
		return
	}
	response.SuccessResponse(c, gin.H{"message": "Terima kasih atas feedback Anda!"}, nil)
}

// ─── Response Management ──────────────────────────────────────────────────────

// ListResponses returns paginated feedback responses with optional filters.
func (h *FeedbackHandler) ListResponses(c *gin.Context) {
	var req dto.ListFeedbackResponsesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationErrorResponse(c, nil)
		return
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 || req.PerPage > 100 {
		req.PerPage = 20
	}

	items, pagination, err := h.uc.ListResponses(c.Request.Context(), &req)
	if err != nil {
		handleFeedbackError(c, err)
		return
	}
	meta := &response.Meta{
		Pagination: response.NewPaginationMeta(pagination.Page, pagination.PerPage, pagination.Total),
	}
	response.SuccessResponse(c, items, meta)
}

// GetResponse returns a single feedback response by ID.
func (h *FeedbackHandler) GetResponse(c *gin.Context) {
	id := c.Param("id")
	item, err := h.uc.GetResponse(c.Request.Context(), id)
	if err != nil {
		handleFeedbackError(c, err)
		return
	}
	response.SuccessResponse(c, item, nil)
}

// ─── Error Mapping ────────────────────────────────────────────────────────────

func handleFeedbackError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrFeedbackFormNotFound):
		response.ErrorResponse(c, http.StatusNotFound, "FEEDBACK_FORM_NOT_FOUND", "Feedback form not found", nil, nil)
	case errors.Is(err, usecase.ErrFeedbackTokenNotFound):
		response.ErrorResponse(c, http.StatusNotFound, "FEEDBACK_TOKEN_NOT_FOUND", "Feedback token not found or invalid", nil, nil)
	case errors.Is(err, usecase.ErrFeedbackTokenExpired):
		response.ErrorResponse(c, http.StatusGone, "FEEDBACK_TOKEN_EXPIRED", "Feedback link has expired", nil, nil)
	case errors.Is(err, usecase.ErrFeedbackTokenUsed):
		response.ErrorResponse(c, http.StatusConflict, "FEEDBACK_TOKEN_USED", "Feedback has already been submitted", nil, nil)
	case errors.Is(err, usecase.ErrFeedbackAlreadySubmitted):
		response.ErrorResponse(c, http.StatusConflict, "FEEDBACK_ALREADY_SUBMITTED", "Feedback has already been submitted for this token", nil, nil)
	case errors.Is(err, usecase.ErrNoActiveForm):
		response.ErrorResponse(c, http.StatusNotFound, "NO_ACTIVE_FEEDBACK_FORM", "No active feedback form configured for this outlet", nil, nil)
	case errors.Is(err, usecase.ErrInvalidFormSchema):
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_FORM_SCHEMA", err.Error(), nil, nil)
	case errors.Is(err, usecase.ErrInvalidCopyRequest):
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid copy request: provide outlet_ids or apply_to_all_outlets", nil, nil)
	case errors.Is(err, usecase.ErrFeedbackForbidden):
		response.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "You do not have permission to access this outlet feedback data", nil, nil)
	default:
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "An internal server error occurred. Our team has been notified", nil, nil)
	}
}

// parseInt is a local helper to avoid importing strconv everywhere.
func parseInt(s string) (int, error) {
	var n int
	_, err := parseIntHelper(s, &n)
	return n, err
}

func parseIntHelper(s string, n *int) (int, error) {
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, errors.New("not a number")
		}
		*n = *n*10 + int(ch-'0')
	}
	return *n, nil
}
