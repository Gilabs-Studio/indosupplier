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

// ScheduleHandler handles HTTP requests for CRM schedules
type ScheduleHandler struct {
	uc usecase.ScheduleUsecase
}

// NewScheduleHandler creates a new schedule handler
func NewScheduleHandler(uc usecase.ScheduleUsecase) *ScheduleHandler {
	return &ScheduleHandler{uc: uc}
}

// Create handles POST request to create a schedule
func (h *ScheduleHandler) Create(c *gin.Context) {
	var req dto.CreateScheduleRequest
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
		handleScheduleError(c, err)
		return
	}

	meta := &response.Meta{}
	if createdBy != "" {
		meta.CreatedBy = createdBy
	}

	response.SuccessResponseCreated(c, result, meta)
}

// GetByID handles GET request to get a schedule by ID
func (h *ScheduleHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	result, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		errors.ErrorResponse(c, "CRM_SCHEDULE_NOT_FOUND", map[string]interface{}{
			"schedule_id": id,
		}, nil)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// List handles GET request to list schedules with filtering and pagination
func (h *ScheduleHandler) List(c *gin.Context) {
	params := repositories.ScheduleListParams{
		Search:     c.Query("search"),
		SortBy:     c.DefaultQuery("sort_by", "scheduled_at"),
		SortDir:    c.DefaultQuery("sort_dir", "asc"),
		EmployeeID: c.Query("employee_id"),
		Status:     c.Query("status"),
		TaskID:     c.Query("task_id"),
		DateFrom:   c.Query("date_from"),
		DateTo:     c.Query("date_to"),
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
		Filters: map[string]interface{}{},
	}

	if params.EmployeeID != "" {
		meta.Filters["employee_id"] = params.EmployeeID
	}
	if params.Status != "" {
		meta.Filters["status"] = params.Status
	}

	response.SuccessResponse(c, results, meta)
}

// Update handles PUT request to update a schedule
func (h *ScheduleHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	var req dto.UpdateScheduleRequest
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
		handleScheduleError(c, err)
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

// Delete handles DELETE request to delete a schedule
func (h *ScheduleHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		handleScheduleError(c, err)
		return
	}

	meta := &response.Meta{}
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(string); ok {
			meta.DeletedBy = uid
		}
	}

	response.SuccessResponseDeleted(c, "crm_schedule", id, meta)
}

// GetFormData handles GET request to get form data for schedule forms
func (h *ScheduleHandler) GetFormData(c *gin.Context) {
	formData, err := h.uc.GetFormData(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, formData, nil)
}

// handleScheduleError maps business errors to appropriate HTTP responses
func handleScheduleError(c *gin.Context, err error) {
	switch err.Error() {
	case "schedule not found":
		errors.ErrorResponse(c, "CRM_SCHEDULE_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "employee not found":
		errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "task not found":
		errors.ErrorResponse(c, "CRM_TASK_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "invalid scheduled_at format, use ISO 8601":
		errors.ErrorResponse(c, "INVALID_FORMAT", map[string]interface{}{
			"message": err.Error(),
			"field":   "scheduled_at",
		}, nil)
	case "invalid end_at format, use ISO 8601":
		errors.ErrorResponse(c, "INVALID_FORMAT", map[string]interface{}{
			"message": err.Error(),
			"field":   "end_at",
		}, nil)
	case "end_at must be after scheduled_at":
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	default:
		errors.InternalServerErrorResponse(c, err.Error())
	}
}
