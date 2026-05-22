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

// ActivityTypeHandler handles HTTP requests for activity types
type ActivityTypeHandler struct {
	uc usecase.ActivityTypeUsecase
}

// NewActivityTypeHandler creates a new activity type handler
func NewActivityTypeHandler(uc usecase.ActivityTypeUsecase) *ActivityTypeHandler {
	return &ActivityTypeHandler{uc: uc}
}

// Create handles POST request to create an activity type
func (h *ActivityTypeHandler) Create(c *gin.Context) {
	var req dto.CreateActivityTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	result, err := h.uc.Create(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			meta.CreatedBy = id
		}
	}

	response.SuccessResponseCreated(c, result, meta)
}

// GetByID handles GET request to get an activity type by ID
func (h *ActivityTypeHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	result, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		errors.ErrorResponse(c, "CRM_ACTIVITY_TYPE_NOT_FOUND", map[string]interface{}{
			"activity_type_id": id,
		}, nil)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// List handles GET request to list activity types
func (h *ActivityTypeHandler) List(c *gin.Context) {
	params := repositories.ListParams{
		Search:  c.Query("search"),
		SortBy:  c.DefaultQuery("sort_by", ""),
		SortDir: c.DefaultQuery("sort_dir", "asc"),
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

	if params.Search != "" {
		meta.Filters["search"] = params.Search
	}

	response.SuccessResponse(c, results, meta)
}

// Update handles PUT request to update an activity type
func (h *ActivityTypeHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	var req dto.UpdateActivityTypeRequest
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
		errors.ErrorResponse(c, "CRM_ACTIVITY_TYPE_NOT_FOUND", map[string]interface{}{
			"activity_type_id": id,
			"message":          err.Error(),
		}, nil)
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

// Delete handles DELETE request to delete an activity type
func (h *ActivityTypeHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		errors.ErrorResponse(c, "CRM_ACTIVITY_TYPE_NOT_FOUND", map[string]interface{}{
			"activity_type_id": id,
		}, nil)
		return
	}

	meta := &response.Meta{}
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(string); ok {
			meta.DeletedBy = uid
		}
	}

	response.SuccessResponseDeleted(c, "crm_activity_type", id, meta)
}
