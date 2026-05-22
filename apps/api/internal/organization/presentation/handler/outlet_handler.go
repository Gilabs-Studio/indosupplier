package handler

import (
	"strconv"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/organization/data/repositories"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
	"github.com/gilabs/gims/api/internal/organization/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// OutletHandler handles HTTP requests for outlets
type OutletHandler struct {
	uc usecase.OutletUsecase
}

// NewOutletHandler creates a new OutletHandler
func NewOutletHandler(uc usecase.OutletUsecase) *OutletHandler {
	return &OutletHandler{uc: uc}
}

// getUserID extracts user ID from gin context
func getOutletUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// Create handles POST /outlets
func (h *OutletHandler) Create(c *gin.Context) {
	var req dto.CreateOutletRequest
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
		if err == usecase.ErrOutletLimitReached {
			errors.ErrorResponse(c, "OUTLET_LIMIT_REACHED", map[string]interface{}{
				"reason": "outlet limit for this tenant has been reached; purchase add-on outlet to continue",
			}, nil)
			return
		}
		errors.ErrorResponse(c, "OUTLET_CREATE_FAILED", map[string]interface{}{
			"message": err.Error(),
		}, nil)
		return
	}

	meta := &response.Meta{CreatedBy: getOutletUserID(c)}
	response.SuccessResponseCreated(c, result, meta)
}

// GetByID handles GET /outlets/:id
func (h *OutletHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	result, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrOutletNotFound {
			errors.ErrorResponse(c, "OUTLET_NOT_FOUND", map[string]interface{}{
				"outlet_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, result, nil)
}

// List handles GET /outlets
func (h *OutletHandler) List(c *gin.Context) {
	params := repositories.OutletListParams{
		Search:      c.Query("search"),
		SortBy:      c.DefaultQuery("sort_by", "name"),
		SortDir:     c.DefaultQuery("sort_dir", "asc"),
		CompanyID:   c.Query("company_id"),
		WarehouseID: c.Query("warehouse_id"),
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

// Update handles PUT /outlets/:id
func (h *OutletHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	var req dto.UpdateOutletRequest
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
		if err == usecase.ErrOutletNotFound {
			errors.ErrorResponse(c, "OUTLET_NOT_FOUND", map[string]interface{}{
				"outlet_id": id,
			}, nil)
			return
		}
		errors.ErrorResponse(c, "OUTLET_UPDATE_FAILED", map[string]interface{}{
			"outlet_id": id,
			"message":   err.Error(),
		}, nil)
		return
	}

	meta := &response.Meta{UpdatedBy: getOutletUserID(c)}
	response.SuccessResponse(c, result, meta)
}

// Delete handles DELETE /outlets/:id
func (h *OutletHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		if err == usecase.ErrOutletNotFound {
			errors.ErrorResponse(c, "OUTLET_NOT_FOUND", map[string]interface{}{
				"outlet_id": id,
			}, nil)
			return
		}
		errors.ErrorResponse(c, "OUTLET_DELETE_FAILED", map[string]interface{}{
			"outlet_id": id,
			"message":   err.Error(),
		}, nil)
		return
	}

	meta := &response.Meta{DeletedBy: getOutletUserID(c)}
	response.SuccessResponseDeleted(c, "outlet", id, meta)
}

// GetFormData handles GET /outlets/form-data
func (h *OutletHandler) GetFormData(c *gin.Context) {
	formData, err := h.uc.GetFormData(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, formData, nil)
}

// GetLimit returns current outlet count and maximum outlet limit for active tenant.
func (h *OutletHandler) GetLimit(c *gin.Context) {
	limit, err := h.uc.GetLimit(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, limit, nil)
}
