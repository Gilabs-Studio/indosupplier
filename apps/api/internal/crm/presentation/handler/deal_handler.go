package handler

import (
	"strconv"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/crm/data/repositories"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
	"github.com/gilabs/gims/api/internal/crm/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// DealHandler handles HTTP requests for deals
type DealHandler struct {
	uc usecase.DealUsecase
}

// NewDealHandler creates a new deal handler
func NewDealHandler(uc usecase.DealUsecase) *DealHandler {
	return &DealHandler{uc: uc}
}

// Create handles POST request to create a deal
func (h *DealHandler) Create(c *gin.Context) {
	var req dto.CreateDealRequest
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
		handleDealError(c, err)
		return
	}

	meta := &response.Meta{}
	if createdBy != "" {
		meta.CreatedBy = createdBy
	}

	response.SuccessResponseCreated(c, result, meta)
}

// GetByID handles GET request to get a deal by ID
func (h *DealHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	result, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		errors.ErrorResponse(c, "DEAL_NOT_FOUND", map[string]interface{}{
			"deal_id": id,
		}, nil)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// List handles GET request to list deals with filtering and pagination
func (h *DealHandler) List(c *gin.Context) {
	params := repositories.DealListParams{
		Search:          c.Query("search"),
		SortBy:          c.DefaultQuery("sort_by", "created_at"),
		SortDir:         c.DefaultQuery("sort_dir", "desc"),
		Status:          c.Query("status"),
		PipelineStageID: c.Query("pipeline_stage_id"),
		CustomerID:      c.Query("customer_id"),
		AssignedTo:      c.Query("assigned_to"),
		LeadID:          c.Query("lead_id"),
		DateFrom:        c.Query("date_from"),
		DateTo:          c.Query("date_to"),
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
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
	}

	response.SuccessResponse(c, results, meta)
}

// ListByStage handles GET request to list deals by stage (Kanban view)
func (h *DealHandler) ListByStage(c *gin.Context) {
	stageID := c.Query("stage_id")
	if stageID == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{
			"message": "stage_id query parameter is required",
		}, nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if perPage > 100 {
		perPage = 100
	}

	params := repositories.DealsByStageParams{
		StageID: stageID,
		Limit:   perPage,
		Offset:  (page - 1) * perPage,
		Search:  c.Query("search"),
		Status:  c.Query("status"),
	}

	results, total, err := h.uc.ListByStage(c.Request.Context(), params)
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
	}

	response.SuccessResponse(c, results, meta)
}

// Update handles PUT request to update a deal
func (h *DealHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	var req dto.UpdateDealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	result, err := h.uc.Update(c.Request.Context(), id, req)
	if err != nil {
		handleDealError(c, err)
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

// Delete handles DELETE request to delete a deal
func (h *DealHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		handleDealError(c, err)
		return
	}

	meta := &response.Meta{}
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(string); ok {
			meta.DeletedBy = uid
		}
	}

	response.SuccessResponseDeleted(c, "crm_deal", id, meta)
}

// MoveStage handles POST request to move a deal to a different pipeline stage
func (h *DealHandler) MoveStage(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	var req dto.MoveDealStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	changedBy := ""
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			changedBy = uid
		}
	}

	result, err := h.uc.MoveStage(c.Request.Context(), id, req, changedBy)
	if err != nil {
		handleDealError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// GetHistory handles GET request to get deal stage transition history
func (h *DealHandler) GetHistory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	history, err := h.uc.GetHistory(c.Request.Context(), id)
	if err != nil {
		handleDealError(c, err)
		return
	}

	response.SuccessResponse(c, history, nil)
}

// GetFormData handles GET request to get form data for deal forms
func (h *DealHandler) GetFormData(c *gin.Context) {
	formData, err := h.uc.GetFormData(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, formData, nil)
}

// GetPipelineSummary handles GET request for pipeline summary statistics
func (h *DealHandler) GetPipelineSummary(c *gin.Context) {
	summary, err := h.uc.GetPipelineSummary(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, summary, nil)
}

// GetForecast handles GET request for deal forecast
func (h *DealHandler) GetForecast(c *gin.Context) {
	forecast, err := h.uc.GetForecast(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, forecast, nil)
}

// ConvertToQuotation handles POST request to convert a won deal into a Sales Quotation
func (h *DealHandler) ConvertToQuotation(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	var req dto.ConvertToQuotationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		// Allow empty body (all fields optional)
		req = dto.ConvertToQuotationRequest{}
	}

	userID := ""
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(string); ok {
			userID = id
		}
	}

	result, err := h.uc.ConvertToQuotation(c.Request.Context(), id, req, userID)
	if err != nil {
		handleDealError(c, err)
		return
	}

	response.SuccessResponseCreated(c, result, nil)
}

// StockCheck handles GET request to check stock availability for deal product items
func (h *DealHandler) StockCheck(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	result, err := h.uc.StockCheck(c.Request.Context(), id)
	if err != nil {
		handleDealError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// SoftDeleteItem handles DELETE request to soft-delete a single deal product item.
func (h *DealHandler) SoftDeleteItem(c *gin.Context) {
	dealID := c.Param("id")
	itemID := c.Param("itemId")
	if dealID == "" || itemID == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "Deal ID and Item ID are required",
		}, nil)
		return
	}

	if err := h.uc.SoftDeleteItem(c.Request.Context(), dealID, itemID); err != nil {
		handleDealError(c, err)
		return
	}

	response.SuccessResponse(c, nil, nil)
}

// RestoreItem handles POST request to restore a soft-deleted deal product item.
func (h *DealHandler) RestoreItem(c *gin.Context) {
	dealID := c.Param("id")
	itemID := c.Param("itemId")
	if dealID == "" || itemID == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "Deal ID and Item ID are required",
		}, nil)
		return
	}

	if err := h.uc.RestoreItem(c.Request.Context(), dealID, itemID); err != nil {
		handleDealError(c, err)
		return
	}

	response.SuccessResponse(c, nil, nil)
}

// GetProductItems handles GET /:id/product-items — returns unified product interest items for the deal.
func (h *DealHandler) GetProductItems(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "Deal ID is required",
		}, nil)
		return
	}

	items, err := h.uc.GetProductItems(c.Request.Context(), id)
	if err != nil {
		handleDealError(c, err)
		return
	}

	response.SuccessResponse(c, items, nil)
}

// handleDealError maps business errors to appropriate HTTP responses
func handleDealError(c *gin.Context, err error) {
	switch err.Error() {
	case "deal not found":
		errors.ErrorResponse(c, "DEAL_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "pipeline stage not found", "invalid pipeline stage":
		errors.ErrorResponse(c, "DEAL_INVALID_STAGE", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "deal already closed":
		errors.ErrorResponse(c, "DEAL_ALREADY_CLOSED", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "close reason required for lost deals":
		errors.ErrorResponse(c, "DEAL_CLOSE_REASON_REQUIRED", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "customer not found":
		errors.ErrorResponse(c, "CUSTOMER_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "contact not found":
		errors.ErrorResponse(c, "CRM_CONTACT_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "assigned employee not found":
		errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "lead not found":
		errors.ErrorResponse(c, "CRM_LEAD_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "invalid expected_close_date format, use YYYY-MM-DD":
		errors.ErrorResponse(c, "INVALID_FORMAT", map[string]interface{}{
			"message": err.Error(),
			"field":   "expected_close_date",
		}, nil)
	case "deal not won":
		errors.ErrorResponse(c, "DEAL_NOT_WON", map[string]interface{}{
			"message": "Deal must be won before converting to quotation",
		}, nil)
	case "deal already converted":
		errors.ErrorResponse(c, "DEAL_ALREADY_CONVERTED", map[string]interface{}{
			"message": "Deal has already been converted to a quotation",
		}, nil)
	case "deal has no items":
		errors.ErrorResponse(c, "DEAL_NO_ITEMS", map[string]interface{}{
			"message": "Deal must have product items before converting to quotation",
		}, nil)
	case "deal customer required":
		errors.ErrorResponse(c, "DEAL_CUSTOMER_REQUIRED", map[string]interface{}{
			"message": "Deal must have a customer before converting to quotation",
		}, nil)
	case "stock check failed":
		errors.InternalServerErrorResponse(c, "Failed to query inventory for stock check")
	default:
		errors.InternalServerErrorResponse(c, err.Error())
	}
}
