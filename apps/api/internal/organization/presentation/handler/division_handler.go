package handler

import (
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
	"github.com/gilabs/gims/api/internal/organization/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// DivisionHandler handles division HTTP requests
type DivisionHandler struct {
	divisionUC usecase.DivisionUsecase
}

// NewDivisionHandler creates a new DivisionHandler
func NewDivisionHandler(divisionUC usecase.DivisionUsecase) *DivisionHandler {
	return &DivisionHandler{divisionUC: divisionUC}
}

// List handles list divisions request
func (h *DivisionHandler) List(c *gin.Context) {
	var req dto.ListDivisionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	divisions, pagination, err := h.divisionUC.List(c.Request.Context(), &req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{
		Pagination: &response.PaginationMeta{
			Page:       pagination.Page,
			PerPage:    pagination.PerPage,
			Total:      pagination.Total,
			TotalPages: pagination.TotalPages,
			HasNext:    pagination.Page < pagination.TotalPages,
			HasPrev:    pagination.Page > 1,
		},
		Filters: map[string]interface{}{},
	}

	if req.Search != "" {
		meta.Filters["search"] = req.Search
	}

	response.SuccessResponse(c, divisions, meta)
}

// GetByID handles get division by ID request
func (h *DivisionHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	division, err := h.divisionUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrDivisionNotFound {
			errors.ErrorResponse(c, "DIVISION_NOT_FOUND", map[string]interface{}{
				"division_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, division, nil)
}

// Create handles create division request
func (h *DivisionHandler) Create(c *gin.Context) {
	var req dto.CreateDivisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	division, err := h.divisionUC.Create(c.Request.Context(), &req)
	if err != nil {
		if err == usecase.ErrDivisionAlreadyExists {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "division",
				"field":    "name",
				"value":    req.Name,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			meta.CreatedBy = id
		}
	}

	response.SuccessResponseCreated(c, division, meta)
}

// Update handles update division request
func (h *DivisionHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateDivisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	division, err := h.divisionUC.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrDivisionNotFound {
			errors.ErrorResponse(c, "DIVISION_NOT_FOUND", map[string]interface{}{
				"division_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrDivisionAlreadyExists {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "division",
				"field":    "name",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			meta.UpdatedBy = uid
		}
	}

	response.SuccessResponse(c, division, meta)
}

// Delete handles delete division request
func (h *DivisionHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.divisionUC.Delete(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrDivisionNotFound {
			errors.ErrorResponse(c, "DIVISION_NOT_FOUND", map[string]interface{}{
				"division_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(string); ok {
			meta.DeletedBy = uid
		}
	}

	response.SuccessResponseDeleted(c, "division", id, meta)
}
