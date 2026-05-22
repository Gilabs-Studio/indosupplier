package handler

import (
	"strconv"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/warehouse/data/repositories"
	"github.com/gilabs/gims/api/internal/warehouse/domain/dto"
	"github.com/gilabs/gims/api/internal/warehouse/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// WarehouseHandler handles HTTP requests for warehouses
type WarehouseHandler struct {
	uc usecase.WarehouseUsecase
}

// NewWarehouseHandler creates a new WarehouseHandler
func NewWarehouseHandler(uc usecase.WarehouseUsecase) *WarehouseHandler {
	return &WarehouseHandler{uc: uc}
}

// getUserID extracts user ID from gin context
func getUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// Create handles POST /warehouses
func (h *WarehouseHandler) Create(c *gin.Context) {
	var req dto.CreateWarehouseRequest
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
		errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
			"resource": "warehouse",
			"field":    "code",
			"message":  err.Error(),
		}, nil)
		return
	}

	meta := &response.Meta{CreatedBy: getUserID(c)}
	response.SuccessResponseCreated(c, result, meta)
}

// GetByID handles GET /warehouses/:id
func (h *WarehouseHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	result, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		errors.ErrorResponse(c, "WAREHOUSE_NOT_FOUND", map[string]interface{}{
			"warehouse_id": id,
		}, nil)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// List handles GET /warehouses
func (h *WarehouseHandler) List(c *gin.Context) {
	params := repositories.WarehouseListParams{
		ListParams: repositories.ListParams{
			Search:  c.Query("search"),
			SortBy:  c.DefaultQuery("sort_by", "name"),
			SortDir: c.DefaultQuery("sort_dir", "asc"),
		},
	}

	// Parse is_active filter
	if isActive := c.Query("is_active"); isActive != "" {
		val := isActive == "true"
		params.IsActive = &val
	}

	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
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

// Update handles PUT /warehouses/:id
func (h *WarehouseHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	var req dto.UpdateWarehouseRequest
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
		errors.ErrorResponse(c, "WAREHOUSE_UPDATE_FAILED", map[string]interface{}{
			"warehouse_id": id,
			"message":      err.Error(),
		}, nil)
		return
	}

	meta := &response.Meta{UpdatedBy: getUserID(c)}
	response.SuccessResponse(c, result, meta)
}

// Delete handles DELETE /warehouses/:id
func (h *WarehouseHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		// Return a specific 422 when the warehouse still holds active stock
		if err == usecase.ErrWarehouseHasStock {
			errors.ErrorResponse(c, "WAREHOUSE_HAS_STOCK", map[string]interface{}{
				"warehouse_id": id,
			}, nil)
			return
		}
		errors.ErrorResponse(c, "WAREHOUSE_DELETE_FAILED", map[string]interface{}{
			"warehouse_id": id,
			"message":      err.Error(),
		}, nil)
		return
	}

	meta := &response.Meta{DeletedBy: getUserID(c)}
	response.SuccessResponseDeleted(c, "warehouse", id, meta)
}
