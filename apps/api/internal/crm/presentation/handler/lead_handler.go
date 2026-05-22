package handler

import (
	stdErrors "errors"
	"strconv"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/crm/data/repositories"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
	"github.com/gilabs/gims/api/internal/crm/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// LeadHandler handles HTTP requests for leads
type LeadHandler struct {
	uc usecase.LeadUsecase
}

// NewLeadHandler creates a new lead handler
func NewLeadHandler(uc usecase.LeadUsecase) *LeadHandler {
	return &LeadHandler{uc: uc}
}

// Create handles POST request to create a lead
func (h *LeadHandler) Create(c *gin.Context) {
	var req dto.CreateLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	createdBy := ""
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			createdBy = id
		}
	}

	result, err := h.uc.Create(c.Request.Context(), req, createdBy)
	if err != nil {
		handleLeadError(c, err)
		return
	}

	meta := &response.Meta{}
	if createdBy != "" {
		meta.CreatedBy = createdBy
	}

	response.SuccessResponseCreated(c, result, meta)
}

// GetByID handles GET request to get a lead by ID
func (h *LeadHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	result, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		errors.ErrorResponse(c, "CRM_LEAD_NOT_FOUND", map[string]interface{}{
			"lead_id": id,
		}, nil)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// List handles GET request to list leads with filtering and pagination
func (h *LeadHandler) List(c *gin.Context) {
	params := repositories.LeadListParams{
		Search:       c.Query("search"),
		SortBy:       c.DefaultQuery("sort_by", "created_at"),
		SortDir:      c.DefaultQuery("sort_dir", "desc"),
		LeadStatusID: c.Query("lead_status_id"),
		LeadSourceID: c.Query("lead_source_id"),
		AssignedTo:   c.Query("assigned_to"),
		DateFrom:     c.Query("date_from"),
		DateTo:       c.Query("date_to"),
	}

	// Score range filters
	if scoreMin := c.Query("score_min"); scoreMin != "" {
		if v, err := strconv.Atoi(scoreMin); err == nil {
			params.ScoreMin = &v
		}
	}
	if scoreMax := c.Query("score_max"); scoreMax != "" {
		if v, err := strconv.Atoi(scoreMax); err == nil {
			params.ScoreMax = &v
		}
	}

	// Conversion filter
	if isConverted := c.Query("is_converted"); isConverted != "" {
		v := isConverted == "true"
		params.IsConverted = &v
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	params.Limit = perPage
	params.Offset = (page - 1) * perPage

	results, total, err := h.uc.List(c.Request.Context(), params)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	meta := &response.Meta{
		Pagination: &response.PaginationMeta{
			Page:       page,
			PerPage:    perPage,
			Total:      int(total),
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
		Filters: map[string]interface{}{},
	}

	if params.Search != "" {
		meta.Filters["search"] = params.Search
	}
	if params.LeadStatusID != "" {
		meta.Filters["lead_status_id"] = params.LeadStatusID
	}
	if params.LeadSourceID != "" {
		meta.Filters["lead_source_id"] = params.LeadSourceID
	}
	if params.AssignedTo != "" {
		meta.Filters["assigned_to"] = params.AssignedTo
	}

	response.SuccessResponse(c, results, meta)
}

// Update handles PUT request to update a lead
func (h *LeadHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	var req dto.UpdateLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	updatedBy := ""
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			updatedBy = uid
		}
	}

	result, err := h.uc.Update(c.Request.Context(), id, req, updatedBy)
	if err != nil {
		handleLeadError(c, err)
		return
	}

	meta := &response.Meta{}
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			meta.UpdatedBy = uid
		}
	}

	response.SuccessResponse(c, result, meta)
}

// Delete handles DELETE request to delete a lead
func (h *LeadHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		handleLeadError(c, err)
		return
	}

	meta := &response.Meta{}
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(string); ok {
			meta.DeletedBy = uid
		}
	}

	response.SuccessResponseDeleted(c, "crm_lead", id, meta)
}

// Convert handles POST request to convert a lead to a deal in the pipeline
func (h *LeadHandler) Convert(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	var req dto.ConvertLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	convertedBy := ""
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			convertedBy = id
		}
	}

	result, err := h.uc.Convert(c.Request.Context(), id, req, convertedBy)
	if err != nil {
		handleLeadError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// GetFormData handles GET request to get form data for leads
func (h *LeadHandler) GetFormData(c *gin.Context) {
	formData, err := h.uc.GetFormData(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, formData, nil)
}

// BulkUpsert handles POST request to bulk upsert leads from automation tools (e.g., n8n).
// Uses email as the deduplication key: existing leads are updated, new ones are created.
func (h *LeadHandler) BulkUpsert(c *gin.Context) {
	var req dto.BulkUpsertLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	createdBy := ""
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			createdBy = id
		}
	}

	result, err := h.uc.BulkUpsert(c.Request.Context(), req, createdBy)
	if err != nil {
		handleLeadError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// GetAnalytics handles GET request to get lead analytics
func (h *LeadHandler) GetAnalytics(c *gin.Context) {
	analytics, err := h.uc.GetAnalytics(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, analytics, nil)
}

// GetUnprocessed handles GET request to retrieve unprocessed leads for n8n automation
// Query params: limit (default 50, max 100)
func (h *LeadHandler) GetUnprocessed(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit := 50
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		if l > 100 {
			limit = 100
		} else {
			limit = l
		}
	}

	leads, err := h.uc.GetUnprocessed(c.Request.Context(), limit)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, leads, nil)
}

// GetProductItems handles GET request to get product items for a lead
func (h *LeadHandler) GetProductItems(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	items, err := h.uc.GetProductItems(c.Request.Context(), id)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, items, nil)
}

// handleLeadError maps business errors to appropriate HTTP responses
func handleLeadError(c *gin.Context, err error) {
	switch err.Error() {
	case "lead not found":
		errors.ErrorResponse(c, "CRM_LEAD_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "lead source not found":
		errors.ErrorResponse(c, "CRM_LEAD_SOURCE_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "lead status not found":
		errors.ErrorResponse(c, "CRM_LEAD_STATUS_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "assigned employee not found":
		errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "pipeline stage not found":
		errors.ErrorResponse(c, "CRM_PIPELINE_STAGE_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "lead already converted":
		errors.ErrorResponse(c, "LEAD_ALREADY_CONVERTED", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "cannot convert a lost lead":
		errors.ErrorResponse(c, "CRM_LEAD_CONVERSION_FAILED", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "cannot update a converted lead":
		errors.ErrorResponse(c, "LEAD_ALREADY_CONVERTED", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "cannot delete a converted lead":
		errors.ErrorResponse(c, "LEAD_ALREADY_CONVERTED", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "cannot manually set converted status, use the convert endpoint":
		errors.ErrorResponse(c, "CRM_LEAD_INVALID_STATUS", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "invalid time_expected format, use YYYY-MM-DD":
		errors.ErrorResponse(c, "INVALID_FORMAT", map[string]interface{}{
			"message": err.Error(),
			"field":   "time_expected",
		}, nil)
	default:
		if stdErrors.Is(err, usecase.ErrCannotManuallySetConvertedStatus) {
			errors.ErrorResponse(c, "CRM_LEAD_INVALID_STATUS", map[string]interface{}{
				"message": err.Error(),
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
	}
}
