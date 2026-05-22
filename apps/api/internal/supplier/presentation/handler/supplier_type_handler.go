package handler

import (
	"strconv"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/supplier/data/repositories"
	"github.com/gilabs/gims/api/internal/supplier/domain/dto"
	"github.com/gilabs/gims/api/internal/supplier/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// SupplierTypeHandler handles HTTP requests for supplier types
type SupplierTypeHandler struct {
	uc usecase.SupplierTypeUsecase
}

// NewSupplierTypeHandler creates a new SupplierTypeHandler
func NewSupplierTypeHandler(uc usecase.SupplierTypeUsecase) *SupplierTypeHandler {
	return &SupplierTypeHandler{uc: uc}
}

// Create handles POST /supplier-types
func (h *SupplierTypeHandler) Create(c *gin.Context) {
	var req dto.CreateSupplierTypeRequest
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

// GetByID handles GET /supplier-types/:id
func (h *SupplierTypeHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	result, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		errors.ErrorResponse(c, "SUPPLIER_TYPE_NOT_FOUND", map[string]interface{}{
			"supplier_type_id": id,
		}, nil)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// List handles GET /supplier-types
func (h *SupplierTypeHandler) List(c *gin.Context) {
	params := repositories.ListParams{
		Search:     c.Query("search"),
		SortBy:     c.DefaultQuery("sort_by", "name"),
		SortDir:    c.DefaultQuery("sort_dir", "asc"),
		ActiveOnly: c.Query("active_only") == "true",
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

// Update handles PUT /supplier-types/:id
func (h *SupplierTypeHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	var req dto.UpdateSupplierTypeRequest
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
		errors.ErrorResponse(c, "SUPPLIER_TYPE_NOT_FOUND", map[string]interface{}{
			"supplier_type_id": id,
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

// Delete handles DELETE /supplier-types/:id
func (h *SupplierTypeHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		errors.ErrorResponse(c, "SUPPLIER_TYPE_NOT_FOUND", map[string]interface{}{
			"supplier_type_id": id,
		}, nil)
		return
	}

	meta := &response.Meta{}
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(string); ok {
			meta.DeletedBy = uid
		}
	}

	response.SuccessResponseDeleted(c, "supplier_type", id, meta)
}
