package handler

import (
	stdErrors "errors"

	coreErrors "github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
	"github.com/gilabs/gims/api/internal/sales/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// SalesTargetHandler handles sales target HTTP requests
type SalesTargetHandler struct {
	targetUC usecase.SalesTargetUsecase
}

// NewSalesTargetHandler creates a new SalesTargetHandler
func NewSalesTargetHandler(targetUC usecase.SalesTargetUsecase) *SalesTargetHandler {
	return &SalesTargetHandler{targetUC: targetUC}
}

// List handles list sales targets request
func (h *SalesTargetHandler) List(c *gin.Context) {
	var req dto.ListSalesTargetsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			coreErrors.HandleValidationError(c, validationErrors)
			return
		}
		coreErrors.InvalidQueryParamResponse(c)
		return
	}

	targets, pagination, err := h.targetUC.List(c.Request.Context(), &req)
	if err != nil {
		coreErrors.InternalServerErrorResponse(c, err.Error())
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

	if req.Year != nil {
		meta.Filters["year"] = *req.Year
	}
	if req.EmployeeID != "" {
		meta.Filters["employee_id"] = req.EmployeeID
	}
	if req.Search != "" {
		meta.Filters["search"] = req.Search
	}

	response.SuccessResponse(c, targets, meta)
}

// ListAvailableEmployees handles list of employees available for sales target creation for a specific year.
func (h *SalesTargetHandler) ListAvailableEmployees(c *gin.Context) {
	var req dto.ListAvailableSalesTargetEmployeesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			coreErrors.HandleValidationError(c, validationErrors)
			return
		}
		coreErrors.InvalidQueryParamResponse(c)
		return
	}

	employees, err := h.targetUC.ListAvailableEmployees(c.Request.Context(), &req)
	if err != nil {
		coreErrors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, employees, nil)
}

// GetByID handles get sales target by ID request
func (h *SalesTargetHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	// Validate UUID to avoid passing non-UUID path segments (eg. "available-employees")
	if _, err := uuid.Parse(id); err != nil {
		coreErrors.ErrorResponse(c, "INVALID_ID", nil, nil)
		return
	}

	target, err := h.targetUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if stdErrors.Is(err, usecase.ErrSalesTargetNotFound) {
			coreErrors.ErrorResponse(c, "SALES_TARGET_NOT_FOUND", map[string]interface{}{
				"target_id": id,
			}, nil)
			return
		}
		coreErrors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, target, nil)
}

// Create handles create sales target request
func (h *SalesTargetHandler) Create(c *gin.Context) {
	var req dto.CreateSalesTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			coreErrors.HandleValidationError(c, validationErrors)
			return
		}
		coreErrors.InvalidRequestBodyResponse(c)
		return
	}

	target, err := h.targetUC.Create(c.Request.Context(), &req)
	if err != nil {
		if stdErrors.Is(err, usecase.ErrSalesTargetConflict) {
			coreErrors.ErrorResponse(c, "SALES_TARGET_CONFLICT", map[string]interface{}{
				"message": err.Error(),
			}, nil)
			return
		}
		coreErrors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			meta.CreatedBy = id
		}
	}

	response.SuccessResponseCreated(c, target, meta)
}

// Update handles update sales target request
func (h *SalesTargetHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateSalesTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			coreErrors.HandleValidationError(c, validationErrors)
			return
		}
		coreErrors.InvalidRequestBodyResponse(c)
		return
	}

	target, err := h.targetUC.Update(c.Request.Context(), id, &req)
	if err != nil {
		if stdErrors.Is(err, usecase.ErrSalesTargetNotFound) {
			coreErrors.ErrorResponse(c, "SALES_TARGET_NOT_FOUND", map[string]interface{}{
				"target_id": id,
			}, nil)
			return
		}
		if stdErrors.Is(err, usecase.ErrSalesTargetConflict) {
			coreErrors.ErrorResponse(c, "SALES_TARGET_CONFLICT", map[string]interface{}{
				"message": err.Error(),
			}, nil)
			return
		}
		coreErrors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			meta.UpdatedBy = uid
		}
	}

	response.SuccessResponse(c, target, meta)
}

// Delete handles delete sales target request
func (h *SalesTargetHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.targetUC.Delete(c.Request.Context(), id)
	if err != nil {
		if stdErrors.Is(err, usecase.ErrSalesTargetNotFound) {
			coreErrors.ErrorResponse(c, "SALES_TARGET_NOT_FOUND", map[string]interface{}{
				"target_id": id,
			}, nil)
			return
		}
		coreErrors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(string); ok {
			meta.DeletedBy = uid
		}
	}

	response.SuccessResponseDeleted(c, "sales_target", id, meta)
}
